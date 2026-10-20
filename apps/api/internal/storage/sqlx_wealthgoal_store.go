package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/ta11y/internal/wealthgoal"
)

// SQLXWealthGoalStore persists monthly goals and reads the marked cashflow money the
// monthly standing is computed from. The standing is a read model over
// cashflow.transactions, so this store reads that table rather than duplicating its
// contents into a second one.
type SQLXWealthGoalStore struct {
	db                *DB
	goalsTable        string
	transactionsTable string
}

var (
	_ wealthgoal.CommandStore = (*SQLXWealthGoalStore)(nil)
	_ wealthgoal.QueryStore   = (*SQLXWealthGoalStore)(nil)
)

// NewSQLXWealthGoalStore creates a wealth-goal store backed by SQLX.
func NewSQLXWealthGoalStore(db *DB) *SQLXWealthGoalStore {
	return &SQLXWealthGoalStore{
		db:                db,
		goalsTable:        qualifyTableAs(db, SchemaWealthGoal, TableWealthGoals, "wealth_goals"),
		transactionsTable: qualifyTable(db, SchemaCashflow, TableTransactions),
	}
}

// SaveGoal stores the share for the goal's month, replacing the one already recorded for
// it. Adjusting the goal twice in the same month is a correction, not a second goal.
func (s *SQLXWealthGoalStore) SaveGoal(ctx context.Context, goal *wealthgoal.Goal) error {
	query := s.db.Rebind(fmt.Sprintf(`
		INSERT INTO %s (id, account_id, share_percent, effective_from, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (account_id, effective_from)
		DO UPDATE SET share_percent = EXCLUDED.share_percent, updated_at = EXCLUDED.updated_at
	`, s.goalsTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(
		ctx,
		query,
		goal.ID,
		goal.AccountID,
		goal.SharePercent,
		goal.EffectiveFrom,
		goal.CreatedAt,
		goal.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("wealth goal store: save goal: %w", err)
	}
	return nil
}

// ListGoals returns every goal an account has set, oldest first, so the month a goal
// applies from can be resolved by walking forward.
func (s *SQLXWealthGoalStore) ListGoals(ctx context.Context, accountID uuid.UUID) ([]*wealthgoal.Goal, error) {
	query := s.db.Rebind(fmt.Sprintf(
		`SELECT id, account_id, share_percent, effective_from, created_at, updated_at
		 FROM %s WHERE account_id = ? ORDER BY effective_from ASC`,
		s.goalsTable,
	))
	goals := []*wealthgoal.Goal{}
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &goals, query, accountID); err != nil {
		return nil, fmt.Errorf("wealth goal store: list goals: %w", err)
	}
	return goals, nil
}

// MonthlyPurposeTotals groups an account's transactions into calendar months and reports
// what was marked as income, what was marked as a contribution, and how many rows still
// carry no purpose. Marked rows count whatever their ignored state -- a transfer to a
// savings account is usually ignored for the cashflow totals and is still a contribution.
// The unmarked count leaves ignored rows out, because the ledger the standing links to
// hides them too.
func (s *SQLXWealthGoalStore) MonthlyPurposeTotals(ctx context.Context, accountID uuid.UUID, from time.Time) ([]wealthgoal.MonthlyPurposeTotals, error) {
	monthExpr := "TO_CHAR(DATE_TRUNC('month', date), 'YYYY-MM-01')"
	if s.db.DriverName() == string(Sqlite) {
		monthExpr = "COALESCE(STRFTIME('%Y-%m-01', date), SUBSTR(CAST(date AS TEXT), 1, 7) || '-01')"
	}

	query := s.db.Rebind(fmt.Sprintf(`
		SELECT
			%s AS month_start,
			COALESCE(SUM(CASE WHEN purpose = 'income' THEN amount_cents ELSE 0 END), 0) AS income_cents,
			COALESCE(SUM(CASE WHEN purpose = 'wealth' THEN amount_cents ELSE 0 END), 0) AS contributed_cents,
			COALESCE(SUM(CASE WHEN COALESCE(purpose, '') = '' AND ignored = ? THEN 1 ELSE 0 END), 0) AS unassigned_count
		FROM %s
		WHERE account_id = ? AND date >= ?
		GROUP BY 1
		ORDER BY 1 ASC
	`, monthExpr, s.transactionsTable))

	type row struct {
		MonthStart       string `db:"month_start"`
		IncomeCents      int64  `db:"income_cents"`
		ContributedCents int64  `db:"contributed_cents"`
		UnassignedCount  int    `db:"unassigned_count"`
	}
	var rows []row
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, query, false, accountID, from); err != nil {
		return nil, fmt.Errorf("wealth goal store: fetch monthly purpose totals: %w", err)
	}

	totals := make([]wealthgoal.MonthlyPurposeTotals, 0, len(rows))
	for _, r := range rows {
		month, err := time.Parse("2006-01-02", r.MonthStart)
		if err != nil {
			return nil, fmt.Errorf("wealth goal store: parse month %q: %w", r.MonthStart, err)
		}
		totals = append(totals, wealthgoal.MonthlyPurposeTotals{
			Month:            month.UTC(),
			IncomeCents:      r.IncomeCents,
			ContributedCents: r.ContributedCents,
			UnassignedCount:  r.UnassignedCount,
		})
	}
	return totals, nil
}
