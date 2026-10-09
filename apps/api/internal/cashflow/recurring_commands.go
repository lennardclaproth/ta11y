package cashflow

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// RecurringCommands exposes the write side of recurring items.
type RecurringCommands struct {
	cs RecurringCommandStore
	qs RecurringQueryStore
}

// RecurringCommandStore persists recurring items, their links and dismissals.
type RecurringCommandStore interface {
	CreateRecurringItem(ctx context.Context, item *RecurringItem) error
	UpdateRecurringItem(ctx context.Context, item *RecurringItem) (int, error)
	EndRecurringItem(ctx context.Context, accountID, id uuid.UUID, from time.Time) (int, error)
	LinkTransactions(ctx context.Context, itemID uuid.UUID, transactionIDs []uuid.UUID) (int, error)
	UnlinkTransaction(ctx context.Context, accountID, itemID, transactionID uuid.UUID) (int, error)
	DismissRecurringSuggestion(ctx context.Context, accountID uuid.UUID, matchKey string, direction CashFlowDirection) error
}

// NewRecurringCommands creates the recurring write-side use cases.
func NewRecurringCommands(cs RecurringCommandStore, qs RecurringQueryStore) *RecurringCommands {
	return &RecurringCommands{cs: cs, qs: qs}
}

// CreateRecurringInput starts a new recurring item from transactions the user pointed at.
type CreateRecurringInput struct {
	Name           string
	Direction      CashFlowDirection
	Rhythm         Rhythm
	TransactionIDs []uuid.UUID
}

// Create starts a recurring item and links the transactions it was pointed at.
// The item's fingerprint comes from the oldest of those transactions, so later
// imports of the same counterparty find their way here on their own.
func (c *RecurringCommands) Create(ctx context.Context, accountID uuid.UUID, input CreateRecurringInput) (*RecurringItem, error) {
	if len(input.TransactionIDs) == 0 {
		return nil, fmt.Errorf("create recurring item: %w", ErrRecurringTransactionsRequired)
	}

	transactions, err := c.qs.ListTransactionsByIDs(ctx, accountID, input.TransactionIDs)
	if err != nil {
		return nil, fmt.Errorf("create recurring item: fetch transactions: %w", err)
	}
	if len(transactions) != len(input.TransactionIDs) {
		return nil, fmt.Errorf("create recurring item: %w", ErrRecurringTransactionNotFound)
	}
	sort.Slice(transactions, func(i, j int) bool { return transactions[i].Date.Before(transactions[j].Date) })

	item, err := NewRecurringItem(accountID, input.Name, input.Direction, input.Rhythm, MatchKeyFor(transactions[0].Description))
	if err != nil {
		return nil, fmt.Errorf("create recurring item: %w", err)
	}
	if err := c.cs.CreateRecurringItem(ctx, item); err != nil {
		return nil, fmt.Errorf("create recurring item: %w", err)
	}
	if _, err := c.cs.LinkTransactions(ctx, item.ID, input.TransactionIDs); err != nil {
		return nil, fmt.Errorf("create recurring item: link transactions: %w", err)
	}
	return item, nil
}

// UpdateRecurringInput renames a recurring item or changes how often it is expected.
type UpdateRecurringInput struct {
	Name   string
	Rhythm Rhythm
}

// Update renames an item and sets its rhythm. The direction is not editable: it
// follows from the transactions already linked, which do not change their minds.
func (c *RecurringCommands) Update(ctx context.Context, accountID, id uuid.UUID, input UpdateRecurringInput) (*RecurringItem, error) {
	item, err := c.load(ctx, accountID, id)
	if err != nil {
		return nil, fmt.Errorf("update recurring item: %w", err)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("update recurring item: %w", ErrRecurringNameRequired)
	}
	item.Name = name
	item.Rhythm = input.Rhythm
	item.UpdatedAt = time.Now().UTC()

	updated, err := c.cs.UpdateRecurringItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("update recurring item: %w", err)
	}
	if updated == 0 {
		return nil, fmt.Errorf("update recurring item: %w", ErrRecurringItemNotFound)
	}
	return item, nil
}

