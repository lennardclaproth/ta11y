package cashflow

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// GetRecurringItemsRequest narrows the overview by item name.
type GetRecurringItemsRequest struct {
	Q string `query:"q"`
}

// RecurringAmountPointResponse is one observed amount and the day it landed.
type RecurringAmountPointResponse struct {
	Date        time.Time `json:"date"`
	AmountCents int64     `json:"amountCents"`
}

// RecurringItemResponse is one recurring item as the overview reads it.
type RecurringItemResponse struct {
	ID              uuid.UUID                      `json:"id"`
	Name            string                         `json:"name"`
	Direction       string                         `json:"direction"`
	Rhythm          string                         `json:"rhythm"`
	LastAmountCents int64                          `json:"lastAmountCents"`
	History         []RecurringAmountPointResponse `json:"history"`
	LastSeen        *time.Time                     `json:"last_seen"`
	NextExpected    *time.Time                     `json:"next_expected"`
	LinkedCount     int                            `json:"linked_count"`
	EndedFrom       *string                        `json:"ended_from"`
}

// RecurringMonthPointResponse is the monthly-equivalent total of one month.
type RecurringMonthPointResponse struct {
	Month        time.Time `json:"month"`
	ExpenseCents int64     `json:"expenseCents"`
	IncomeCents  int64     `json:"incomeCents"`
}

// RecurringOverviewResponse returns the whole recurring page in one read.
type RecurringOverviewResponse struct {
	Expenses []RecurringItemResponse `json:"expenses"`
	Income   []RecurringItemResponse `json:"income"`
	Ended    []RecurringItemResponse `json:"ended"`
	// Monthly totals spread a quarterly amount over three months and a yearly one
	// over twelve; the page says so where it shows them.
	MonthlyExpenseCents int64                         `json:"monthlyExpenseCents"`
	MonthlyIncomeCents  int64                         `json:"monthlyIncomeCents"`
	Series              []RecurringMonthPointResponse `json:"series"`
}

