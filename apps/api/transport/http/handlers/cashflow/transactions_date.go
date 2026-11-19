package cashflow

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// ChangeTransactionDateRequest moves one manual transaction to another date.
type ChangeTransactionDateRequest struct {
	ID   uuid.UUID `json:"id"`
	Date string    `json:"date"`
}

func (r ChangeTransactionDateRequest) isValid() map[string]string {
	problems := make(map[string]string)
	if r.ID == uuid.Nil {
		problems["id"] = "id is required"
	}
	if strings.TrimSpace(r.Date) == "" {
		problems["date"] = "date is required"
	}
	return problems
}

// ChangeTransactionDateResponse returns the transaction on its new date.
type ChangeTransactionDateResponse struct {
	ID   uuid.UUID `json:"id"`
	Date time.Time `json:"date"`
}

// ChangeTransactionDate moves a manual cashflow transaction to another date.
//
// @Summary     Change a cashflow transaction date
// @Description Moves a manually entered cashflow transaction to another date, today or earlier. Imported transactions keep their statement date.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body ChangeTransactionDateRequest true "Change date request"
// @Success     200 {object} ChangeTransactionDateResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     409 {object} map[string]string "Conflict"
// @Failure     422 {object} map[string]string "Unprocessable entity"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/date [post]
// @Tags        Transactions
func ChangeTransactionDate(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, err := httpx.JSONDecode[ChangeTransactionDateRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		if problems := req.isValid(); len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		tx, err := commands.ChangeDate(r.Context(), accountID, req.ID, req.Date)
		if err != nil {
			writeChangeDateError(w, r, log, err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, ChangeTransactionDateResponse{ID: tx.ID, Date: tx.Date})
	})
}

func writeChangeDateError(w http.ResponseWriter, r *http.Request, log logging.Logger, err error) {
	switch {
	case errors.Is(err, cashflow.ErrNoTransactionFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"id": cashflow.ErrNoTransactionFound.Error()})
	case errors.Is(err, cashflow.ErrCashflowDateNotEditable):
		_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"transaction": cashflow.ErrCashflowDateNotEditable.Error()})
	case errors.Is(err, cashflow.ErrCashflowDateInFuture):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"date": cashflow.ErrCashflowDateInFuture.Error()})
	case errors.Is(err, cashflow.ErrManualCashflowInvalidDate):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"date": cashflow.ErrManualCashflowInvalidDate.Error()})
	case errors.Is(err, cashflow.ErrDuplicateTransaction):
		_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"transaction": "duplicate transaction"})
	default:
		log.Error(r.Context(), "cashflow change date: failed to move transaction", err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to change the transaction date"})
	}
}
