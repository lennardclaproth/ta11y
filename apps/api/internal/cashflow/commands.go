package cashflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/money"
)

type Commands struct {
	cs  CommandStore
	qs  QueryStore
	aec accountExistenceChecker
}

type accountExistenceChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

// CommandStore persists cashflow transaction mutations.
type CommandStore interface {
	CreateTransactions(ctx context.Context, txs []*Transaction) (int, error)
	UpdateDate(ctx context.Context, accountID, id uuid.UUID, date time.Time, checksum string) (int, error)
	UpdateTagByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID, tag string) (int, error)
	UpdateTagByFilter(ctx context.Context, filters TransactionFilters, tag string) (int, error)
	UpdateIgnoredByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID, ignored bool) (int, error)
	UpdateIgnoredByFilter(ctx context.Context, filters TransactionFilters, ignored bool) (int, error)
	UpdatePurposeByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID, purpose Purpose) (int, error)
	UpdatePurposeByFilter(ctx context.Context, filters TransactionFilters, purpose Purpose) (int, error)
}

const (
	manualCashflowBatchMaxSize = 100
)

// NewCommands creates cashflow write-side use cases.
func NewCommands(
	cStore CommandStore,
	qStore QueryStore,
	aec accountExistenceChecker,
) *Commands {
	return &Commands{
		cs:  cStore,
		qs:  qStore,
		aec: aec,
	}
}

type TransactionData struct {
	Description string
	Note        string
	Source      string
	Direction   CashFlowDirection
	Amount      money.Price
	Date        time.Time
	AccountType *AccountType
	Tag         string
	// RowNumber is the source row number (e.g. CSV line). It records where the row
	// came from; it deliberately no longer feeds the dedup checksum, because an
	// overlapping export moves every row to a different line. Manual entries leave it
	// zero and receive a generated row number.
	RowNumber int
	// DedupSeq distinguishes rows with identical content inside one import file and
	// feeds the dedup checksum in RowNumber's place. Importers set it from
	// importer.DedupSequencer; manual entries leave it zero.
	DedupSeq int
}

// NewTransactionData validates and maps manual cashflow input into transaction data.
func NewTransactionData(
	dateRaw, amountRaw, typeRaw, descriptionRaw, noteRaw, tagRaw, vendorRaw, source string,
	rowNumber *int,
) (TransactionData, error) {
	txDate, err := parseTransactionDate(dateRaw)
	if err != nil {
		return TransactionData{}, err
	}

	amount, err := money.ParsePrice(amountRaw)
	if err != nil {
		return TransactionData{}, err
	}

	direction, err := ParseDirection(typeRaw)
	if err != nil {
		return TransactionData{}, err
	}

	description := strings.TrimSpace(descriptionRaw)
	if description == "" {
		return TransactionData{}, ErrManualCashflowDescriptionRequired
	}
	note := strings.TrimSpace(noteRaw)
	if note == "" {
		return TransactionData{}, ErrManualCashflowNoteRequired
	}
	tag := strings.TrimSpace(tagRaw)
	if tag == "" {
		return TransactionData{}, ErrManualCashflowTagRequired
	}

	if vendorName := strings.TrimSpace(vendorRaw); vendorName != "" {
		source = "manual:" + vendorName
	}

	return TransactionData{
		Description: description,
		Note:        note,
		Source:      source,
		Direction:   *direction,
		Amount:      amount,
		Date:        txDate,
		Tag:         tag,
	}, nil
}

// parseTransactionDate reads a YYYY-MM-DD day and refuses anything after today. A
// cashflow transaction records something that already happened, so a future day is
// rejected at the boundary rather than left to surface as an empty month later.
func parseTransactionDate(raw string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, ErrManualCashflowInvalidDate
	}
	parsed = parsed.UTC()
	if parsed.After(date.StartOfDayUTC(time.Now())) {
		return time.Time{}, ErrCashflowDateInFuture
	}
	return parsed, nil
}

