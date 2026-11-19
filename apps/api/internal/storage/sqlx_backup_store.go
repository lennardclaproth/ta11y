package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

// SQLXBackupStore reads an account's source records for a backup and merges a backup
// back in. It spans the feature schemas on purpose: a restore is one transaction, and
// its inserts write every column — including the ones the features' own inserts leave
// to later updates, such as a transaction's ignore rule.
type SQLXBackupStore struct {
	db *DB

	importsTable            string
	vendorsTable            string
	cashflowTable           string
	ignoreRulesTable        string
	recurringItemsTable     string
	recurringLinksTable     string
	recurringDismissalTable string
	portfolioTable          string
	classesTable            string
	itemsTable              string
	mutationsTable          string
	purchasesTable          string
	listingsTable           string
	eodsTable               string
	providersTable          string
}

var (
	_ backup.AccountReader = (*SQLXBackupStore)(nil)
	_ backup.EODReader     = (*SQLXBackupStore)(nil)
	_ backup.AccountWriter = (*SQLXBackupStore)(nil)
	_ backup.EODWriter     = (*SQLXBackupStore)(nil)
)

// NewSQLXBackupStore creates a backup store backed by SQLX.
func NewSQLXBackupStore(db *DB) *SQLXBackupStore {
	return &SQLXBackupStore{
		db:                      db,
		importsTable:            qualifyTable(db, SchemaImports, TableImports),
		vendorsTable:            qualifyTable(db, SchemaVendors, TableVendors),
		cashflowTable:           qualifyTable(db, SchemaCashflow, TableTransactions),
		ignoreRulesTable:        qualifyTableAs(db, SchemaCashflow, TableIgnoreRules, "cashflow_ignore_rules"),
		recurringItemsTable:     qualifyTableAs(db, SchemaCashflow, TableRecurringItems, "cashflow_recurring_items"),
		recurringLinksTable:     qualifyTableAs(db, SchemaCashflow, TableRecurringLinks, "cashflow_recurring_links"),
		recurringDismissalTable: qualifyTableAs(db, SchemaCashflow, TableRecurringDismissals, "cashflow_recurring_dismissals"),
		portfolioTable:          qualifyTableAs(db, SchemaPortfolio, TableTransactions, "portfolio_transactions"),
		classesTable:            qualifyTableAs(db, SchemaAssets, TableAssetClasses, "asset_classes"),
		itemsTable:              qualifyTableAs(db, SchemaAssets, TableAssetItems, "asset_items"),
		mutationsTable:          qualifyTableAs(db, SchemaAssets, TableAssetMutations, "asset_mutations"),
		purchasesTable:          qualifyTableAs(db, SchemaAssets, TableAssetPurchases, "asset_purchases"),
		listingsTable:           qualifyTable(db, SchemaMarketData, TableListings),
		eodsTable:               qualifyTable(db, SchemaMarketData, TableEOD),
		providersTable:          qualifyTable(db, SchemaMarketData, TableProviders),
	}
}

// Do runs fn in one database transaction.
func (s *SQLXBackupStore) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.db.WithTx(ctx, fn)
}

func (s *SQLXBackupStore) selectAll(ctx context.Context, dest any, what, query string, args ...any) error {
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), dest, s.db.Rebind(query), args...); err != nil {
		return fmt.Errorf("backup store: %s: %w", what, err)
	}
	return nil
}

// --- Reading an account -------------------------------------------------------------

// AccountImports returns the account's imports of one type with their vendor.
func (s *SQLXBackupStore) AccountImports(ctx context.Context, accID uuid.UUID, importType importer.ImportType) ([]backup.ImportRecord, error) {
	var out []backup.ImportRecord
	err := s.selectAll(ctx, &out, "imports", fmt.Sprintf(`
		SELECT i.*, v.name AS vendor_name
		FROM %s i JOIN %s v ON v.id = i.vendor_id
		WHERE i.account_id = ? AND i.type = ?
		ORDER BY i.created_at ASC
	`, s.importsTable, s.vendorsTable), accID, importType)
	return out, err
}

