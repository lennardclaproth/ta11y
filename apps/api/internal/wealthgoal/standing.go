package wealthgoal

import (
	"sort"
	"time"

	"github.com/lennardclaproth/ta11y/internal/date"
)

// MonthResult is how a calendar month ended up against the goal that applied in it.
type MonthResult string

const (
	// ResultMet means the month put at least the goal's share of its income towards wealth.
	ResultMet MonthResult = "met"
	// ResultMissed means it did not.
	ResultMissed MonthResult = "missed"
	// ResultInProgress is the month that is still running. It is never scored and never
	// counts towards the streak -- an import that has not happened yet would break it.
	ResultInProgress MonthResult = "in_progress"
	// ResultNotScored is a month that ended before the account had any goal. Its numbers
	// are readable, but there was nothing to judge them against.
	ResultNotScored MonthResult = "not_scored"
)

// MonthlyPurposeTotals is what cashflow marked in one calendar month.
type MonthlyPurposeTotals struct {
	Month time.Time
	// IncomeCents is the money marked as income, 1e6-scaled like every amount in the API.
	IncomeCents int64
	// ContributedCents is the money marked as a contribution towards wealth.
	ContributedCents int64
	// UnassignedCount is how many transactions in the month nobody has pointed at yet.
	// Ignored rows are left out: they are out of the ledger the jump lands in too.
	UnassignedCount int
}

// MonthStanding is one calendar month scored against the goal that applied in it.
type MonthStanding struct {
	Month            time.Time
	IncomeCents      int64
	ContributedCents int64
	// GoalPercent is the goal in force in this month, not today's goal.
	GoalPercent     int
	Result          MonthResult
	UnassignedCount int
}

// Standing is the monthly scoreboard: the goal in force today, the months that have
// something in them (newest first), and the run of met months.
type Standing struct {
	Goal          *Goal
	Months        []MonthStanding
	CurrentStreak int
	BestStreak    int
}

// buildStanding scores every month that holds marked or unmarked money against the goal in
// force in it. Months are returned newest first; months without any transaction at all are
// absent rather than scored as missed, because nothing happened in them to judge.
func buildStanding(totals []MonthlyPurposeTotals, goals []*Goal, now time.Time) Standing {
	currentMonth := date.StartOfMonthUTC(now)
	ordered := append([]*Goal(nil), goals...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].EffectiveFrom.Before(ordered[j].EffectiveFrom)
	})

	months := make([]MonthStanding, 0, len(totals))
	for _, total := range totals {
		month := date.StartOfMonthUTC(total.Month)
		if month.After(currentMonth) {
			// A month that has not started yet is not part of the standing. An import can
			// carry a date in the future, and scoring it would put a second in_progress
			// month in front of the running one.
			continue
		}
		goal := goalAt(ordered, month)
		if goal == nil {
			// Before the first goal there is nothing to score against. The month still
			// belongs in the standing so its numbers can be read, with no result.
			months = append(months, MonthStanding{
				Month:            month,
				IncomeCents:      total.IncomeCents,
				ContributedCents: total.ContributedCents,
				Result:           ResultNotScored,
				UnassignedCount:  total.UnassignedCount,
			})
			continue
		}
		months = append(months, MonthStanding{
			Month:            month,
			IncomeCents:      total.IncomeCents,
			ContributedCents: total.ContributedCents,
			GoalPercent:      goal.SharePercent,
			Result:           result(month, currentMonth, total, goal.SharePercent),
			UnassignedCount:  total.UnassignedCount,
		})
	}

	// The month you are in is the one the card leads with, so it is always present once a
	// goal exists -- an empty running month reads as "nothing marked yet", not as a gap.
	if currentGoal := goalAt(ordered, currentMonth); currentGoal != nil && !holdsMonth(months, currentMonth) {
		months = append(months, MonthStanding{
			Month:       currentMonth,
			GoalPercent: currentGoal.SharePercent,
			Result:      ResultInProgress,
		})
	}

	sort.Slice(months, func(i, j int) bool { return months[i].Month.After(months[j].Month) })

	current, best := streaks(months)
	return Standing{
		Goal:          goalAt(ordered, currentMonth),
		Months:        months,
		CurrentStreak: current,
		BestStreak:    best,
	}
}

// result scores one month. A month with nothing marked as income is missed rather than met:
// the goal is a share of income, and no income means none of it was put towards wealth.
func result(month, currentMonth time.Time, total MonthlyPurposeTotals, goalPercent int) MonthResult {
	if !month.Before(currentMonth) {
		return ResultInProgress
	}
	if total.IncomeCents <= 0 {
		return ResultMissed
	}
	// Compared before rounding, so a month that is a fraction of a percent short is short.
	if total.ContributedCents*100 >= total.IncomeCents*int64(goalPercent) {
		return ResultMet
	}
	return ResultMissed
}

func holdsMonth(months []MonthStanding, month time.Time) bool {
	for _, candidate := range months {
		if candidate.Month.Equal(month) {
			return true
		}
	}
	return false
}

// goalAt returns the goal in force in month: the latest one that had started by then.
func goalAt(ascending []*Goal, month time.Time) *Goal {
	var found *Goal
	for _, goal := range ascending {
		if goal.EffectiveFrom.After(month) {
			break
		}
		found = goal
	}
	return found
}

// streaks walks the scored months newest first. The running month is skipped rather than
// treated as a break -- it has not finished, so it cannot have been missed.
func streaks(months []MonthStanding) (current, best int) {
	counting := true
	run := 0
	for _, month := range months {
		switch month.Result {
		case ResultInProgress, ResultNotScored:
			continue
		case ResultMet:
			run++
			if run > best {
				best = run
			}
			if counting {
				current = run
			}
		default:
			run = 0
			counting = false
		}
	}
	return current, best
}
