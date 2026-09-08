//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// seedCashflowAccount creates the account plus its cashflow projection, which the
// transactions table is FK-bound to.
func seedCashflowAccount(t *testing.T, db *storage.DB, email string) uuid.UUID {
	t.Helper()

	accountID := seedAccount(t, db, email, email)

	// The transactions table is FK-bound to the cashflow projection, which is normally
	// created by the AccountCreated event handler. Inserted directly here so the test
	// exercises the store rather than the event bus.
	if _, err := db.ExecContext(t.Context(),
		db.Rebind(`INSERT INTO cashflow_accounts (id, account_id, created_at, updated_at) VALUES (?, ?, ?, ?)`),
		uuid.New(), accountID, time.Now().UTC(), time.Now().UTC(),
	); err != nil {
		t.Fatalf("create cashflow projection: %v", err)
	}
	return accountID
}

func seedTransaction(t *testing.T, db *storage.DB, accountID uuid.UUID, description string) uuid.UUID {
	t.Helper()

	store := storage.NewSQLXCashflowStore(db)
	tx := &cashflow.Transaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		Description: description,
		Source:      "manual",
		Direction:   cashflow.CashOut,
		AmountCents: money.Price(1000),
		Date:        time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if _, err := store.CreateTransactions(t.Context(), []*cashflow.Transaction{tx}); err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	return tx.ID
}

// Listing must return only the caller's rows. Before account scoping the cashflow
// queries carried no account predicate at all, so every account saw everything.
func TestCashflowListIsScopedToTheAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		seedTransaction(t, db, mine, "my groceries")
		seedTransaction(t, db, theirs, "their rent")

		store := storage.NewSQLXCashflowStore(db)
		result, err := store.ListTransactions(t.Context(), cashflow.TransactionListQuery{
			AccountID: mine,
			Sort:      sorting.Sort{Field: "date", Direction: sorting.DESC},
		})
		if err != nil {
			t.Fatalf("list transactions: %v", err)
		}

		if result.Total != 1 {
			t.Fatalf("total = %d, want only the caller's single transaction", result.Total)
		}
		for _, tx := range result.Transactions {
			if tx.AccountID != mine {
				t.Fatalf("listing leaked a transaction belonging to %v", tx.AccountID)
			}
		}
	})
}

func TestCashflowAnalyticsAreScopedToTheAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		seedTransaction(t, db, mine, "my groceries")
		seedTransaction(t, db, theirs, "their rent")
		seedTransaction(t, db, theirs, "their car")

		store := storage.NewSQLXCashflowStore(db)
		points, err := store.GetMonthlyAnalytics(t.Context(), cashflow.AnalyticsFilter{AccountID: mine})
		if err != nil {
			t.Fatalf("monthly analytics: %v", err)
		}

		var total int64
		for _, p := range points {
			total += p.OutgoingCents
		}
		if total != 1000 {
			t.Fatalf("outgoing total = %d, want 1000 (only the caller's transaction)", total)
		}
	})
}

// Tagging by id must not reach another account's rows even when the id is known, which
// is exactly what a client replaying an id from elsewhere would attempt.
func TestCashflowTagByIDsCannotCrossAccounts(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		theirTx := seedTransaction(t, db, theirs, "their rent")

		store := storage.NewSQLXCashflowStore(db)
		updated, err := store.UpdateTagByIDs(t.Context(), mine, []uuid.UUID{theirTx}, "hijacked")
		if err != nil {
			t.Fatalf("update tag: %v", err)
		}
		if updated != 0 {
			t.Fatalf("updated = %d, want 0 -- another account's transaction was modified", updated)
		}

		result, err := store.ListTransactions(t.Context(), cashflow.TransactionListQuery{
			AccountID: theirs,
			Sort:      sorting.Sort{Field: "date", Direction: sorting.DESC},
		})
		if err != nil {
			t.Fatalf("list transactions: %v", err)
		}
		for _, tx := range result.Transactions {
			if tx.Tag == "hijacked" {
				t.Fatal("another account's transaction was tagged")
			}
		}
	})
}

func TestCashflowIgnoreByIDsCannotCrossAccounts(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		theirTx := seedTransaction(t, db, theirs, "their rent")

		store := storage.NewSQLXCashflowStore(db)
		updated, err := store.UpdateIgnoredByIDs(t.Context(), mine, []uuid.UUID{theirTx}, true)
		if err != nil {
			t.Fatalf("update ignored: %v", err)
		}
		if updated != 0 {
			t.Fatalf("updated = %d, want 0 -- another account's transaction was modified", updated)
		}
	})
}

// A filtered bulk update is the widest-reaching mutation in the feature: unscoped, one
// request would retag every account's transactions at once.
func TestCashflowTagByFilterCannotCrossAccounts(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		seedTransaction(t, db, mine, "my groceries")
		seedTransaction(t, db, theirs, "their rent")

		store := storage.NewSQLXCashflowStore(db)
		updated, err := store.UpdateTagByFilter(t.Context(), cashflow.TransactionFilters{AccountID: mine}, "mine")
		if err != nil {
			t.Fatalf("update tag by filter: %v", err)
		}
		if updated != 1 {
			t.Fatalf("updated = %d, want 1 -- the filter reached beyond the caller's account", updated)
		}
	})
}

func TestCashflowCountByFilterIsScopedToTheAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		seedTransaction(t, db, mine, "my groceries")
		seedTransaction(t, db, theirs, "their rent")
		seedTransaction(t, db, theirs, "their car")

		store := storage.NewSQLXCashflowStore(db)
		count, err := store.CountByFilter(t.Context(), cashflow.TransactionFilters{AccountID: mine})
		if err != nil {
			t.Fatalf("count by filter: %v", err)
		}
		if count != 1 {
			t.Fatalf("count = %d, want 1", count)
		}
	})
}
