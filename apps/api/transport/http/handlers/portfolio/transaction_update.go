package portfolio

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/logging"
	portfoliodomain "github.com/lennardclaproth/ta11y/internal/portfolio"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// UpdatePortfolioTransactionRequest replaces what was entered for one manual transaction.
// The vendor it was entered under stays as it was.
type UpdatePortfolioTransactionRequest struct {
	OccurredAt  string     `json:"occurred_at"`
	Type        string     `json:"type"`
	ListingID   *uuid.UUID `json:"listing_id,omitempty"`
	Amount      string     `json:"amount"`
	Quantity    *string    `json:"quantity,omitempty"`
	Description *string    `json:"description,omitempty"`
}

func (r UpdatePortfolioTransactionRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.OccurredAt) == "" {
		problems["occurred_at"] = "occurred_at is required"
	}
	if strings.TrimSpace(r.Type) == "" {
		problems["type"] = "type is required"
	}
	if strings.TrimSpace(r.Amount) == "" {
		problems["amount"] = "amount is required"
	}
	return len(problems) == 0, problems
}

// UpdatePortfolioTransaction edits a manual portfolio transaction and rebuilds the portfolio.
//
// @Summary Edit a manual portfolio transaction
// @Description Replaces the date, type, listing, amount, quantity and description of a manually entered portfolio transaction under the same rules as creating one, then rebuilds positions and snapshots in the background. Imported transactions are refused. The response answers before the rebuild runs; a portfolio.rebuilt websocket event follows when it is done.
// @Tags portfolio
// @Accept json
// @Produce json
// @Param transaction_id path string true "Transaction ID"
// @Param request body UpdatePortfolioTransactionRequest true "Edited transaction"
// @Success 200 {object} ManualPortfolioTransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/transactions/{transaction_id} [put]
func UpdatePortfolioTransaction(
	log logging.Logger,
	commands *portfoliodomain.Commands,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		id, err := uuid.Parse(r.PathValue("transaction_id"))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"transaction_id": "transaction_id must be a uuid"})
			return
		}
		req, err := httpx.JSONDecode[UpdatePortfolioTransactionRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		if ok, problems := req.isValid(); !ok {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}
		if commands == nil {
			log.Error(r.Context(), "portfolio update transaction: commands are not configured", errors.New("portfolio commands not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update the transaction"})
			return
		}

		result, err := commands.UpdateTransaction(r.Context(), portfoliodomain.ManualTransactionUpdateInput{
			AccountID:   accountID,
			ID:          id,
			OccurredAt:  req.OccurredAt,
			Type:        req.Type,
			ListingID:   req.ListingID,
			Amount:      req.Amount,
			Quantity:    req.Quantity,
			Description: req.Description,
		})
		if err != nil {
			writeEditPortfolioTransactionError(w, r, log, err, "failed to update the transaction")
			return
		}
		_ = httpx.JSONEncode(w, http.StatusOK, toManualPortfolioTransactionResponse(result))
	})
}

// writeEditPortfolioTransactionError maps the refusals of an edit or delete to the field
// that caused them. failure is the generic message for anything unexpected.
func writeEditPortfolioTransactionError(w http.ResponseWriter, r *http.Request, log logging.Logger, err error, failure string) {
	switch {
	case errors.Is(err, portfoliodomain.ErrTransactionNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"transaction_id": portfoliodomain.ErrTransactionNotFound.Error()})
	case errors.Is(err, portfoliodomain.ErrTransactionNotEditable):
		_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"transaction": portfoliodomain.ErrTransactionNotEditable.Error()})
	case errors.Is(err, portfoliodomain.ErrManualListingNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"listing_id": portfoliodomain.ErrManualListingNotFound.Error()})
	case errors.Is(err, portfoliodomain.ErrDuplicateTransaction):
		_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"transaction": "duplicate transaction"})
	case isManualValidationErr(err):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"transaction": err.Error()})
	default:
		log.Error(r.Context(), "portfolio edit transaction: "+failure, err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": failure})
	}
}
