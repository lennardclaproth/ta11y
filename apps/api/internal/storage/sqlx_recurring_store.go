package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
)

// SQLXRecurringStore persists and reads recurring cashflow items, the
// transactions linked to them, and the suggestions that were dismissed. It
// satisfies the cashflow recurring command and query contracts.
type SQLXRecurringStore struct {
	db               *DB
	itemsTable       string
	linksTable       string
	dismissalsTable  string
	transactionTable string
}

var (
	_ cashflow.RecurringCommandStore = (*SQLXRecurringStore)(nil)
	_ cashflow.RecurringQueryStore   = (*SQLXRecurringStore)(nil)
	_ cashflow.RecurringUnitOfWork   = (*SQLXRecurringStore)(nil)
)

// NewSQLXRecurringStore creates a recurring-item store backed by SQLX.
func NewSQLXRecurringStore(db *DB) *SQLXRecurringStore {
	return &SQLXRecurringStore{
		db:               db,
		itemsTable:       qualifyTableAs(db, SchemaCashflow, TableRecurringItems, "cashflow_recurring_items"),
		linksTable:       qualifyTableAs(db, SchemaCashflow, TableRecurringLinks, "cashflow_recurring_links"),
		dismissalsTable:  qualifyTableAs(db, SchemaCashflow, TableRecurringDismissals, "cashflow_recurring_dismissals"),
		transactionTable: qualifyTable(db, SchemaCashflow, TableTransactions),
	}
}

// Do runs fn within a single database transaction.
func (s *SQLXRecurringStore) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	return s.db.WithTx(ctx, fn)
}

// CreateRecurringItem inserts a recurring item, mapping a name already in use to
// the feature's "already exists" error.
func (s *SQLXRecurringStore) CreateRecurringItem(ctx context.Context, item *cashflow.RecurringItem) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, name, direction, rhythm, match_key, ended_from, created_at, updated_at)
		VALUES (:id, :account_id, :name, :direction, :rhythm, :match_key, :ended_from, :created_at, :updated_at)
	`, s.itemsTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, item); err != nil {
		if isUniqueViolation(err) {
			return cashflow.ErrRecurringItemExists
		}
		return fmt.Errorf("recurring store: insert item: %w", err)
	}
	return nil
}

// UpdateRecurringItem stores a changed name and rhythm within the owning account.
func (s *SQLXRecurringStore) UpdateRecurringItem(ctx context.Context, item *cashflow.RecurringItem) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`UPDATE %s SET name = ?, rhythm = ?, updated_at = ? WHERE account_id = ? AND id = ?`,
		s.itemsTable,
	))
	updated, err := s.exec(ctx, query, item.Name, item.Rhythm, item.UpdatedAt, item.AccountID, item.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, cashflow.ErrRecurringItemExists
		}
		return 0, err
	}
	return updated, nil
}

// EndRecurringItem records the month an item stops being expected from.
func (s *SQLXRecurringStore) EndRecurringItem(ctx context.Context, accountID, id uuid.UUID, from time.Time) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`UPDATE %s SET ended_from = ?, updated_at = ? WHERE account_id = ? AND id = ?`,
		s.itemsTable,
	))
	return s.exec(ctx, query, from, time.Now().UTC(), accountID, id)
}

// LinkTransactions attaches transactions to an item, leaving rows that already
// belong to one where they are: a transaction has at most one item, and the
// database holds that rule.
func (s *SQLXRecurringStore) LinkTransactions(ctx context.Context, itemID uuid.UUID, transactionIDs []uuid.UUID) (int, error) {
	if len(transactionIDs) == 0 {
		return 0, nil
	}

	type link struct {
		ID            uuid.UUID `db:"id"`
		ItemID        uuid.UUID `db:"item_id"`
		TransactionID uuid.UUID `db:"transaction_id"`
		CreatedAt     time.Time `db:"created_at"`
	}
	now := time.Now().UTC()
	links := make([]link, 0, len(transactionIDs))
	for _, transactionID := range transactionIDs {
		links = append(links, link{ID: uuid.New(), ItemID: itemID, TransactionID: transactionID, CreatedAt: now})
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, item_id, transaction_id, created_at)
		VALUES (:id, :item_id, :transaction_id, :created_at)
		ON CONFLICT (transaction_id) DO NOTHING
	`, s.linksTable)
	res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, links)
	if err != nil {
		return 0, fmt.Errorf("recurring store: link transactions: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("recurring store: rows affected: %w", err)
	}
	return int(affected), nil
}

// UnlinkTransaction detaches one transaction from an item. The account predicate
// runs through the item, so another account's link cannot be removed by guessing ids.
func (s *SQLXRecurringStore) UnlinkTransaction(ctx context.Context, accountID, itemID, transactionID uuid.UUID) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		DELETE FROM %s
		WHERE item_id = ? AND transaction_id = ?
		  AND item_id IN (SELECT id FROM %s WHERE account_id = ?)
	`, s.linksTable, s.itemsTable))
	return s.exec(ctx, query, itemID, transactionID, accountID)
}

// DismissRecurringSuggestion records a refused pattern. Refusing one twice is not
// an error, so an existing row is left alone.
func (s *SQLXRecurringStore) DismissRecurringSuggestion(ctx context.Context, accountID uuid.UUID, matchKey string, direction cashflow.CashFlowDirection) error {
	query := s.db.Rebind(fmt.Sprintf(`
		INSERT INTO %s (id, account_id, match_key, direction, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (account_id, match_key, direction) DO NOTHING
	`, s.dismissalsTable))
	if _, err := s.exec(ctx, query, uuid.New(), accountID, matchKey, direction, time.Now().UTC()); err != nil {
		return err
	}
	return nil
}

// ListRecurringItems returns every recurring item of an account, newest first.
func (s *SQLXRecurringStore) ListRecurringItems(ctx context.Context, accountID uuid.UUID) ([]*cashflow.RecurringItem, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`SELECT * FROM %s WHERE account_id = ? ORDER BY created_at DESC`,
		s.itemsTable,
	))
	items := []*cashflow.RecurringItem{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &items, query, accountID); err != nil {
		return nil, fmt.Errorf("recurring store: list items: %w", err)
	}
	return items, nil
}

// GetRecurringItem returns one item within an account, or nil when it holds none with that id.
func (s *SQLXRecurringStore) GetRecurringItem(ctx context.Context, accountID, id uuid.UUID) (*cashflow.RecurringItem, error) {
	query := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? AND id = ?`, s.itemsTable))
	var item cashflow.RecurringItem
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &item, query, accountID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("recurring store: fetch item: %w", err)
	}
	return &item, nil
}

