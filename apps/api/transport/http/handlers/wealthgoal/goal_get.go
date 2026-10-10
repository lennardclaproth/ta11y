// Package wealthgoal exposes the monthly wealth-growth goal [034] over HTTP: the one goal
// an account sets, and the monthly standing scored against it.
package wealthgoal

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/wealthgoal"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// GoalResponse is the monthly goal in force. `goal` is null when the account has never set
// one, which is the first run the web app sends to the goal dialog first.
type GoalResponse struct {
	Goal *Goal `json:"goal"`
}

// Goal is the share of income an account puts towards wealth, from one month onwards.
type Goal struct {
	SharePercent int `json:"share_percent"`
	// EffectiveFrom is the first of the calendar month the goal started applying in,
	// as "YYYY-MM-DD".
	EffectiveFrom string `json:"effective_from"`
}

// GetGoal returns the monthly goal in force for the signed-in account.
//
// @Summary     Get the monthly wealth goal
// @Description The share of income that should go towards wealth each month. `goal` is null when none has been set.
// @Tags        WealthGoal
// @Produce     application/json
// @Success     200 {object} GoalResponse
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /wealth-goal [get]
func GetGoal(log logging.Logger, queries *wealthgoal.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		goal, err := queries.CurrentGoal(r.Context(), accountID)
		if err != nil {
			log.Error(r.Context(), "wealth goal: failed to read the current goal", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get the monthly goal"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, GoalResponse{Goal: toGoalResponse(goal)})
	})
}

func toGoalResponse(goal *wealthgoal.Goal) *Goal {
	if goal == nil {
		return nil
	}
	return &Goal{
		SharePercent:  goal.SharePercent,
		EffectiveFrom: goal.EffectiveFrom.Format("2006-01-02"),
	}
}
