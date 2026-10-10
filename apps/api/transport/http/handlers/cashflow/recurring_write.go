package cashflow

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// CreateRecurringItemRequest starts a recurring item from the transactions the
// user pointed at on Cashflow.
type CreateRecurringItemRequest struct {
	Name      string      `json:"name"`
	Direction string      `json:"direction"`
	Rhythm    string      `json:"rhythm"`
	IDs       []uuid.UUID `json:"ids"`
}

// UpdateRecurringItemRequest renames an item or changes how often it is expected.
type UpdateRecurringItemRequest struct {
	Name   string `json:"name"`
	Rhythm string `json:"rhythm"`
}

// LinkRecurringTransactionsRequest adds transactions to an existing item.
type LinkRecurringTransactionsRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

// EndRecurringItemRequest stops an item from a month ("YYYY-MM").
type EndRecurringItemRequest struct {
	From string `json:"from"`
}

// ConfirmRecurringSuggestionRequest turns a suggested pattern into a real item.
type ConfirmRecurringSuggestionRequest struct {
	MatchKey  string `json:"match_key"`
	Name      string `json:"name"`
	Direction string `json:"direction"`
	Rhythm    string `json:"rhythm"`
}

// DismissRecurringSuggestionRequest refuses a suggested pattern for good.
type DismissRecurringSuggestionRequest struct {
	MatchKey  string `json:"match_key"`
	Direction string `json:"direction"`
}

// RecurringMutationResponse reports how many transactions a mutation touched.
type RecurringMutationResponse struct {
	LinkedCount int `json:"linked_count"`
}

// CreateRecurringItem starts a recurring item and links the selected transactions.
//
// @Summary     Create a recurring item
// @Description Starts a recurring item from one or more cashflow transactions. The tag and ignored status of those transactions are left untouched.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       payload body CreateRecurringItemRequest true "Create recurring item request"
// @Success     201 {object} RecurringItemResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     409 {object} map[string]string "Conflict"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring [post]
func CreateRecurringItem(log logging.Logger, commands *cashflow.RecurringCommands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, ok := decodeRecurring[CreateRecurringItemRequest](w, r)
		if !ok {
			return
		}

		direction, rhythm, problems := parseRecurringShape(req.Direction, req.Rhythm)
		if strings.TrimSpace(req.Name) == "" {
			problems["name"] = cashflow.ErrRecurringNameRequired.Error()
		}
		if len(req.IDs) == 0 {
			problems["ids"] = cashflow.ErrRecurringTransactionsRequired.Error()
		}
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		item, err := commands.Create(r.Context(), accountID, cashflow.CreateRecurringInput{
			Name:           req.Name,
			Direction:      direction,
			Rhythm:         rhythm,
			TransactionIDs: req.IDs,
		})
		if err != nil {
			writeRecurringError(w, r, log, "failed to create the recurring item", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusCreated, toRecurringItemResponse(cashflow.RecurringItemView{Item: item}))
	})
}

// UpdateRecurringItem renames an item and sets how often it is expected.
//
// @Summary     Update a recurring item
// @Description Renames a recurring item and sets its rhythm. The direction follows from the linked transactions and is not editable.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       item_id path string true "Recurring item id"
// @Param       payload body UpdateRecurringItemRequest true "Update recurring item request"
// @Success     200 {object} RecurringItemResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     409 {object} map[string]string "Conflict"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/{item_id} [patch]
func UpdateRecurringItem(log logging.Logger, commands *cashflow.RecurringCommands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, itemID, ok := recurringScope(w, r)
		if !ok {
			return
		}

		req, ok := decodeRecurring[UpdateRecurringItemRequest](w, r)
		if !ok {
			return
		}

		rhythm, err := cashflow.ParseRhythm(req.Rhythm)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"rhythm": err.Error()})
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"name": cashflow.ErrRecurringNameRequired.Error()})
			return
		}

		item, err := commands.Update(r.Context(), accountID, itemID, cashflow.UpdateRecurringInput{
			Name:   req.Name,
			Rhythm: rhythm,
		})
		if err != nil {
			writeRecurringError(w, r, log, "failed to update the recurring item", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toRecurringItemResponse(cashflow.RecurringItemView{Item: item}))
	})
}

