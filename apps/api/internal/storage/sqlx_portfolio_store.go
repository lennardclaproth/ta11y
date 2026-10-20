package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
	"github.com/lennardclaproth/ta11y/internal/sorting"
)

// SQLXPortfolioStore persists and reads the portfolio aggregate: account projection,
// transactions, positions, position snapshots, and portfolio snapshots. It satisfies
// the portfolio command, query, builder, and lock contracts.
type SQLXPortfolioStore struct {
	db                 *DB
	accountsTable      string
	transactionsTable  string
	positionsTable     string
	posSnapshotsTable  string
	portSnapshotsTable string
	listingsTable      string
}

var (
	_ portfolio.CommandStore      = (*SQLXPortfolioStore)(nil)
	_ portfolio.QueryStore        = (*SQLXPortfolioStore)(nil)
	_ portfolio.PositionStore     = (*SQLXPortfolioStore)(nil)
	_ portfolio.PortfolioStore    = (*SQLXPortfolioStore)(nil)
	_ portfolio.TransactionStore  = (*SQLXPortfolioStore)(nil)
	_ portfolio.TransactionReader = (*SQLXPortfolioStore)(nil)
	_ portfolio.Locker            = (*SQLXPortfolioStore)(nil)
)

// NewSQLXPortfolioStore creates a portfolio store backed by SQLX.
func NewSQLXPortfolioStore(db *DB) *SQLXPortfolioStore {
	return &SQLXPortfolioStore{
		db:                 db,
		accountsTable:      qualifyTableAs(db, SchemaPortfolio, TableAccounts, "portfolio_accounts"),
		transactionsTable:  qualifyTableAs(db, SchemaPortfolio, TableTransactions, "portfolio_transactions"),
		positionsTable:     qualifyTable(db, SchemaPortfolio, TablePositions),
		posSnapshotsTable:  qualifyTable(db, SchemaPortfolio, TablePosSnapshots),
		portSnapshotsTable: qualifyTable(db, SchemaPortfolio, TablePortSnapshots),
		listingsTable:      qualifyTable(db, SchemaMarketData, TableListings),
	}
}

// --- Account projection -----------------------------------------------------

// CreateAccount inserts the portfolio account projection, ignoring an existing row.
func (s *SQLXPortfolioStore) CreateAccount(ctx context.Context, acc *portfolio.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, building, created_at, updated_at)
		VALUES (:id, :account_id, :building, :created_at, :updated_at)
		ON CONFLICT (account_id) DO NOTHING
	`, s.accountsTable)
	_, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, acc)
	return err
}

// TryAcquireBuildLock atomically claims the account's build lock, returning true on
// success or portfolio.ErrAccountNotFound when the projection does not exist.
func (s *SQLXPortfolioStore) TryAcquireBuildLock(ctx context.Context, id uuid.UUID) (bool, error) {
	executor := s.db.GetExecutor(ctx)
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s SET building = ?, updated_at = ?
		WHERE account_id = ? AND building = ?
	`, s.accountsTable))
	res, err := executor.ExecContext(ctx, query, true, time.Now().UTC(), id, false)
	if err != nil {
		return false, fmt.Errorf("portfolio store: acquire build lock: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("portfolio store: acquire build lock rows: %w", err)
	}
	if n == 1 {
		return true, nil
	}

	var exists int
	checkQuery := s.db.Rebind(fmt.Sprintf("SELECT 1 FROM %s WHERE account_id = ?", s.accountsTable))
	if err := sqlx.GetContext(ctx, executor, &exists, checkQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, portfolio.ErrAccountNotFound
		}
		return false, fmt.Errorf("portfolio store: check account: %w", err)
	}
	return false, nil
}

