package wealthgoal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// defaultStandingMonths is how far back the standing reaches when the caller does not say.
	defaultStandingMonths = 12
	// maxStandingMonths bounds the scoreboard; it is read, not exported or charted.
	maxStandingMonths = 60
)

// QueryStore reads goals and the marked money the standing is computed from.
type QueryStore interface {
	// ListGoals returns every goal an account has set, oldest first.
	ListGoals(ctx context.Context, accountID uuid.UUID) ([]*Goal, error)
	// MonthlyPurposeTotals returns, per calendar month from `from` onwards, what the
	// account marked as income, what it marked as a contribution, and how many of its
	// transactions carry no purpose yet. Months without any transaction are absent.
	MonthlyPurposeTotals(ctx context.Context, accountID uuid.UUID, from time.Time) ([]MonthlyPurposeTotals, error)
}

// Queries exposes the read-side wealth-goal use cases.
type Queries struct {
	qs QueryStore
}

// NewQueries creates wealth-goal read-side use cases.
func NewQueries(qs QueryStore) *Queries {
	return &Queries{qs: qs}
}

// CurrentGoal returns the goal in force this month, or nil when the account has never set
// one -- the first run, where the standing has nothing to compare against.
func (q *Queries) CurrentGoal(ctx context.Context, accountID uuid.UUID) (*Goal, error) {
	goals, err := q.qs.ListGoals(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("current monthly goal: %w", err)
	}
	return goalAt(goals, StartOfMonth(time.Now())), nil
}

// Standing scores the last `months` calendar months against the goal that applied in each,
// newest first, and reports the run of met months. `months` is clamped to a sane window.
func (q *Queries) Standing(ctx context.Context, accountID uuid.UUID, months int) (*Standing, error) {
	if months <= 0 {
		months = defaultStandingMonths
	}
	if months > maxStandingMonths {
		months = maxStandingMonths
	}

	goals, err := q.qs.ListGoals(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("monthly standing: %w", err)
	}

	from := StartOfMonth(time.Now()).AddDate(0, -(months - 1), 0)
	totals, err := q.qs.MonthlyPurposeTotals(ctx, accountID, from)
	if err != nil {
		return nil, fmt.Errorf("monthly standing: %w", err)
	}

	standing := buildStanding(totals, goals, time.Now())
	return &standing, nil
}
