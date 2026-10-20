//go:build integration

package integration

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/storage"
	"github.com/lennardclaproth/ta11y/migrations"
)

// The cashflow projection is what every cashflow transaction references; without it
// the insert fails on the account foreign key. Creating it twice must be harmless,
// because the account.created handler can be redelivered.
func TestCashflowAccountProjectionEnablesTransactions(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "cashflow@example.com", "cashflow@example.com")
		store := storage.NewSQLXCashflowStore(db)

		for range 2 {
			if err := store.CreateAccount(t.Context(), cashflow.NewAccount(accountID)); err != nil {
				t.Fatalf("create cashflow projection: %v", err)
			}
		}

		var projections int
		if err := db.GetContext(t.Context(), &projections,
			db.Rebind(`SELECT COUNT(*) FROM cashflow_accounts WHERE account_id = ?`), accountID,
		); err != nil {
			t.Fatalf("count projections: %v", err)
		}
		if projections != 1 {
			t.Fatalf("expected one cashflow projection, got %d", projections)
		}

		seedTransaction(t, db, accountID, "groceries")
	})
}

// An account created before the cashflow account.created handler existed has no
// projection; the backfill migration gives it one with a well-formed id.
func TestCashflowAccountsBackfillMigration(t *testing.T) {
	const backfillVersion = 20261012120000

	db := storage.NewDB(filepath.Join(t.TempDir(), "backfill.db"), storage.Sqlite)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test db: %v", err)
		}
	})
	provider, err := goose.NewProvider(goose.DialectSQLite3, db.DB.DB, migrations.GetFS(storage.Sqlite))
	if err != nil {
		t.Fatalf("create migration provider: %v", err)
	}
	if _, err := provider.UpTo(t.Context(), backfillVersion-1); err != nil {
		t.Fatalf("migrate to before the backfill: %v", err)
	}

	accountID := seedAccount(t, db, "existing@example.com", "existing@example.com")

	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatalf("run backfill: %v", err)
	}

	var id string
	if err := db.GetContext(t.Context(), &id,
		db.Rebind(`SELECT id FROM cashflow_accounts WHERE account_id = ?`), accountID,
	); err != nil {
		t.Fatalf("read backfilled projection: %v", err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(id) {
		t.Fatalf("expected a version-4 UUID id, got %q", id)
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("backfilled id does not parse: %v", err)
	}

	seedTransaction(t, db, accountID, "groceries")
}