// CashflowTransactions returns every ledger row of the account.
func (s *SQLXBackupStore) CashflowTransactions(ctx context.Context, accID uuid.UUID) ([]*cashflow.Transaction, error) {
	var out []*cashflow.Transaction
	err := s.selectAll(ctx, &out, "cashflow transactions",
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? ORDER BY date ASC, row_number ASC`, s.cashflowTable), accID)
	return out, err
}

// IgnoreRules returns the account's ignore rules, oldest first.
func (s *SQLXBackupStore) IgnoreRules(ctx context.Context, accID uuid.UUID) ([]*cashflow.IgnoreRule, error) {
	var out []*cashflow.IgnoreRule
	err := s.selectAll(ctx, &out, "ignore rules",
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? ORDER BY created_at ASC`, s.ignoreRulesTable), accID)
	return out, err
}

// RecurringItems returns the account's recurring items, oldest first.
func (s *SQLXBackupStore) RecurringItems(ctx context.Context, accID uuid.UUID) ([]*cashflow.RecurringItem, error) {
	var out []*cashflow.RecurringItem
	err := s.selectAll(ctx, &out, "recurring items",
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? ORDER BY created_at ASC`, s.recurringItemsTable), accID)
	return out, err
}

// RecurringLinks returns which ledger rows are linked to which of the account's items.
func (s *SQLXBackupStore) RecurringLinks(ctx context.Context, accID uuid.UUID) ([]backup.LinkRecord, error) {
	var out []backup.LinkRecord
	err := s.selectAll(ctx, &out, "recurring links", fmt.Sprintf(`
		SELECT l.item_id, l.transaction_id
		FROM %s l JOIN %s i ON i.id = l.item_id
		WHERE i.account_id = ?
		ORDER BY l.created_at ASC
	`, s.recurringLinksTable, s.recurringItemsTable), accID)
	return out, err
}

// RecurringDismissals returns the suggestions the account refused.
func (s *SQLXBackupStore) RecurringDismissals(ctx context.Context, accID uuid.UUID) ([]cashflow.DismissedSuggestion, error) {
	var out []cashflow.DismissedSuggestion
	err := s.selectAll(ctx, &out, "recurring dismissals",
		fmt.Sprintf(`SELECT match_key, direction FROM %s WHERE account_id = ? ORDER BY created_at ASC`, s.recurringDismissalTable), accID)
	return out, err
}

// PortfolioTransactions returns every portfolio transaction of the account.
func (s *SQLXBackupStore) PortfolioTransactions(ctx context.Context, accID uuid.UUID) ([]*portfolio.Transaction, error) {
	var out []*portfolio.Transaction
	err := s.selectAll(ctx, &out, "portfolio transactions",
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? ORDER BY occurred_at ASC, row_number ASC`, s.portfolioTable), accID)
	return out, err
}

// ManualAssetClasses returns the account's classes other than the portfolio's.
func (s *SQLXBackupStore) ManualAssetClasses(ctx context.Context, accID uuid.UUID) ([]*assets.Class, error) {
	var out []*assets.Class
	err := s.selectAll(ctx, &out, "asset classes",
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? AND source = ? ORDER BY created_at ASC`, s.classesTable),
		accID, assets.ClassSourceManual)
	return out, err
}

// AssetClasses returns all of the account's classes.
func (s *SQLXBackupStore) AssetClasses(ctx context.Context, accID uuid.UUID) ([]*assets.Class, error) {
	var out []*assets.Class
	err := s.selectAll(ctx, &out, "asset classes",
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? ORDER BY created_at ASC`, s.classesTable), accID)
	return out, err
}

// AssetItems returns the account's items with the listing each is priced from.
func (s *SQLXBackupStore) AssetItems(ctx context.Context, accID uuid.UUID) ([]backup.ItemRecord, error) {
	var out []backup.ItemRecord
	err := s.selectAll(ctx, &out, "asset items", fmt.Sprintf(`
		SELECT i.*, l.symbol AS listing_symbol, l.source AS listing_source
		FROM %s i LEFT JOIN %s l ON l.id = i.listing_id
		WHERE i.account_id = ?
		ORDER BY i.created_at ASC
	`, s.itemsTable, s.listingsTable), accID)
	return out, err
}

