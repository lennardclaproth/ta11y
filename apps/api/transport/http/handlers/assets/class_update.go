package assets

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

type UpdateClassRequest struct {
	ID       uuid.UUID `json:"id"`
	Name     *string   `json:"name,omitempty"`
	Archived *bool     `json:"archived,omitempty"`
}

func (r UpdateClassRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)

	if r.Name != nil {
		lenName := len(strings.TrimSpace(*r.Name))
		if lenName < 1 || lenName > 255 {
			problems["name"] = "name must be between 1 and 255 characters long"
		}
	}
	return len(problems) == 0, problems
}

// UpdateClass updates mutable class fields.
//
// @Summary Update class
// @Description Updates manual class name and archived state.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body UpdateClassRequest true "Update asset class payload"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/classes [patch]
func UpdateClass(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		req, err := httpx.JSONDecode[UpdateClassRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		// The name-length rule was written but never applied: isValid had no caller, so any
		// name -- empty, or far past the column's limit -- reached the command untouched.
		if isValid, problems := req.isValid(); !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		switch err = commands.UpdateClass(r.Context(), accountID, req.ID, req.Name, req.Archived); {
		case errors.Is(err, assets.ErrClassAccountMismatch), errors.Is(err, assets.ErrClassNotFound):
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "class not found"})
		case err != nil:
			log.Error(r.Context(), "An error occurred while updating a class", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update class"})
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
}
