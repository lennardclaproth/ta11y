package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

type Commands struct {
	cs  CommandStore
	qs  TransactionReader
	mdq marketdata.Queries
	vq  vendor.Queries
	rb  rebuilder

	rebuilds *rebuildQueue
}

// CommandStore persists portfolio accounts and transactions.
type CommandStore interface {
	CreateAccount(ctx context.Context, acc *Account) error
	CreateTransaction(ctx context.Context, tx *Transaction) error
	CreateTransactions(ctx context.Context, txs []*Transaction) (int, error)
	UpdateTransactionOccurredAt(ctx context.Context, accountID, id uuid.UUID, occurredAt time.Time, checksum string) (int, error)
	// UpdateManualTransaction stores the edited content of a manual transaction and
	// reports how many rows changed; imported rows are never touched.
	UpdateManualTransaction(ctx context.Context, tx *Transaction) (int, error)
	// DeleteManualTransaction removes a manual transaction and reports how many rows went;
	// imported rows are never touched.
	DeleteManualTransaction(ctx context.Context, accountID, id uuid.UUID) (int, error)
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
// performance and net worth on screen are being brought up to date.
type RebuildOutcome string

const (
	// RebuildStarted means a rebuild was scheduled in the background. It is followed by a
	// portfolio.rebuilt event once the account has caught up with this write.
	RebuildStarted RebuildOutcome = "started"
	// RebuildSkipped means no rebuilder is wired, so nothing was scheduled.
	RebuildSkipped RebuildOutcome = "skipped"
)

// RebuildResult reports the outcome of the rebuild a write asked for.
type RebuildResult struct {
	Outcome RebuildOutcome
}

// NewCommands constructs the portfolio write-side use cases. The rebuilder may be nil,
// in which case manual writes report their rebuild as skipped; log records background
// rebuilds that fail and may be nil.
func NewCommands(cs CommandStore, qs TransactionReader, mdq marketdata.Queries, vq vendor.Queries, rb rebuilder, log logging.Logger) *Commands {
	c := &Commands{cs: cs, qs: qs, mdq: mdq, vq: vq, rb: rb}
	if rb != nil {
		c.rebuilds = newRebuildQueue(rb, log)
	}
	return c
}

// rebuild schedules a background rebuild for a write that changed the account's
// transactions. The write answers without waiting: a full rebuild can take seconds, and
// the portfolio.rebuilt event tells clients when the account has caught up.
func (c *Commands) rebuild(ctx context.Context, accountID uuid.UUID) RebuildResult {
	if c.rebuilds == nil {
		return RebuildResult{Outcome: RebuildSkipped}
	}
	c.rebuilds.request(ctx, accountID)
	return RebuildResult{Outcome: RebuildStarted}
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

// ManualTransactionCreateResult is what a manual write produced: the stored transaction,
// the listing it names, its signed amount and the outcome of the rebuild it asked for.
// UpdateTransaction answers with it too.
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

	data, listingID, err := c.manualTransactionData(ctx, string(v.Name), manualFields{
		OccurredAt:  input.OccurredAt,
		Type:        input.Type,
		ListingID:   input.ListingID,
		Amount:      input.Amount,
		Quantity:    input.Quantity,
		Description: input.Description,
	})
	if err != nil {
		return nil, err
	}

	tx, err := NewManualTransaction(data, input.AccountID)
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

// manualFields are the parts of a manual transaction a person types in. Create and edit
// share them so both paths enforce the same rules.
type manualFields struct {
	OccurredAt  string
	Type        string
	ListingID   *uuid.UUID
	Amount      string
	Quantity    *string
	Description *string
}

// manualTransactionData validates manual fields and resolves the listing into the
// TransactionData a manual transaction is built from, along with the listing it names.
func (c *Commands) manualTransactionData(ctx context.Context, source string, f manualFields) (TransactionData, *uuid.UUID, error) {
	occurredAt, err := parseOccurredAt(f.OccurredAt)
	if err != nil {
		return TransactionData{}, nil, err
	}

	txType, err := parseManualType(f.Type)
	if err != nil {
		return TransactionData{}, nil, err
	}

	amount, err := parseDecimalString(f.Amount, ErrManualInvalidAmount)
	if err != nil {
		return TransactionData{}, nil, err
	}
	if txType == TxCash {
		if amount == 0 {
			return TransactionData{}, nil, ErrManualCashAmountMustBeNonZero
		}
	} else if amount <= 0 {
		return TransactionData{}, nil, ErrManualNonCashAmountMustBePos
	}

	quantity, err := parseManualQuantity(txType, f.Quantity)
	if err != nil {
		return TransactionData{}, nil, err
	}

	listingID, isin, symbol, err := c.resolveManualListing(ctx, txType, f.ListingID)
	if err != nil {
		return TransactionData{}, nil, err
	}

	unitPrice := 0.0
	if txType == TxBuy || txType == TxSell {
		unitPrice = amount / quantity
	}

	description := ""
	if f.Description != nil {
		description = strings.TrimSpace(*f.Description)
	}

	return TransactionData{
		Source:      source,
		OccurredAt:  occurredAt,
		ISIN:        isin,
		Symbol:      symbol,
		Description: description,
		Type:        txType,
		Quantity:    quantity,
		Price:       unitPrice,
		Amount:      amount,
	}, listingID, nil
}

// ManualTransactionUpdateInput replaces what was entered for one manual transaction. The
// vendor it was entered under is not part of it and stays as it was.
type ManualTransactionUpdateInput struct {
	AccountID   uuid.UUID
	ID          uuid.UUID
	OccurredAt  string
	Type        string
	ListingID   *uuid.UUID
	Amount      string
	Quantity    *string
	Description *string
}

// UpdateTransaction replaces the fields of one manually entered transaction under the same
// rules as creating one, then asks for a rebuild because the edited row changes the stream
// positions and snapshots are computed from. Imported transactions are refused: they stay
// as the statement records them, and their checksum is what recognises them on re-import.
func (c *Commands) UpdateTransaction(ctx context.Context, input ManualTransactionUpdateInput) (*ManualTransactionCreateResult, error) {
	current, err := c.editableTransaction(ctx, input.AccountID, input.ID)
	if err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	data, listingID, err := c.manualTransactionData(ctx, current.Source, manualFields{
		OccurredAt:  input.OccurredAt,
		Type:        input.Type,
		ListingID:   input.ListingID,
		Amount:      input.Amount,
		Quantity:    input.Quantity,
		Description: input.Description,
	})
	if err != nil {
		return nil, err
	}

	// Built as a fresh manual transaction so the checksum follows the edited content the
	// same way it does on create; only the identity of the row is carried over.
	edited, err := NewManualTransaction(data, input.AccountID)
	if err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}
	edited.ID = current.ID
	edited.CreatedAt = current.CreatedAt

	clash, err := c.qs.TransactionByChecksum(ctx, input.AccountID, edited.Checksum)
	if err != nil {
		return nil, fmt.Errorf("update transaction: check for duplicate: %w", err)
	}
	if clash != nil && clash.ID != current.ID {
		return nil, fmt.Errorf("update transaction: %w", ErrDuplicateTransaction)
	}

	updated, err := c.cs.UpdateManualTransaction(ctx, edited)
	if err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}
	if updated == 0 {
		return nil, fmt.Errorf("update transaction: %w", ErrTransactionNotFound)
	}

	return &ManualTransactionCreateResult{
		Transaction:  edited,
		ListingID:    listingID,
		SignedAmount: SignedAmountForRead(edited.Type, edited.Quantity, edited.AmountCents.Float64()),
		Rebuild:      c.rebuild(ctx, input.AccountID),
	}, nil
}

