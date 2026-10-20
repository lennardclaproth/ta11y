package assets

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// UpdatePurchase corrects one purchase of a daily-priced item.
//
// @Summary Edit purchase
// @Description Replaces the date, quantity and unit price of a purchase. The item's worth history is rebuilt from its purchases afterwards.
// @Tags assets
// @Accept json
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Param purchase_id path string true "Purchase ID"
// @Param request body PurchaseRequest true "Corrected purchase"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id}/purchases/{purchase_id} [put]
func UpdatePurchase(log logging.Logger, commands assets.Commands) http.Handler {
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
		req, err := httpx.JSONDecode[PurchaseRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		problems := make(map[string]string)
		req.validate(problems, "")
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}
		err = commands.UpdatePurchase(r.Context(), accountID, assetID, purchaseID, req.toInput())
		if writeEditError(w, r, log, "update purchase", err) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
