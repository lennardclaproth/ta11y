package wealthgoal

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/wealthgoal"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// GetStandingRequest bounds how far back the standing reaches.
type GetStandingRequest struct {
	Months int `query:"months"`
}

// MonthStandingResponse is one calendar month scored against the goal that applied in it.
type MonthStandingResponse struct {
	// Month is the first of the calendar month, as "YYYY-MM-DD".
	Month string `json:"month"`
	// IncomeCents and ContributedCents are 1e6-scaled, like every amount in this API.
	IncomeCents      int64 `json:"income_cents"`
	ContributedCents int64 `json:"contributed_cents"`
	// GoalPercent is the goal in force in this month, not today's goal.
	GoalPercent int `json:"goal_percent"`
	// Result is one of met, missed, in_progress, not_scored.
	Result string `json:"result"`
	// UnassignedCount is how many transactions in the month carry no purpose yet.
	UnassignedCount int `json:"unassigned_count"`
}

// StandingResponse is the monthly scoreboard: the goal in force today, the months that
// hold something (newest first), and the run of met months.
type StandingResponse struct {
	Goal          *Goal                   `json:"goal"`
	Months        []MonthStandingResponse `json:"months"`
	CurrentStreak int                     `json:"current_streak"`
	BestStreak    int                     `json:"best_streak"`
}

// GetStanding returns the monthly standing for the signed-in account.
//
// @Summary     Get the monthly wealth-goal standing
// @Description Per calendar month: what was marked as income, what went towards wealth, the goal that applied, and whether the month was met. The running month is never scored, and the streak counts only finished months.
// @Tags        WealthGoal
// @Produce     application/json
// @Param       months query int false "How many calendar months back to score (default 12, max 60)"
// @Success     200 {object} StandingResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /wealth-goal/standing [get]
func GetStanding(log logging.Logger, queries *wealthgoal.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, err := httpx.DecodeQuery[GetStandingRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}
		if req.Months < 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"months": "months must be greater than or equal to 0"})
			return
		}

		standing, err := queries.Standing(r.Context(), accountID, req.Months)
		if err != nil {
			log.Error(r.Context(), "wealth goal: failed to read the monthly standing", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get the monthly standing"})
			return
		}

		months := make([]MonthStandingResponse, 0, len(standing.Months))
		for _, month := range standing.Months {
			months = append(months, MonthStandingResponse{
				Month:            month.Month.Format("2006-01-02"),
				IncomeCents:      month.IncomeCents,
				ContributedCents: month.ContributedCents,
				GoalPercent:      month.GoalPercent,
				Result:           string(month.Result),
				UnassignedCount:  month.UnassignedCount,
			})
		}

		_ = httpx.JSONEncode(w, http.StatusOK, StandingResponse{
			Goal:          toGoalResponse(standing.Goal),
			Months:        months,
			CurrentStreak: standing.CurrentStreak,
			BestStreak:    standing.BestStreak,
		})
	})
}