// ReleaseBuildLock clears the account's build lock.
func (s *SQLXPortfolioStore) ReleaseBuildLock(ctx context.Context, id uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET building = ?, updated_at = ? WHERE account_id = ?`, s.accountsTable))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, false, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("portfolio store: release build lock: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("portfolio store: release build lock rows: %w", err)
	}
	if n == 0 {
		return portfolio.ErrAccountNotFound
	}
	return nil
}

// --- Transactions -----------------------------------------------------------

const portfolioTxColumns = `
	id, account_id, import_id, origin, source, occurred_at, position_id, isin, symbol,
	description, type, quantity, unit_price, amount_cents, checksum, row_number,
	created_at, updated_at
`

const portfolioTxValues = `
	:id, :account_id, :import_id, :origin, :source, :occurred_at, :position_id, :isin, :symbol,
	:description, :type, :quantity, :unit_price, :amount_cents, :checksum, :row_number,
	:created_at, :updated_at
`

// CreateTransaction persists one portfolio transaction, mapping a duplicate checksum
// to portfolio.ErrDuplicateTransaction.
func (s *SQLXPortfolioStore) CreateTransaction(ctx context.Context, tx *portfolio.Transaction) error {
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, s.transactionsTable, portfolioTxColumns, portfolioTxValues)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, tx); err != nil {
		if isUniqueViolation(err) {
			return portfolio.ErrDuplicateTransaction
		}
		return fmt.Errorf("portfolio store: create transaction: %w", err)
	}
	return nil
}

// CreateTransactions persists a batch of portfolio transactions in a single insert,
// skipping rows whose checksum already exists, and returns the number inserted.
func (s *SQLXPortfolioStore) CreateTransactions(ctx context.Context, txs []*portfolio.Transaction) (int, error) {
	if len(txs) == 0 {
		return 0, nil
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (checksum) DO NOTHING
	`, s.transactionsTable, portfolioTxColumns, portfolioTxValues)
	res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, txs)
	if err != nil {
		return 0, fmt.Errorf("portfolio store: bulk insert transactions: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("portfolio store: rows affected: %w", err)
	}
	return int(affected), nil
}

// Transaction returns one portfolio transaction within an account, or nil when the
// account does not hold a transaction with that id.
func (s *SQLXPortfolioStore) Transaction(ctx context.Context, accountID, id uuid.UUID) (*portfolio.Transaction, error) {
	query := s.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE account_id = ? AND id = ?", s.transactionsTable))
	return s.transaction(ctx, query, accountID, id)
}

// TransactionByChecksum returns the transaction holding the given dedup checksum within
// an account, or nil when no row holds it.
func (s *SQLXPortfolioStore) TransactionByChecksum(ctx context.Context, accountID uuid.UUID, checksum string) (*portfolio.Transaction, error) {
	query := s.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE account_id = ? AND checksum = ?", s.transactionsTable))
	return s.transaction(ctx, query, accountID, checksum)
}

func (s *SQLXPortfolioStore) transaction(ctx context.Context, query string, args ...any) (*portfolio.Transaction, error) {
	var tx portfolio.Transaction
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &tx, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("portfolio store: fetch transaction: %w", err)
	}
	return &tx, nil
}

// UpdateTransactionOccurredAt moves one transaction to another date, storing the
// checksum recomputed for it. A checksum already held by another row surfaces as the
// feature's duplicate error.
func (s *SQLXPortfolioStore) UpdateTransactionOccurredAt(ctx context.Context, accountID, id uuid.UUID, occurredAt time.Time, checksum string) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`UPDATE %s SET occurred_at = ?, checksum = ?, updated_at = ? WHERE account_id = ? AND id = ?`,
		s.transactionsTable,
	))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, occurredAt, checksum, time.Now().UTC(), accountID, id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, portfolio.ErrDuplicateTransaction
		}
		return 0, fmt.Errorf("portfolio store: update transaction date: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("portfolio store: rows affected: %w", err)
	}
	return int(affected), nil
}

