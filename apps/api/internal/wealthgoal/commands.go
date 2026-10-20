package wealthgoal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CommandStore persists monthly goals.
type CommandStore interface {
	// SaveGoal stores the goal for its account and month, replacing the share already
	// recorded for that month.
	SaveGoal(ctx context.Context, goal *Goal) error
}

// Commands exposes the write-side wealth-goal use cases.
type Commands struct {
	cs CommandStore
}

// NewCommands creates wealth-goal write-side use cases.
func NewCommands(cs CommandStore) *Commands {
	return &Commands{cs: cs}
}

// SetGoal records the share of income the account wants to put towards wealth. It applies
// from the current calendar month onwards: months that have already been scored keep the
// goal they were judged by, so raising the bar never turns a met month into a missed one.
func (c *Commands) SetGoal(ctx context.Context, accountID uuid.UUID, sharePercent int) (*Goal, error) {
	goal, err := NewGoal(accountID, sharePercent, time.Now())
	if err != nil {
		return nil, fmt.Errorf("set monthly goal: %w", err)
	}
	if err := c.cs.SaveGoal(ctx, goal); err != nil {
		return nil, fmt.Errorf("set monthly goal: %w", err)
	}
	return goal, nil
}
