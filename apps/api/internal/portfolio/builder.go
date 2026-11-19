package portfolio

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/sorting"

	"github.com/lennardclaproth/ta11y/internal/money"
)

type Builder struct {
	mdq *marketdata.Queries
	pfs PortfolioStore
	pss PositionStore
	ts  TransactionStore
	lk  Locker
	bus eventbus.Bus
}

type PositionStore interface {
	GetLastSnapshot(ctx context.Context, positionID uuid.UUID) (*PositionSnapshot, error)
	CreatePositionSnapshots(ctx context.Context, snaps []*PositionSnapshot) error
	CreateMany(ctx context.Context, positions []*Position) error
	UpdatePositions(ctx context.Context, transactions []Transaction) error
}

type PortfolioStore interface {
	CreatePortfolioSnapshots(ctx context.Context, snaps []*PortfolioSnapshot) error
	Clean(ctx context.Context, accID uuid.UUID) error
	// Do runs fn in one database transaction that the stores join through its context.
	Do(ctx context.Context, fn func(txCtx context.Context) error) error
}

type TransactionStore interface {
	TransactionsForPosition(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]Transaction, error)
	TransactionsForAccount(ctx context.Context, accID uuid.UUID, sort string) ([]Transaction, error)
}

type Locker interface {
	TryAcquireBuildLock(ctx context.Context, accID uuid.UUID) (bool, error)
	ReleaseBuildLock(ctx context.Context, accID uuid.UUID) error
}

// NewBuilder constructs a portfolio Builder. The bus may be nil, in which case
// Rebuilt events are not published.
func NewBuilder(
	mdq *marketdata.Queries,
	pfs PortfolioStore,
	pss PositionStore,
	ts TransactionStore,
	lk Locker,
	bus eventbus.Bus,
) *Builder {
	return &Builder{
		mdq: mdq,
		pfs: pfs,
		pss: pss,
		ts:  ts,
		lk:  lk,
		bus: bus,
	}
}

