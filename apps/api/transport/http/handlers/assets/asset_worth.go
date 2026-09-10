package assets

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/money"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

type SetAssetWorthRequest struct {
	Worth         string  `json:"worth"`
	EffectiveDate string  `json:"effective_date"`
	Note          *string `json:"note,omitempty"`
}

func (r SetAssetWorthRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)

	_, err := money.ParsePrice(r.Worth)
	if err != nil {
		problems["worth"] = "invalid worth format"
	}

	_, err = time.Parse(time.DateOnly, r.EffectiveDate)
	if err != nil {
		problems["effective_date"] = "effective_date must be in YYYY-MM-DD format"
	}

	return len(problems) == 0, problems
}

// SetAssetWorth sets absolute worth for an asset.
//
// @Summary Set asset item worth
// @Description Replaces item worth using a non-future effective date.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body SetAssetWorthRequest true "Set worth payload"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id}/worth [put]
func SetAssetWorth(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		// get asset_id from path and validate
		assetID, err := uuid.Parse(r.PathValue("asset_id"))
		if err != nil || assetID == uuid.Nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
				"asset_id": "asset_id must be a valid UUID",
			})
			return
		}
		req, err := httpx.JSONDecode[SetAssetWorthRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}
		//nolint:errcheck
		worth, _ := money.ParsePrice(req.Worth)
		//nolint:errcheck
		effectiveDate, _ := time.Parse(time.DateOnly, req.EffectiveDate)

		err = commands.UpdateAssetWorth(
			r.Context(),
			accountID,
			assetID,
			worth,
			assets.ChangeTypeSet,
			nil,
			effectiveDate,
			req.Note,
		)

		switch {
		// We use the same not found response for both asset and class not
		// found to avoid leaking existence of asset IDs.
		case errors.Is(err, assets.ErrAssetNotFound) || errors.Is(err, assets.ErrClassAccountMismatch):
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
			return
		case errors.Is(err, assets.ErrClassReserved):
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "asset class is reserved and cannot be modified"})
			return
		case err != nil:
			log.Error(r.Context(), "set asset worth: failed to update asset worth", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update asset worth"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// AdjustAssetWorthRequest applies a directional delta to an item worth.
type AdjustAssetWorthRequest struct {
	Direction     string  `json:"direction"`
	Amount        string  `json:"amount"`
	EffectiveDate string  `json:"effective_date"`
	Note          *string `json:"note,omitempty"`
}

func (r AdjustAssetWorthRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.Direction != string(assets.ChangeDirectionIncrease) && r.Direction != string(assets.ChangeDirectionDecrease) {
		problems["direction"] = "direction must be either INCREASE or DECREASE"
	}

	_, err := money.ParsePrice(r.Amount)
	if err != nil {
		problems["worth"] = "invalid worth format"
	}

	_, err = time.Parse(time.DateOnly, r.EffectiveDate)
	if err != nil {
		problems["effective_date"] = "effective_date must be in YYYY-MM-DD format"
	}
	return len(problems) == 0, problems
}

// AdjustAssetWorth adjusts item worth by directional amount.
//
// @Summary Adjust asset worth
// @Description Applies increase/decrease delta using non-future effective date.
// @Tags assets
// @Accept json
// @Produce json
// @param asset_id path string true "Asset ID"
// @Param request body AdjustAssetWorthRequest true "Adjust worth payload"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id}/adjust [put]
func AdjustAssetWorth(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		// get asset_id from path and validate
		assetID, err := uuid.Parse(r.PathValue("asset_id"))
		if err != nil || assetID == uuid.Nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{
				"asset_id": "asset_id must be a valid UUID",
			})
			return
		}
		req, err := httpx.JSONDecode[AdjustAssetWorthRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		// isValid has already parsed both, so neither error can fire here. The layout was
		// written as "2026-01-02" -- a date rather than Go's reference layout -- so it never
		// matched and every adjustment was stored with the zero time; time.DateOnly is the
		// stdlib constant for this format and cannot be mistyped into an invalid layout.
		amount, _ := money.ParsePrice(req.Amount)
		effectiveDate, _ := time.Parse(time.DateOnly, req.EffectiveDate)
		changeDir := assets.ChangeDirectionIncrease
		if req.Direction == string(assets.ChangeDirectionDecrease) {
			changeDir = assets.ChangeDirectionDecrease
		}

		err = commands.UpdateAssetWorth(
			r.Context(),
			accountID,
			assetID,
			amount,
			assets.ChangeTypeAdjust,
			&changeDir,
			effectiveDate,
			req.Note,
		)
		switch {
		// We use the same not found response for both asset and class not
		// found to avoid leaking existence of asset IDs.
		case errors.Is(err, assets.ErrAssetNotFound) || errors.Is(err, assets.ErrClassAccountMismatch):
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
			return
		case errors.Is(err, assets.ErrClassReserved):
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "asset class is reserved and cannot be modified"})
			return
		case err != nil:
			log.Error(r.Context(), "adjust asset worth: failed to update asset worth", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update asset worth"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
