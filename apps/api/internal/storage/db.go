package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"go.elastic.co/apm/module/apmsql/v2"
	_ "go.elastic.co/apm/module/apmsql/v2/pq"
	_ "modernc.org/sqlite"
)

type ConnectionType string

const (
	Sqlite   ConnectionType = "sqlite3"
	Postgres ConnectionType = "postgres"
)

const (
	SchemaVendors    = "vendor"
	SchemaCashflow   = "cashflow"
	SchemaPortfolio  = "portfolio"
	SchemaAccount    = "account"
	SchemaImports    = "import"
	SchemaMarketData = "marketdata"
	SchemaAssets     = "assets"
)

const (
	TableVendors         = "vendors"
	TableTransactions    = "transactions"
	TableImports         = "imports"
	TableListings        = "listings"
	TableEOD             = "eods"
	TableEODUploads      = "eod_uploads"
	TableProviders       = "providers"
	TableProviderListing = "provider_listings"
	TableCatalogueSyncs  = "catalogue_syncs"
	TableAccounts        = "accounts"
	TableIdentities      = "identities"
	TableSessions        = "sessions"
	TablePositions       = "positions"
	TablePosSnapshots    = "position_snapshots"
	TablePortSnapshots   = "portfolio_snapshots"
	TableAssetClasses    = "classes"
	TableAssetItems      = "items"
	TableAssetMutations  = "mutations"
	TableAssetPurchases  = "purchases"
	TableAssetSnapshot   = "snapshots"

	TableRecurringItems      = "recurring_items"
	TableRecurringLinks      = "recurring_links"
	TableRecurringDismissals = "recurring_dismissals"
)

type DB struct {
	*sqlx.DB
}

type txContextKey struct{}

var transactionContextKey = txContextKey{}

func NewDB(connStr string, connType ConnectionType) *DB {
	var (
		db  *sql.DB
		err error
	)
	if connType == Sqlite {
		db, err = sql.Open("sqlite", sqliteDSN(connStr))
	} else {
		db, err = apmsql.Open(string(connType), connStr)
	}
	if err != nil {
		panic(fmt.Errorf("db: failed to open connection to database: %w", err))
	}

	sqlxDB := sqlx.NewDb(db, string(connType))

	return &DB{DB: sqlxDB}
}

// sqliteBusyTimeout is how long a SQLite connection waits for a lock before it
// gives up. SQLite locks the whole database for a write, and the event handlers
// write on their own pooled connections while a request reads on another; with
// the default timeout of zero the loser fails instantly with SQLITE_BUSY instead
// of waiting out a write that takes milliseconds. Postgres needs no equivalent.
const sqliteBusyTimeout = 5 * time.Second

// sqliteDSN adds the busy timeout to a SQLite connection string, preserving any
// parameters the configured string already carries.
func sqliteDSN(connStr string) string {
	separator := "?"
	if strings.Contains(connStr, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%s_pragma=busy_timeout(%d)", connStr, separator, sqliteBusyTimeout.Milliseconds())
}

func qualifyTable(db *DB, schema, table string) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return table
	}
	return fmt.Sprintf("%s.%s", schema, table)
}

// qualifyTableAs resolves a table name across dialects when the SQLite name is not
// the bare Postgres table name. On Postgres it returns the schema-qualified name; on
// SQLite (which has no schemas) it returns the flattened, prefixed name the
// migrations use (e.g. portfolio.accounts -> portfolio_accounts).
func qualifyTableAs(db *DB, schema, pgTable, sqliteTable string) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return sqliteTable
	}
	return fmt.Sprintf("%s.%s", schema, pgTable)
}

func (db *DB) GetExecutor(ctx context.Context) sqlx.ExtContext {
	tx, ok := ctx.Value(transactionContextKey).(*sqlx.Tx)
	if ok {
		return tx
	}
	return db
}

// WithTx executes fn in a database transaction and commits when fn returns nil.
func (db *DB) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}

	ctxTx := context.WithValue(ctx, transactionContextKey, tx)
	if err := fn(ctxTx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("db: rollback tx: %w (original error: %v)", rollbackErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: commit tx: %w", err)
	}
	return nil
}
