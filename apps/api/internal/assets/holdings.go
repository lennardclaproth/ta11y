package assets

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/sorting"
)

// eodWindowLimit bounds one price read. A holding's history runs from its first
// purchase to today, so this is a guard against an absurd range rather than a page
// size: fifty years of daily prices still fit under it.
const eodWindowLimit = 20000

// HoldingsStore reads what a daily-priced item is built from and rewrites the
// mutations derived from it.
type HoldingsStore interface {
	DailyPricedAssets(ctx context.Context, accID uuid.UUID) ([]*Asset, error)
	PurchasesForAccount(ctx context.Context, accID uuid.UUID) ([]*Purchase, error)
	// ManualMutations returns the account's mutations for items that are not daily
	// priced, oldest first. They are the part of a class total this rebuild does not
	// own and must carry alongside its own.
	ManualMutations(ctx context.Context, accID uuid.UUID) ([]*Mutation, error)
	// DeleteDerivedMutations drops every mutation belonging to a daily-priced item.
	// Those rows are derived, so each rebuild replaces them wholesale.
	DeleteDerivedMutations(ctx context.Context, accID uuid.UUID) error
	CreateMutation(ctx context.Context, mutation *Mutation) error
	SetWorth(ctx context.Context, asset *Asset) error
}

// HoldingsSyncer rewrites the worth history of every daily-priced item in an
// account from its purchases and its listing's daily prices.
//
// It is the counterpart of SyncPortfolio: where that one mirrors the portfolio into
// the read-only PORTFOLIO class, this one mirrors "quantity held times the price
// that day" into ordinary classes. Both end in the same place -- one SET mutation
// per item per day -- so the snapshot Builder and the net worth it feeds need no
// knowledge of either.
type HoldingsSyncer struct {
	hs  HoldingsStore
	mdq *marketdata.Queries
	uow UnitOfWork
}

// NewHoldingsSyncer constructs the daily-priced holdings syncer.
func NewHoldingsSyncer(hs HoldingsStore, mdq *marketdata.Queries, uow UnitOfWork) *HoldingsSyncer {
	return &HoldingsSyncer{hs: hs, mdq: mdq, uow: uow}
}