// LinkRecurringTransactions adds transactions to an existing recurring item.
//
// @Summary     Link transactions to a recurring item
// @Description Links cashflow transactions to an existing recurring item. An ended item takes no new transactions.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       item_id path string true "Recurring item id"
// @Param       payload body LinkRecurringTransactionsRequest true "Link transactions request"
// @Success     200 {object} RecurringMutationResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     422 {object} map[string]string "Unprocessable entity"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/{item_id}/transactions [post]
func LinkRecurringTransactions(log logging.Logger, commands *cashflow.RecurringCommands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, itemID, ok := recurringScope(w, r)
		if !ok {
			return
		}

		req, ok := decodeRecurring[LinkRecurringTransactionsRequest](w, r)
		if !ok {
			return
		}
		if len(req.IDs) == 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"ids": cashflow.ErrRecurringTransactionsRequired.Error()})
			return
		}

		linked, err := commands.AddTransactions(r.Context(), accountID, itemID, req.IDs)
		if err != nil {
			writeRecurringError(w, r, log, "failed to link the transactions", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, RecurringMutationResponse{LinkedCount: linked})
	})
}

// UnlinkRecurringTransaction detaches one transaction from a recurring item.
//
// @Summary     Unlink a transaction from a recurring item
// @Description Removes one transaction from a recurring item. The transaction itself stays in Cashflow.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       item_id path string true "Recurring item id"
// @Param       transaction_id path string true "Cashflow transaction id"
// @Success     200 {object} map[string]string "OK"
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/{item_id}/transactions/{transaction_id} [delete]
func UnlinkRecurringTransaction(log logging.Logger, commands *cashflow.RecurringCommands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, itemID, ok := recurringScope(w, r)
		if !ok {
			return
		}

		transactionID, err := uuid.Parse(r.PathValue("transaction_id"))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"transaction_id": "transaction_id must be a valid UUID"})
			return
		}

		if err := commands.RemoveTransaction(r.Context(), accountID, itemID, transactionID); err != nil {
			writeRecurringError(w, r, log, "failed to unlink the transaction", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, struct{}{})
	})
}

// EndRecurringItem stops a recurring item from a chosen month.
//
// @Summary     End a recurring item
// @Description Stops a recurring item from the first day of the given month ("YYYY-MM"). Transactions linked before it stay linked; nothing after it is linked automatically.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       item_id path string true "Recurring item id"
// @Param       payload body EndRecurringItemRequest true "End recurring item request"
// @Success     200 {object} RecurringItemResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/{item_id}/end [post]
func EndRecurringItem(log logging.Logger, commands *cashflow.RecurringCommands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, itemID, ok := recurringScope(w, r)
		if !ok {
			return
		}

		req, ok := decodeRecurring[EndRecurringItemRequest](w, r)
		if !ok {
			return
		}

		item, err := commands.End(r.Context(), accountID, itemID, req.From)
		if err != nil {
			writeRecurringError(w, r, log, "failed to end the recurring item", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toRecurringItemResponse(cashflow.RecurringItemView{Item: item}))
	})
}

// ConfirmRecurringSuggestion turns a suggested pattern into a confirmed item.
//
// @Summary     Confirm a recurring suggestion
// @Description Creates the recurring item a suggestion proposed and links every transaction it was found in. A suggestion counts for nothing until this is called.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       payload body ConfirmRecurringSuggestionRequest true "Confirm suggestion request"
// @Success     201 {object} RecurringItemResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     409 {object} map[string]string "Conflict"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/suggestions/confirm [post]
func ConfirmRecurringSuggestion(log logging.Logger, commands *cashflow.RecurringCommands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, ok := decodeRecurring[ConfirmRecurringSuggestionRequest](w, r)
		if !ok {
			return
		}

		direction, rhythm, problems := parseRecurringShape(req.Direction, req.Rhythm)
		if strings.TrimSpace(req.MatchKey) == "" {
			problems["match_key"] = "match_key is required"
		}
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		item, err := commands.ConfirmSuggestion(r.Context(), accountID, cashflow.ConfirmSuggestionInput{
			MatchKey:  req.MatchKey,
			Name:      req.Name,
			Direction: direction,
			Rhythm:    rhythm,
		})
		if err != nil {
			writeRecurringError(w, r, log, "failed to confirm the suggestion", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusCreated, toRecurringItemResponse(cashflow.RecurringItemView{Item: item}))
	})
}