// UpdateManualTransaction stores the edited content of one manual transaction. The
// position link is cleared because the instrument may have changed; the rebuild that
// follows every manual write sets it again. A checksum already held by another row
// surfaces as the feature's duplicate error.
func (s *SQLXPortfolioStore) UpdateManualTransaction(ctx context.Context, tx *portfolio.Transaction) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s SET
			occurred_at = ?, isin = ?, symbol = ?, description = ?, type = ?,
			quantity = ?, unit_price = ?, amount_cents = ?, checksum = ?,
			position_id = NULL, updated_at = ?
		WHERE account_id = ? AND id = ? AND origin = ?
	`, s.transactionsTable))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query,
		tx.OccurredAt, tx.ISIN, tx.Symbol, tx.Description, string(tx.Type),
		tx.Quantity, tx.UnitPrice, tx.AmountCents, tx.Checksum,
		time.Now().UTC(), tx.AccountID, tx.ID, string(portfolio.TransactionOriginManual),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, portfolio.ErrDuplicateTransaction
		}
		return 0, fmt.Errorf("portfolio store: update manual transaction: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("portfolio store: rows affected: %w", err)
	}
	return int(affected), nil
}

// DeleteManualTransaction removes one manual transaction from an account.
func (s *SQLXPortfolioStore) DeleteManualTransaction(ctx context.Context, accountID, id uuid.UUID) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`DELETE FROM %s WHERE account_id = ? AND id = ? AND origin = ?`,
		s.transactionsTable,
	))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, accountID, id, string(portfolio.TransactionOriginManual))
	if err != nil {
		return 0, fmt.Errorf("portfolio store: delete manual transaction: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("portfolio store: rows affected: %w", err)
	}
	return int(affected), nil
}

// TransactionsForAccount returns an account's transactions ordered by occurrence
// (sort "asc"/"desc").
func (s *SQLXPortfolioStore) TransactionsForAccount(ctx context.Context, accID uuid.UUID, sort string) ([]portfolio.Transaction, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE account_id = ?
		ORDER BY occurred_at %s, row_number %s
	`, s.transactionsTable, normalizeSortOrder(sort), normalizeSortOrder(sort)))
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &txs, query, accID); err != nil {
		return nil, fmt.Errorf("portfolio store: transactions for account: %w", err)
	}
	return txs, nil
}

// ImportedProducts returns the distinct instruments one import contributed, with how
// many of its rows mentioned each. Cash rows are excluded: they identify no
// instrument, so they can never be missing a listing.
func (s *SQLXPortfolioStore) ImportedProducts(ctx context.Context, accountID, importID uuid.UUID) ([]portfolio.ImportedProduct, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT MAX(description) AS name, isin, symbol, COUNT(1) AS transactions
		FROM %s
		WHERE account_id = ? AND import_id = ? AND type <> ?
			AND (isin IS NOT NULL OR symbol IS NOT NULL)
		GROUP BY isin, symbol
		ORDER BY name ASC
	`, s.transactionsTable))
	products := make([]portfolio.ImportedProduct, 0)
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &products, query, accountID, importID, portfolio.TxCash); err != nil {
		return nil, fmt.Errorf("portfolio store: imported products: %w", err)
	}
	return products, nil
}

// TransactionsForPosition returns a position's transactions on or after the optional
// from date, ordered by occurrence.
func (s *SQLXPortfolioStore) TransactionsForPosition(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]portfolio.Transaction, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE position_id = ?`, s.transactionsTable)
	args := []any{positionID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	query += " ORDER BY occurred_at ASC, row_number ASC"
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &txs, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("portfolio store: transactions for position: %w", err)
	}
	return txs, nil
}

// UpdatePositions maps each transaction to its resolved position.
func (s *SQLXPortfolioStore) UpdatePositions(ctx context.Context, transactions []portfolio.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET position_id = ?, updated_at = ? WHERE id = ?`, s.transactionsTable))
	executor := s.db.GetExecutor(ctx)
	now := time.Now().UTC()
	for i := range transactions {
		tx := transactions[i]
		if tx.PositionID == nil {
			continue
		}
		if _, err := executor.ExecContext(ctx, query, *tx.PositionID, now, tx.ID); err != nil {
			return fmt.Errorf("portfolio store: update position mapping: %w", err)
		}
	}
	return nil
}

// --- Positions --------------------------------------------------------------

// CreateMany persists a batch of positions.
func (s *SQLXPortfolioStore) CreateMany(ctx context.Context, positions []*portfolio.Position) error {
	if len(positions) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, isin, symbol, listing_id,
			open_date, close_date, quantity, cost_basis,
			fees, income, taxes, realized_pnl,
			created_at, updated_at
		) VALUES (
			:id, :account_id, :isin, :symbol, :listing_id,
			:open_date, :close_date, :quantity, :cost_basis,
			:fees, :income, :taxes, :realized_pnl,
			:created_at, :updated_at
		)
	`, s.positionsTable)
	executor := s.db.GetExecutor(ctx)
	for _, pos := range positions {
		if pos.CreatedAt.IsZero() {
			pos.CreatedAt = time.Now().UTC()
		}
		if pos.UpdatedAt.IsZero() {
			pos.UpdatedAt = time.Now().UTC()
		}
		if _, err := sqlx.NamedExecContext(ctx, executor, query, pos); err != nil {
			return fmt.Errorf("portfolio store: create position: %w", err)
		}
	}
	return nil
}

