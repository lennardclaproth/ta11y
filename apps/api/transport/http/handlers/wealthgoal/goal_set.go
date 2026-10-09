package wealthgoal

import (
	"errors"
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/wealthgoal"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// SetGoalRequest sets the share of income that should go towards wealth each month.
type SetGoalRequest struct {
	SharePercent *int `json:"share_percent"`
}

// SetGoal records the monthly goal, applying from the current calendar month.
//
// @Summary     Set the monthly wealth goal
// @Description Sets the share of income that should go towards wealth each month. The new goal applies from the current calendar month; months already scored keep the goal they were judged by.
// @Tags        WealthGoal
// @Accept      application/json
// @Produce     application/json
// @Param       payload body SetGoalRequest true "Monthly goal"
// @Success     200 {object} GoalResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /wealth-goal [put]
func SetGoal(log logging.Logger, commands *wealthgoal.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, err := httpx.JSONDecode[SetGoalRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		if req.SharePercent == nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"share_percent": "share_percent is required"})
			return
		}

		goal, err := commands.SetGoal(r.Context(), accountID, *req.SharePercent)
		if err != nil {
			if errors.Is(err, wealthgoal.ErrInvalidSharePercent) {
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"share_percent": wealthgoal.ErrInvalidSharePercent.Error()})
				return
			}
			log.Error(r.Context(), "wealth goal: failed to save the goal", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to save the monthly goal"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, GoalResponse{Goal: toGoalResponse(goal)})
	})
}
