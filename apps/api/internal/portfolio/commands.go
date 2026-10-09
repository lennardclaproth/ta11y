package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

type Commands struct {
	cs  CommandStore
	qs  TransactionReader
	mdq marketdata.Queries
	vq  vendor.Queries
	rb  rebuilder
}

// CommandStore persists portfolio accounts and transactions.
type CommandStore interface {
	CreateAccount(ctx context.Context, acc *Account) error
	CreateTransaction(ctx context.Context, tx *Transaction) error
	CreateTransactions(ctx context.Context, txs []*Transaction) (int, error)
	UpdateTransactionOccurredAt(ctx context.Context, accountID, id uuid.UUID, occurredAt time.Time, checksum string) (int, error)
}

// TransactionReader reads the single transactions a write has to inspect first.
type TransactionReader interface {
	Transaction(ctx context.Context, accountID, id uuid.UUID) (*Transaction, error)
	TransactionByChecksum(ctx context.Context, accountID uuid.UUID, checksum string) (*Transaction, error)
}

// rebuilder recomputes positions and snapshots for an account. Writing a transaction by
// hand changes the stream a rebuild reads, so the write asks for one; the rebuild itself
// is unchanged and still refuses to run twice at once.
type rebuilder interface {
	Build(ctx context.Context, accountID uuid.UUID) error
}

// RebuildOutcome reports what happened to the rebuild a write asked for. The write has
// already succeeded in every case -- the outcome only says whether the positions,
// performance and net worth on screen have caught up with it yet.
type RebuildOutcome string

const (
	// RebuildCompleted means the rebuild ran and the account is up to date.
	RebuildCompleted RebuildOutcome = "completed"
	// RebuildInProgress means another rebuild held the lock, so this one did not run.
	RebuildInProgress RebuildOutcome = "in_progress"
	// RebuildSkipped means there was nothing to rebuild yet, or no rebuilder is wired.
	RebuildSkipped RebuildOutcome = "skipped"
	// RebuildFailed means the rebuild was attempted and errored; Err carries the reason.
	RebuildFailed RebuildOutcome = "failed"
)

// RebuildResult reports the outcome of the rebuild a write asked for.
type RebuildResult struct {
	Outcome RebuildOutcome
	Err     error
}

// NewCommands constructs the portfolio write-side use cases. The rebuilder may be nil,
// in which case manual writes report their rebuild as skipped.
func NewCommands(cs CommandStore, qs TransactionReader, mdq marketdata.Queries, vq vendor.Queries, rb rebuilder) *Commands {
	return &Commands{cs: cs, qs: qs, mdq: mdq, vq: vq, rb: rb}
}

const (
	// rebuildReportTimeout caps how long a write waits for the rebuild it asked for before
	// it answers. It stays under the transport's 30s write timeout, so the response to a
	// write that did succeed still reaches the client.
	rebuildReportTimeout = 20 * time.Second
	// rebuildTimeout bounds a detached rebuild so one that hangs cannot hold the build lock
	// for the lifetime of the process.
	rebuildTimeout = 30 * time.Minute
)

// rebuild asks for a rebuild and classifies the outcome. A rebuild that cannot run is
// not a failed write: the transaction is stored either way and the caller reports that
// the portfolio has not caught up.
func (c *Commands) rebuild(ctx context.Context, accountID uuid.UUID) RebuildResult {
	if c.rb == nil {
		return RebuildResult{Outcome: RebuildSkipped}
	}

	// The rebuild deliberately outlives the request. Build clears the account's positions
	// before it recomputes them, so a client that navigates away mid-write would otherwise
	// leave the account emptied until someone rebuilds by hand.
	buildCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rebuildTimeout)
	done := make(chan error, 1)
	go func() {
		defer cancel()
		done <- c.rb.Build(buildCtx, accountID)
	}()

	// Waiting out a long rebuild would run past the write timeout and lose the response.
	// Reporting that it is still running keeps the reply inside the window and leaves the
	// existing rebuild action as the way out.
	reportCtx, stopWaiting := context.WithTimeout(ctx, rebuildReportTimeout)
	defer stopWaiting()

	select {
	case err := <-done:
		return classifyRebuild(err)
	case <-reportCtx.Done():
		return RebuildResult{Outcome: RebuildInProgress}
	}
}

