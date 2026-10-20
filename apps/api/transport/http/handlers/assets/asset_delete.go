package assets

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// DeleteAsset removes an asset item.
//
// @Summary Delete asset item
// @Description Removes an item in a manual class together with its worth changes and purchases. The class history and net worth are rebuilt without it.
// @Tags assets
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id} [delete]
func DeleteAsset(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		assetID, ok := pathID(w, r, "asset_id")
		if !ok {
			return
		}
		if writeEditError(w, r, log, "delete asset", commands.DeleteAsset(r.Context(), accountID, assetID)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