// SyncAccount rebuilds the derived mutations for every daily-priced item in the
// account. It is safe to run repeatedly: the derived rows are deleted and written
// again from the purchases and prices that hold at that moment.
func (s *HoldingsSyncer) SyncAccount(ctx context.Context, accID uuid.UUID) error {
	holdings, err := s.hs.DailyPricedAssets(ctx, accID)
	if err != nil {
		return fmt.Errorf("sync holdings: failed to list daily-priced items: %w", err)
	}
	if len(holdings) == 0 {
		// Nothing derived can exist without a daily-priced item, so there is also
		// nothing to clean up.
		return nil
	}

	purchases, err := s.hs.PurchasesForAccount(ctx, accID)
	if err != nil {
		return fmt.Errorf("sync holdings: failed to list purchases: %w", err)
	}
	manual, err := s.hs.ManualMutations(ctx, accID)
	if err != nil {
		return fmt.Errorf("sync holdings: failed to list manual mutations: %w", err)
	}

	prices := make(map[uuid.UUID]map[string]money.Price, len(holdings))
	for _, holding := range holdings {
		if !holding.IsDailyPriced() {
			continue
		}
		series, err := s.pricesForHolding(ctx, *holding.ListingID, earliestPurchaseDay(purchases, holding.ID))
		if err != nil {
			return err
		}
		prices[holding.ID] = series
	}

	today := date.StartOfDayUTC(time.Now().UTC())
	plan := planHoldingMutations(accID, holdings, purchases, manual, prices, today)
	if len(plan.mutations) == 0 {
		return nil
	}

	if err := s.uow.Do(ctx, func(txCtx context.Context) error {
		if err := s.hs.DeleteDerivedMutations(txCtx, accID); err != nil {
			return err
		}
		for _, mutation := range plan.mutations {
			if err := s.hs.CreateMutation(txCtx, mutation); err != nil {
				return err
			}
		}
		for _, holding := range holdings {
			holding.CurrentWorth = plan.currentWorth[holding.ID]
			if err := s.hs.SetWorth(txCtx, holding); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("sync holdings: failed to execute transaction: %w", err)
	}
	return nil
}

// pricesForHolding reads the listing's closing prices from the first purchase
// onwards, keyed by day. Reading through marketdata.Queries is what keeps the
// provider fetch bounded to the window the holding actually needs and refreshes a
// stale listing on the way.
func (s *HoldingsSyncer) pricesForHolding(
	ctx context.Context,
	listingID uuid.UUID,
	from *time.Time,
) (map[string]money.Price, error) {
	if s.mdq == nil || from == nil {
		return nil, nil
	}
	result, err := s.mdq.GetEODByListing(ctx, listingID, from, nil, eodWindowLimit, 0, sorting.ASC)
	if err != nil {
		// A listing that cannot be read leaves the holding valued at what was paid
		// for it (see holdingValue); that is stale, but it is not zero, and it does
		// not fail a rebuild that the rest of the account depends on.
		return nil, nil
	}
	series := make(map[string]money.Price, len(result.Data))
	for _, row := range result.Data {
		series[dayKey(row.Date)] = row.Close
	}
	return series, nil
}

// holdingPlan is the outcome of a rebuild: the derived mutations to write and the
// worth each item should carry afterwards.
type holdingPlan struct {
	mutations    []*Mutation
	currentWorth map[uuid.UUID]money.Price
}

// planHoldingMutations walks every day from the first purchase in the account to
// today and emits one SET mutation per daily-priced item per day.
//
// A mutation records the class total as well as the item's worth, and the snapshot
// Builder reads that total. A class can hold manual items next to daily-priced
// ones, so the total written here is the manual worth carried forward to that day
// plus every daily-priced item in the class on that day -- otherwise a price move
// would silently erase the manual items beside it.
func planHoldingMutations(
	accID uuid.UUID,
	holdings []*Asset,
	purchases []*Purchase,
	manual []*Mutation,
	prices map[uuid.UUID]map[string]money.Price,
	today time.Time,
) holdingPlan {
	plan := holdingPlan{currentWorth: make(map[uuid.UUID]money.Price, len(holdings))}

	byAsset := purchasesByAsset(purchases)
	start := earliestDay(byAsset)
	if start == nil {
		return plan
	}

	// The day range spans the whole account, because a class total has to be known for every
	// day any item in it moved. An individual item's history still starts at its own first
	// purchase: writing it from the account's earliest day would back-date a holding bought
	// last week with years of zero-worth rows, and read back as a flat line before it existed.
	firstDay := make(map[uuid.UUID]*time.Time, len(holdings))
	for _, holding := range holdings {
		firstDay[holding.ID] = earliestDay(map[uuid.UUID][]*Purchase{holding.ID: byAsset[holding.ID]})
	}

	manualTotals := manualTotalsByClassAndDay(manual)
	manualCarried := make(map[uuid.UUID]money.Price, len(manualTotals))
	lastPrice := make(map[uuid.UUID]money.Price, len(holdings))
	previousWorth := make(map[uuid.UUID]money.Price, len(holdings))
	now := time.Now().UTC()
	note := "derived from purchases and the daily price"

	for day := *start; !day.After(today); day = day.AddDate(0, 0, 1) {
		key := dayKey(day)
		for classID, series := range manualTotals {
			if total, ok := series[key]; ok {
				manualCarried[classID] = total
			}
		}

		worthToday := make(map[uuid.UUID]money.Price, len(holdings))
		classTotals := make(map[uuid.UUID]money.Price, len(holdings))
		for _, holding := range holdings {
			quantity, paid := holdingPosition(byAsset[holding.ID], day)
			if price, ok := prices[holding.ID][key]; ok {
				lastPrice[holding.ID] = price
			}
			worth := holdingValue(quantity, paid, lastPrice[holding.ID])
			worthToday[holding.ID] = worth
			classTotals[holding.ClassID] += worth
		}
		for classID := range classTotals {
			classTotals[classID] += manualCarried[classID]
		}

		for _, holding := range holdings {
			if start := firstDay[holding.ID]; start == nil || day.Before(*start) {
				continue
			}
			worth := worthToday[holding.ID]
			plan.mutations = append(plan.mutations, &Mutation{
				ID:              uuid.New(),
				AccountID:       accID,
				ClassID:         holding.ClassID,
				AssetID:         holding.ID,
				ChangeType:      ChangeTypeSet,
				Amount:          worth,
				PreviousWorth:   previousWorth[holding.ID],
				NewWorth:        worth,
				ClassTotalWorth: classTotals[holding.ClassID],
				EffectiveDate:   day,
				Note:            &note,
				CreatedAt:       now,
			})
			previousWorth[holding.ID] = worth
			plan.currentWorth[holding.ID] = worth
		}
	}
	return plan
}

// holdingValue is quantity times the last known price, falling back to what was
// paid while no price is known yet. A purchase recorded before the provider has a
// price for that day would otherwise read as a holding worth nothing.
func holdingValue(quantity float64, paid, price money.Price) money.Price {
	if quantity <= 0 {
		return 0
	}
	if price <= 0 {
		return paid
	}
	return money.Price(float64(price) * quantity)
}

// holdingPosition returns how much was held and how much had been paid for it by
// the end of the given day.
func holdingPosition(purchases []*Purchase, day time.Time) (float64, money.Price) {
	quantity := 0.0
	paid := money.Price(0)
	for _, purchase := range purchases {
		if purchase == nil || purchase.PurchasedOn.After(day) {
			continue
		}
		quantity += purchase.Quantity
		paid += purchase.Paid()
	}
	return quantity, paid
}

// manualTotalsByClassAndDay replays the mutations of the items this rebuild does
// not own into a per-class total for every day one of them changed.
func manualTotalsByClassAndDay(mutations []*Mutation) map[uuid.UUID]map[string]money.Price {
	byClass := make(map[uuid.UUID]map[string]money.Price)
	worthByItem := make(map[uuid.UUID]map[uuid.UUID]money.Price)

	for _, mutation := range mutations {
		if mutation == nil {
			continue
		}
		items := worthByItem[mutation.ClassID]
		if items == nil {
			items = make(map[uuid.UUID]money.Price)
			worthByItem[mutation.ClassID] = items
		}
		items[mutation.AssetID] = mutation.NewWorth

		total := money.Price(0)
		for _, worth := range items {
			total += worth
		}
		days := byClass[mutation.ClassID]
		if days == nil {
			days = make(map[string]money.Price)
			byClass[mutation.ClassID] = days
		}
		// Mutations arrive oldest first, so the last row of a day wins.
		days[dayKey(mutation.EffectiveDate)] = total
	}
	return byClass
}

func purchasesByAsset(purchases []*Purchase) map[uuid.UUID][]*Purchase {
	byAsset := make(map[uuid.UUID][]*Purchase)
	for _, purchase := range purchases {
		if purchase == nil {
			continue
		}
		byAsset[purchase.AssetID] = append(byAsset[purchase.AssetID], purchase)
	}
	for id := range byAsset {
		rows := byAsset[id]
		sort.SliceStable(rows, func(i, j int) bool {
			return rows[i].PurchasedOn.Before(rows[j].PurchasedOn)
		})
	}
	return byAsset
}

// earliestDay is the first day any purchase in the account was made, and therefore
// the first day a derived worth exists.
func earliestDay(byAsset map[uuid.UUID][]*Purchase) *time.Time {
	var earliest *time.Time
	for _, rows := range byAsset {
		if len(rows) == 0 {
			continue
		}
		day := date.StartOfDayUTC(rows[0].PurchasedOn)
		if earliest == nil || day.Before(*earliest) {
			earliest = &day
		}
	}
	return earliest
}

// earliestPurchaseDay is the window start for one holding's price history: prices
// older than its first purchase are never used, so they are never fetched.
func earliestPurchaseDay(purchases []*Purchase, assetID uuid.UUID) *time.Time {
	var earliest *time.Time
	for _, purchase := range purchases {
		if purchase == nil || purchase.AssetID != assetID {
			continue
		}
		day := date.StartOfDayUTC(purchase.PurchasedOn)
		if earliest == nil || day.Before(*earliest) {
			earliest = &day
		}
	}
	return earliest
}

func dayKey(t time.Time) string {
	return date.StartOfDayUTC(t).Format(time.DateOnly)
}
