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

// SQLXCashflowIgnoreRuleStore persists cashflow ignore rules and applies them to
// transactions. It satisfies the cashflow ignore-rule command and query contracts.
type SQLXCashflowIgnoreRuleStore struct {
	db                *DB
	tableName         string
	transactionsTable string
}

var (
	_ cashflow.IgnoreRuleCommandStore = (*SQLXCashflowIgnoreRuleStore)(nil)
	_ cashflow.IgnoreRuleQueryStore   = (*SQLXCashflowIgnoreRuleStore)(nil)
)

// NewSQLXCashflowIgnoreRuleStore creates an ignore-rule store backed by SQLX.
func NewSQLXCashflowIgnoreRuleStore(db *DB) *SQLXCashflowIgnoreRuleStore {
	return &SQLXCashflowIgnoreRuleStore{
		db:                db,
		tableName:         qualifyTableAs(db, SchemaCashflow, TableIgnoreRules, "cashflow_ignore_rules"),
		transactionsTable: qualifyTable(db, SchemaCashflow, TableTransactions),
	}
}

// CreateIgnoreRule stores a new ignore rule.
func (s *SQLXCashflowIgnoreRuleStore) CreateIgnoreRule(ctx context.Context, rule *cashflow.IgnoreRule) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, name, match_field, contains, direction,
			source, enabled, ignored_total, last_applied_at, created_at, updated_at
		) VALUES (
			:id, :account_id, :name, :match_field, :contains, :direction,
			:source, :enabled, :ignored_total, :last_applied_at, :created_at, :updated_at
		)
	`, s.tableName)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, rule); err != nil {
		return fmt.Errorf("cashflow ignore rule store: insert: %w", err)
	}
	return nil
}

// UpdateIgnoreRule rewrites a rule's matching fields. Its counters are left alone:
// they record what the rule has done, not what it now matches.
func (s *SQLXCashflowIgnoreRuleStore) UpdateIgnoreRule(ctx context.Context, rule *cashflow.IgnoreRule) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET name = ?, match_field = ?, contains = ?, direction = ?, source = ?, enabled = ?, updated_at = ?
		WHERE account_id = ? AND id = ?
	`, s.tableName))
	return s.exec(ctx, query,
		rule.Name, rule.MatchField, rule.Contains, rule.Direction,
		rule.Source, rule.Enabled, rule.UpdatedAt, rule.AccountID, rule.ID,
	)
}

// DeleteIgnoreRule removes one of the account's rules. Transactions it ignored keep
// their state; the foreign key nulls their reference to it.
func (s *SQLXCashflowIgnoreRuleStore) DeleteIgnoreRule(ctx context.Context, accountID, id uuid.UUID) (int, error) {
	query := s.db.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE account_id = ? AND id = ?`, s.tableName))
	return s.exec(ctx, query, accountID, id)
}

// ListIgnoreRules returns the account's rules, newest first.
func (s *SQLXCashflowIgnoreRuleStore) ListIgnoreRules(ctx context.Context, accountID uuid.UUID) ([]*cashflow.IgnoreRule, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`SELECT * FROM %s WHERE account_id = ? ORDER BY created_at DESC`, s.tableName,
	))
	rules := []*cashflow.IgnoreRule{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rules, query, accountID); err != nil {
		return nil, fmt.Errorf("cashflow ignore rule store: list: %w", err)
	}
	return rules, nil
}

// GetIgnoreRule returns one of the account's rules, or nil when it holds no rule with
// that id.
func (s *SQLXCashflowIgnoreRuleStore) GetIgnoreRule(ctx context.Context, accountID, id uuid.UUID) (*cashflow.IgnoreRule, error) {
	query := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ? AND id = ?`, s.tableName))
	var rule cashflow.IgnoreRule
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &rule, query, accountID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("cashflow ignore rule store: fetch: %w", err)
	}
	return &rule, nil
}

// CountIgnoreRuleTargets counts what applying the rule would actually ignore: rows it
// matches that are not ignored yet and whose ignored state was not decided by hand.
func (s *SQLXCashflowIgnoreRuleStore) CountIgnoreRuleTargets(ctx context.Context, filters cashflow.TransactionFilters) (int, error) {
	whereClause, args := s.targetWhereClause(filters, nil)
	query := s.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s%s", s.transactionsTable, whereClause))
	total := 0
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &total, query, args...); err != nil {
		return 0, fmt.Errorf("cashflow ignore rule store: count targets: %w", err)
	}
	return total, nil
}

