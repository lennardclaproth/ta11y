//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
	"github.com/lennardclaproth/ta11y/internal/storage"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

// seedLedgerAccount creates an account with the cashflow, portfolio and assets
// projections its records are bound to.
func seedLedgerAccount(t *testing.T, db *storage.DB, email string) uuid.UUID {
	t.Helper()
	ctx := t.Context()
	accID := seedAccount(t, db, email, email)
	if err := storage.NewSQLXCashflowStore(db).CreateAccount(ctx, cashflow.NewAccount(accID)); err != nil {
		t.Fatalf("create cashflow projection: %v", err)
	}
	if err := storage.NewSQLXPortfolioStore(db).CreateAccount(ctx, portfolio.NewAccount(accID)); err != nil {
		t.Fatalf("create portfolio projection: %v", err)
	}
	if err := storage.NewSQLXAssetsStore(db).CreateAccount(ctx, assets.NewAccount(accID)); err != nil {
		t.Fatalf("create assets projection: %v", err)
	}
	return accID
}

type backupFixture struct {
	accID         uuid.UUID
	rent, coffee  *cashflow.Transaction
	listing       uuid.UUID
	portfolioDate time.Time
}

// seedBackupFixture gives an account one record of every kind a backup carries: an
// imported and a manual ledger row (one ignored by a rule, one tagged and linked to a
// recurring item), a dismissal, an imported portfolio buy, and a hand-kept asset class
// with a valued item and a daily-priced one.
func seedBackupFixture(t *testing.T, db *storage.DB) backupFixture {
	t.Helper()
	ctx := t.Context()
	accID := seedLedgerAccount(t, db, "source@example.com")

	ing, err := vendor.NewVendor(vendor.VendorING, vendor.VendorTypeBank)
	if err != nil {
		t.Fatalf("build vendor: %v", err)
	}
	degiro, err := vendor.NewVendor(vendor.VendorDEGIRO, vendor.VendorTypeBrokerage)
	if err != nil {
		t.Fatalf("build vendor: %v", err)
	}
	vendors := storage.NewSQLXVendorStore(db)
	for _, v := range []*vendor.Vendor{ing, degiro} {
		if err := vendors.Create(ctx, v); err != nil {
			t.Fatalf("create vendor: %v", err)
		}
	}
	imports := storage.NewSQLXImporterStore(db)
	cashflowImport := importer.NewImport(*ing, "ing.csv", accID)
	cashflowImport.Status = importer.ImportStatusCompleted
	portfolioImport := importer.NewImport(*degiro, "degiro.csv", accID)
	portfolioImport.Type = importer.ImportTypePortfolio
	portfolioImport.Status = importer.ImportStatusCompleted
	for _, imp := range []*importer.Import{cashflowImport, portfolioImport} {
		if err := imports.Create(ctx, imp); err != nil {
			t.Fatalf("create import: %v", err)
		}
	}

	draft, err := cashflow.NewIgnoreRuleDraft("Card settlement", "description", "SETTLEMENT", "out", "", true)
	if err != nil {
		t.Fatalf("new ignore rule draft: %v", err)
	}
	rule := cashflow.NewIgnoreRule(accID, draft)
	if err := storage.NewSQLXCashflowIgnoreRuleStore(db).CreateIgnoreRule(ctx, rule); err != nil {
		t.Fatalf("create ignore rule: %v", err)
	}

	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	rent, err := cashflow.NewTransaction("Rent", "", "ING", "housing", cashflow.CashOut, 95000, day, 1, 0, &cashflowImport.ID, nil, accID)
	if err != nil {
		t.Fatalf("build transaction: %v", err)
	}
	settlement, err := cashflow.NewTransaction("CARD SETTLEMENT", "", "ING", "", cashflow.CashOut, 20000, day, 2, 0, &cashflowImport.ID, nil, accID)
	if err != nil {
		t.Fatalf("build transaction: %v", err)
	}
	settlement.Ignored = true
	settlement.IgnoredByRuleID = &rule.ID
	coffee, err := cashflow.NewTransaction("Coffee", "", cashflow.SourceManual, "", cashflow.CashOut, 350, day, 0, 0, nil, nil, accID)
	if err != nil {
		t.Fatalf("build transaction: %v", err)
	}
	// The backup store's insert is the one that writes the ignore rule a row carries.
	if err := storage.NewSQLXBackupStore(db).InsertCashflowTransactions(ctx, []*cashflow.Transaction{rent, settlement, coffee}); err != nil {
		t.Fatalf("create transactions: %v", err)
	}

	recurring := storage.NewSQLXRecurringStore(db)
	item, err := cashflow.NewRecurringItem(accID, "Rent", cashflow.CashOut, cashflow.RhythmMonthly, "rent")
	if err != nil {
		t.Fatalf("build recurring item: %v", err)
	}
	if err := recurring.CreateRecurringItem(ctx, item); err != nil {
		t.Fatalf("create recurring item: %v", err)
	}
	if _, err := recurring.LinkTransactions(ctx, item.ID, []uuid.UUID{rent.ID}); err != nil {
		t.Fatalf("link transaction: %v", err)
	}
	if err := recurring.DismissRecurringSuggestion(ctx, accID, "coffee", cashflow.CashOut); err != nil {
		t.Fatalf("dismiss suggestion: %v", err)
	}

	isin := "NL0009272749"
	buy, err := portfolio.NewTransaction(portfolio.TransactionData{
		Source: "DEGIRO", OccurredAt: day, ISIN: &isin, Type: portfolio.TxBuy, Quantity: 3, Price: 80.5,
	}, 4, portfolioImport.ID, &accID, nil)
	if err != nil {
		t.Fatalf("build portfolio transaction: %v", err)
	}
	if _, err := storage.NewSQLXPortfolioStore(db).CreateTransactions(ctx, []*portfolio.Transaction{buy}); err != nil {
		t.Fatalf("create portfolio transaction: %v", err)
	}

	listing, err := marketdata.NewListing("VWRL.AS", "Vanguard FTSE All-World", marketdata.SourceMarketStack)
	if err != nil {
		t.Fatalf("build listing: %v", err)
	}
	if err := storage.NewSQLXMarketDataStore(db).Create(ctx, listing); err != nil {
		t.Fatalf("create listing: %v", err)
	}

	assetStore := storage.NewSQLXAssetsStore(db)
	class, err := assets.NewClass(accID, nil, "Savings")
	if err != nil {
		t.Fatalf("build class: %v", err)
	}
	if err := assetStore.CreateClass(ctx, class); err != nil {
		t.Fatalf("create class: %v", err)
	}
	now := time.Now().UTC()
	deposit := &assets.Asset{ID: uuid.New(), ClassID: class.ID, AccountID: accID, Name: "Deposit", CurrentWorth: 500000, CreatedAt: now, UpdatedAt: now}
	fund := &assets.Asset{ID: uuid.New(), ClassID: class.ID, AccountID: accID, Name: "Fund", ListingID: &listing.ID, CreatedAt: now, UpdatedAt: now}
	for _, a := range []*assets.Asset{deposit, fund} {
		if err := assetStore.CreateAsset(ctx, a); err != nil {
			t.Fatalf("create asset: %v", err)
		}
	}
	if err := assetStore.CreateMutation(ctx, &assets.Mutation{
		ID: uuid.New(), AccountID: accID, ClassID: class.ID, AssetID: deposit.ID, ChangeType: assets.ChangeTypeSet,
		Amount: 500000, NewWorth: 500000, ClassTotalWorth: 500000, EffectiveDate: day, CreatedAt: now,
	}); err != nil {
		t.Fatalf("create mutation: %v", err)
	}
	if err := assetStore.CreatePurchase(ctx, &assets.Purchase{
		ID: uuid.New(), AccountID: accID, AssetID: fund.ID, PurchasedOn: day, Quantity: 2, UnitPrice: money.Price(10000), CreatedAt: now,
	}); err != nil {
		t.Fatalf("create purchase: %v", err)
	}

	return backupFixture{accID: accID, rent: rent, coffee: coffee, listing: listing.ID, portfolioDate: day}
}