// ListRecurringLinks returns every link of an account with the transaction fields
// the item screens read, oldest first.
func (s *SQLXRecurringStore) ListRecurringLinks(ctx context.Context, accountID uuid.UUID) ([]cashflow.RecurringLink, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT l.item_id, l.transaction_id, t.date, t.amount_cents, t.description, t.source
		FROM %s l
		JOIN %s i ON i.id = l.item_id
		JOIN %s t ON t.id = l.transaction_id
		WHERE i.account_id = ?
		ORDER BY t.date ASC
	`, s.linksTable, s.itemsTable, s.transactionTable))
	links := []cashflow.RecurringLink{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &links, query, accountID); err != nil {
		return nil, fmt.Errorf("recurring store: list links: %w", err)
	}
	return links, nil
}

// ListRecurringLinksForItem returns the links of one item, oldest first. The
// account predicate runs through the item, so another account's links stay out.
func (s *SQLXRecurringStore) ListRecurringLinksForItem(ctx context.Context, accountID, itemID uuid.UUID) ([]cashflow.RecurringLink, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT l.item_id, l.transaction_id, t.date, t.amount_cents, t.description, t.source
		FROM %s l
		JOIN %s i ON i.id = l.item_id
		JOIN %s t ON t.id = l.transaction_id
		WHERE i.account_id = ? AND l.item_id = ?
		ORDER BY t.date ASC
	`, s.linksTable, s.itemsTable, s.transactionTable))
	links := []cashflow.RecurringLink{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &links, query, accountID, itemID); err != nil {
		return nil, fmt.Errorf("recurring store: list item links: %w", err)
	}
	return links, nil
}

// ListUnlinkedTransactions returns the transactions suggestions may be drawn from:
// an account's own rows that are neither ignored nor already linked to an item.
func (s *SQLXRecurringStore) ListUnlinkedTransactions(ctx context.Context, accountID uuid.UUID) ([]*cashflow.Transaction, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT t.* FROM %s t
		WHERE t.account_id = ?
		  AND t.ignored = ?
		  AND NOT EXISTS (SELECT 1 FROM %s l WHERE l.transaction_id = t.id)
		ORDER BY t.date ASC
	`, s.transactionTable, s.linksTable))
	return s.selectTransactions(ctx, query, accountID, false)
}

// ListDismissedSuggestions returns the patterns an account refused.
func (s *SQLXRecurringStore) ListDismissedSuggestions(ctx context.Context, accountID uuid.UUID) ([]cashflow.DismissedSuggestion, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`SELECT match_key, direction FROM %s WHERE account_id = ?`,
		s.dismissalsTable,
	))
	dismissed := []cashflow.DismissedSuggestion{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &dismissed, query, accountID); err != nil {
		return nil, fmt.Errorf("recurring store: list dismissals: %w", err)
	}
	return dismissed, nil
}

// ListTransactionsByIDs returns the account's transactions among the given ids.
func (s *SQLXRecurringStore) ListTransactionsByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID) ([]*cashflow.Transaction, error) {
	if len(ids) == 0 {
		return []*cashflow.Transaction{}, nil
	}
	query, args, err := sqlx.In(
		fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? AND id IN (?)`, s.transactionTable),
		accountID, ids,
	)
	if err != nil {
		return nil, fmt.Errorf("recurring store: expand ids: %w", err)
	}
	return s.selectTransactions(ctx, s.db.Rebind(query), args...)
}

// ListImportedTransactions returns the rows one import added, skipping the ones
// already linked to an item.
func (s *SQLXRecurringStore) ListImportedTransactions(ctx context.Context, accountID, importID uuid.UUID) ([]*cashflow.Transaction, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT t.* FROM %s t
		WHERE t.account_id = ? AND t.import_id = ?
		  AND NOT EXISTS (SELECT 1 FROM %s l WHERE l.transaction_id = t.id)
		ORDER BY t.date ASC
	`, s.transactionTable, s.linksTable))
	return s.selectTransactions(ctx, query, accountID, importID)
}

func (s *SQLXRecurringStore) selectTransactions(ctx context.Context, query string, args ...any) ([]*cashflow.Transaction, error) {
	transactions := []*cashflow.Transaction{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &transactions, query, args...); err != nil {
		return nil, fmt.Errorf("recurring store: fetch transactions: %w", err)
	}
	return transactions, nil
}

func (s *SQLXRecurringStore) exec(ctx context.Context, query string, args ...any) (int, error) {
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("recurring store: execute statement: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("recurring store: rows affected: %w", err)
	}
	return int(affected), nil
}
