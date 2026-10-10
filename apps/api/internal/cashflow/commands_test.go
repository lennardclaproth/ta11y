package cashflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// dateStore is a CommandStore/QueryStore pair holding a single transaction, enough to
// drive ChangeDate without a database.
type dateStore struct {
	current    *Transaction
	byChecksum map[string]*Transaction

	updatedDate     time.Time
	updatedChecksum string
}

func (s *dateStore) GetTransaction(_ context.Context, _, _ uuid.UUID) (*Transaction, error) {
	return s.current, nil
}

func (s *dateStore) GetTransactionByChecksum(_ context.Context, _ uuid.UUID, checksum string) (*Transaction, error) {
	return s.byChecksum[checksum], nil
}

func (s *dateStore) UpdateDate(_ context.Context, _, _ uuid.UUID, date time.Time, checksum string) (int, error) {
	s.updatedDate = date
	s.updatedChecksum = checksum
	return 1, nil
}

func (s *dateStore) CreateTransactions(_ context.Context, _ []*Transaction) (int, error) {
	return 0, nil
}
func (s *dateStore) UpdateTagByIDs(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ string) (int, error) {
	return 0, nil
}
func (s *dateStore) UpdateTagByFilter(_ context.Context, _ TransactionFilters, _ string) (int, error) {
	return 0, nil
}
func (s *dateStore) UpdateIgnoredByIDs(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ bool) (int, error) {
	return 0, nil
}
func (s *dateStore) UpdateIgnoredByFilter(_ context.Context, _ TransactionFilters, _ bool) (int, error) {
	return 0, nil
}
func (s *dateStore) GetMonthlyAnalytics(_ context.Context, _ AnalyticsFilter) ([]MonthlyAnalyticsPoint, error) {
	return nil, nil
}
func (s *dateStore) GetTagDistribution(_ context.Context, _ AnalyticsFilter) (*TagDistribution, error) {
	return nil, nil
}
func (s *dateStore) ListTransactions(_ context.Context, _ TransactionListQuery) (*TransactionListResult, error) {
	return nil, nil
}
func (s *dateStore) CountByFilter(_ context.Context, _ TransactionFilters) (int, error) {
	return 0, nil
}

func manualTransaction(t *testing.T, accID uuid.UUID, source string, day time.Time) *Transaction {
	t.Helper()
	amount, err := money.NewPrice(12.5)
	if err != nil {
		t.Fatalf("new price: %v", err)
	}
	tx, err := NewTransaction("Bike repair", "Entered late", source, "household", CashOut, amount, day, 7, 7, nil, nil, accID)
	if err != nil {
		t.Fatalf("new transaction: %v", err)
	}
	return tx
}

func TestNewTransactionDataRejectsFutureDate(t *testing.T) {
	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	_, err := NewTransactionData(tomorrow, "10", "expense", "Bike repair", "Entered late", "household", "", SourceManual, nil)
	if !errors.Is(err, ErrCashflowDateInFuture) {
		t.Fatalf("expected ErrCashflowDateInFuture, got %v", err)
	}
}

func TestChangeDateMovesManualTransaction(t *testing.T) {
	accID := uuid.New()
	current := manualTransaction(t, accID, SourceManual, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &dateStore{current: current, byChecksum: map[string]*Transaction{}}
	commands := NewCommands(store, store, nil, nil, nil)

	moved, err := commands.ChangeDate(context.Background(), accID, current.ID, "2026-07-14")
	if err != nil {
		t.Fatalf("change date: %v", err)
	}
	if want := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC); !moved.Date.Equal(want) {
		t.Fatalf("expected date %s, got %s", want, moved.Date)
	}
	if !store.updatedDate.Equal(moved.Date) {
		t.Fatalf("expected the stored date to be %s, got %s", moved.Date, store.updatedDate)
	}
	// The checksum carries the date, so a moved row must not keep the old one.
	if store.updatedChecksum == current.Checksum || store.updatedChecksum != moved.Checksum {
		t.Fatalf("expected a recomputed checksum, got %q", store.updatedChecksum)
	}
}

func TestChangeDateRefusesImportedTransaction(t *testing.T) {
	accID := uuid.New()
	current := manualTransaction(t, accID, "ing", time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &dateStore{current: current, byChecksum: map[string]*Transaction{}}
	commands := NewCommands(store, store, nil, nil, nil)

	if _, err := commands.ChangeDate(context.Background(), accID, current.ID, "2026-07-14"); !errors.Is(err, ErrCashflowDateNotEditable) {
		t.Fatalf("expected ErrCashflowDateNotEditable, got %v", err)
	}
}

func TestChangeDateRefusesFutureDate(t *testing.T) {
	accID := uuid.New()
	current := manualTransaction(t, accID, SourceManual, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &dateStore{current: current, byChecksum: map[string]*Transaction{}}
	commands := NewCommands(store, store, nil, nil, nil)

	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	if _, err := commands.ChangeDate(context.Background(), accID, current.ID, tomorrow); !errors.Is(err, ErrCashflowDateInFuture) {
		t.Fatalf("expected ErrCashflowDateInFuture, got %v", err)
	}
}

func TestChangeDateRefusesDuplicate(t *testing.T) {
	accID := uuid.New()
	current := manualTransaction(t, accID, SourceManual, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	clash := current.MovedTo(time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC))
	clash.ID = uuid.New()
	store := &dateStore{current: current, byChecksum: map[string]*Transaction{clash.Checksum: clash}}
	commands := NewCommands(store, store, nil, nil, nil)

	if _, err := commands.ChangeDate(context.Background(), accID, current.ID, "2026-07-14"); !errors.Is(err, ErrDuplicateTransaction) {
		t.Fatalf("expected ErrDuplicateTransaction, got %v", err)
	}
	if !store.updatedDate.IsZero() {
		t.Fatal("expected a refused change not to reach the store")
	}
}