func classifyRebuild(err error) RebuildResult {
	switch {
	case err == nil:
		return RebuildResult{Outcome: RebuildCompleted}
	case errors.Is(err, ErrBuildInProgress):
		return RebuildResult{Outcome: RebuildInProgress}
	case errors.Is(err, ErrPortfolioNoSnapshots), errors.Is(err, ErrAccountNotFound):
		return RebuildResult{Outcome: RebuildSkipped}
	default:
		return RebuildResult{Outcome: RebuildFailed, Err: err}
	}
}

type ManualTransactionInput struct {
	AccountID   uuid.UUID
	VendorID    uuid.UUID
	OccurredAt  string
	Type        string
	ListingID   *uuid.UUID
	Amount      string
	Quantity    *string
	Description *string
}

type ManualTransactionCreateResult struct {
	Transaction  *Transaction
	ListingID    *uuid.UUID
	SignedAmount float64
	Rebuild      RebuildResult
}

// parseOccurredAt reads a YYYY-MM-DD day and refuses anything after today. A portfolio
// transaction records a trade that already happened, so a future day is rejected at the
// boundary rather than left to distort the snapshots a rebuild derives from it.
func parseOccurredAt(raw string) (time.Time, error) {
	occurredAt, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, ErrManualInvalidOccurredAt
	}
	occurredAt = occurredAt.UTC()
	if occurredAt.After(date.StartOfDayUTC(time.Now())) {
		return time.Time{}, ErrManualOccurredAtInFuture
	}
	return occurredAt, nil
}

// ChangeTransactionDate moves one manually entered transaction to another day, leaving
// every other field as entered, then asks for a rebuild because the moved row lands
// elsewhere in the stream positions and snapshots are computed from.
func (c *Commands) ChangeTransactionDate(ctx context.Context, accountID, id uuid.UUID, occurredAtRaw string) (*Transaction, RebuildResult, error) {
	occurredAt, err := parseOccurredAt(occurredAtRaw)
	if err != nil {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: %w", err)
	}
	if c.qs == nil {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: no transaction reader configured")
	}

	current, err := c.qs.Transaction(ctx, accountID, id)
	if err != nil {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: fetch transaction: %w", err)
	}
	if current == nil {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: %w", ErrTransactionNotFound)
	}
	if !current.IsManual() {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: %w", ErrTransactionDateNotEditable)
	}
	if date.SameDayUTC(current.OccurredAt, occurredAt) {
		return current, RebuildResult{Outcome: RebuildSkipped}, nil
	}

	moved := current.MovedTo(occurredAt)
	clash, err := c.qs.TransactionByChecksum(ctx, accountID, moved.Checksum)
	if err != nil {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: check for duplicate: %w", err)
	}
	if clash != nil && clash.ID != current.ID {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: %w", ErrDuplicateTransaction)
	}

	updated, err := c.cs.UpdateTransactionOccurredAt(ctx, accountID, id, moved.OccurredAt, moved.Checksum)
	if err != nil {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: %w", err)
	}
	if updated == 0 {
		return nil, RebuildResult{}, fmt.Errorf("change transaction date: %w", ErrTransactionNotFound)
	}

	return moved, c.rebuild(ctx, accountID), nil
}

