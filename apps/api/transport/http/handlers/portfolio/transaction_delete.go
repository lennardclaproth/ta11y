package portfolio

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/logging"
	portfoliodomain "github.com/lennardclaproth/ta11y/internal/portfolio"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// DeletePortfolioTransactionResponse confirms the removal and reports the rebuild.
type DeletePortfolioTransactionResponse struct {
	ID uuid.UUID `json:"id"`
	// Rebuild reports the rebuild this write scheduled:
	// "started" (recalculating in the background; a portfolio.rebuilt websocket event
	// follows when done) or "skipped" (nothing was scheduled).
	Rebuild string `json:"rebuild"`
}

// DeletePortfolioTransaction removes a manual portfolio transaction and rebuilds the portfolio.
//
// @Summary Delete a manual portfolio transaction
// @Description Removes a manually entered portfolio transaction, then rebuilds positions and snapshots in the background. Imported transactions are refused, since the next upload of the same statement would bring them back.
// @Tags portfolio
// @Produce json
// @Param transaction_id path string true "Transaction ID"
// @Success 200 {object} DeletePortfolioTransactionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /portfolio/transactions/{transaction_id} [delete]
func DeletePortfolioTransaction(
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
		if commands == nil {
			log.Error(r.Context(), "portfolio delete transaction: commands are not configured", errors.New("portfolio commands not configured"))
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete the transaction"})
			return
		}

		rebuild, err := commands.DeleteTransaction(r.Context(), accountID, id)
		if err != nil {
			writeEditPortfolioTransactionError(w, r, log, err, "failed to delete the transaction")
			return
		}
		_ = httpx.JSONEncode(w, http.StatusOK, DeletePortfolioTransactionResponse{
			ID:      id,
			Rebuild: string(rebuild.Outcome),
		})
	})
}
