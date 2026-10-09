// Package wealthgoal keeps the one monthly goal an account sets -- the share of the income
// it marks that should go towards wealth -- and scores each calendar month against the goal
// that applied in it. The money itself lives in cashflow: a transaction marked as income or
// as a contribution is what the standing counts.
package wealthgoal

import (
	"time"

	"github.com/google/uuid"
)

// Goal is the share of income an account wants to put towards wealth, from one month
// onwards. Adjusting the goal writes a new row for the month it is adjusted in, so a month
// that has already been scored keeps the goal it was judged by.
type Goal struct {
	ID            uuid.UUID `db:"id"`
	AccountID     uuid.UUID `db:"account_id"`
	SharePercent  int       `db:"share_percent"`
	EffectiveFrom time.Time `db:"effective_from"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// NewGoal validates a share and builds the goal that applies from effectiveFrom's calendar
// month onwards.
func NewGoal(accountID uuid.UUID, sharePercent int, effectiveFrom time.Time) (*Goal, error) {
	if sharePercent < 0 || sharePercent > 100 {
		return nil, ErrInvalidSharePercent
	}
	now := time.Now().UTC()
	return &Goal{
		ID:            uuid.New(),
		AccountID:     accountID,
		SharePercent:  sharePercent,
		EffectiveFrom: StartOfMonth(effectiveFrom),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// StartOfMonth is the first day of t's calendar month in UTC. A month in this feature is
// always the calendar month of the transaction date -- there is no custom period.
func StartOfMonth(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
}
