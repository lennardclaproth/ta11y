//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
	"github.com/lennardclaproth/ta11y/internal/storage"
)

// seedPortfolioAccount creates an account with its portfolio projection, which the
// transactions table references.
func seedPortfolioAccount(t *testing.T, db *storage.DB) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	accountID := seedAccount(t, db, "investor", "investor@example.com")
	if _, err := db.GetExecutor(ctx).ExecContext(ctx,
		db.Rebind(`INSERT INTO portfolio_accounts (id, account_id) VALUES (?, ?)`),
		uuid.New(), accountID,
	); err != nil {
		t.Fatalf("seed portfolio account: %v", err)
	}
	return accountID
}

// TestPortfolioTransactionsShowAndFindTheLinkedListingSymbol covers a row that carries only
// an ISIN, as imported rows do: once a rebuild maps it to a position linked to a listing,
// the list reports that listing's symbol and searching by it finds the row.
func TestPortfolioTransactionsShowAndFindTheLinkedListingSymbol(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXPortfolioStore(db)
		accountID := seedPortfolioAccount(t, db)
		listing := newListing(t, storage.NewSQLXMarketDataStore(db), ctx, "TDT.AS")

		isin := "NL0009272749"
		tx, err := portfolio.NewManualTransaction(portfolio.TransactionData{
			Source:     "DEGIRO",
			OccurredAt: time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC),
			ISIN:       &isin,
			Type:       portfolio.TxBuy,
			Quantity:   3,
			Price:      99.71,
			Amount:     299.13,
		}, accountID)
		if err != nil {
			t.Fatalf("new transaction: %v", err)
		}
		if err := store.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("create transaction: %v", err)
		}
		position := &portfolio.Position{
			ID:        uuid.New(),
			AccountID: accountID,
			ISIN:      &isin,
			ListingID: &listing.ID,
			OpenDate:  tx.OccurredAt,
		}
		if err := store.CreateMany(ctx, []*portfolio.Position{position}); err != nil {
			t.Fatalf("create position: %v", err)
		}
		tx.PositionID = &position.ID
		if err := store.UpdatePositions(ctx, []portfolio.Transaction{*tx}); err != nil {
			t.Fatalf("map transaction to position: %v", err)
		}

		result, err := store.FetchForAccount(ctx, portfolio.TransactionListQuery{AccountID: accountID, Q: "tdt"})
		if err != nil {
			t.Fatalf("fetch transactions: %v", err)
		}
		if result.Total != 1 || len(result.Transactions) != 1 {
			t.Fatalf("want the row found by its listing symbol, got total %d", result.Total)
		}
		got := result.Transactions[0]
		if got.Symbol != nil {
			t.Fatalf("want the stored row left without a symbol, got %q", *got.Symbol)
		}
		if got.ListingSymbol == nil || *got.ListingSymbol != "TDT.AS" {
			t.Fatalf("want listing symbol TDT.AS, got %v", got.ListingSymbol)
		}
	})
}

// TestManualPortfolioTransactionUpdateAndDelete covers the edit and delete a manual
// transaction allows, against the real schema.
func TestManualPortfolioTransactionUpdateAndDelete(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXPortfolioStore(db)
		accountID := seedPortfolioAccount(t, db)

		tx, err := portfolio.NewManualTransaction(portfolio.TransactionData{
			Source:     "DEGIRO",
			OccurredAt: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
			Type:       portfolio.TxCash,
			Amount:     500,
		}, accountID)
		if err != nil {
			t.Fatalf("new transaction: %v", err)
		}
		if err := store.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("create transaction: %v", err)
		}
		position := &portfolio.Position{ID: uuid.New(), AccountID: accountID, OpenDate: tx.OccurredAt}
		if err := store.CreateMany(ctx, []*portfolio.Position{position}); err != nil {
			t.Fatalf("create position: %v", err)
		}
		tx.PositionID = &position.ID
		if err := store.UpdatePositions(ctx, []portfolio.Transaction{*tx}); err != nil {
			t.Fatalf("map transaction to position: %v", err)
		}

		edited, err := portfolio.NewManualTransaction(portfolio.TransactionData{
			Source:      "DEGIRO",
			OccurredAt:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Description: "Deposit",
			Type:        portfolio.TxCash,
			Amount:      750,
		}, accountID)
		if err != nil {
			t.Fatalf("new edited transaction: %v", err)
		}
		edited.ID = tx.ID

		updated, err := store.UpdateManualTransaction(ctx, edited)
		if err != nil || updated != 1 {
			t.Fatalf("update manual transaction: updated %d, err %v", updated, err)
		}
		stored, err := store.Transaction(ctx, accountID, tx.ID)
		if err != nil || stored == nil {
			t.Fatalf("read edited transaction: %v", err)
		}
		if stored.AmountCents != edited.AmountCents || stored.Description != "Deposit" ||
			!stored.OccurredAt.Equal(edited.OccurredAt) || stored.Checksum != edited.Checksum {
			t.Fatalf("want the edit stored, got amount %d, description %q, date %s", stored.AmountCents, stored.Description, stored.OccurredAt)
		}
		if stored.PositionID != nil {
			t.Fatal("want the position link cleared for the next rebuild to set")
		}

		deleted, err := store.DeleteManualTransaction(ctx, accountID, tx.ID)
		if err != nil || deleted != 1 {
			t.Fatalf("delete manual transaction: deleted %d, err %v", deleted, err)
		}
		gone, err := store.Transaction(ctx, accountID, tx.ID)
		if err != nil {
			t.Fatalf("read deleted transaction: %v", err)
		}
		if gone != nil {
			t.Fatal("want the transaction removed")
		}
	})
}