func (b *Builder) buildPositionSnapshots(
	ctx context.Context,
	accID uuid.UUID,
	pos *Position,
) ([]*PositionSnapshot, error) {
	// Determine start date from when on we want the snapshots to be built.
	var startDate time.Time
	if pos.ListingID == nil {
		return nil, nil // if we don't have a listing for this position, we can't build snapshots, so we just return
	}
	// Get the listing for the position to get the historical market prices
	listing, err := b.mdq.Listing(ctx, *pos.ListingID)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: fetch listing: %w", err)
	}
	if listing == nil {
		return nil, nil // if we don't have a listing for this position, we can't build snapshots, so we just return
	}
	if listing.Symbol == "" {
		return nil, nil // if we don't have a symbol for this listing, we can't build snapshots, so we just return
	}
	// Setup the time window of a position. The start date is the position open date, and the end date is either
	// the position close date or today (if still open).
	startDate = pos.OpenDate
	endExclusive := date.StartOfDayUTC(time.Now())
	if pos.CloseDate != nil && date.StartOfDayUTC(*pos.CloseDate).AddDate(0, 0, 1).Before(endExclusive) {
		endExclusive = date.StartOfDayUTC(*pos.CloseDate).AddDate(0, 0, 1)
	}
	lps, err := b.pss.GetLastSnapshot(ctx, pos.ID)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: get last snapshot: %w", err)
	}
	if lps != nil && date.StartOfDayUTC(lps.OccurredAt).Before(endExclusive) {
		startDate = date.StartOfDayUTC(lps.OccurredAt).AddDate(0, 0, 1)
	}
	// Get all transactions for the position, sorted by time
	ts, err := b.ts.TransactionsForPosition(ctx, pos.ID, &startDate)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: get transactions: %w", err)
	}
	// Get historical market data for the listing from the position open date to today
	eods, err := b.mdq.GetEODByListing(ctx, *pos.ListingID, &startDate, nil, 0, 0, sorting.ASC)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: get eods: %w", err)
	}
	// Prices from the split date onward are post-split, so the quantity has to move
	// with them; otherwise every snapshot after a split understates the position.
	splits, err := b.mdq.SplitsForListing(ctx, *pos.ListingID)
	if err != nil {
		return nil, fmt.Errorf("build position snapshots: get splits: %w", err)
	}
	splitsByDay := splitFactorsByDay(splits)
	// Initialize iterators for transactions and EOD data, and an accumulator for the position state
	txIdx := 0
	dIdx := 0
	var prevSnapshot *PositionSnapshot
	var snapshots []*PositionSnapshot
	acc := PositionAcc{}
	// If we have a last snapshot, we should start with that as the initial state for the position,
	// and then apply any transactions that occurred after that snapshot to get to the current state.
	if lps != nil {
		acc.CostBasis = lps.CostBasis
		acc.Quantity = lps.Quantity
		acc.RealizedPnL = lps.RealizedPnL
		acc.Income = lps.Income
		acc.Fees = lps.Fees
		acc.Taxes = lps.Taxes
		prevSnapshot = lps
	}
	// Walk through the days from the start date to today, and build a snapshot for each day,
	// based on the transactions that occurred on that day and the market price for that day.
	for d := date.StartOfDayUTC(startDate); d.Before(endExclusive); d = d.AddDate(0, 0, 1) {
		// Set day end for the current day, this is used to determine which transactions
		// to include in the snapshot for this day (i.e. all transactions that occurred
		// before the end of the day)
		dayEnd := date.EndOfDayUTC(d)
		unitPriceSet := false
		var unitPrice money.Price
		// The split rescales what was carried in from the previous day. It runs before
		// this day's transactions because a trade on the ex-date is already post-split.
		if factor, ok := splitsByDay[d]; ok {
			acc.applySplit(factor)
		}
		// apply all tx up to this day
		for txIdx < len(ts) && !ts[txIdx].OccurredAt.UTC().After(dayEnd) {
			tx := ts[txIdx]
			if err := acc.ApplyTx(tx); err != nil {
				return nil, fmt.Errorf("build position snapshots: apply tx: %w", err)
			}
			// If the transaction is a buy or sell, we can calculate the market value based on the
			// transaction price, this is an approximation but it's better than carrying forward
			// the previous market value without any adjustments.
			if tx.Type == TxBuy || tx.Type == TxSell {
				unitPrice = tx.UnitPrice
				unitPriceSet = true
			}
			txIdx++
		}
		// find marketclose for this day (advance pointer)
		// If EOD data is available, calculate market value based on the close price for that day.
		for dIdx < len(eods.Data) && date.StartOfDayUTC(eods.Data[dIdx].Date).Before(d) {
			dIdx++
		}
		if dIdx < len(eods.Data) && date.SameDayUTC(eods.Data[dIdx].Date, d) {
			// We prioritize the close price for the day for market value calculation, if available. If
			// not, we fallback to the open price. If there is no open and close price available, we keep
			// the previous market value (if any) to avoid having a gap in the market value for this day.
			if eods.Data[dIdx].Close > 0 {
				unitPrice = eods.Data[dIdx].Close
				unitPriceSet = true
			}
			if !unitPriceSet && eods.Data[dIdx].Open > 0 {
				unitPrice = eods.Data[dIdx].Open
				unitPriceSet = true
			}
		}
		// If we don't have EOD data for this day, we carry forward the previous market value.
		// If a buy or sell transaction occurred on this day we can calculate the market value based on the
		// transaction price, this is an approximation but it's better than carrying forward the previous market
		// value without any adjustments.
		if !unitPriceSet && prevSnapshot != nil {
			unitPrice = prevSnapshot.UnitPrice
		}
		snap, err := NewPositionSnapshot(
			pos.ID,
			accID,
			listing.ID,
			listing.Symbol,
			listing.Name,
			acc.Quantity,
			unitPrice,
			acc.CostBasis,
			acc.RealizedPnL,
			acc.Income,
			acc.Fees,
			acc.Taxes,
			d,
			prevSnapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("build position snapshots: new snapshot: %w", err)
		}
		prevSnapshot = snap
		snapshots = append(snapshots, snap)
	}

	return snapshots, nil
}

// withSplitEvents returns the transaction stream with each instrument's share splits
// interleaved in chronological order.
//
// Instruments are resolved to listings by the same identity the position cycles use, so
// a split is only injected for something the account actually traded. An instrument
// that maps to no listing simply contributes no splits, exactly as before.
func (b *Builder) withSplitEvents(ctx context.Context, transactions []Transaction) ([]Transaction, error) {
	instruments := collectInstruments(transactions)
	if len(instruments) == 0 {
		return transactions, nil
	}

	var events []Transaction
	for identity, inst := range instruments {
		listing, err := b.mdq.ListingByIdentity(ctx, normalizedID(inst.isin), normalizedID(inst.symbol))
		if err != nil || listing == nil {
			// No listing means no price history and therefore no splits to replay.
			continue
		}
		splits, err := b.mdq.SplitsForListing(ctx, listing.ID)
		if err != nil {
			return nil, fmt.Errorf("load splits for %s: %w", identity, err)
		}
		events = append(events, splitEvents(splits, inst.isin, inst.symbol)...)
	}

	return mergeChronologically(transactions, events), nil
}