// ChangeDate moves one manually entered transaction to another day, leaving every other
// field as entered. The dedup checksum carries the date, so it is recomputed and checked
// against the account's existing rows before the update.
func (c *Commands) ChangeDate(ctx context.Context, accID, id uuid.UUID, dateRaw string) (*Transaction, error) {
	newDate, err := parseTransactionDate(dateRaw)
	if err != nil {
		return nil, fmt.Errorf("change date: %w", err)
	}

	current, err := c.qs.GetTransaction(ctx, accID, id)
	if err != nil {
		return nil, fmt.Errorf("change date: fetch transaction: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("change date: %w", ErrNoTransactionFound)
	}
	if !current.IsManual() {
		return nil, fmt.Errorf("change date: %w", ErrCashflowDateNotEditable)
	}
	if date.SameDayUTC(current.Date, newDate) {
		return current, nil
	}

	moved := current.MovedTo(newDate)
	clash, err := c.qs.GetTransactionByChecksum(ctx, accID, moved.Checksum)
	if err != nil {
		return nil, fmt.Errorf("change date: check for duplicate: %w", err)
	}
	if clash != nil && clash.ID != current.ID {
		return nil, fmt.Errorf("change date: %w", ErrDuplicateTransaction)
	}

	updated, err := c.cs.UpdateDate(ctx, accID, id, moved.Date, moved.Checksum)
	if err != nil {
		return nil, fmt.Errorf("change date: %w", err)
	}
	if updated == 0 {
		return nil, fmt.Errorf("change date: %w", ErrNoTransactionFound)
	}
	return moved, nil
}

// CreateManyResult reports the outcome of a batch cashflow create.
type CreateManyResult struct {
	// Transactions are the transactions built from the input. On a successful
	// create with no duplicates these are the persisted rows.
	Transactions []*Transaction
	// Imported is the number of rows newly inserted.
	Imported int
	// Duplicates is the number of rows skipped because they already existed.
	Duplicates int
}

// CreateMany validates a batch of cashflow transactions and persists them with a single
// bulk insert, skipping and counting rows that already exist. It serves both manual
// entry and CSV imports: manual callers pass a nil import ID and are capped at
// manualCashflowBatchMaxSize; import callers pass the import ID and set RowNumber and
// DedupSeq on each TransactionData. Rows without a row number fall back to a generated
// manual one, and rows without a sequence fall back to that row number.
func (c *Commands) CreateMany(ctx context.Context, accID uuid.UUID, impID *uuid.UUID, transactions []TransactionData) (CreateManyResult, error) {
	if len(transactions) == 0 {
		return CreateManyResult{}, fmt.Errorf("create many: %w", ErrTransactionsRequired)
	}
	if impID == nil && len(transactions) > manualCashflowBatchMaxSize {
		return CreateManyResult{}, fmt.Errorf("create many: %w", ErrTransactionLimitExceeded)
	}

	exists, err := c.aec.Exists(ctx, accID)
	if err != nil {
		return CreateManyResult{}, fmt.Errorf("create many: failed to check if account exists: %w", err)
	}
	if !exists {
		return CreateManyResult{}, fmt.Errorf("create many: %w", ErrAccountNotFound)
	}

	txs := make([]*Transaction, 0, len(transactions))
	for i, row := range transactions {
		rowNumber := row.RowNumber
		if rowNumber == 0 {
			rowNumber = manualCashflowRowNumber(i)
		}
		// Manual entries carry no sequence; their row number kept rows apart before
		// and still does, so nothing about manual deduplication changes.
		dedupSeq := row.DedupSeq
		if dedupSeq == 0 {
			dedupSeq = rowNumber
		}
		tx, err := NewTransaction(
			row.Description,
			row.Note,
			row.Source,
			row.Tag,
			row.Direction,
			row.Amount,
			row.Date,
			rowNumber,
			dedupSeq,
			impID,
			row.AccountType,
			accID,
		)
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

type ExecutionMode string

const (
	TagByFilterModeSync  ExecutionMode = "sync"
	TagByFilterModeAsync ExecutionMode = "async"
)

type BulkTagResult struct {
	Mode         ExecutionMode
	UpdatedCount int
	TotalMatched int
}

type TagByFilterCommand struct {
	Tag       string
	AccountID *uuid.UUID
	Filters   TransactionFilters
}

// TagByID applies a tag to one cashflow transaction.
func (c *Commands) TagByID(ctx context.Context, accountID, id uuid.UUID, tag string) error {
	_, err := c.TagByIDs(ctx, accountID, []uuid.UUID{id}, tag)
	if err != nil {
		return fmt.Errorf("cashflow tag by id: %w", err)
	}
	return nil
}

// TagByIDs applies a tag to the selected cashflow transactions.
func (c *Commands) TagByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID, tag string) (int, error) {
	updated, err := c.cs.UpdateTagByIDs(ctx, accountID, ids, tag)
	if err != nil {
		return 0, fmt.Errorf("cashflow tag by ids: %w", err)
	}
	return updated, nil
}

// TagByFilter applies or schedules tagging based on total matched rows and async policy.
func (c *Commands) TagByFilter(ctx context.Context, tag string, accID uuid.UUID, filters TransactionFilters) (BulkTagResult, error) {
	// The account is authoritative over anything the filters carried in.
	filters.AccountID = accID
	total, err := c.qs.CountByFilter(ctx, filters)
	if err != nil {
		return BulkTagResult{}, fmt.Errorf("cashflow tag by filter count: %w", err)
	}
	// TODO: implement async tagging
	updated, err := c.cs.UpdateTagByFilter(ctx, filters, tag)
	if err != nil {
		return BulkTagResult{}, fmt.Errorf("cashflow tag by filter update: %w", err)
	}
	return BulkTagResult{
		Mode:         TagByFilterModeSync,
		UpdatedCount: updated,
		TotalMatched: total,
	}, nil
}

// IgnoreByIDs sets the ignored flag for the selected cashflow transactions.
func (c *Commands) IgnoreByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID, ignored bool) (int, error) {
	updated, err := c.cs.UpdateIgnoredByIDs(ctx, accountID, ids, ignored)
	if err != nil {
		return 0, fmt.Errorf("cashflow ignore by ids: %w", err)
	}
	return updated, nil
}

// IgnoreByFilter sets the ignored flag for cashflow transactions matching the supplied filters.
func (c *Commands) IgnoreByFilter(ctx context.Context, filters TransactionFilters, ignored bool) (int, error) {
	updated, err := c.cs.UpdateIgnoredByFilter(ctx, filters, ignored)
	if err != nil {
		return 0, fmt.Errorf("cashflow ignore by filter: %w", err)
	}
	return updated, nil
}

// MarkPurposeResult reports what a purpose mutation did. Matched counts the rows the
// caller pointed at, Updated only those the purpose could actually apply to: a selection
// may hold both directions, and income only sticks to incoming money.
type MarkPurposeResult struct {
	Matched int
	Updated int
}

// MarkPurposeByIDs sets what the selected transactions count as towards the monthly goal.
// Rows whose direction does not allow the purpose are left alone, which the result reports.
func (c *Commands) MarkPurposeByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID, purpose Purpose) (MarkPurposeResult, error) {
	updated, err := c.cs.UpdatePurposeByIDs(ctx, accountID, ids, purpose)
	if err != nil {
		return MarkPurposeResult{}, fmt.Errorf("cashflow mark purpose by ids: %w", err)
	}
	return MarkPurposeResult{Matched: len(ids), Updated: updated}, nil
}

// MarkPurposeByFilter sets what every transaction matching the filters counts as towards
// the monthly goal. The account is authoritative over anything the filters carried in.
func (c *Commands) MarkPurposeByFilter(ctx context.Context, accountID uuid.UUID, filters TransactionFilters, purpose Purpose) (MarkPurposeResult, error) {
	filters.AccountID = accountID
	matched, err := c.qs.CountByFilter(ctx, filters)
	if err != nil {
		return MarkPurposeResult{}, fmt.Errorf("cashflow mark purpose by filter count: %w", err)
	}
	updated, err := c.cs.UpdatePurposeByFilter(ctx, filters, purpose)
	if err != nil {
		return MarkPurposeResult{}, fmt.Errorf("cashflow mark purpose by filter: %w", err)
	}
	return MarkPurposeResult{Matched: matched, Updated: updated}, nil
}