// TestManualPortfolioTransactionWritesLeaveImportedRowsAlone covers the store's own guard:
// an imported row is neither edited nor deleted, even when asked for by id.
func TestManualPortfolioTransactionWritesLeaveImportedRowsAlone(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXPortfolioStore(db)
		accountID := seedPortfolioAccount(t, db)

		vendorID, importID := uuid.New(), uuid.New()
		exec := db.GetExecutor(ctx)
		if _, err := exec.ExecContext(ctx,
			db.Rebind(`INSERT INTO vendors (id, name, type) VALUES (?, ?, ?)`),
			vendorID, "DEGIRO", "brokerage",
		); err != nil {
			t.Fatalf("seed vendor: %v", err)
		}
		if _, err := exec.ExecContext(ctx,
			db.Rebind(`INSERT INTO imports (id, vendor_id, type, path) VALUES (?, ?, ?, ?)`),
			importID, vendorID, "portfolio", "statement.csv",
		); err != nil {
			t.Fatalf("seed import: %v", err)
		}

		isin := "NL0009272749"
		imported, err := portfolio.NewTransaction(portfolio.TransactionData{
			Source:     "DEGIRO",
			OccurredAt: time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC),
			ISIN:       &isin,
			Type:       portfolio.TxBuy,
			Quantity:   3,
			Price:      99.71,
			Amount:     299.13,
		}, 7, importID, &accountID, nil)
		if err != nil {
			t.Fatalf("new imported transaction: %v", err)
		}
		if err := store.CreateTransaction(ctx, imported); err != nil {
			t.Fatalf("create imported transaction: %v", err)
		}

		edit := *imported
		edit.Description = "changed"
		edit.Checksum = "edited-checksum"
		if updated, err := store.UpdateManualTransaction(ctx, &edit); err != nil || updated != 0 {
			t.Fatalf("want an imported row left unedited, got %d rows, err %v", updated, err)
		}
		if deleted, err := store.DeleteManualTransaction(ctx, accountID, imported.ID); err != nil || deleted != 0 {
			t.Fatalf("want an imported row left undeleted, got %d rows, err %v", deleted, err)
		}
		stored, err := store.Transaction(ctx, accountID, imported.ID)
		if err != nil || stored == nil {
			t.Fatalf("read imported transaction: %v", err)
		}
		if stored.Checksum != imported.Checksum || stored.Description != imported.Description {
			t.Fatal("want the imported row exactly as stored")
		}
	})
}

// countRows counts an account's rows in one of the portfolio tables.
func countRows(t *testing.T, db *storage.DB, table string, accountID uuid.UUID) int {
	t.Helper()
	var n int
	if err := db.GetContext(context.Background(), &n, db.Rebind("SELECT COUNT(1) FROM "+table+" WHERE account_id = ?"), accountID); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// TestPortfolioRebuildWritesTheWholeHistory runs a real rebuild over enough days that the
// snapshot inserts span several batches, twice, to cover the transaction it runs in and
// the clean that starts each one.
func TestPortfolioRebuildWritesTheWholeHistory(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXPortfolioStore(db)
		mdStore := storage.NewSQLXMarketDataStore(db)
		accountID := seedPortfolioAccount(t, db)

		isin := "NL0009272749"
		listing := &marketdata.Listing{ID: uuid.New(), Symbol: "TDT.AS", Name: "VanEck AEX", Source: marketdata.SourceMarketStack, ISIN: &isin}
		if err := mdStore.Create(ctx, listing); err != nil {
			t.Fatalf("create listing: %v", err)
		}
		const days = 700
		today := date.StartOfDayUTC(time.Now())
		opened := today.AddDate(0, 0, -days)
		eods := make([]*marketdata.EOD, 0, days)
		for d := opened; d.Before(today); d = d.AddDate(0, 0, 1) {
			eods = append(eods, &marketdata.EOD{ID: uuid.New(), ListingID: listing.ID, Date: d, Open: money.Price(80_000_000), Close: money.Price(81_000_000)})
		}
		if _, err := mdStore.CreateEODs(ctx, eods); err != nil {
			t.Fatalf("create eods: %v", err)
		}
		tx, err := portfolio.NewManualTransaction(portfolio.TransactionData{
			Source: "DEGIRO", OccurredAt: opened, ISIN: &isin, Type: portfolio.TxBuy, Quantity: 10, Price: 80, Amount: 800,
		}, accountID)
		if err != nil {
			t.Fatalf("new transaction: %v", err)
		}
		if err := store.CreateTransaction(ctx, tx); err != nil {
			t.Fatalf("create transaction: %v", err)
		}

		builder := portfolio.NewBuilder(marketdata.NewQueries(mdStore, nil), store, store, store, store, nil)
		for run := 1; run <= 2; run++ {
			if err := builder.Build(ctx, accountID); err != nil {
				t.Fatalf("rebuild %d: %v", run, err)
			}
			if got := countRows(t, db, "position_snapshots", accountID); got != days {
				t.Fatalf("rebuild %d: want %d position snapshots, got %d", run, days, got)
			}
			if got := countRows(t, db, "portfolio_snapshots", accountID); got != days {
				t.Fatalf("rebuild %d: want %d portfolio snapshots, got %d", run, days, got)
			}
			if got := countRows(t, db, "positions", accountID); got != 1 {
				t.Fatalf("rebuild %d: want 1 position, got %d", run, got)
			}
		}
		acquired, err := store.TryAcquireBuildLock(ctx, accountID)
		if err != nil || !acquired {
			t.Fatalf("want the build lock released after a rebuild, acquired %v, err %v", acquired, err)
		}
	})
}
