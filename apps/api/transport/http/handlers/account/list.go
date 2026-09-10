package account

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// AccountResponse is a single account returned by the list endpoint.
type AccountResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Admin marks accounts the API admits to admin-only routes; it answers 403 to the rest.
	Admin      bool    `json:"admin"`
	ExternalID *string `json:"external_id,omitempty"`
}

// List returns all accounts.
//
// @Summary List accounts
// @Description Returns all accounts.
// @Tags accounts
// @Produce json
// @Success 200 {array} AccountResponse
// @Failure 500 {object} map[string]string
// @Router /accounts [get]
func List(log logging.Logger, queries *account.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accounts, err := queries.List(r.Context())
		if err != nil {
			log.Error(r.Context(), "list accounts: failed to list accounts", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to list accounts"})
			return
		}

		res := make([]AccountResponse, 0, len(accounts))
		for _, acc := range accounts {
			res = append(res, AccountResponse{
				ID:         acc.ID.String(),
				Name:       acc.Name,
				Admin:      acc.Admin,
				ExternalID: acc.ExternalID,
			})
		}
		_ = httpx.JSONEncode(w, http.StatusOK, res)
	})
}
