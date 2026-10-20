package assets

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// DeleteMutation removes one worth change of an item whose worth is set by hand.
//
// @Summary Delete worth change
// @Description Removes a worth change and replays the class history without it. An item keeps at least one change; deleting its last one answers 409, and the item itself is deleted instead.
// @Tags assets
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Param mutation_id path string true "Worth change ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id}/mutations/{mutation_id} [delete]
func DeleteMutation(log logging.Logger, commands assets.Commands) http.Handler {
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
		err := commands.DeleteMutation(r.Context(), accountID, assetID, mutationID)
		if writeEditError(w, r, log, "delete worth change", err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