// GetLastSnapshot returns the most recent position snapshot, or (nil, nil) when none.
func (s *SQLXPortfolioStore) GetLastSnapshot(ctx context.Context, positionID uuid.UUID) (*portfolio.PositionSnapshot, error) {
	var snap portfolio.PositionSnapshot
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s WHERE position_id = ? ORDER BY occurred_at DESC LIMIT 1
	`, s.posSnapshotsTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &snap, query, positionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("portfolio store: get last snapshot: %w", err)
	}
	return &snap, nil
}

// snapshotInsertBatch bounds the rows in one multi-row insert. A rebuild writes a
// snapshot per position per day, thousands of rows, and one statement per row made the
// round trips the cost of a rebuild. 500 rows of the widest snapshot (24 columns) stay
// far below the 65535 parameters Postgres accepts in one statement.
const snapshotInsertBatch = 500

// insertInBatches runs a named insert over rows, snapshotInsertBatch rows per statement.
func insertInBatches[T any](ctx context.Context, exec sqlx.ExtContext, query string, rows []T) error {
	for start := 0; start < len(rows); start += snapshotInsertBatch {
		end := min(start+snapshotInsertBatch, len(rows))
		if _, err := sqlx.NamedExecContext(ctx, exec, query, rows[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// CreatePositionSnapshots persists position snapshots in multi-row inserts, mapping the
// per-day unique violation to portfolio.ErrPositionSnapshotAlreadyExists.
func (s *SQLXPortfolioStore) CreatePositionSnapshots(ctx context.Context, snaps []*portfolio.PositionSnapshot) error {
	now := time.Now().UTC()
	for _, snap := range snaps {
		if snap.CreatedAt.IsZero() {
			snap.CreatedAt = now
		}
		snap.OccurredAt = date.StartOfDayUTC(snap.OccurredAt)
		snap.UpdatedAt = now
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, position_id, symbol, name, listing_id, occurred_at,
			quantity, unit_price, average_price, market_value,
			cost_basis, income, fees, taxes,
			total_pnl, total_pnl_pct, realized_pnl, unrealized_pnl, unrealized_pnl_pct,
			daily_delta_pnl, daily_delta_pnl_pct,
			created_at, updated_at
		) VALUES (
			:id, :account_id, :position_id, :symbol, :name, :listing_id, :occurred_at,
			:quantity, :unit_price, :average_price, :market_value,
			:cost_basis, :income, :fees, :taxes,
			:total_pnl, :total_pnl_pct, :realized_pnl, :unrealized_pnl, :unrealized_pnl_pct,
			:daily_delta_pnl, :daily_delta_pnl_pct,
			:created_at, :updated_at
		)
	`, s.posSnapshotsTable)
	if err := insertInBatches(ctx, s.db.GetExecutor(ctx), query, snaps); err != nil {
		if isUniqueViolation(err) {
			return portfolio.ErrPositionSnapshotAlreadyExists
		}
		return fmt.Errorf("portfolio store: create position snapshots: %w", err)
	}
	return nil
}

// --- Portfolio snapshots ----------------------------------------------------