// AddTransactions links further transactions to an existing item.
func (c *RecurringCommands) AddTransactions(ctx context.Context, accountID, id uuid.UUID, transactionIDs []uuid.UUID) (int, error) {
	if len(transactionIDs) == 0 {
		return 0, fmt.Errorf("link recurring transactions: %w", ErrRecurringTransactionsRequired)
	}

	item, err := c.load(ctx, accountID, id)
	if err != nil {
		return 0, fmt.Errorf("link recurring transactions: %w", err)
	}
	if item.IsEnded() {
		return 0, fmt.Errorf("link recurring transactions: %w", ErrRecurringItemEnded)
	}

	transactions, err := c.qs.ListTransactionsByIDs(ctx, accountID, transactionIDs)
	if err != nil {
		return 0, fmt.Errorf("link recurring transactions: fetch transactions: %w", err)
	}
	if len(transactions) != len(transactionIDs) {
		return 0, fmt.Errorf("link recurring transactions: %w", ErrRecurringTransactionNotFound)
	}

	linked, err := c.cs.LinkTransactions(ctx, item.ID, transactionIDs)
	if err != nil {
		return 0, fmt.Errorf("link recurring transactions: %w", err)
	}
	return linked, nil
}

// RemoveTransaction unlinks one transaction from an item. The transaction itself
// stays in Cashflow untouched — unlinking is how a double import is corrected.
func (c *RecurringCommands) RemoveTransaction(ctx context.Context, accountID, id, transactionID uuid.UUID) error {
	removed, err := c.cs.UnlinkTransaction(ctx, accountID, id, transactionID)
	if err != nil {
		return fmt.Errorf("unlink recurring transaction: %w", err)
	}
	if removed == 0 {
		return fmt.Errorf("unlink recurring transaction: %w", ErrRecurringTransactionNotFound)
	}
	return nil
}

// End stops an item from a month. Everything linked before that month stays
// linked and keeps its history; nothing after it is linked automatically again.
func (c *RecurringCommands) End(ctx context.Context, accountID, id uuid.UUID, monthRaw string) (*RecurringItem, error) {
	month, err := ParseEndMonth(monthRaw)
	if err != nil {
		return nil, fmt.Errorf("end recurring item: %w", err)
	}

	item, err := c.load(ctx, accountID, id)
	if err != nil {
		return nil, fmt.Errorf("end recurring item: %w", err)
	}

	ended, err := c.cs.EndRecurringItem(ctx, accountID, id, month)
	if err != nil {
		return nil, fmt.Errorf("end recurring item: %w", err)
	}
	if ended == 0 {
		return nil, fmt.Errorf("end recurring item: %w", ErrRecurringItemNotFound)
	}

	item.EndedFrom = &month
	item.UpdatedAt = time.Now().UTC()
	return item, nil
}

// ConfirmSuggestionInput turns a suggested pattern into a real item, with whatever
// name and rhythm the user settled on.
type ConfirmSuggestionInput struct {
	MatchKey  string
	Name      string
	Direction CashFlowDirection
	Rhythm    Rhythm
}

