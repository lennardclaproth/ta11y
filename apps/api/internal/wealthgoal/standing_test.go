package wealthgoal

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func month(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse month %q: %v", value, err)
	}
	return parsed.UTC()
}

func goalFrom(t *testing.T, from string, percent int) *Goal {
	t.Helper()
	goal, err := NewGoal(uuid.New(), percent, month(t, from))
	if err != nil {
		t.Fatalf("new goal: %v", err)
	}
	return goal
}

// A month met its goal when the share it put towards wealth reaches the goal that applied
// in it, compared before rounding so a fraction short is short.
func TestBuildStandingScoresMonthsAgainstTheGoalInForce(t *testing.T) {
	now := month(t, "2026-07-10")
	goals := []*Goal{
		goalFrom(t, "2026-01-01", 25),
		goalFrom(t, "2026-03-01", 30),
	}
	totals := []MonthlyPurposeTotals{
		{Month: month(t, "2026-01-01"), IncomeCents: 1000, ContributedCents: 250},
		{Month: month(t, "2026-03-01"), IncomeCents: 1000, ContributedCents: 299},
		{Month: month(t, "2026-04-01"), IncomeCents: 1000, ContributedCents: 300},
		{Month: month(t, "2026-07-01"), IncomeCents: 1000, ContributedCents: 10},
	}

	standing := buildStanding(totals, goals, now)

	want := []struct {
		month  string
		goal   int
		result MonthResult
	}{
		{"2026-07-01", 30, ResultInProgress},
		{"2026-04-01", 30, ResultMet},
		{"2026-03-01", 30, ResultMissed},
		{"2026-01-01", 25, ResultMet},
	}
	if len(standing.Months) != len(want) {
		t.Fatalf("months = %d, want %d", len(standing.Months), len(want))
	}
	for i, expected := range want {
		got := standing.Months[i]
		if got.Month != month(t, expected.month) {
			t.Errorf("months[%d].Month = %s, want %s", i, got.Month, expected.month)
		}
		if got.GoalPercent != expected.goal {
			t.Errorf("months[%d].GoalPercent = %d, want %d", i, got.GoalPercent, expected.goal)
		}
		if got.Result != expected.result {
			t.Errorf("months[%d].Result = %s, want %s", i, got.Result, expected.result)
		}
	}
	if standing.Goal == nil || standing.Goal.SharePercent != 30 {
		t.Errorf("standing goal = %v, want the 30%% goal", standing.Goal)
	}
}

// A month with nothing marked as income is missed: the goal is a share of income, and no
// income means nothing was put towards it.
func TestBuildStandingTreatsAMonthWithoutIncomeAsMissed(t *testing.T) {
	now := month(t, "2026-07-10")
	totals := []MonthlyPurposeTotals{
		{Month: month(t, "2026-06-01"), IncomeCents: 0, ContributedCents: 500},
	}

	standing := buildStanding(totals, []*Goal{goalFrom(t, "2026-01-01", 30)}, now)

	if standing.Months[1].Result != ResultMissed {
		t.Errorf("June result = %s, want %s", standing.Months[1].Result, ResultMissed)
	}
}

// The running month is always present once a goal exists, so the card that leads with it
// has something to show before anything is marked.
func TestBuildStandingAlwaysHoldsTheRunningMonth(t *testing.T) {
	now := month(t, "2026-07-10")

	standing := buildStanding(nil, []*Goal{goalFrom(t, "2026-01-01", 30)}, now)

	if len(standing.Months) != 1 {
		t.Fatalf("months = %d, want 1", len(standing.Months))
	}
	if standing.Months[0].Month != month(t, "2026-07-01") {
		t.Errorf("month = %s, want 2026-07-01", standing.Months[0].Month)
	}
	if standing.Months[0].Result != ResultInProgress {
		t.Errorf("result = %s, want %s", standing.Months[0].Result, ResultInProgress)
	}
}