// CreatePortfolioSnapshots upserts portfolio snapshots keyed by (account, day) in
// multi-row inserts.
func (s *SQLXPortfolioStore) CreatePortfolioSnapshots(ctx context.Context, snapshots []*portfolio.PortfolioSnapshot) error {
	now := time.Now().UTC()
	for _, snapshot := range snapshots {
		if snapshot.CreatedAt.IsZero() {
			snapshot.CreatedAt = now
		}
		snapshot.OccurredAt = date.StartOfDayUTC(snapshot.OccurredAt)
		snapshot.UpdatedAt = now
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, occurred_at,
			market_value, cost_basis,
			realized_pnl, unrealized_pnl, unrealized_pnl_pct, income, fees, taxes, cash_balance,
			total_pnl, total_pnl_pct,
			daily_delta_pnl, daily_delta_pnl_pct, time_weighted_return_pct,
			net_cashflow, cumulative_net_cashflow,
			created_at, updated_at
		) VALUES (
			:id, :account_id, :occurred_at,
			:market_value, :cost_basis,
			:realized_pnl, :unrealized_pnl, :unrealized_pnl_pct, :income, :fees, :taxes, :cash_balance,
			:total_pnl, :total_pnl_pct,
			:daily_delta_pnl, :daily_delta_pnl_pct, :time_weighted_return_pct,
			:net_cashflow, :cumulative_net_cashflow,
			:created_at, :updated_at
		)
		ON CONFLICT (account_id, occurred_at) DO UPDATE SET
			market_value = EXCLUDED.market_value,
			cost_basis = EXCLUDED.cost_basis,
			realized_pnl = EXCLUDED.realized_pnl,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			unrealized_pnl_pct = EXCLUDED.unrealized_pnl_pct,
			income = EXCLUDED.income,
			fees = EXCLUDED.fees,
			taxes = EXCLUDED.taxes,
			cash_balance = EXCLUDED.cash_balance,
			total_pnl = EXCLUDED.total_pnl,
			total_pnl_pct = EXCLUDED.total_pnl_pct,
			daily_delta_pnl = EXCLUDED.daily_delta_pnl,
			daily_delta_pnl_pct = EXCLUDED.daily_delta_pnl_pct,
			time_weighted_return_pct = EXCLUDED.time_weighted_return_pct,
			net_cashflow = EXCLUDED.net_cashflow,
			cumulative_net_cashflow = EXCLUDED.cumulative_net_cashflow,
			updated_at = EXCLUDED.updated_at
	`, s.portSnapshotsTable)
	if err := insertInBatches(ctx, s.db.GetExecutor(ctx), query, snapshots); err != nil {
		return fmt.Errorf("portfolio store: create portfolio snapshots: %w", err)
	}
	return nil
}

// Do runs fn within a single database transaction; stores called with the context fn
// receives take part in it.
func (s *SQLXPortfolioStore) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return s.db.WithTx(ctx, fn)
}

// Clean removes all built portfolio data for an account ahead of a full rebuild:
// position snapshots, portfolio snapshots, and positions.
func (s *SQLXPortfolioStore) Clean(ctx context.Context, accID uuid.UUID) error {
	executor := s.db.GetExecutor(ctx)
	for _, table := range []string{s.posSnapshotsTable, s.portSnapshotsTable, s.positionsTable} {
		query := s.db.Rebind(fmt.Sprintf("DELETE FROM %s WHERE account_id = ?", table))
		if _, err := executor.ExecContext(ctx, query, accID); err != nil {
			return fmt.Errorf("portfolio store: clean %s: %w", table, err)
		}
	}
	return nil
}

// --- Queries ----------------------------------------------------------------

// FetchForAccount returns portfolio transactions for an account with optional
// filters, date sorting, and offset pagination.
func (s *SQLXPortfolioStore) FetchForAccount(
	ctx context.Context,
	query portfolio.TransactionListQuery,
) (*portfolio.TransactionListResult, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 25
	}

	whereClause, args := buildPortfolioTransactionWhereClause(query)
	executor := s.db.GetExecutor(ctx)

	// Imported rows usually carry only an ISIN, so the symbol of the listing their position
	// is linked to stands in for one, both in the result and in the search.
	joins := fmt.Sprintf(`
		LEFT JOIN %s p ON p.id = t.position_id
		LEFT JOIN %s l ON l.id = p.listing_id`, s.positionsTable, s.listingsTable)

	totalQuery := s.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s t%s%s", s.transactionsTable, joins, whereClause))
	total := 0
	if err := sqlx.GetContext(ctx, executor, &total, totalQuery, args...); err != nil {
		return nil, fmt.Errorf("portfolio store: count transactions: %w", err)
	}

	sortOrder := normalizeSortOrder(string(query.SortOrder))
	dataQuery := s.db.Rebind(fmt.Sprintf(`
		SELECT
			t.*,
			p.listing_id,
			l.symbol AS listing_symbol
		FROM %s t%s
		%s
		ORDER BY t.occurred_at %s, t.row_number %s, t.created_at %s, t.id %s
		LIMIT ?
		OFFSET ?
	`, s.transactionsTable, joins, whereClause, sortOrder, sortOrder, sortOrder, sortOrder))
	dataArgs := append(append([]any{}, args...), limit, query.Offset)

	var rows []portfolio.TransactionWithListingID
	if err := sqlx.SelectContext(ctx, executor, &rows, dataQuery, dataArgs...); err != nil {
		return nil, fmt.Errorf("portfolio store: fetch transactions: %w", err)
	}

	return &portfolio.TransactionListResult{
		Total:        total,
		Transactions: rows,
	}, nil
}

// SnapshotsForAccount returns portfolio snapshots for an account within an optional
// date range, ordered by occurrence (default ascending), optionally paginated.
func (s *SQLXPortfolioStore) SnapshotsForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	limit, offset *int,
	from, to *time.Time,
	sort *sorting.Direction,
) ([]*portfolio.PortfolioSnapshot, error) {
	dir := "ASC"
	if sort != nil {
		dir = normalizeSortOrder(string(*sort))
	}
	query := fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ?`, s.portSnapshotsTable)
	args := []any{accountID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND occurred_at <= ?"
		args = append(args, *to)
	}
	query += fmt.Sprintf(" ORDER BY occurred_at %s, id ASC", dir)
	if limit != nil {
		query += " LIMIT ?"
		args = append(args, *limit)
		if offset != nil {
			query += " OFFSET ?"
			args = append(args, *offset)
		}
	}

	var snapshots []*portfolio.PortfolioSnapshot
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &snapshots, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("portfolio store: snapshots for account: %w", err)
	}
	return snapshots, nil
}

