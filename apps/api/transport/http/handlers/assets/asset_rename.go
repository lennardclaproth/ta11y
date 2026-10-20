package assets

import (
	"net/http"
	"strings"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// RenameAssetRequest is an item's new name.
type RenameAssetRequest struct {
	Name string `json:"name"`
}

func (r RenameAssetRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	name := strings.TrimSpace(r.Name)
	if name == "" {
		problems["name"] = "name is required"
	} else if len(name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}
	return len(problems) == 0, problems
}

// RenameAsset renames an asset item.
//
// @Summary Rename asset item
// @Description Changes the name of an item in a manual class, daily-priced or not.
// @Tags assets
// @Accept json
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Param request body RenameAssetRequest true "New name"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id} [patch]
func RenameAsset(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		assetID, ok := pathID(w, r, "asset_id")
		if !ok {
			return
		}
		req, err := httpx.JSONDecode[RenameAssetRequest](r)
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
		if writeEditError(w, r, log, "rename asset", commands.RenameAsset(r.Context(), accountID, assetID, req.Name)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
