package assets

import (
	"net/http"
	"time"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/money"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// UpdateMutationRequest is a corrected worth change. Amount is the new worth for a
// change that set the worth, and the amount added or withdrawn for one that adjusted
// it; direction only applies to the latter.
type UpdateMutationRequest struct {
	Amount        string  `json:"amount"`
	Direction     *string `json:"direction,omitempty"`
	EffectiveDate string  `json:"effective_date"`
	Note          *string `json:"note,omitempty"`
}

func (r UpdateMutationRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if amount, err := money.ParsePrice(r.Amount); err != nil {
		problems["amount"] = "amount must be a valid decimal string"
	} else if amount < 0 {
		problems["amount"] = "amount cannot be negative"
	}
	if r.Direction != nil &&
		*r.Direction != string(assets.ChangeDirectionIncrease) &&
		*r.Direction != string(assets.ChangeDirectionDecrease) {
		problems["direction"] = "direction must be either INCREASE or DECREASE"
	}
	if _, err := time.Parse(time.DateOnly, r.EffectiveDate); err != nil {
		problems["effective_date"] = "effective_date must be in YYYY-MM-DD format"
	}
	return len(problems) == 0, problems
}

// UpdateMutation corrects one worth change of an item whose worth is set by hand.
//
// @Summary Edit worth change
// @Description Corrects the amount, direction, date or note of a worth change. Its kind stays the same. The class history is replayed in date order afterwards, so a changed date or amount carries through to every later change and to net worth.
// @Tags assets
// @Accept json
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Param mutation_id path string true "Worth change ID"
// @Param request body UpdateMutationRequest true "Corrected worth change"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id}/mutations/{mutation_id} [put]
func UpdateMutation(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		assetID, ok := pathID(w, r, "asset_id")
		if !ok {
			return
		}
		mutationID, ok := pathID(w, r, "mutation_id")
		if !ok {
			return
		}
		req, err := httpx.JSONDecode[UpdateMutationRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		if isValid, problems := req.isValid(); !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		// isValid has parsed both already.
		amount, _ := money.ParsePrice(req.Amount)
		effectiveDate, _ := time.Parse(time.DateOnly, req.EffectiveDate)
		edit := assets.MutationEdit{Amount: amount, EffectiveDate: effectiveDate, Note: req.Note}
		if req.Direction != nil {
			direction := assets.ChangeDirection(*req.Direction)
			edit.Direction = &direction
		}
		err = commands.UpdateMutation(r.Context(), accountID, assetID, mutationID, edit)
		if writeEditError(w, r, log, "update worth change", err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
