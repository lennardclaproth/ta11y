package portfolio

import (
	"context"
	"errors"
	"sync"
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
	edited            *Transaction
	deletedID         uuid.UUID
}

func (s *transactionDateStore) UpdateManualTransaction(_ context.Context, tx *Transaction) (int, error) {
	s.edited = tx
	return 1, nil
}

func (s *transactionDateStore) DeleteManualTransaction(_ context.Context, _, id uuid.UUID) (int, error) {
	s.deletedID = id
	return 1, nil
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

// stubRebuilder records the rebuilds the background queue runs, answering each with the
// next of errs (nil once they run out) and signalling built after every call.
type stubRebuilder struct {
	mu    sync.Mutex
	calls int
	errs  []error
	built chan struct{}
}

func newStubRebuilder(errs ...error) *stubRebuilder {
	return &stubRebuilder{errs: errs, built: make(chan struct{}, 16)}
}

func (r *stubRebuilder) Build(_ context.Context, _ uuid.UUID) error {
	r.mu.Lock()
	r.calls++
	var err error
	if len(r.errs) > 0 {
		err, r.errs = r.errs[0], r.errs[1:]
	}
	r.mu.Unlock()
	r.built <- struct{}{}
	return err
}

func (r *stubRebuilder) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

// waitForBuilds blocks until the queue has run n rebuilds.
func (r *stubRebuilder) waitForBuilds(t *testing.T, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		select {
		case <-r.built:
		case <-time.After(5 * time.Second):
			t.Fatalf("expected %d background rebuilds, saw %d", n, i)
		}
	}
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
	rb := newStubRebuilder()
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

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
	if rebuild.Outcome != RebuildStarted {
		t.Fatalf("expected the rebuild to be started in the background, got %q", rebuild.Outcome)
	}
	rb.waitForBuilds(t, 1)
}

func TestBackgroundRebuildWaitsOutARebuildStartedElsewhere(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	// The first attempt finds an import's rebuild holding the lock; that one may have read
	// the transactions before this write, so the queue tries again rather than giving up.
	rb := newStubRebuilder(ErrBuildInProgress)
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)
	commands.rebuilds.retryDelay = time.Millisecond

	if _, _, err := commands.ChangeTransactionDate(context.Background(), accID, current.ID, "2026-07-14"); err != nil {
		t.Fatalf("expected the move to succeed even with a rebuild running, got %v", err)
	}
	rb.waitForBuilds(t, 2)
}

// gatedRebuilder blocks its first Build until released, so a test can pile requests up
// behind a running rebuild.
type gatedRebuilder struct {
	mu      sync.Mutex
	calls   int
	started chan struct{}
	release chan struct{}
	done    chan struct{}
}

func (r *gatedRebuilder) Build(_ context.Context, _ uuid.UUID) error {
	r.mu.Lock()
	r.calls++
	first := r.calls == 1
	r.mu.Unlock()
	if first {
		r.started <- struct{}{}
		<-r.release
	}
	r.done <- struct{}{}
	return nil
}