// DeleteTransaction removes one manually entered transaction, then asks for a rebuild.
// Imported transactions are refused: removing one would only last until the same
// statement is uploaded again.
func (c *Commands) DeleteTransaction(ctx context.Context, accountID, id uuid.UUID) (RebuildResult, error) {
	if _, err := c.editableTransaction(ctx, accountID, id); err != nil {
		return RebuildResult{}, fmt.Errorf("delete transaction: %w", err)
	}
	deleted, err := c.cs.DeleteManualTransaction(ctx, accountID, id)
	if err != nil {
		return RebuildResult{}, fmt.Errorf("delete transaction: %w", err)
	}
	if deleted == 0 {
		return RebuildResult{}, fmt.Errorf("delete transaction: %w", ErrTransactionNotFound)
	}
	return c.rebuild(ctx, accountID), nil
}

// editableTransaction returns the account's transaction when it was entered by hand.
func (c *Commands) editableTransaction(ctx context.Context, accountID, id uuid.UUID) (*Transaction, error) {
	if c.qs == nil {
		return nil, fmt.Errorf("no transaction reader configured")
	}
	current, err := c.qs.Transaction(ctx, accountID, id)
	if err != nil {
		return nil, fmt.Errorf("fetch transaction: %w", err)
	}
	if current == nil {
		return nil, ErrTransactionNotFound
	}
	if !current.IsManual() {
		return nil, ErrTransactionNotEditable
	}
	return current, nil
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