// ApplyIgnoreRule ignores what the rule matches and has not been decided by hand, and
// adds the number it ignored to the rule's own total. A non-nil importID narrows it to
// the rows one import brought in.
func (s *SQLXCashflowIgnoreRuleStore) ApplyIgnoreRule(ctx context.Context, rule *cashflow.IgnoreRule, importID *uuid.UUID) (int, error) {
	whereClause, whereArgs := s.targetWhereClause(rule.Filters(), importID)
	now := time.Now().UTC()
	update := s.db.Rebind(fmt.Sprintf(
		`UPDATE %s SET ignored = ?, ignored_by_rule_id = ?, updated_at = ?%s`,
		s.transactionsTable, whereClause,
	))
	args := append([]any{true, rule.ID, now}, whereArgs...)

	ignored, err := s.exec(ctx, update, args...)
	if err != nil {
		return 0, err
	}
	if ignored == 0 {
		return 0, nil
	}

	touch := s.db.Rebind(fmt.Sprintf(
		`UPDATE %s SET ignored_total = ignored_total + ?, last_applied_at = ?, updated_at = ? WHERE id = ?`,
		s.tableName,
	))
	if _, err := s.exec(ctx, touch, ignored, now, now, rule.ID); err != nil {
		return ignored, err
	}
	rule.IgnoredTotal += ignored
	rule.LastAppliedAt = &now
	return ignored, nil
}

// ListIgnoredByRule returns what each rule ignored in one import, newest first, capped
// at perRule rows per group alongside the group's full total. A row restored by hand
// stays in its group: the page has to be able to say the rule caught it and that it
// was put back.
func (s *SQLXCashflowIgnoreRuleStore) ListIgnoredByRule(ctx context.Context, accountID, importID uuid.UUID, perRule int) ([]cashflow.IgnoredRuleGroup, error) {
	rules, err := s.ListIgnoreRules(ctx, accountID)
	if err != nil {
		return nil, err
	}

	totals, err := s.countIgnoredPerRule(ctx, accountID, importID)
	if err != nil {
		return nil, err
	}

	groups := make([]cashflow.IgnoredRuleGroup, 0, len(totals))
	for _, rule := range rules {
		total, ok := totals[rule.ID]
		if !ok {
			continue
		}
		rows, err := s.ignoredByRule(ctx, accountID, importID, rule.ID, perRule)
		if err != nil {
			return nil, err
		}
		groups = append(groups, cashflow.IgnoredRuleGroup{Rule: rule, Total: total, Transactions: rows})
	}
	return groups, nil
}

func (s *SQLXCashflowIgnoreRuleStore) countIgnoredPerRule(ctx context.Context, accountID, importID uuid.UUID) (map[uuid.UUID]int, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT ignored_by_rule_id AS rule_id, COUNT(1) AS total
		FROM %s
		WHERE account_id = ? AND import_id = ? AND ignored_by_rule_id IS NOT NULL
		GROUP BY ignored_by_rule_id
	`, s.transactionsTable))

	type row struct {
		RuleID uuid.UUID `db:"rule_id"`
		Total  int       `db:"total"`
	}
	var rows []row
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, query, accountID, importID); err != nil {
		return nil, fmt.Errorf("cashflow ignore rule store: count ignored per rule: %w", err)
	}

	totals := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		totals[r.RuleID] = r.Total
	}
	return totals, nil
}

func (s *SQLXCashflowIgnoreRuleStore) ignoredByRule(ctx context.Context, accountID, importID, ruleID uuid.UUID, limit int) ([]*cashflow.Transaction, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE account_id = ? AND import_id = ? AND ignored_by_rule_id = ?
		ORDER BY date DESC
		LIMIT ?
	`, s.transactionsTable))
	transactions := []*cashflow.Transaction{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &transactions, query, accountID, importID, ruleID, limit); err != nil {
		return nil, fmt.Errorf("cashflow ignore rule store: list ignored by rule: %w", err)
	}
	return transactions, nil
}

// targetWhereClause narrows the ledger filters a rule carries to the rows a rule is
// still allowed to touch: not ignored yet, and not decided by hand.
func (s *SQLXCashflowIgnoreRuleStore) targetWhereClause(filters cashflow.TransactionFilters, importID *uuid.UUID) (string, []any) {
	hideIgnored := true
	filters.HideIgnored = &hideIgnored
	filters.ImportID = importID

	whereClause, args := buildCashflowWhereClause(cashflowQueryFromFilters(filters))
	condition := "ignore_overridden = ?"
	args = append(args, false)
	if whereClause == "" {
		return " WHERE " + condition, args
	}
	return whereClause + " AND " + condition, args
}

func (s *SQLXCashflowIgnoreRuleStore) exec(ctx context.Context, query string, args ...any) (int, error) {
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("cashflow ignore rule store: exec: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("cashflow ignore rule store: rows affected: %w", err)
	}
	return int(affected), nil
}