func TestBackgroundRebuildFoldsRequestsIntoOneFollowUp(t *testing.T) {
	accID := uuid.New()
	rb := &gatedRebuilder{started: make(chan struct{}, 1), release: make(chan struct{}), done: make(chan struct{}, 8)}
	queue := newRebuildQueue(rb, nil)

	queue.request(context.Background(), accID)
	<-rb.started
	// Three writes land while the first rebuild runs; one follow-up covers all of them.
	for i := 0; i < 3; i++ {
		queue.request(context.Background(), accID)
	}
	close(rb.release)

	for i := 0; i < 2; i++ {
		select {
		case <-rb.done:
		case <-time.After(5 * time.Second):
			t.Fatalf("expected two rebuilds, saw %d", i)
		}
	}
	select {
	case <-rb.done:
		t.Fatal("expected the requests made mid-run to fold into a single follow-up")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestChangeTransactionDateRebuildOutlivesTheRequest(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	rb := &contextRebuilder{started: make(chan context.Context, 1), release: make(chan struct{})}
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, rebuild, err := commands.ChangeTransactionDate(ctx, accID, current.ID, "2026-07-14")
	if err != nil {
		t.Fatalf("change transaction date: %v", err)
	}
	if rebuild.Outcome != RebuildStarted {
		t.Fatalf("expected the write to report the rebuild as started, got %q", rebuild.Outcome)
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
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, newStubRebuilder(), nil)

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
	rb := newStubRebuilder()
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

	if _, _, err := commands.ChangeTransactionDate(context.Background(), accID, current.ID, "2026-07-14"); !errors.Is(err, ErrDuplicateTransaction) {
		t.Fatalf("expected ErrDuplicateTransaction, got %v", err)
	}
	if !store.updatedOccurredAt.IsZero() {
		t.Fatal("expected a refused change not to reach the store")
	}
	if rb.callCount() != 0 {
		t.Fatal("expected a refused change not to start a rebuild")
	}
}

func importedTx(t *testing.T, accID uuid.UUID) *Transaction {
	t.Helper()
	tx, err := NewTransaction(TransactionData{
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
	return tx
}

// cashEdit is an edit that needs no listing lookup, so it runs without a market data store.
func cashEdit(accID, id uuid.UUID, amount string) ManualTransactionUpdateInput {
	description := "  top-up  "
	return ManualTransactionUpdateInput{
		AccountID:   accID,
		ID:          id,
		OccurredAt:  "2026-07-14",
		Type:        "CASH",
		Amount:      amount,
		Description: &description,
	}
}

func TestUpdateTransactionReplacesManualFieldsAndRebuilds(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	rb := newStubRebuilder()
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

	result, err := commands.UpdateTransaction(context.Background(), cashEdit(accID, current.ID, "-250"))
	if err != nil {
		t.Fatalf("update transaction: %v", err)
	}
	edited := store.edited
	if edited == nil {
		t.Fatal("expected the edit to reach the store")
	}
	if edited.ID != current.ID || !edited.CreatedAt.Equal(current.CreatedAt) {
		t.Fatal("expected the edit to keep the row's identity")
	}
	if edited.Type != TxCash || edited.Symbol != nil || edited.Description != "top-up" {
		t.Fatalf("expected the edited fields to be stored, got type %q symbol %v description %q", edited.Type, edited.Symbol, edited.Description)
	}
	if edited.Source != current.Source || !edited.IsManual() {
		t.Fatal("expected the vendor and origin to stay as entered")
	}
	if edited.Checksum == current.Checksum {
		t.Fatal("expected the checksum to follow the edited content")
	}
	if result.SignedAmount != -250 {
		t.Fatalf("expected a signed amount of -250, got %v", result.SignedAmount)
	}
	if result.Rebuild.Outcome != RebuildStarted {
		t.Fatalf("expected a background rebuild, got %q", result.Rebuild.Outcome)
	}
	rb.waitForBuilds(t, 1)
}

func TestUpdateTransactionRefusesImportedTransaction(t *testing.T) {
	accID := uuid.New()
	imported := importedTx(t, accID)
	store := &transactionDateStore{current: imported, byChecksum: map[string]*Transaction{}}
	rb := newStubRebuilder()
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

	if _, err := commands.UpdateTransaction(context.Background(), cashEdit(accID, imported.ID, "100")); !errors.Is(err, ErrTransactionNotEditable) {
		t.Fatalf("expected ErrTransactionNotEditable, got %v", err)
	}
	if store.edited != nil || rb.callCount() != 0 {
		t.Fatal("expected a refused edit neither to reach the store nor to rebuild")
	}
}

func TestUpdateTransactionAppliesTheCreateRules(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, newStubRebuilder(), nil)

	if _, err := commands.UpdateTransaction(context.Background(), cashEdit(accID, current.ID, "0")); !errors.Is(err, ErrManualCashAmountMustBeNonZero) {
		t.Fatalf("expected ErrManualCashAmountMustBeNonZero, got %v", err)
	}
	if store.edited != nil {
		t.Fatal("expected an invalid edit not to reach the store")
	}
}

func TestUpdateTransactionRefusesDuplicate(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	clash, err := NewManualTransaction(TransactionData{
		Source:      current.Source,
		OccurredAt:  time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
		Description: "top-up",
		Type:        TxCash,
		Amount:      -250,
	}, accID)
	if err != nil {
		t.Fatalf("new manual transaction: %v", err)
	}
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{clash.Checksum: clash}}
	rb := newStubRebuilder()
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

	if _, err := commands.UpdateTransaction(context.Background(), cashEdit(accID, current.ID, "-250")); !errors.Is(err, ErrDuplicateTransaction) {
		t.Fatalf("expected ErrDuplicateTransaction, got %v", err)
	}
	if store.edited != nil || rb.callCount() != 0 {
		t.Fatal("expected a refused edit neither to reach the store nor to rebuild")
	}
}

func TestDeleteTransactionRemovesManualTransactionAndRebuilds(t *testing.T) {
	accID := uuid.New()
	current := manualTx(t, accID, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	store := &transactionDateStore{current: current, byChecksum: map[string]*Transaction{}}
	rb := newStubRebuilder()
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

	rebuild, err := commands.DeleteTransaction(context.Background(), accID, current.ID)
	if err != nil {
		t.Fatalf("delete transaction: %v", err)
	}
	if store.deletedID != current.ID {
		t.Fatal("expected the transaction to be deleted")
	}
	if rebuild.Outcome != RebuildStarted {
		t.Fatalf("expected a background rebuild, got %q", rebuild.Outcome)
	}
	rb.waitForBuilds(t, 1)
}

func TestDeleteTransactionRefusesImportedTransaction(t *testing.T) {
	accID := uuid.New()
	imported := importedTx(t, accID)
	store := &transactionDateStore{current: imported, byChecksum: map[string]*Transaction{}}
	rb := newStubRebuilder()
	commands := NewCommands(store, store, marketdata.Queries{}, vendor.Queries{}, rb, nil)

	if _, err := commands.DeleteTransaction(context.Background(), accID, imported.ID); !errors.Is(err, ErrTransactionNotEditable) {
		t.Fatalf("expected ErrTransactionNotEditable, got %v", err)
	}
	if store.deletedID != uuid.Nil || rb.callCount() != 0 {
		t.Fatal("expected a refused delete neither to reach the store nor to rebuild")
	}
}
