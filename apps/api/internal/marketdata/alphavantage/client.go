// Package alphavantage fetches daily crypto prices from Alpha Vantage.
//
// The provider itself already exists in the app (keys, quota, base URI); what was
// missing was a fetcher, so a listing whose source is alpha_vantage could never
// accumulate end-of-day data. This package supplies exactly that for digital
// currencies quoted in euro — the one instrument kind the provider delivers in a
// currency the app can use without conversion.
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"go.elastic.co/apm/module/apmhttp/v2"
)

// ProviderStore reads the provider record and books the request against its quota.
type ProviderStore interface {
	GetByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error)
	DeductTokens(ctx context.Context, name marketdata.ProviderName, count int32) error
}

// Client is the Alpha Vantage end-of-day fetcher for digital currencies.
type Client struct {
	http         *http.Client
	ps           ProviderStore
	providerName marketdata.ProviderName
}

var (
	_ marketdata.EODFetcher     = (*Client)(nil)
	_ marketdata.QuoteCatalogue = (*Client)(nil)
)

// NewClient constructs an Alpha Vantage client that books its requests against
// the given provider's quota.
func NewClient(ps ProviderStore, providerName marketdata.ProviderName) *Client {
	return &Client{
		http: &http.Client{
			Transport: apmhttp.WrapRoundTripper(http.DefaultTransport),
			Timeout:   20 * time.Second,
		},
		ps:           ps,
		providerName: providerName,
	}
}

// dailyResponse is the DIGITAL_CURRENCY_DAILY payload. The provider answers an
// error with HTTP 200 and a body carrying one of the message fields below, so a
// successful status is not on its own a successful fetch.
type dailyResponse struct {
	ErrorMessage string                       `json:"Error Message"`
	Note         string                       `json:"Note"`
	Information  string                       `json:"Information"`
	Series       map[string]map[string]string `json:"Time Series (Digital Currency Daily)"`
}

// GetEOD yields the daily series for every requested symbol within [from, to].
//
// The provider returns a symbol's whole history in a single response, so unlike
// the paginated MarketStack fetcher this costs exactly one request per symbol
// regardless of the window. That matters: the free tier allows only a few dozen
// requests a day, and a holding's history reaches back to its first purchase.
func (c *Client) GetEOD(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[marketdata.EOD, error] {
	return func(yield func(marketdata.EOD, error) bool) {
		for _, symbol := range symbols {
			currency, ok := DigitalCurrencyBySymbol(symbol)
			if !ok {
				yield(marketdata.EOD{}, fmt.Errorf("alphavantage: %q is not a supported digital currency in %s", symbol, Market))
				return
			}
			series, err := c.fetchDaily(ctx, currency)
			if err != nil {
				yield(marketdata.EOD{}, err)
				return
			}
			if !yieldSeries(yield, symbol, series, from, to) {
				return
			}
		}
	}
}

// fetchDaily performs the one request a symbol costs and books a token for it.
func (c *Client) fetchDaily(ctx context.Context, currency DigitalCurrency) (map[string]map[string]string, error) {
	req, err := c.constructDailyRequest(ctx, currency)
	if err != nil {
		return nil, err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("alphavantage: execute request: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	closeErr := res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("alphavantage: read response body: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("alphavantage: close response body: %w", closeErr)
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("alphavantage: received status %d", res.StatusCode)
	}

	var parsed dailyResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("alphavantage: unmarshal response: %w", err)
	}
	// The quota message arrives as a 200 with no series, so it has to be read out
	// of the body rather than the status; booking the token first keeps the local
	// count honest even when the answer is a refusal.
	if err := c.ps.DeductTokens(ctx, c.providerName, 1); err != nil {
		return nil, fmt.Errorf("alphavantage: deduct token: %w", err)
	}
	if msg := firstNonEmpty(parsed.ErrorMessage, parsed.Note, parsed.Information); msg != "" && len(parsed.Series) == 0 {
		return nil, fmt.Errorf("alphavantage: provider refused the request: %s", msg)
	}
	if len(parsed.Series) == 0 {
		return nil, fmt.Errorf("alphavantage: no daily series for %s", currency.Symbol())
	}
	return parsed.Series, nil
}

func (c *Client) constructDailyRequest(ctx context.Context, currency DigitalCurrency) (*http.Request, error) {
	provider, err := c.ps.GetByName(ctx, c.providerName)
	if err != nil {
		return nil, fmt.Errorf("alphavantage: fetch provider %s: %w", c.providerName, err)
	}
	if provider == nil {
		return nil, fmt.Errorf("alphavantage: provider %s not found", c.providerName)
	}
	if !provider.IsAPIIngestion() {
		return nil, fmt.Errorf("alphavantage: provider %s is not in API ingestion mode", c.providerName)
	}
	if provider.BaseURI == nil || *provider.BaseURI == "" {
		return nil, fmt.Errorf("alphavantage: provider %s has no base URI", c.providerName)
	}
	if provider.ApiKey == nil || *provider.ApiKey == "" {
		return nil, fmt.Errorf("alphavantage: provider %s has no api key", c.providerName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *provider.BaseURI+"/query", nil)
	if err != nil {
		return nil, fmt.Errorf("alphavantage: construct request: %w", err)
	}
	q := req.URL.Query()
	q.Add("function", "DIGITAL_CURRENCY_DAILY")
	q.Add("symbol", currency.Code)
	q.Add("market", Market)
	q.Add("apikey", *provider.ApiKey)
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "application/json")
	return req, nil
}

// yieldSeries converts the provider's date-keyed map into EOD rows inside the
// requested window, oldest first so a consumer that stops early keeps the
// earliest history rather than a random slice of it.
func yieldSeries(
	yield func(marketdata.EOD, error) bool,
	symbol string,
	series map[string]map[string]string,
	from, to *time.Time,
) bool {
	days := make([]string, 0, len(series))
	for day := range series {
		days = append(days, day)
	}
	sort.Strings(days)

	for _, day := range days {
		parsedDay, err := time.ParseInLocation(time.DateOnly, day, time.UTC)
		if err != nil {
			continue
		}
		if from != nil && parsedDay.Before(date.StartOfDayUTC(*from)) {
			continue
		}
		if to != nil && parsedDay.After(date.StartOfDayUTC(*to)) {
			continue
		}
		row := series[day]
		// Crypto trades continuously, so there is no corporate action to report:
		// the split factor is always 1.
		eod, err := marketdata.NewEOD(
			symbol,
			parsedDay,
			parseQuote(row, "1. open"),
			parseQuote(row, "4. close"),
			parseQuote(row, "2. high"),
			parseQuote(row, "3. low"),
			int64(parseQuote(row, "5. volume")),
			1,
		)
		if err != nil {
			continue
		}
		if !yield(eod, nil) {
			return false
		}
	}
	return true
}

// parseQuote reads one numeric field, returning 0 when the provider omits it.
func parseQuote(row map[string]string, key string) float64 {
	value, err := strconv.ParseFloat(row[key], 64)
	if err != nil {
		return 0
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