// ManualAssetMutations returns the mutations entered by hand: those of items in
// hand-kept classes that are not priced from a listing.
func (s *SQLXBackupStore) ManualAssetMutations(ctx context.Context, accID uuid.UUID) ([]*assets.Mutation, error) {
	var out []*assets.Mutation
	err := s.selectAll(ctx, &out, "asset mutations", fmt.Sprintf(`
		SELECT %s
		FROM %s m
		JOIN %s i ON i.id = m.item_id
		JOIN %s c ON c.id = i.class_id
		WHERE m.account_id = ? AND i.listing_id IS NULL AND c.source = ?
		ORDER BY m.effective_date ASC, m.created_at ASC
	`, assetMutationSelectColumnsM, s.mutationsTable, s.itemsTable, s.classesTable), accID, assets.ClassSourceManual)
	return out, err
}

// AssetPurchases returns the account's recorded purchases.
func (s *SQLXBackupStore) AssetPurchases(ctx context.Context, accID uuid.UUID) ([]*assets.Purchase, error) {
	var out []*assets.Purchase
	err := s.selectAll(ctx, &out, "asset purchases",
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? ORDER BY purchased_on ASC, created_at ASC`, s.purchasesTable), accID)
	return out, err
}

// --- Reading market data ------------------------------------------------------------

// Listings returns every listing, with or without end-of-day prices.
func (s *SQLXBackupStore) Listings(ctx context.Context) ([]*marketdata.Listing, error) {
	var out []*marketdata.Listing
	err := s.selectAll(ctx, &out, "listings",
		fmt.Sprintf(`SELECT * FROM %s ORDER BY source ASC, symbol ASC`, s.listingsTable))
	return out, err
}

// ListingEODs returns a listing's end-of-day prices, oldest first.
func (s *SQLXBackupStore) ListingEODs(ctx context.Context, listingID uuid.UUID) ([]*marketdata.EOD, error) {
	var out []*marketdata.EOD
	err := s.selectAll(ctx, &out, "eods", fmt.Sprintf(`
		SELECT
			id, listing_id, symbol, date,
			open_cents AS open, high_cents AS high, low_cents AS low, close_cents AS close,
			volume, split_factor, created_at, updated_at
		FROM %s
		WHERE listing_id = ?
		ORDER BY date ASC
	`, s.eodsTable), listingID)
	return out, err
}

// --- Resolving references -----------------------------------------------------------

// VendorIDs returns every vendor by name.
func (s *SQLXBackupStore) VendorIDs(ctx context.Context) (map[string]uuid.UUID, error) {
	var rows []struct {
		ID   uuid.UUID `db:"id"`
		Name string    `db:"name"`
	}
	if err := s.selectAll(ctx, &rows, "vendors", fmt.Sprintf(`SELECT id, name FROM %s`, s.vendorsTable)); err != nil {
		return nil, err
	}
	out := make(map[string]uuid.UUID, len(rows))
	for _, r := range rows {
		out[r.Name] = r.ID
	}
	return out, nil
}

// ListingIDs returns every listing by symbol and source.
func (s *SQLXBackupStore) ListingIDs(ctx context.Context) (map[backup.ListingRef]uuid.UUID, error) {
	var rows []struct {
		ID     uuid.UUID         `db:"id"`
		Symbol string            `db:"symbol"`
		Source marketdata.Source `db:"source"`
	}
	if err := s.selectAll(ctx, &rows, "listings", fmt.Sprintf(`SELECT id, symbol, source FROM %s`, s.listingsTable)); err != nil {
		return nil, err
	}
	out := make(map[backup.ListingRef]uuid.UUID, len(rows))
	for _, r := range rows {
		out[backup.ListingRef{Symbol: r.Symbol, Source: r.Source}] = r.ID
	}
	return out, nil
}

// ProviderIDs returns one provider per name, preferring the one GetProviderByName
// would pick: a manual provider, else an API key with quota left.
func (s *SQLXBackupStore) ProviderIDs(ctx context.Context) (map[marketdata.ProviderName]uuid.UUID, error) {
	var rows []struct {
		ID   uuid.UUID               `db:"id"`
		Name marketdata.ProviderName `db:"name"`
	}
	if err := s.selectAll(ctx, &rows, "providers", fmt.Sprintf(`
		SELECT id, name
		FROM %s
		WHERE ingestion_mode = 'MANUAL' OR ingestion_mode = 'API'
		ORDER BY
			CASE WHEN ingestion_mode = 'MANUAL' THEN 0 ELSE 1 END ASC,
			CASE WHEN total > 0 AND remaining <= 0 THEN 1 ELSE 0 END ASC,
			remaining DESC,
			used ASC,
			api_key ASC
	`, s.providersTable)); err != nil {
		return nil, err
	}
	out := make(map[marketdata.ProviderName]uuid.UUID, len(rows))
	for _, r := range rows {
		if _, ok := out[r.Name]; !ok {
			out[r.Name] = r.ID
		}
	}
	return out, nil
}

// CashflowChecksums returns the id of each ledger row holding one of the checksums.
func (s *SQLXBackupStore) CashflowChecksums(ctx context.Context, checksums []string) (map[string]uuid.UUID, error) {
	return s.checksums(ctx, s.cashflowTable, checksums)
}

// PortfolioChecksums returns the id of each portfolio transaction holding one of the checksums.
func (s *SQLXBackupStore) PortfolioChecksums(ctx context.Context, checksums []string) (map[string]uuid.UUID, error) {
	return s.checksums(ctx, s.portfolioTable, checksums)
}

// checksumLookupBatch keeps one IN list well under the parameter limits of both dialects.
const checksumLookupBatch = 500

func (s *SQLXBackupStore) checksums(ctx context.Context, table string, checksums []string) (map[string]uuid.UUID, error) {
	out := make(map[string]uuid.UUID, len(checksums))
	for start := 0; start < len(checksums); start += checksumLookupBatch {
		end := min(start+checksumLookupBatch, len(checksums))
		query, args, err := sqlx.In(fmt.Sprintf(`SELECT id, checksum FROM %s WHERE checksum IN (?)`, table), checksums[start:end])
		if err != nil {
			return nil, fmt.Errorf("backup store: checksums: %w", err)
		}
		var rows []struct {
			ID       uuid.UUID `db:"id"`
			Checksum string    `db:"checksum"`
		}
		if err := s.selectAll(ctx, &rows, "checksums", query, args...); err != nil {
			return nil, err
		}
		for _, r := range rows {
			out[r.Checksum] = r.ID
		}
	}
	return out, nil
}

// --- Writing ------------------------------------------------------------------------

// insertRows runs a named insert over rows, snapshotInsertBatch rows per statement, and
// returns how many rows it added: an insert that skips conflicts adds fewer than it got.
func insertRows[T any](ctx context.Context, s *SQLXBackupStore, what, query string, rows []T) (int, error) {
	added := 0
	for start := 0; start < len(rows); start += snapshotInsertBatch {
		end := min(start+snapshotInsertBatch, len(rows))
		res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, rows[start:end])
		if err != nil {
			return added, fmt.Errorf("backup store: insert %s: %w", what, err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return added, fmt.Errorf("backup store: insert %s: rows affected: %w", what, err)
		}
		added += int(n)
	}
	return added, nil
}

// InsertImports writes recreated import records.
func (s *SQLXBackupStore) InsertImports(ctx context.Context, imports []*importer.Import) error {
	_, err := insertRows(ctx, s, "imports", fmt.Sprintf(`
		INSERT INTO %s (
			id, vendor_id, account_id, listing_id, type, source, path,
			status, status_msg, duplicates, total_rows, imported, failed, auto_ignored,
			created_at, updated_at
		) VALUES (
			:id, :vendor_id, :account_id, :listing_id, :type, :source, :path,
			:status, :status_msg, :duplicates, :total_rows, :imported, :failed, :auto_ignored,
			:created_at, :updated_at
		)
	`, s.importsTable), imports)
	return err
}

// InsertCashflowTransactions writes ledger rows with their ignored state and ignore rule.
func (s *SQLXBackupStore) InsertCashflowTransactions(ctx context.Context, txs []*cashflow.Transaction) error {
	_, err := insertRows(ctx, s, "cashflow transactions", fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, description, note, source, amount_cents,
			direction, date, checksum, created_at, updated_at, tag,
			row_number, ignored, import_id, account_type, ignored_by_rule_id, ignore_overridden
		) VALUES (
			:id, :account_id, :description, :note, :source, :amount_cents,
			:direction, :date, :checksum, :created_at, :updated_at, :tag,
			:row_number, :ignored, :import_id, :account_type, :ignored_by_rule_id, :ignore_overridden
		)
	`, s.cashflowTable), txs)
	return err
}