// PositionsWithLatestSnapshot returns account positions joined with their latest
// position snapshot, optionally including closed positions.
func (s *SQLXPortfolioStore) PositionsWithLatestSnapshot(
	ctx context.Context,
	accID uuid.UUID,
	includeClosed bool,
) ([]*portfolio.PositionWithLatestSnapshot, error) {
	query := fmt.Sprintf(`
		SELECT
			p.id,
			p.symbol,
			ls.name,
			p.quantity,
			p.cost_basis,
			p.realized_pnl,
			ls.market_value,
			ls.unrealized_pnl_pct,
			ls.occurred_at AS last_snapshot_at,
			p.open_date,
			p.close_date,
			(p.close_date IS NOT NULL) AS is_closed
		FROM %s p
		LEFT JOIN %s ls
			ON ls.id = (
				SELECT ps.id
				FROM %s ps
				WHERE ps.position_id = p.id
				ORDER BY ps.occurred_at DESC, ps.id DESC
				LIMIT 1
			)
		WHERE p.account_id = ?
	`, s.positionsTable, s.posSnapshotsTable, s.posSnapshotsTable)
	args := []any{accID}
	if !includeClosed {
		query += " AND p.close_date IS NULL"
	}
	query += " ORDER BY p.open_date ASC, p.id ASC"

	var rows []*portfolio.PositionWithLatestSnapshot
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("portfolio store: positions with latest snapshot: %w", err)
	}
	return rows, nil
}

func buildPortfolioTransactionWhereClause(query portfolio.TransactionListQuery) (string, []any) {
	conditions := []string{"t.account_id = ?"}
	args := []any{query.AccountID}

	if query.From != nil {
		conditions = append(conditions, "t.occurred_at >= ?")
		args = append(args, *query.From)
	}
	if query.To != nil {
		conditions = append(conditions, "t.occurred_at <= ?")
		args = append(args, *query.To)
	}
	if query.Type != nil {
		conditions = append(conditions, "t.type = ?")
		args = append(args, string(*query.Type))
	}
	if query.Origin != nil {
		conditions = append(conditions, "t.origin = ?")
		args = append(args, string(*query.Origin))
	}

	appendContains := func(column string, value string) {
		v := strings.TrimSpace(value)
		if v == "" {
			return
		}
		conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE ?", column))
		args = append(args, "%"+strings.ToLower(v)+"%")
	}

	appendContains("t.source", query.Source)
	if listing := strings.TrimSpace(query.Listing); listing != "" {
		pattern := "%" + strings.ToLower(listing) + "%"
		conditions = append(conditions, "(LOWER(COALESCE(NULLIF(TRIM(t.symbol), ''), l.symbol, '')) LIKE ? OR LOWER(COALESCE(t.isin, '')) LIKE ?)")
		args = append(args, pattern, pattern)
	}
	if q := strings.TrimSpace(query.Q); q != "" {
		pattern := "%" + strings.ToLower(q) + "%"
		conditions = append(
			conditions,
			"(LOWER(t.description) LIKE ? OR LOWER(t.source) LIKE ? OR LOWER(COALESCE(NULLIF(TRIM(t.symbol), ''), l.symbol, '')) LIKE ? OR LOWER(COALESCE(t.isin, '')) LIKE ?)",
		)
		args = append(args, pattern, pattern, pattern, pattern)
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}
