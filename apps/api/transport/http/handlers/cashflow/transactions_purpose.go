package cashflow

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// MarkPurposeBySelectionRequest points selected transactions at the monthly goal.
type MarkPurposeBySelectionRequest struct {
	Purpose string      `json:"purpose"`
	IDs     []uuid.UUID `json:"ids"`
}

// MarkPurposeByFilterRequest points every transaction matching the filters at the goal.
type MarkPurposeByFilterRequest struct {
	Purpose string             `json:"purpose"`
	Filters TransactionFilters `json:"filters"`
}

// MarkPurposeResponse reports the result of a purpose mutation. `matched_count` is what the
// caller pointed at and `updated_count` what the purpose could apply to: income only sticks
// to incoming money and a contribution only to outgoing money.
type MarkPurposeResponse struct {
	UpdatedCount int    `json:"updated_count"`
	MatchedCount int    `json:"matched_count"`
	Status       string `json:"status"`
}

func (r MarkPurposeBySelectionRequest) isValid() map[string]string {
	problems := make(map[string]string)
	if len(r.IDs) == 0 {
		problems["ids"] = "ids is required"
		return problems
	}
	for i, id := range r.IDs {
		if id == uuid.Nil {
			problems[fmt.Sprintf("ids[%d]", i)] = "id must be a valid UUID"
		}
	}
	return problems
}

// MarkPurposeBySelection sets what the selected transactions count as towards the goal.
//
// @Summary     Mark selected cashflow transactions for the monthly goal
// @Description Set purpose=income/wealth/none for selected transaction IDs. Income only applies to incoming money and wealth only to outgoing money; rows of the other direction are left unchanged.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body MarkPurposeBySelectionRequest true "Selection purpose request"
// @Success     200 {object} MarkPurposeResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/purpose/selection [post]
// @Tags        Transactions
func MarkPurposeBySelection(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, err := httpx.JSONDecode[MarkPurposeBySelectionRequest](r)
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

		purpose, err := cashflow.ParsePurpose(req.Purpose)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"purpose": err.Error()})
			return
		}

		result, err := commands.MarkPurposeByIDs(r.Context(), accountID, req.IDs, purpose)
		if err != nil {
			log.Error(r.Context(), "cashflow mark purpose selection: failed to update transactions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to mark selected cashflow transactions"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, markPurposeResponse(result, purpose))
	})
}

// MarkPurposeByFilter sets what every transaction matching the filter counts as.
//
// @Summary     Mark filtered cashflow transactions for the monthly goal
// @Description Set purpose=income/wealth/none for all transactions matching the supplied filter. Income only applies to incoming money and wealth only to outgoing money.
// @Accept      application/json
// @Produce     application/json
// @Param       payload body MarkPurposeByFilterRequest true "Filter purpose request"
// @Success     200 {object} MarkPurposeResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/transactions/purpose/filter [post]
// @Tags        Transactions
func MarkPurposeByFilter(log logging.Logger, commands *cashflow.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, err := httpx.JSONDecode[MarkPurposeByFilterRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}

		purpose, err := cashflow.ParsePurpose(req.Purpose)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"purpose": err.Error()})
			return
		}

		filters, problems := req.Filters.ToAppFilters()
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		result, err := commands.MarkPurposeByFilter(r.Context(), accountID, filters, purpose)
		if err != nil {
			log.Error(r.Context(), "cashflow mark purpose filter: failed to update transactions", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to mark filtered cashflow transactions"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, markPurposeResponse(result, purpose))
	})
}

func markPurposeResponse(result cashflow.MarkPurposeResult, purpose cashflow.Purpose) MarkPurposeResponse {
	return MarkPurposeResponse{
		UpdatedCount: result.Updated,
		MatchedCount: result.Matched,
		Status:       fmt.Sprintf("marked %d of %d transactions as %s", result.Updated, result.Matched, purposeStatusLabel(purpose)),
	}
}

func purposeStatusLabel(purpose cashflow.Purpose) string {
	switch purpose {
	case cashflow.PurposeIncome:
		return "income"
	case cashflow.PurposeWealth:
		return "a contribution towards wealth"
	default:
		return "not assigned"
	}
}