// InsertIgnoreRules writes ignore rules with what they have done so far.
func (s *SQLXBackupStore) InsertIgnoreRules(ctx context.Context, rules []*cashflow.IgnoreRule) error {
	_, err := insertRows(ctx, s, "ignore rules", fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, name, match_field, contains, direction,
			source, enabled, ignored_total, last_applied_at, created_at, updated_at
		) VALUES (
			:id, :account_id, :name, :match_field, :contains, :direction,
			:source, :enabled, :ignored_total, :last_applied_at, :created_at, :updated_at
		)
	`, s.ignoreRulesTable), rules)
	return err
}

// InsertRecurringItems writes recurring items.
func (s *SQLXBackupStore) InsertRecurringItems(ctx context.Context, items []*cashflow.RecurringItem) error {
	_, err := insertRows(ctx, s, "recurring items", fmt.Sprintf(`
		INSERT INTO %s (id, account_id, name, direction, rhythm, match_key, ended_from, created_at, updated_at)
		VALUES (:id, :account_id, :name, :direction, :rhythm, :match_key, :ended_from, :created_at, :updated_at)
	`, s.recurringItemsTable), items)
	return err
}

type recurringLinkRow struct {
	ID            uuid.UUID `db:"id"`
	ItemID        uuid.UUID `db:"item_id"`
	TransactionID uuid.UUID `db:"transaction_id"`
	CreatedAt     time.Time `db:"created_at"`
}

// InsertRecurringLinks links ledger rows to items, leaving a row that belongs to an
// item already where it is.
func (s *SQLXBackupStore) InsertRecurringLinks(ctx context.Context, links []backup.LinkRecord) (int, error) {
	now := time.Now().UTC()
	rows := make([]recurringLinkRow, 0, len(links))
	for _, l := range links {
		rows = append(rows, recurringLinkRow{ID: uuid.New(), ItemID: l.ItemID, TransactionID: l.TransactionID, CreatedAt: now})
	}
	return insertRows(ctx, s, "recurring links", fmt.Sprintf(`
		INSERT INTO %s (id, item_id, transaction_id, created_at)
		VALUES (:id, :item_id, :transaction_id, :created_at)
		ON CONFLICT (transaction_id) DO NOTHING
	`, s.recurringLinksTable), rows)
}

type recurringDismissalRow struct {
	ID        uuid.UUID                  `db:"id"`
	AccountID uuid.UUID                  `db:"account_id"`
	MatchKey  string                     `db:"match_key"`
	Direction cashflow.CashFlowDirection `db:"direction"`
	CreatedAt time.Time                  `db:"created_at"`
}

// InsertDismissals records refused suggestions the account had not refused yet.
func (s *SQLXBackupStore) InsertDismissals(ctx context.Context, accID uuid.UUID, dismissals []cashflow.DismissedSuggestion) (int, error) {
	now := time.Now().UTC()
	rows := make([]recurringDismissalRow, 0, len(dismissals))
	for _, d := range dismissals {
		rows = append(rows, recurringDismissalRow{ID: uuid.New(), AccountID: accID, MatchKey: d.MatchKey, Direction: d.Direction, CreatedAt: now})
	}
	return insertRows(ctx, s, "recurring dismissals", fmt.Sprintf(`
		INSERT INTO %s (id, account_id, match_key, direction, created_at)
		VALUES (:id, :account_id, :match_key, :direction, :created_at)
		ON CONFLICT (account_id, match_key, direction) DO NOTHING
	`, s.recurringDismissalTable), rows)
}

// InsertPortfolioTransactions writes portfolio transactions.
func (s *SQLXBackupStore) InsertPortfolioTransactions(ctx context.Context, txs []*portfolio.Transaction) error {
	_, err := insertRows(ctx, s, "portfolio transactions",
		fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, s.portfolioTable, portfolioTxColumns, portfolioTxValues), txs)
	return err
}

// InsertAssetClasses writes asset classes.
func (s *SQLXBackupStore) InsertAssetClasses(ctx context.Context, classes []*assets.Class) error {
	_, err := insertRows(ctx, s, "asset classes", fmt.Sprintf(`
		INSERT INTO %s (id, account_id, name, source, archived, created_at, updated_at)
		VALUES (:id, :account_id, :name, :source, :archived, :created_at, :updated_at)
	`, s.classesTable), classes)
	return err
}

// InsertAssetItems writes asset items.
func (s *SQLXBackupStore) InsertAssetItems(ctx context.Context, items []*assets.Asset) error {
	_, err := insertRows(ctx, s, "asset items", fmt.Sprintf(`
		INSERT INTO %s (id, class_id, account_id, name, current_worth, archived, listing_id, created_at, updated_at)
		VALUES (:id, :class_id, :account_id, :name, :current_worth, :archived, :listing_id, :created_at, :updated_at)
	`, s.itemsTable), items)
	return err
}

// InsertAssetMutations writes worth mutations with their running totals as they were.
func (s *SQLXBackupStore) InsertAssetMutations(ctx context.Context, mutations []*assets.Mutation) error {
	_, err := insertRows(ctx, s, "asset mutations", fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, class_id, item_id, change_type, direction, amount,
			previous_worth, new_worth, class_total_worth, effective_date, note, created_at
		) VALUES (
			:id, :account_id, :class_id, :asset_id, :change_type, :direction, :amount,
			:previous_worth, :new_worth, :class_total_worth, :effective_date, :note, :created_at
		)
	`, s.mutationsTable), mutations)
	return err
}