func (c *Commands) CreateTransaction(ctx context.Context, input ManualTransactionInput) (*ManualTransactionCreateResult, error) {
	v, err := c.vq.GetById(ctx, input.VendorID)
	if err != nil {
		if errors.Is(err, vendor.ErrVendorNotFound) {
			return nil, ErrManualVendorNotFound
		}
		return nil, fmt.Errorf("manual transaction: fetch vendor: %w", err)
	}
	if !v.Active {
		return nil, ErrManualVendorNotActive
	}
	if v.Type != vendor.VendorTypeBrokerage {
		return nil, ErrManualVendorTypeNotSupported
	}

	occurredAt, err := parseOccurredAt(input.OccurredAt)
	if err != nil {
		return nil, err
	}

	txType, err := parseManualType(input.Type)
	if err != nil {
		return nil, err
	}

	amount, err := parseDecimalString(input.Amount, ErrManualInvalidAmount)
	if err != nil {
		return nil, err
	}
	if txType == TxCash {
		if amount == 0 {
			return nil, ErrManualCashAmountMustBeNonZero
		}
	} else if amount <= 0 {
		return nil, ErrManualNonCashAmountMustBePos
	}

	quantity, err := parseManualQuantity(txType, input.Quantity)
	if err != nil {
		return nil, err
	}

	listingID, isin, symbol, err := c.resolveManualListing(ctx, txType, input.ListingID)
	if err != nil {
		return nil, err
	}

	unitPrice := 0.0
	if txType == TxBuy || txType == TxSell {
		unitPrice = amount / quantity
	}

	description := ""
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}

	tx, err := NewManualTransaction(TransactionData{
		Source:      string(v.Name),
		OccurredAt:  occurredAt,
		ISIN:        isin,
		Symbol:      symbol,
		Description: description,
		Type:        txType,
		Quantity:    quantity,
		Price:       unitPrice,
		Amount:      amount,
	}, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("manual transaction: create transaction: %w", err)
	}

	if err := c.cs.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	signedAmount := tx.AmountCents.Float64()
	if tx.Type == TxCash && tx.Quantity < 0 {
		signedAmount = -signedAmount
	}

	return &ManualTransactionCreateResult{
		Transaction:  tx,
		ListingID:    listingID,
		SignedAmount: signedAmount,
		Rebuild:      c.rebuild(ctx, input.AccountID),
	}, nil
}

func (c *Commands) resolveManualListing(ctx context.Context, txType TransactionType, listingID *uuid.UUID) (*uuid.UUID, *string, *string, error) {
	if txType == TxCash {
		if listingID != nil {
			return nil, nil, nil, ErrManualListingForbidden
		}
		return nil, nil, nil, nil
	}
	if listingID == nil {
		return nil, nil, nil, ErrManualListingRequired
	}

	listing, err := c.mdq.Listing(ctx, *listingID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("manual transaction: fetch listing: %w", err)
	}
	if listing == nil {
		return nil, nil, nil, ErrManualListingNotFound
	}

	var isin *string
	if listing.ISIN != nil {
		v := strings.TrimSpace(*listing.ISIN)
		if v != "" {
			isin = &v
		}
	}

	var symbol *string
	if v := strings.TrimSpace(listing.Symbol); v != "" {
		symbol = &v
	}

	if isin == nil && symbol == nil {
		return nil, nil, nil, ErrManualListingIdentityMissing
	}
	return listingID, isin, symbol, nil
}

func (c *Commands) CreateAccount(ctx context.Context, accountID uuid.UUID) (*Account, error) {
	acc := NewAccount(accountID)
	if err := c.cs.CreateAccount(ctx, acc); err != nil {
		return nil, fmt.Errorf("create account: failed to store account: %w", err)
	}
	return acc, nil
}

// CreateManyResult reports the outcome of a batch portfolio create.
type CreateManyResult struct {
	// Transactions are the transactions built from the input.
	Transactions []*Transaction
	// Imported is the number of rows newly inserted.
	Imported int
	// Duplicates is the number of rows skipped because they already existed.
	Duplicates int
}

// CreateMany builds a batch of imported portfolio transactions and persists them with a
// single bulk insert, skipping and counting rows that already exist. Each TransactionData
// carries its source row number, which orders rows that share a day, and its DedupSeq,
// which is what deduplication keys on.
func (c *Commands) CreateMany(ctx context.Context, importID uuid.UUID, accountID *uuid.UUID, rows []TransactionData) (CreateManyResult, error) {
	if len(rows) == 0 {
		return CreateManyResult{}, nil
	}

	txs := make([]*Transaction, 0, len(rows))
	for _, row := range rows {
		tx, err := NewTransaction(row, row.RowNumber, importID, accountID, nil)
		if err != nil {
			return CreateManyResult{}, fmt.Errorf("create many: build transaction: %w", err)
		}
		txs = append(txs, tx)
	}

	inserted, err := c.cs.CreateTransactions(ctx, txs)
	if err != nil {
		return CreateManyResult{}, fmt.Errorf("create many: %w", err)
	}

	return CreateManyResult{
		Transactions: txs,
		Imported:     inserted,
		Duplicates:   len(txs) - inserted,
	}, nil
}
