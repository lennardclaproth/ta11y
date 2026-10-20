package cashflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
)

// recordingCashflowStore records the account id of the created projection.
type recordingCashflowStore struct{ createdAccountID uuid.UUID }

func (s *recordingCashflowStore) CreateAccount(_ context.Context, acc *cashflow.Account) error {
	s.createdAccountID = acc.AccountID
	return nil
}
func (s *recordingCashflowStore) CreateTransactions(_ context.Context, _ []*cashflow.Transaction) (int, error) {
	return 0, nil
}
func (s *recordingCashflowStore) UpdateDate(_ context.Context, _, _ uuid.UUID, _ time.Time, _ string) (int, error) {
	return 0, nil
}
func (s *recordingCashflowStore) UpdateTagByIDs(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ string) (int, error) {
	return 0, nil
}
func (s *recordingCashflowStore) UpdateTagByFilter(_ context.Context, _ cashflow.TransactionFilters, _ string) (int, error) {
	return 0, nil
}
func (s *recordingCashflowStore) UpdateIgnoredByIDs(_ context.Context, _ uuid.UUID, _ []uuid.UUID, _ bool) (int, error) {
	return 0, nil
}
func (s *recordingCashflowStore) UpdateIgnoredByFilter(_ context.Context, _ cashflow.TransactionFilters, _ bool) (int, error) {
	return 0, nil
}

func TestAccountCreatedCreatesCashflowProjection(t *testing.T) {
	store := &recordingCashflowStore{}
	commands := cashflow.NewCommands(store, nil, nil)
	handler := NewAccountCreatedHandler(commands, nil)

	accID := uuid.New()
	if err := handler.Handle(context.Background(), account.Created{AccID: accID}, eventbus.Metadata{}); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if store.createdAccountID != accID {
		t.Fatalf("expected cashflow projection for %s, got %s", accID, store.createdAccountID)
	}
}
