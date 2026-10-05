package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

// transactionDateStore is a CommandStore/TransactionReader pair holding a single
// transaction, enough to drive ChangeTransactionDate without a database.
type transactionDateStore struct {
	current    *Transaction
	byChecksum map[string]*Transaction

	updatedOccurredAt time.Time
	updatedChecksum   string
}

func (s *transactionDateStore) Transaction(_ context.Context, _, _ uuid.UUID) (*Transaction, error) {
	return s.current, nil
}

func (s *transactionDateStore) TransactionByChecksum(_ context.Context, _ uuid.UUID, checksum string) (*Transaction, error) {
	return s.byChecksum[checksum], nil
}

func (s *transactionDateStore) UpdateTransactionOccurredAt(_ context.Context, _, _ uuid.UUID, occurredAt time.Time, checksum string) (int, error) {
	s.updatedOccurredAt = occurredAt
	s.updatedChecksum = checksum
	return 1, nil
}

func (s *transactionDateStore) CreateAccount(_ context.Context, _ *Account) error { return nil }
func (s *transactionDateStore) CreateTransaction(_ context.Context, _ *Transaction) error {
	return nil
}
func (s *transactionDateStore) CreateTransactions(_ context.Context, _ []*Transaction) (int, error) {
	return 0, nil
}

// stubRebuilder records the rebuild a write asks for and answers with a fixed error.
type stubRebuilder struct {
	calls int
	err   error
}

func (r *stubRebuilder) Build(_ context.Context, _ uuid.UUID) error {
	r.calls++
	return r.err
}

// contextRebuilder hands back the context its Build was called with and blocks until
// released, so a test can tell a detached rebuild from one that dies with the request.
type contextRebuilder struct {
	started chan context.Context
	release chan struct{}
}

func (r *contextRebuilder) Build(ctx context.Context, _ uuid.UUID) error {
	r.started <- ctx
	<-r.release
	return nil
}

func manualTx(t *testing.T, accID uuid.UUID, day time.Time) *Transaction {
	t.Helper()
	tx, err := NewManualTransaction(TransactionData{
		Source:     "degiro",
		OccurredAt: day,
		Symbol:     ptrString("NWIF"),
		Type:       TxBuy,
		Quantity:   10,
		Price:      90,
		Amount:     900,
	}, accID)
	if err != nil {
		t.Fatalf("new manual transaction: %v", err)
	}
	return tx
}

func TestCreateTransactionRejectsFutureOccurredAt(t *testing.T) {
	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	if _, err := parseOccurredAt(tomorrow); !errors.Is(err, ErrManualOccurredAtInFuture) {
		t.Fatalf("expected ErrManualOccurredAtInFuture, got %v", err)
	}
}

func TestChangeTransactionDateMovesManualTransactionAndRebuilds(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	rb := &stubRebuilder{}
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb)

	moved, rebuild, err := commands.ChangeTransactionDate(context.Background(), accID, current.ID, "2026-07-14")
	if err != nil {
		t.Fatalf("change transaction date: %v", err)
	}
	if want := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC); !moved.OccurredAt.Equal(want) {
		t.Fatalf("expected occurred_at %s, got %s", want, moved.OccurredAt)
	}
	if !store.updatedOccurredAt.Equal(moved.OccurredAt) {
		t.Fatalf("expected the stored date to be %s, got %s", moved.OccurredAt, store.updatedOccurredAt)
	}
	if store.updatedChecksum == current.Checksum {
		t.Fatal("expected the checksum to follow the new date")
	}
	if rb.calls != 1 {
		t.Fatalf("expected exactly one rebuild, got %d", rb.calls)
	}
	if rebuild.Outcome != RebuildCompleted {
		t.Fatalf("expected a completed rebuild, got %q", rebuild.Outcome)
	}
}

func TestChangeTransactionDateReportsRebuildAlreadyRunning(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, &stubRebuilder{err: ErrBuildInProgress})

	_, rebuild, err := commands.ChangeTransactionDate(context.Background(), accID, current.ID, "2026-07-14")
	if err != nil {
		t.Fatalf("expected the move to succeed even with a rebuild running, got %v", err)
	}
	if rebuild.Outcome != RebuildInProgress {
		t.Fatalf("expected in_progress, got %q", rebuild.Outcome)
	}
}

func TestChangeTransactionDateRebuildOutlivesTheRequest(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	rb := &contextRebuilder{started: make(chan context.Context, 1), release: make(chan struct{})}
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, rebuild, err := commands.ChangeTransactionDate(ctx, accID, current.ID, "2026-07-14")
	if err != nil {
		t.Fatalf("change transaction date: %v", err)
	}
	if rebuild.Outcome != RebuildInProgress {
		t.Fatalf("expected the write to report the rebuild as still running, got %q", rebuild.Outcome)
	}

	select {
	case buildCtx := <-rb.started:
		if buildCtx.Err() != nil {
			t.Fatalf("expected the rebuild to run on a live context, got %v", buildCtx.Err())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("expected the rebuild to start even though the request was cancelled")
	}
	close(rb.release)
}

func TestChangeTransactionDateRefusesImportedTransaction(t *testing.T) {
	accID := uuid.New()
	imported, err := NewTransaction(TransactionData{
		Source:     "degiro",
		OccurredAt: time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
		Symbol:     ptrString("NWIF"),
		Type:       TxBuy,
		Quantity:   10,
		Price:      90,
		Amount:     900,
	}, 4, uuid.New(), &accID, nil)
	if err != nil {
		t.Fatalf("new transaction: %v", err)
	}
	store := &transactionDateStore{current: imported, byChecksum: map[string]*Transaction{}}
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, &stubRebuilder{})

	if _, _, err := commands.ChangeTransactionDate(context.Background(), accID, imported.ID, "2026-07-14"); !errors.Is(err, ErrTransactionDateNotEditable) {
		t.Fatalf("expected ErrTransactionDateNotEditable, got %v", err)
	}
}

func TestChangeTransactionDateRefusesDuplicate(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	clash := current.MovedTo(time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC))
	clash.ID = uuid.New()
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{clash.Checksum: clash}}
	rb := &stubRebuilder{}
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb)

	if _, _, err := commands.ChangeTransactionDate(context.Background(), accID, current.ID, "2026-07-14"); !errors.Is(err, ErrDuplicateTransaction) {
		t.Fatalf("expected ErrDuplicateTransaction, got %v", err)
	}
	if !store.updatedOccurredAt.IsZero() {
		t.Fatal("expected a refused change not to reach the store")
	}
	if rb.calls != 0 {
		t.Fatal("expected a refused change not to start a rebuild")
	}
}
