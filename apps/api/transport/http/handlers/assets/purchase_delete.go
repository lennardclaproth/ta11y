package assets

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// DeletePurchase removes one purchase of a daily-priced item.
//
// @Summary Delete purchase
// @Description Removes a purchase and rebuilds the item's worth history without it. An item keeps at least one purchase; deleting its last one answers 409, and the item itself is deleted instead.
// @Tags assets
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Param purchase_id path string true "Purchase ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id}/purchases/{purchase_id} [delete]
func DeletePurchase(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		assetID, ok := pathID(w, r, "asset_id")
		if !ok {
			return
		}
		purchaseID, ok := pathID(w, r, "purchase_id")
		if !ok {
			return
		}
		err := commands.DeletePurchase(r.Context(), accountID, assetID, purchaseID)
		if writeEditError(w, r, log, "delete purchase", err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