// RecurringTransactionResponse is one transaction linked to an item.
type RecurringTransactionResponse struct {
	ID          uuid.UUID `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	AmountCents int64     `json:"amountCents"`
	Source      string    `json:"source"`
}

// RecurringItemDetailResponse is one item with the transactions behind it.
type RecurringItemDetailResponse struct {
	RecurringItemResponse
	Transactions []RecurringTransactionResponse `json:"transactions"`
}

// RecurringSuggestionResponse is a pattern found in the account's history.
type RecurringSuggestionResponse struct {
	MatchKey    string    `json:"match_key"`
	Name        string    `json:"name"`
	Direction   string    `json:"direction"`
	Rhythm      string    `json:"rhythm"`
	AmountCents int64     `json:"amountCents"`
	Matches     int       `json:"matches"`
	Since       time.Time `json:"since"`
	Sample      string    `json:"sample"`
}

// RecurringSuggestionsResponse returns the patterns awaiting a decision.
type RecurringSuggestionsResponse struct {
	Data []RecurringSuggestionResponse `json:"data"`
}

// GetRecurringItems returns the account's recurring items grouped into running
// expenses, running income and ended items, with the monthly totals and trend.
//
// @Summary     Get recurring items
// @Description Returns running recurring expenses, running recurring income and ended items, with monthly-equivalent totals and a per-month trend.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       q query string false "Narrow the groups by item name"
// @Success     200 {object} RecurringOverviewResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring [get]
func GetRecurringItems(log logging.Logger, queries *cashflow.RecurringQueries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		req, err := httpx.DecodeQuery[GetRecurringItemsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}

		overview, err := queries.Overview(r.Context(), accountID, req.Q)
		if err != nil {
			log.Error(r.Context(), "recurring overview: failed to read items", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get recurring items"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, RecurringOverviewResponse{
			Expenses:            toRecurringItemResponses(overview.Expenses),
			Income:              toRecurringItemResponses(overview.Income),
			Ended:               toRecurringItemResponses(overview.Ended),
			MonthlyExpenseCents: overview.MonthlyExpenseCents,
			MonthlyIncomeCents:  overview.MonthlyIncomeCents,
			Series:              toRecurringMonthResponses(overview.Series),
		})
	})
}

// GetRecurringItem returns one recurring item with its amount history and the
// transactions linked to it.
//
// @Summary     Get one recurring item
// @Description Returns a recurring item with its observed amounts over time and the transactions linked to it, newest first.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Param       item_id path string true "Recurring item id"
// @Success     200 {object} RecurringItemDetailResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/{item_id} [get]
func GetRecurringItem(log logging.Logger, queries *cashflow.RecurringQueries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		itemID, err := uuid.Parse(r.PathValue("item_id"))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"item_id": "item_id must be a valid UUID"})
			return
		}

		detail, err := queries.Item(r.Context(), accountID, itemID)
		if err != nil {
			if errors.Is(err, cashflow.ErrRecurringItemNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"item_id": cashflow.ErrRecurringItemNotFound.Error()})
				return
			}
			log.Error(r.Context(), "recurring item: failed to read item", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get the recurring item"})
			return
		}

		transactions := make([]RecurringTransactionResponse, 0, len(detail.Transactions))
		for _, link := range detail.Transactions {
			transactions = append(transactions, RecurringTransactionResponse{
				ID:          link.TransactionID,
				Date:        link.Date,
				Description: link.Description,
				AmountCents: int64(link.AmountCents),
				Source:      link.Source,
			})
		}

		_ = httpx.JSONEncode(w, http.StatusOK, RecurringItemDetailResponse{
			RecurringItemResponse: toRecurringItemResponse(detail.RecurringItemView),
			Transactions:          transactions,
		})
	})
}

// GetRecurringSuggestions returns the recurring patterns found in the account's
// history. A suggestion is a proposal only: nothing is added until it is confirmed.
//
// @Summary     Get recurring suggestions
// @Description Returns patterns found in transactions that are neither ignored nor already linked, leaving out dismissed patterns and patterns that already belong to an item.
// @Tags        Recurring
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} RecurringSuggestionsResponse
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /cashflow/recurring/suggestions [get]
func GetRecurringSuggestions(log logging.Logger, queries *cashflow.RecurringQueries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}

		suggestions, err := queries.Suggestions(r.Context(), accountID)
		if err != nil {
			log.Error(r.Context(), "recurring suggestions: failed to read history", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get recurring suggestions"})
			return
		}

		data := make([]RecurringSuggestionResponse, 0, len(suggestions))
		for _, suggestion := range suggestions {
			data = append(data, RecurringSuggestionResponse{
				MatchKey:    suggestion.MatchKey,
				Name:        suggestion.Name,
				Direction:   string(suggestion.Direction),
				Rhythm:      string(suggestion.Rhythm),
				AmountCents: int64(suggestion.AmountCents),
				Matches:     suggestion.Matches,
				Since:       suggestion.Since,
				Sample:      suggestion.Sample,
			})
		}

		_ = httpx.JSONEncode(w, http.StatusOK, RecurringSuggestionsResponse{Data: data})
	})
}

func toRecurringItemResponses(views []cashflow.RecurringItemView) []RecurringItemResponse {
	out := make([]RecurringItemResponse, 0, len(views))
	for _, view := range views {
		out = append(out, toRecurringItemResponse(view))
	}
	return out
}

func toRecurringItemResponse(view cashflow.RecurringItemView) RecurringItemResponse {
	history := make([]RecurringAmountPointResponse, 0, len(view.History))
	for _, point := range view.History {
		history = append(history, RecurringAmountPointResponse{
			Date:        point.Date,
			AmountCents: int64(point.AmountCents),
		})
	}

	var endedFrom *string
	if view.Item.EndedFrom != nil {
		month := view.Item.EndedFrom.Format("2006-01")
		endedFrom = &month
	}

	return RecurringItemResponse{
		ID:              view.Item.ID,
		Name:            view.Item.Name,
		Direction:       string(view.Item.Direction),
		Rhythm:          string(view.Item.Rhythm),
		LastAmountCents: int64(view.LastAmountCents),
		History:         history,
		LastSeen:        view.LastSeen,
		NextExpected:    view.NextExpected,
		LinkedCount:     view.LinkedCount,
		EndedFrom:       endedFrom,
	}
}

func toRecurringMonthResponses(points []cashflow.RecurringMonthPoint) []RecurringMonthPointResponse {
	out := make([]RecurringMonthPointResponse, 0, len(points))
	for _, point := range points {
		out = append(out, RecurringMonthPointResponse{
			Month:        point.Month,
			ExpenseCents: point.ExpenseCents,
			IncomeCents:  point.IncomeCents,
		})
	}
	return out
}