// roundTrip writes a file to JSON and reads it back, as a download and upload would.
func roundTrip[T any](t *testing.T, file *T) *T {
	t.Helper()
	raw, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("marshal backup: %v", err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal backup: %v", err)
	}
	return &out
}

// backupHarness wires the backup feature to one database, without a bus.
func backupHarness(db *storage.DB) (*storage.SQLXBackupStore, *backup.Queries, *backup.Commands) {
	store := storage.NewSQLXBackupStore(db)
	return store, backup.NewQueries(store, store), backup.NewCommands(store, store, nil)
}

func TestCashflowBackupRestoresIntoAnotherAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		fx := seedBackupFixture(t, db)
		store, queries, commands := backupHarness(db)

		exported, err := queries.ExportCashflow(ctx, fx.accID)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		file := roundTrip(t, exported)
		target := seedLedgerAccount(t, db, "target@example.com")

		dry, err := commands.ImportCashflow(ctx, target, file, true)
		if err != nil {
			t.Fatalf("dry run: %v", err)
		}
		if dry.Transactions.New != 3 {
			t.Fatalf("dry run counts: %+v", dry)
		}
		if txs, _ := store.CashflowTransactions(ctx, target); len(txs) != 0 {
			t.Fatalf("dry run wrote %d transactions", len(txs))
		}

		result, err := commands.ImportCashflow(ctx, target, file, false)
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		want := backup.CashflowResult{
			Imports: backup.Count{New: 1}, Transactions: backup.Count{New: 3}, IgnoreRules: backup.Count{New: 1},
			RecurringItems: backup.Count{New: 1}, RecurringLinks: backup.Count{New: 1}, Dismissals: backup.Count{New: 1},
		}
		if !reflect.DeepEqual(*result, want) {
			t.Fatalf("import counts:\n got %+v\nwant %+v", *result, want)
		}

		txs, err := store.CashflowTransactions(ctx, target)
		if err != nil {
			t.Fatalf("read restored transactions: %v", err)
		}
		byDescription := map[string]*cashflow.Transaction{}
		for _, tx := range txs {
			byDescription[tx.Description] = tx
		}
		// A restored row has to carry the checksum a fresh import of the same statement
		// into this account would produce, or that import would duplicate it.
		fresh, err := cashflow.NewTransaction("Rent", "", "ING", "housing", cashflow.CashOut, 95000, fx.portfolioDate, 1, 0, nil, nil, target)
		if err != nil {
			t.Fatalf("build transaction: %v", err)
		}
		if got := byDescription["Rent"]; got == nil || got.Checksum != fresh.Checksum || got.Tag != "housing" || got.ImportID == nil {
			t.Fatalf("restored rent row: %+v", got)
		}
		rules, err := store.IgnoreRules(ctx, target)
		if err != nil || len(rules) != 1 {
			t.Fatalf("restored rules: %v, %v", rules, err)
		}
		if got := byDescription["CARD SETTLEMENT"]; got == nil || !got.Ignored || got.IgnoredByRuleID == nil || *got.IgnoredByRuleID != rules[0].ID {
			t.Fatalf("restored settlement row: %+v", got)
		}
		links, err := store.RecurringLinks(ctx, target)
		if err != nil || len(links) != 1 || links[0].TransactionID != byDescription["Rent"].ID {
			t.Fatalf("restored links: %v, %v", links, err)
		}

		again, err := commands.ImportCashflow(ctx, target, file, false)
		if err != nil {
			t.Fatalf("second import: %v", err)
		}
		if again.Transactions.New != 0 || again.IgnoreRules.New != 0 || again.RecurringLinks.New != 0 || again.Imports.New != 0 {
			t.Fatalf("second import added records: %+v", again)
		}
	})
}