func (b *Builder) buildPositions(ctx context.Context, accID uuid.UUID) ([]*Position, error) {
	// Load the event stream (transactions) in chronological order.
	ts, err := b.ts.TransactionsForAccount(ctx, accID, "ASC")
	if err != nil {
		return nil, fmt.Errorf("build positions: failed toload transactions: %w", err)
	}
	// Splits belong in the same chronological stream as the trades: a sell after a
	// split is quoted in post-split shares, so the holding has to be rescaled before
	// that sell is matched against it, or the position closes at the wrong quantity
	// and books the wrong realized profit.
	ts, err = b.withSplitEvents(ctx, ts)
	if err != nil {
		return nil, fmt.Errorf("build positions: %w", err)
	}
	// One active lifecycle ("cycle") per canonical instrument key at a time.
	// key = ISIN if available; otherwise symbol (with alias promotion later).
	activeByKey := make(map[string]*Position)
	// alias map to stitch symbol-only transactions to an ISIN once the ISIN appears later.
	// Example: first tx has Symbol="AAPL" only; later tx has ISIN="US0378331005".
	aliases := make(map[string]string) // symbol -> isin
	// We persist every created cycle (open or closed).
	allPositions := make([]*Position, 0)
	for i := range ts {
		tx := &ts[i]
		// Cash transactions don't map to security positions.
		if tx.Type == TxCash {
			continue
		}
		// Determine canonical key and a symbolKey (used for promotion).
		key, symbolKey, err := canonicalPositionKey(*tx, aliases)
		if err != nil {
			return nil, fmt.Errorf("build positions: canonical key: %w", err)
		}
		// Fetch active position cycle; if ISIN appears later, promote symbol-cycle to ISIN key.
		pos := getOrPromoteCycle(activeByKey, key, symbolKey, *tx)
		// Apply transaction according to type. This helper is where the branching lives,
		// so the loop stays easy to read.
		pos, created, err := applyTxToCycle(accID, tx, key, symbolKey, aliases, activeByKey, pos)
		if err != nil {
			return nil, fmt.Errorf("build positions: apply transaction: %w", err)
		}
		if created {
			allPositions = append(allPositions, pos)
		}
	}
	// Listing mapping runs on every rebuild, so a listing added after an import is
	// picked up the next time Rebuild portfolio runs. The match is exact (ISIN, or the
	// symbol when there is no ISIN): no match leaves the position unlinked rather than
	// attaching it to a near-neighbour, and an unlinked position still counts as a
	// position — it just contributes no market value.
	posList := make([]*Position, 0, len(allPositions))
	for _, pos := range allPositions {
		listing, err := b.mdq.ListingByIdentity(ctx, normalizedID(pos.ISIN), normalizedID(pos.Symbol))
		if err == nil && listing != nil {
			pos.linkToListing(listing)
		}
		posList = append(posList, pos)
	}
	// Persist positions (cycles).
	if err := b.pss.CreateMany(ctx, posList); err != nil {
		return nil, fmt.Errorf("build positions: persist positions: %w", err)
	}
	// Persist transaction->position mapping. Synthetic split events are filtered out:
	// they have no stored row to map, and the store keys the update by transaction id.
	if err := b.pss.UpdatePositions(ctx, storedTransactions(ts)); err != nil {
		return nil, fmt.Errorf("build positions: persist transaction mapping: %w", err)
	}
	return posList, nil
}

// Build builds the portfolio based on the transactions.
func (b *Builder) Build(ctx context.Context, accID uuid.UUID) (err error) {
	// The lock is taken outside the transaction below, so a concurrent rebuild sees it
	// straight away rather than once this one commits.
	acquired, err := b.lk.TryAcquireBuildLock(ctx, accID)
	if err != nil {
		return fmt.Errorf("portfolio build failed to acquire build lock: %w", err)
	}
	if !acquired {
		return ErrBuildInProgress
	}
	defer func() {
		// Ensure we always clear the syncing flag, even if the request context is canceled.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if releaseErr := b.lk.ReleaseBuildLock(cleanupCtx, accID); releaseErr != nil {
			wrapped := fmt.Errorf("portfolio build failed to release build lock: %w", releaseErr)
			if err == nil {
				err = wrapped
			} else {
				err = errors.Join(err, wrapped)
			}
		}
	}()

	// The rebuild replaces the account's positions and snapshots wholesale, so it runs as
	// one transaction: readers keep the previous result until it commits, and a rebuild
	// that fails or is cut short leaves that result in place instead of a half-built one.
	var noSnapshots bool
	if err := b.pfs.Do(ctx, func(txCtx context.Context) error {
		built, err := b.rebuild(txCtx, accID)
		noSnapshots = !built
		return err
	}); err != nil {
		return err
	}
	// Nothing priced is still a finished rebuild: the cleared positions are committed so
	// they no longer show, and the caller learns there is nothing to chart.
	if noSnapshots {
		return ErrPortfolioNoSnapshots
	}
	b.publishRebuilt(ctx, accID)
	return nil
}

