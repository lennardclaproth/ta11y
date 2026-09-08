package marketstack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/lennardclaproth/ta11y/internal/marketdata"
)

var _ marketdata.TickerSearcher = (*MarketStackClient)(nil)

// tickersPath is the v2 catalogue endpoint. v1 called it "tickers"; on v2 that name
// 404s, so the two are not interchangeable.
const tickersPath = "tickerslist"

// marketstackTickersResponse mirrors the /v2/tickerslist envelope. Pagination.Total is
// the provider's full match count, which is routinely far larger than one page.
// Name is a plain string because the provider returns null for some symbols and
// JSON null decodes to the empty string, which the catalogue already treats as
// "unnamed" rather than distinguishing it from an empty name.
type marketstackTickersResponse struct {
	Pagination struct {
		Limit  int `json:"limit"`
		Total  int `json:"total"`
		Offset int `json:"offset"`
		Count  int `json:"count"`
	} `json:"pagination"`
	Data []struct {
		Name          string `json:"name"`
		Ticker        string `json:"ticker"`
		HasIntraday   bool   `json:"has_intraday"`
		HasEOD        bool   `json:"has_eod"`
		StockExchange struct {
			Name    string `json:"name"`
			Acronym string `json:"acronym"`
			MIC     string `json:"mic"`
		} `json:"stock_exchange"`
	} `json:"data"`
}

// SearchTickers queries the MarketStack ticker catalogue, following the format:
//
// 'https://api.marketstack.com/v2/tickerslist?search=asml&limit=100&offset=0&access_key=XXXX'
//
// The v2 catalogue endpoint is 'tickerslist'; 'tickers' is the v1 name and answers
// 404 "Route not found" on v2, which surfaced as a 500 from catalogue search.
//
// An empty query returns the provider's catalogue in popularity order, which is
// what the bounded seed sync pages through. No exchange filter is sent, so results
// span every exchange and the MIC on each row is what distinguishes the variants
// of one company (ASML on XNAS versus ASML.XAMS on Euronext Amsterdam).
//
// One call costs one provider request and is metered as such.
func (c *MarketStackClient) SearchTickers(
	ctx context.Context,
	query string,
	limit, offset int,
) (marketdata.TickerPage, error) {
	req, err := c.constructTickersRequest(ctx, query, limit, offset)
	if err != nil {
		return marketdata.TickerPage{}, fmt.Errorf("SearchTickers failed to construct request: %w", err)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return marketdata.TickerPage{}, fmt.Errorf("SearchTickers failed to execute request: %w", err)
	}
	bodyBytes, err := io.ReadAll(res.Body)
	closeErr := res.Body.Close()
	if err != nil {
		return marketdata.TickerPage{}, fmt.Errorf("SearchTickers failed to read response body: %w", err)
	}
	if closeErr != nil {
		return marketdata.TickerPage{}, fmt.Errorf("SearchTickers failed to close response body: %w", closeErr)
	}
	if res.StatusCode >= 300 {
		return marketdata.TickerPage{}, fmt.Errorf("SearchTickers received status %d: %s", res.StatusCode, string(bodyBytes))
	}

	var parsed marketstackTickersResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return marketdata.TickerPage{}, fmt.Errorf("SearchTickers failed to unmarshal response: %w", err)
	}

	// The request is billed whether or not the caller finds what it wanted.
	if err := c.ps.DeductTokens(ctx, c.providerName, 1); err != nil {
		return marketdata.TickerPage{}, fmt.Errorf("SearchTickers failed to deduct tokens: %w", err)
	}

	entries := make([]marketdata.ProviderListing, 0, len(parsed.Data))
	for _, d := range parsed.Data {
		symbol := strings.TrimSpace(d.Ticker)
		if symbol == "" {
			continue // a row without a symbol is not addressable
		}
		entries = append(entries, marketdata.ProviderListing{
			Source:       marketdata.SourceMarketStack,
			Symbol:       symbol,
			Name:         optional(d.Name),
			ExchangeName: optional(d.StockExchange.Name),
			ExchangeMIC:  optional(d.StockExchange.MIC),
			HasEOD:       d.HasEOD,
			HasIntraday:  d.HasIntraday,
		})
	}

	return marketdata.TickerPage{
		Data:   entries,
		Limit:  parsed.Pagination.Limit,
		Offset: parsed.Pagination.Offset,
		Total:  parsed.Pagination.Total,
	}, nil
}

// constructTickersRequest builds the HTTP request for searching MarketStack tickers.
func (c *MarketStackClient) constructTickersRequest(
	ctx context.Context,
	query string,
	limit, offset int,
) (*http.Request, error) {
	provider, err := c.ps.GetByName(ctx, c.providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch provider %s: %w", c.providerName, err)
	}
	if provider == nil {
		return nil, fmt.Errorf("provider %s not found", c.providerName)
	}
	if !provider.IsAPIIngestion() {
		return nil, fmt.Errorf("provider %s is not API ingestion mode", c.providerName)
	}
	if provider.BaseURI == nil || *provider.BaseURI == "" {
		return nil, fmt.Errorf("provider %s base URI is empty", c.providerName)
	}
	if provider.ApiKey == nil || *provider.ApiKey == "" {
		return nil, fmt.Errorf("provider %s api key is empty", c.providerName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *provider.BaseURI+"/"+tickersPath, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("search", query)
	q.Add("limit", strconv.Itoa(limit))
	q.Add("offset", strconv.Itoa(offset))
	q.Add("access_key", *provider.ApiKey)
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	return req, nil
}

// optional maps a blank provider string to nil so absent metadata stays absent
// rather than being stored as an empty value.
func optional(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
