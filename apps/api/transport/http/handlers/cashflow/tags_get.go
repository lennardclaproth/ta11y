package cashflow

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// CashflowTagResponse is one tag the account has used.
type CashflowTagResponse struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// CashflowTagsResponse lists the account's tags, most used first.
type CashflowTagsResponse struct {
	Data []CashflowTagResponse `json:"data"`
}

// GetCashflowTags lists every tag the account has used.
//
// @Summary     List cashflow tags
// @Description Returns every tag the account has put on a cashflow transaction, with the number of transactions carrying it, most used first. Not limited to a period; untagged transactions are not a tag and are left out.
// @Tags        Transactions
// @Produce     application/json
// @Success     200 {object} CashflowTagsResponse
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/tags [get]
func GetCashflowTags(log logging.Logger, queries *cashflow.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		tags, err := queries.Tags(r.Context(), accountID)
		if err != nil {
			log.Error(r.Context(), "cashflow tags: failed to list tags", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to list cashflow tags"})
			return
		}

		data := make([]CashflowTagResponse, 0, len(tags))
		for _, tag := range tags {
			data = append(data, CashflowTagResponse{Tag: tag.Tag, Count: tag.Count})
		}
		_ = httpx.JSONEncode(w, http.StatusOK, CashflowTagsResponse{Data: data})
	})
}