// ConfirmSuggestion creates the item a suggestion proposed and links every
// transaction the suggestion was found in. Nothing becomes an item without this.
func (c *RecurringCommands) ConfirmSuggestion(ctx context.Context, accountID uuid.UUID, input ConfirmSuggestionInput) (*RecurringItem, error) {
	matchKey := strings.TrimSpace(input.MatchKey)
	if matchKey == "" {
		return nil, fmt.Errorf("confirm recurring suggestion: %w", ErrRecurringSuggestionNotFound)
	}

	transactions, err := c.qs.ListUnlinkedTransactions(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("confirm recurring suggestion: fetch transactions: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(transactions))
	var oldest *Transaction
	for _, tx := range transactions {
		if tx.Direction != input.Direction || MatchKeyFor(tx.Description) != matchKey {
			continue
		}
		ids = append(ids, tx.ID)
		if oldest == nil || tx.Date.Before(oldest.Date) {
			oldest = tx
		}
	}
	if oldest == nil {
		return nil, fmt.Errorf("confirm recurring suggestion: %w", ErrRecurringSuggestionNotFound)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = SuggestedNameFor(matchKey)
	}

	item, err := NewRecurringItem(accountID, name, input.Direction, input.Rhythm, matchKey)
	if err != nil {
		return nil, fmt.Errorf("confirm recurring suggestion: %w", err)
	}
	if err := c.cs.CreateRecurringItem(ctx, item); err != nil {
		return nil, fmt.Errorf("confirm recurring suggestion: %w", err)
	}
	if _, err := c.cs.LinkTransactions(ctx, item.ID, ids); err != nil {
		return nil, fmt.Errorf("confirm recurring suggestion: link transactions: %w", err)
	}
	return item, nil
}

// DismissSuggestion refuses a suggested pattern for good.
func (c *RecurringCommands) DismissSuggestion(ctx context.Context, accountID uuid.UUID, matchKey string, direction CashFlowDirection) error {
	key := strings.TrimSpace(matchKey)
	if key == "" {
		return fmt.Errorf("dismiss recurring suggestion: %w", ErrRecurringSuggestionNotFound)
	}
	if direction != CashIn && direction != CashOut {
		return fmt.Errorf("dismiss recurring suggestion: %w", ErrUnsupportedDirection)
	}
	if err := c.cs.DismissRecurringSuggestion(ctx, accountID, key, direction); err != nil {
		return fmt.Errorf("dismiss recurring suggestion: %w", err)
	}
	return nil
}

// LinkImported attaches the rows of a finished import to the items they belong to.
// Only items the user already confirmed take part: an import never creates an
// item, and an ended item never takes another row.
func (c *RecurringCommands) LinkImported(ctx context.Context, accountID, importID uuid.UUID) (int, error) {
	items, err := c.qs.ListRecurringItems(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("link imported transactions: items: %w", err)
	}
	if len(items) == 0 {
		return 0, nil
	}

	imported, err := c.qs.ListImportedTransactions(ctx, accountID, importID)
	if err != nil {
		return 0, fmt.Errorf("link imported transactions: rows: %w", err)
	}
	if len(imported) == 0 {
		return 0, nil
	}

	lastAmounts, err := c.lastAmounts(ctx, accountID)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, item := range items {
		if item.IsEnded() {
			continue
		}
		ids := make([]uuid.UUID, 0, len(imported))
		for _, tx := range imported {
			if matchesItem(item, lastAmounts[item.ID], tx) {
				ids = append(ids, tx.ID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		linked, err := c.cs.LinkTransactions(ctx, item.ID, ids)
		if err != nil {
			return total, fmt.Errorf("link imported transactions: %w", err)
		}
		total += linked
	}
	return total, nil
}

// lastAmounts is the most recent amount per item, which the import match compares against.
func (c *RecurringCommands) lastAmounts(ctx context.Context, accountID uuid.UUID) (map[uuid.UUID]money.Price, error) {
	links, err := c.qs.ListRecurringLinks(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("link imported transactions: links: %w", err)
	}

	latest := make(map[uuid.UUID]time.Time, len(links))
	amounts := make(map[uuid.UUID]money.Price, len(links))
	for _, link := range links {
		if seen, ok := latest[link.ItemID]; ok && !link.Date.After(seen) {
			continue
		}
		latest[link.ItemID] = link.Date
		amounts[link.ItemID] = link.AmountCents
	}
	return amounts, nil
}

func (c *RecurringCommands) load(ctx context.Context, accountID, id uuid.UUID) (*RecurringItem, error) {
	item, err := c.qs.GetRecurringItem(ctx, accountID, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrRecurringItemNotFound
	}
	return item, nil
}