// A transaction dated in a month that has not started yet -- an import carries no date
// validation -- must not add a second in_progress month in front of the running one.
func TestBuildStandingLeavesOutMonthsThatHaveNotStarted(t *testing.T) {
	now := month(t, "2026-07-10")
	totals := []MonthlyPurposeTotals{
		{Month: month(t, "2026-07-01"), IncomeCents: 1000, ContributedCents: 300},
		{Month: month(t, "2027-01-01"), IncomeCents: 1000, ContributedCents: 0},
	}

	standing := buildStanding(totals, []*Goal{goalFrom(t, "2026-01-01", 30)}, now)

	if len(standing.Months) != 1 {
		t.Fatalf("months = %d, want 1", len(standing.Months))
	}
	if standing.Months[0].Month != month(t, "2026-07-01") {
		t.Errorf("month = %s, want 2026-07-01", standing.Months[0].Month)
	}
	if standing.Months[0].Result != ResultInProgress {
		t.Errorf("result = %s, want %s", standing.Months[0].Result, ResultInProgress)
	}
}

// Without a goal there is nothing to score against, and nothing is invented.
func TestBuildStandingWithoutAGoalScoresNothing(t *testing.T) {
	now := month(t, "2026-07-10")
	totals := []MonthlyPurposeTotals{
		{Month: month(t, "2026-06-01"), IncomeCents: 1000, ContributedCents: 500},
	}

	standing := buildStanding(totals, nil, now)

	if standing.Goal != nil {
		t.Errorf("goal = %v, want nil", standing.Goal)
	}
	if standing.Months[0].Result != ResultNotScored {
		t.Errorf("result = %s, want %s", standing.Months[0].Result, ResultNotScored)
	}
	if standing.CurrentStreak != 0 || standing.BestStreak != 0 {
		t.Errorf("streaks = %d/%d, want 0/0", standing.CurrentStreak, standing.BestStreak)
	}
}

// The streak counts finished months only: the running month never breaks it, and a month
// that still holds unassigned transactions counts with the result it has.
func TestStreaksCountFinishedMonthsOnly(t *testing.T) {
	now := month(t, "2026-07-10")
	goals := []*Goal{goalFrom(t, "2026-01-01", 30)}
	totals := []MonthlyPurposeTotals{
		{Month: month(t, "2026-01-01"), IncomeCents: 1000, ContributedCents: 400},
		{Month: month(t, "2026-02-01"), IncomeCents: 1000, ContributedCents: 100},
		{Month: month(t, "2026-03-01"), IncomeCents: 1000, ContributedCents: 300},
		{Month: month(t, "2026-04-01"), IncomeCents: 1000, ContributedCents: 350},
		{Month: month(t, "2026-05-01"), IncomeCents: 1000, ContributedCents: 300, UnassignedCount: 2},
		{Month: month(t, "2026-06-01"), IncomeCents: 1000, ContributedCents: 320},
		{Month: month(t, "2026-07-01"), IncomeCents: 1000, ContributedCents: 0},
	}

	standing := buildStanding(totals, goals, now)

	if standing.CurrentStreak != 4 {
		t.Errorf("current streak = %d, want 4", standing.CurrentStreak)
	}
	if standing.BestStreak != 4 {
		t.Errorf("best streak = %d, want 4", standing.BestStreak)
	}
}

func TestNewGoalRefusesAShareOutsideZeroToHundred(t *testing.T) {
	for _, share := range []int{-1, 101} {
		if _, err := NewGoal(uuid.New(), share, time.Now()); err != ErrInvalidSharePercent {
			t.Errorf("NewGoal(%d) error = %v, want %v", share, err, ErrInvalidSharePercent)
		}
	}
}

// A goal applies from the first of the month it was set in, whatever day that was.
func TestNewGoalStartsAtTheFirstOfItsMonth(t *testing.T) {
	goal, err := NewGoal(uuid.New(), 30, month(t, "2026-07-23"))
	if err != nil {
		t.Fatalf("new goal: %v", err)
	}
	if goal.EffectiveFrom != month(t, "2026-07-01") {
		t.Errorf("effective from = %s, want 2026-07-01", goal.EffectiveFrom)
	}
}