func TestPortfolioBackupRestoresIntoAnotherAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		fx := seedBackupFixture(t, db)
		store, queries, commands := backupHarness(db)

		exported, err := queries.ExportPortfolio(ctx, fx.accID)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		target := seedLedgerAccount(t, db, "target@example.com")

		result, err := commands.ImportPortfolio(ctx, target, roundTrip(t, exported), false)
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if result.Imports != (backup.Count{New: 1}) || result.Transactions != (backup.Count{New: 1}) {
			t.Fatalf("import counts: %+v", result)
		}
		txs, err := store.PortfolioTransactions(ctx, target)
		if err != nil || len(txs) != 1 || txs[0].ImportID == nil || txs[0].Origin != portfolio.TransactionOriginImport {
			t.Fatalf("restored portfolio transactions: %v, %v", txs, err)
		}

		again, err := commands.ImportPortfolio(ctx, target, roundTrip(t, exported), false)
		if err != nil || again.Transactions.New != 0 {
			t.Fatalf("second import added transactions: %+v, %v", again, err)
		}
	})
}

func TestAssetsBackupRestoresIntoAnotherAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		fx := seedBackupFixture(t, db)
		store, queries, commands := backupHarness(db)

		exported, err := queries.ExportAssets(ctx, fx.accID)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		target := seedLedgerAccount(t, db, "target@example.com")

		result, err := commands.ImportAssets(ctx, target, roundTrip(t, exported), false)
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		want := backup.AssetsResult{
			Classes: backup.Count{New: 1}, Items: backup.Count{New: 2},
			Mutations: backup.Count{New: 1}, Purchases: backup.Count{New: 1},
		}
		if !reflect.DeepEqual(*result, want) {
			t.Fatalf("import counts:\n got %+v\nwant %+v", *result, want)
		}
		items, err := store.AssetItems(ctx, target)
		if err != nil || len(items) != 2 {
			t.Fatalf("restored items: %v, %v", items, err)
		}
		for _, item := range items {
			if item.Name == "Fund" && (item.ListingID == nil || *item.ListingID != fx.listing) {
				t.Fatalf("restored fund lost its listing: %+v", item)
			}
		}

		again, err := commands.ImportAssets(ctx, target, roundTrip(t, exported), false)
		if err != nil || again.Items != (backup.Count{Existing: 2}) || again.Mutations.New != 0 {
			t.Fatalf("second import added items: %+v, %v", again, err)
		}
	})
}

