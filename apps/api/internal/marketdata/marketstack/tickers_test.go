package marketstack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lennardclaproth/ta11y/internal/marketdata"
)

// stubProviderStore returns a provider pointed at the test server.
type stubProviderStore struct {
	baseURI  string
	deducted int32
}

func (s *stubProviderStore) GetByName(_ context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
	key := "test-key"
	return &marketdata.Provider{
		Name:          name,
		IngestionMode: marketdata.ProviderIngestionModeAPI,
		BaseURI:       &s.baseURI,
		ApiKey:        &key,
	}, nil
}

func (s *stubProviderStore) DeductTokens(_ context.Context, _ marketdata.ProviderName, count int32) error {
	s.deducted += count
	return nil
}

const tickersListBody = `{
  "pagination": {"limit": 2, "offset": 0, "count": 2, "total": 31},
  "data": [
    {"name":"ASML Holding NV","ticker":"ASML","has_intraday":false,"has_eod":true,
     "stock_exchange":{"name":"NASDAQ - ALL MARKETS","acronym":"NASDAQ","mic":"XNAS"}},
    {"name":"ASML HOLDING","ticker":"ASML.XAMS","has_intraday":false,"has_eod":true,
     "stock_exchange":{"name":"EURONEXT - EURONEXT AMSTERDAM","acronym":"","mic":"XAMS"}}
  ]
}`

// The v2 catalogue endpoint is "tickerslist". Requesting v1's "tickers" against v2
// answers 404 "Route not found", which surfaced as a 500 from catalogue search, so the
// path is pinned here rather than left to a string literal in the request builder.
func TestSearchTickersCallsTheV2TickersListEndpoint(t *testing.T) {
	var gotPath, gotSearch string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotSearch = r.URL.Query().Get("search")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tickersListBody))
	}))
	defer server.Close()

	store := &stubProviderStore{baseURI: server.URL + "/v2"}
	client := NewMarketStackClient(store, marketdata.ProviderMarketStack)

	page, err := client.SearchTickers(t.Context(), "asml", 2, 0)
	if err != nil {
		t.Fatalf("SearchTickers: %v", err)
	}

	if gotPath != "/v2/tickerslist" {
		t.Fatalf("path = %q, want /v2/tickerslist", gotPath)
	}
	if gotSearch != "asml" {
		t.Fatalf("search = %q, want asml", gotSearch)
	}
	if page.Total != 31 {
		t.Fatalf("total = %d, want the provider's full match count 31", page.Total)
	}
	if len(page.Data) != 2 {
		t.Fatalf("entries = %d, want 2", len(page.Data))
	}
	if store.deducted != 1 {
		t.Fatalf("deducted = %d, want 1 request billed", store.deducted)
	}

	// The MIC is what distinguishes the listing variants of one company.
	amsterdam := page.Data[1]
	if amsterdam.Symbol != "ASML.XAMS" {
		t.Fatalf("symbol = %q, want ASML.XAMS", amsterdam.Symbol)
	}
	if amsterdam.ExchangeMIC == nil || *amsterdam.ExchangeMIC != "XAMS" {
		t.Fatalf("exchange MIC = %v, want XAMS", amsterdam.ExchangeMIC)
	}
	if !amsterdam.HasEOD {
		t.Fatal("expected the Amsterdam listing to report end-of-day availability")
	}
}

// A non-2xx from the provider must surface as an error rather than being parsed as an
// empty page, which is how the 404 stayed invisible until it reached the HTTP layer.
func TestSearchTickersReportsAProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"not_found_error","message":"Route not found"}}`))
	}))
	defer server.Close()

	store := &stubProviderStore{baseURI: server.URL + "/v2"}
	client := NewMarketStackClient(store, marketdata.ProviderMarketStack)

	if _, err := client.SearchTickers(t.Context(), "asml", 2, 0); err == nil {
		t.Fatal("expected an error for a 404 from the provider")
	}
	if store.deducted != 0 {
		t.Fatalf("deducted = %d, want 0 -- a failed request must not be billed", store.deducted)
	}
}