// rebuild clears and recomputes the account's positions and snapshots. It reports false
// when no position produced a snapshot, in which case only the cleared state is written.
func (b *Builder) rebuild(ctx context.Context, accID uuid.UUID) (bool, error) {
	if err := b.pfs.Clean(ctx, accID); err != nil {
		return false, fmt.Errorf("portfolio build failed to clean: %w", err)
	}
	positions, err := b.buildPositions(ctx, accID)
	if err != nil {
		return false, fmt.Errorf("portfolio build failed to build positions: %w", err)
	}
	var positionSnapshots []*PositionSnapshot
	for _, ps := range positions {
		snapshots, err := b.buildPositionSnapshots(ctx, accID, ps)
		if err != nil {
			return false, fmt.Errorf("portfolio build failed to build position snapshots for position %s: %w", ps.ID, err)
		}
		positionSnapshots = append(positionSnapshots, snapshots...)
	}
	if len(positionSnapshots) == 0 {
		return false, nil
	}
	if err := b.pss.CreatePositionSnapshots(ctx, positionSnapshots); err != nil {
		return false, fmt.Errorf("portfolio build failed to persist position snapshots: %w", err)
	}
	snapshots := aggregatePortfolioSnapshots(accID, positionSnapshots, date.StartOfDayUTC(time.Now()))
	if err := b.pfs.CreatePortfolioSnapshots(ctx, snapshots); err != nil {
		return false, fmt.Errorf("portfolio build failed to persist portfolio snapshots: %w", err)
	}
	return true, nil
}

// linkToListing attaches the position to its listing. Statement rows often identify an
// instrument by ISIN alone, so a position without a symbol takes the listing's ticker.
func (p *Position) linkToListing(listing *marketdata.Listing) {
	p.ListingID = &listing.ID
	if p.Symbol != nil && strings.TrimSpace(*p.Symbol) != "" {
		return
	}
	if symbol := strings.TrimSpace(listing.Symbol); symbol != "" {
		p.Symbol = &symbol
	}
}

// aggregatePortfolioSnapshots builds one account snapshot per day, from the earliest
// position snapshot up to endExclusive. Position snapshots arrive grouped per position, so
// they are ordered by day first. Each position then contributes its most recent snapshot on
// or before that day: a position without a snapshot that day carries forward, and a closed
// position keeps its realized result, income and costs at zero market value.
func aggregatePortfolioSnapshots(
	accID uuid.UUID,
	positionSnapshots []*PositionSnapshot,
	endExclusive time.Time,
) []*PortfolioSnapshot {
	if len(positionSnapshots) == 0 {
		return nil
	}
	ordered := make([]*PositionSnapshot, len(positionSnapshots))
	copy(ordered, positionSnapshots)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].OccurredAt.Before(ordered[j].OccurredAt)
	})

	latest := make(map[uuid.UUID]*PositionSnapshot)
	var positionOrder []uuid.UUID
	var snapshots []*PortfolioSnapshot
	var prevSnapshot *PortfolioSnapshot
	idx := 0
	for d := date.StartOfDayUTC(ordered[0].OccurredAt); d.Before(endExclusive); d = d.AddDate(0, 0, 1) {
		dayEnd := date.EndOfDayUTC(d)
		for idx < len(ordered) && !ordered[idx].OccurredAt.UTC().After(dayEnd) {
			ps := ordered[idx]
			if _, seen := latest[ps.PositionID]; !seen {
				positionOrder = append(positionOrder, ps.PositionID)
			}
			latest[ps.PositionID] = ps
			idx++
		}
		pss := make([]*PositionSnapshot, 0, len(positionOrder))
		for _, id := range positionOrder {
			pss = append(pss, latest[id])
		}
		snap := NewPortfolioSnapshot(accID, d, 0, pss, prevSnapshot)
		snapshots = append(snapshots, snap)
		prevSnapshot = snap
	}
	return snapshots
}

// publishRebuilt announces that an account's portfolio was rebuilt. It is a
// no-op when no bus is configured.
func (b *Builder) publishRebuilt(ctx context.Context, accID uuid.UUID) {
	if b.bus == nil {
		return
	}
	_ = b.bus.Publish(ctx, TopicRebuilt, Rebuilt{AccID: accID})
}
