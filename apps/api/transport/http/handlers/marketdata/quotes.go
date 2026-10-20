package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// SearchQuotesRequest filters the daily-priced instrument catalogue.
type SearchQuotesRequest struct {
	Q     string `query:"q"`
	Limit int    `query:"limit"`
}

func (r SearchQuotesRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.Limit < 0 {
		problems["limit"] = "limit cannot be negative"
	}
	return len(problems) == 0, problems
}

// QuoteResponse is one instrument the holdings picker can offer.
type QuoteResponse struct {
	Symbol    string     `json:"symbol"`
	Name      string     `json:"name"`
	Kind      string     `json:"kind"`
	Currency  string     `json:"currency"`
	ListingID *uuid.UUID `json:"listing_id,omitempty"`
	// Price and PriceDate are only present for an instrument someone already
	// tracks; an untracked one has no price history yet.
	Price     *string `json:"price,omitempty"`
	PriceDate *string `json:"price_date,omitempty"`
	// Selectable is false for an instrument no provider quotes in euro. Reason
	// then carries the explanation the picker shows on the blocked row.
	Selectable bool   `json:"selectable"`
	Reason     string `json:"reason,omitempty"`
}

// SearchQuotes lists the instruments that can back a daily-priced holding.
//
// @Summary Search daily-priced instruments
// @Description Returns the instruments an asset item can be linked to for daily pricing, selectable ones first. An instrument no configured provider quotes in euro is returned unselectable with the reason attached rather than omitted. This never calls a provider.
// @Tags marketdata
// @Accept json
// @Produce json
// @Param q query string false "Case-insensitive partial query over symbol and name"
// @Param limit query int false "Page size (max 50, default 50)"
// @Success 200 {array} QuoteResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/quotes [get]
func SearchQuotes(log logging.Logger, quotes *marketdata.Quotes) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[SearchQuotesRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		results, err := quotes.Search(r.Context(), req.Q, req.Limit)
		if err != nil {
			log.Error(r.Context(), "search quotes: failed to search quotes", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to search quotes"})
			return
		}

		rows := make([]QuoteResponse, 0, len(results))
		for _, quote := range results {
			rows = append(rows, toQuoteResponse(quote))
		}
		_ = httpx.JSONEncode(w, http.StatusOK, rows)
	})
}

func toQuoteResponse(quote marketdata.Quote) QuoteResponse {
	row := QuoteResponse{
		Symbol:     quote.Symbol,
		Name:       quote.Name,
		Kind:       string(quote.Kind),
		Currency:   string(quote.Currency),
		ListingID:  quote.ListingID,
		Selectable: quote.Selectable,
		Reason:     quote.Reason,
	}
	if quote.Price != nil {
		price := quote.Price.String()
		row.Price = &price
	}
	if quote.PriceDate != nil {
		day := quote.PriceDate.Format(time.DateOnly)
		row.PriceDate = &day
	}
	return row
}