// DismissRecurringSuggestion refuses a suggested pattern for good.
//
// @Summary     Dismiss a recurring suggestion
// @Description Refuses a suggested pattern. A dismissed suggestion is not offered again.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       payload body DismissRecurringSuggestionRequest true "Dismiss suggestion request"
// @Success     200 {object} map[string]string "OK"
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/suggestions/dismiss [post]
func DismissRecurringSuggestion(log logging.Logger, commands *cashflow.RecurringCommands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, ok := decodeRecurring[DismissRecurringSuggestionRequest](w, r)
		if !ok {
			return
		}

		direction, err := cashflow.ParseDirection(req.Direction)
		if err != nil || direction == nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"direction": "direction must be either in or out"})
			return
		}

		if err := commands.DismissSuggestion(r.Context(), accountID, req.MatchKey, *direction); err != nil {
			writeRecurringError(w, r, log, "failed to dismiss the suggestion", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, struct{}{})
	})
}

// recurringScope resolves the account from the session and the item from the path.
func recurringScope(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	accountID, ok := httpx.AccountID(w, r)
	if !ok {
		return uuid.Nil, uuid.Nil, false
	}
	itemID, err := uuid.Parse(r.PathValue("item_id"))
	if err != nil {
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"item_id": "item_id must be a valid UUID"})
		return uuid.Nil, uuid.Nil, false
	}
	return accountID, itemID, true
}

func decodeRecurring[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	req, err := httpx.JSONDecode[T](r)
	if err != nil {
		if httpx.WriteDecodeError(w, err) {
			return req, false
		}
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		return req, false
	}
	return req, true
}

// parseRecurringShape validates the two vocabularies a recurring item is built on.
func parseRecurringShape(directionRaw, rhythmRaw string) (cashflow.CashFlowDirection, cashflow.Rhythm, map[string]string) {
	problems := make(map[string]string)

	var direction cashflow.CashFlowDirection
	parsed, err := cashflow.ParseDirection(directionRaw)
	if err != nil || parsed == nil {
		problems["direction"] = "direction must be either in or out"
	} else {
		direction = *parsed
	}

	rhythm, err := cashflow.ParseRhythm(rhythmRaw)
	if err != nil {
		problems["rhythm"] = err.Error()
	}

	return direction, rhythm, problems
}

// writeRecurringError maps the feature's refusals onto status codes; anything else
// is logged and reported as a server error.
func writeRecurringError(w http.ResponseWriter, r *http.Request, log logging.Logger, message string, err error) {
	switch {
	case errors.Is(err, cashflow.ErrRecurringItemNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"item_id": cashflow.ErrRecurringItemNotFound.Error()})
	case errors.Is(err, cashflow.ErrRecurringTransactionNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"ids": cashflow.ErrRecurringTransactionNotFound.Error()})
	case errors.Is(err, cashflow.ErrRecurringSuggestionNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"match_key": cashflow.ErrRecurringSuggestionNotFound.Error()})
	case errors.Is(err, cashflow.ErrRecurringItemExists):
		_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"name": cashflow.ErrRecurringItemExists.Error()})
	case errors.Is(err, cashflow.ErrRecurringItemEnded):
		_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"item_id": cashflow.ErrRecurringItemEnded.Error()})
	case errors.Is(err, cashflow.ErrRecurringInvalidMonth):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"from": cashflow.ErrRecurringInvalidMonth.Error()})
	case errors.Is(err, cashflow.ErrRecurringNameRequired):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"name": cashflow.ErrRecurringNameRequired.Error()})
	case errors.Is(err, cashflow.ErrRecurringTransactionsRequired):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"ids": cashflow.ErrRecurringTransactionsRequired.Error()})
	default:
		log.Error(r.Context(), message, err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": message})
	}
}
