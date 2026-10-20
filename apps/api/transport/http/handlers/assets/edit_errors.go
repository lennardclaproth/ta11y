package assets

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// pathID reads a UUID path value, answering 400 itself when it is not one.
func pathID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil || id == uuid.Nil {
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{name: name + " must be a valid UUID"})
		return uuid.Nil, false
	}
	return id, true
}

// invalidEdits are the domain refusals of an edit's values, answered with their own words.
var invalidEdits = []error{
	assets.ErrMutationAmount,
	assets.ErrMutationDirection,
	assets.ErrMutationInFuture,
	assets.ErrOpeningNotSet,
	assets.ErrPurchaseQuantityInvalid,
	assets.ErrPurchaseUnitPriceInvalid,
	assets.ErrPurchaseDateInFuture,
}

// matching returns the first target err wraps, or nil.
func matching(err error, targets ...error) error {
	for _, target := range targets {
		if errors.Is(err, target) {
			return target
		}
	}
	return nil
}

// writeEditError maps the errors the item, worth-change and purchase edits share. It
// writes the response and reports whether there was an error to write.
func writeEditError(w http.ResponseWriter, r *http.Request, log logging.Logger, action string, err error) bool {
	switch {
	case err == nil:
		return false
	// The same not-found answer covers an unknown item and someone else's item, so a
	// request cannot probe which IDs exist.
	case errors.Is(err, assets.ErrAssetNotFound), errors.Is(err, assets.ErrClassAccountMismatch):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
	case errors.Is(err, assets.ErrMutationNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "worth change not found"})
	case errors.Is(err, assets.ErrPurchaseNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "purchase not found"})
	case errors.Is(err, assets.ErrClassReserved), errors.Is(err, assets.ErrClassNotManual):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "items in the portfolio class follow the portfolio and cannot be edited"})
	case errors.Is(err, assets.ErrAssetDailyPriced):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "this item follows a daily price; edit its purchases instead"})
	case errors.Is(err, assets.ErrAssetNotDailyPriced):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "this item has no purchases; edit its worth changes instead"})
	case matching(err, assets.ErrLastMutation, assets.ErrLastPurchase) != nil:
		_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{
			"error": matching(err, assets.ErrLastMutation, assets.ErrLastPurchase).Error(),
		})
	case errors.Is(err, assets.ErrAssetNameEmpty):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"name": "name is required"})
	case matching(err, invalidEdits...) != nil:
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": matching(err, invalidEdits...).Error()})
	default:
		log.Error(r.Context(), action+": unexpected failure", err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to " + action})
	}
	return true
}