func TestCashflowBackupRefusesMissingVendors(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		fx := seedBackupFixture(t, db)
		store, queries, commands := backupHarness(db)

		file, err := queries.ExportCashflow(ctx, fx.accID)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		for i := range file.Imports {
			file.Imports[i].Vendor = "Unknown Bank"
		}
		target := seedLedgerAccount(t, db, "target@example.com")

		result, err := commands.ImportCashflow(ctx, target, file, false)
		if !errors.Is(err, backup.ErrMissingReferences) {
			t.Fatalf("expected missing references, got %v", err)
		}
		if len(result.Missing.Vendors) != 1 || result.Missing.Vendors[0] != "Unknown Bank" {
			t.Fatalf("missing vendors: %v", result.Missing.Vendors)
		}
		if txs, _ := store.CashflowTransactions(ctx, target); len(txs) != 0 {
			t.Fatalf("refused import wrote %d transactions", len(txs))
		}
	})
}

func TestAssetsBackupRefusesMissingListings(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		fx := seedBackupFixture(t, db)
		_, queries, commands := backupHarness(db)

		file, err := queries.ExportAssets(ctx, fx.accID)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		for ci := range file.Classes {
			for ii := range file.Classes[ci].Items {
				if l := file.Classes[ci].Items[ii].Listing; l != nil {
					l.Symbol = "MISSING.AS"
				}
			}
		}
		target := seedLedgerAccount(t, db, "target@example.com")

		result, err := commands.ImportAssets(ctx, target, file, false)
		if !errors.Is(err, backup.ErrMissingReferences) || len(result.Missing.Listings) != 1 {
			t.Fatalf("expected the missing listing to be named, got %+v, %v", result, err)
		}
	})
}

func TestBackupRefusesAFileOfAnotherKind(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		fx := seedBackupFixture(t, db)
		_, queries, commands := backupHarness(db)

		exported, err := queries.ExportPortfolio(ctx, fx.accID)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		raw, err := json.Marshal(exported)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var asCashflow backup.CashflowFile
		if err := json.Unmarshal(raw, &asCashflow); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, err := commands.ImportCashflow(ctx, fx.accID, &asCashflow, true); !errors.Is(err, backup.ErrWrongKind) {
			t.Fatalf("expected a portfolio file to be refused as cashflow, got %v", err)
		}
	})
}

func TestEODBackupMergesDaysAndKeepsSplits(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store, queries, commands := backupHarness(db)
		listingID := seedListing(t, db, "ASML.AS")
		day := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		seedEOD(t, db, listingID, "ASML.AS", day, 600, 1)

		exported, err := queries.ExportEODs(ctx)
		if err != nil {
			t.Fatalf("export: %v", err)
		}
		file := roundTrip(t, exported)
		if len(file.Listings) != 1 || len(file.Listings[0].EODs) != 1 {
			t.Fatalf("exported listings: %+v", file.Listings)
		}
		next := file.Listings[0].EODs[0]
		next.Date = backup.Day{Time: day.AddDate(0, 0, 1)}
		next.SplitFactor = 4
		file.Listings[0].EODs = append(file.Listings[0].EODs, next)

		result, err := commands.ImportEODs(ctx, file, false)
		if err != nil {
			t.Fatalf("import: %v", err)
		}
		if result.EODs != (backup.Count{New: 1, Existing: 1}) || result.Listings != (backup.Count{New: 1}) {
			t.Fatalf("import counts: %+v", result)
		}
		eods, err := store.ListingEODs(ctx, listingID)
		if err != nil || len(eods) != 2 || eods[1].SplitFactor != 4 {
			t.Fatalf("eods after import: %v, %v", eods, err)
		}

		file.Listings[0].Symbol = "MISSING.AS"
		result, err = commands.ImportEODs(ctx, file, false)
		if !errors.Is(err, backup.ErrMissingReferences) || len(result.Missing.Listings) != 1 {
			t.Fatalf("expected the missing listing to be named, got %+v, %v", result, err)
		}
	})
}