// InsertAssetPurchases writes recorded purchases.
func (s *SQLXBackupStore) InsertAssetPurchases(ctx context.Context, purchases []*assets.Purchase) error {
	_, err := insertRows(ctx, s, "asset purchases", fmt.Sprintf(`
		INSERT INTO %s (id, account_id, item_id, purchased_on, quantity, unit_price, created_at)
		VALUES (:id, :account_id, :item_id, :purchased_on, :quantity, :unit_price, :created_at)
	`, s.purchasesTable), purchases)
	return err
}

// InsertListings writes new listings with their instrument data and provider. The
// accumulated range stays empty: nothing has been synced from the provider yet.
func (s *SQLXBackupStore) InsertListings(ctx context.Context, listings []*marketdata.Listing) error {
	query := s.db.Rebind(fmt.Sprintf(`
		INSERT INTO %s (
			id, symbol, name, source, type, ticker, isin, description, exchange, region,
			currency, provider, active, should_accumulate, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, s.listingsTable))
	for _, l := range listings {
		if _, err := s.db.GetExecutor(ctx).ExecContext(ctx, query,
			l.ID, l.Symbol, l.Name, l.Source, l.Type, l.Ticker, l.ISIN, l.Description, l.Exchange, l.Region,
			l.Currency, l.ProviderID, l.Active, l.ShouldAccumulate, l.CreatedAt, l.UpdatedAt,
		); err != nil {
			return fmt.Errorf("backup store: insert listing %s: %w", l.Symbol, err)
		}
	}
	return nil
}

// InsertEODs writes end-of-day prices with their split factor, skipping a day the
// listing has, and returns how many days it added.
func (s *SQLXBackupStore) InsertEODs(ctx context.Context, eods []*marketdata.EOD) (int, error) {
	return insertRows(ctx, s, "eods", fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (listing_id, date) DO NOTHING
	`, s.eodsTable, eodInsertColumns, eodInsertValues), eods)
}
