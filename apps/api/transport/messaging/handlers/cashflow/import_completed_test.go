package cashflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	"github.com/lennardclaproth/ta11y/internal/importer"
)

// recordingStore satisfies the recurring store contracts and records which
// transactions were linked, so the test can observe what the handler did.
type recordingStore struct {
	items    []*cashflow.RecurringItem
	links    []cashflow.RecurringLink
	imported []*cashflow.Transaction
	linked   []uuid.UUID
}

func (s *recordingStore) CreateRecurringItem(_ context.Context, _ *cashflow.RecurringItem) error {
	return nil
}
func (s *recordingStore) UpdateRecurringItem(_ context.Context, _ *cashflow.RecurringItem) (int, error) {
	return 0, nil
}
func (s *recordingStore) EndRecurringItem(_ context.Context, _, _ uuid.UUID, _ time.Time) (int, error) {
	return 0, nil
}
func (s *recordingStore) LinkTransactions(_ context.Context, _ uuid.UUID, ids []uuid.UUID) (int, error) {
	s.linked = append(s.linked, ids...)
	return len(ids), nil
}
func (s *recordingStore) UnlinkTransaction(_ context.Context, _, _, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (s *recordingStore) DismissRecurringSuggestion(_ context.Context, _ uuid.UUID, _ string, _ cashflow.CashFlowDirection) error {
	return nil
}
func (s *recordingStore) ListRecurringItems(_ context.Context, _ uuid.UUID) ([]*cashflow.RecurringItem, error) {
	return s.items, nil
}
func (s *recordingStore) GetRecurringItem(_ context.Context, _, _ uuid.UUID) (*cashflow.RecurringItem, error) {
	return nil, nil
}
func (s *recordingStore) ListRecurringLinks(_ context.Context, _ uuid.UUID) ([]cashflow.RecurringLink, error) {
	return s.links, nil
}
func (s *recordingStore) ListRecurringLinksForItem(_ context.Context, _, itemID uuid.UUID) ([]cashflow.RecurringLink, error) {
	own := []cashflow.RecurringLink{}
	for _, link := range s.links {
		if link.ItemID == itemID {
			own = append(own, link)
		}
	}
	return own, nil
}
func (s *recordingStore) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}
func (s *recordingStore) ListUnlinkedTransactions(_ context.Context, _ uuid.UUID) ([]*cashflow.Transaction, error) {
	return nil, nil
}
func (s *recordingStore) ListDismissedSuggestions(_ context.Context, _ uuid.UUID) ([]cashflow.DismissedSuggestion, error) {
	return nil, nil
}
func (s *recordingStore) ListTransactionsByIDs(_ context.Context, _ uuid.UUID, _ []uuid.UUID) ([]*cashflow.Transaction, error) {
	return nil, nil
}
func (s *recordingStore) ListImportedTransactions(_ context.Context, _, _ uuid.UUID) ([]*cashflow.Transaction, error) {
	return s.imported, nil
}

func storeWithOneItem() (*recordingStore, *cashflow.Transaction) {
	item := &cashflow.RecurringItem{
		ID:        uuid.New(),
		Name:      "Pixel Stream",
		Direction: cashflow.CashOut,
		Rhythm:    cashflow.RhythmMonthly,
		MatchKey:  cashflow.MatchKeyFor("PIXEL STREAM MONTHLY"),
	}
	match := &cashflow.Transaction{
		ID:          uuid.New(),
		Description: "PIXEL STREAM MONTHLY 1018",
		Direction:   cashflow.CashOut,
		AmountCents: 1200,
		Date:        time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC),
	}
	return &recordingStore{
		items: []*cashflow.RecurringItem{item},
		links: []cashflow.RecurringLink{{
			ItemID:        item.ID,
			TransactionID: uuid.New(),
			Date:          time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC),
			AmountCents:   1200,
		}},
		imported: []*cashflow.Transaction{match},
	}, match
}

func TestImportCompletedLinksCashflowRowsToConfirmedItems(t *testing.T) {
	store, match := storeWithOneItem()
	handler := NewImportCompletedHandler(cashflow.NewRecurringCommands(store, store, store), nil)

	accID := uuid.New()
	err := handler.Handle(context.Background(), importer.Completed{
		ImportID:  uuid.New(),
		Type:      importer.ImportTypeCashflow,
		AccountID: &accID,
	}, eventbus.Metadata{})

	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.linked) != 1 || store.linked[0] != match.ID {
		t.Fatalf("linked %v, want the matching imported row", store.linked)
	}
}

func TestImportCompletedIgnoresOtherImportTypes(t *testing.T) {
	store, _ := storeWithOneItem()
	handler := NewImportCompletedHandler(cashflow.NewRecurringCommands(store, store, store), nil)

	accID := uuid.New()
	err := handler.Handle(context.Background(), importer.Completed{
		ImportID:  uuid.New(),
		Type:      importer.ImportTypePortfolio,
		AccountID: &accID,
	}, eventbus.Metadata{})

	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(store.linked) != 0 {
		t.Fatalf("linked %d rows for a portfolio import, want none", len(store.linked))
	}
}
