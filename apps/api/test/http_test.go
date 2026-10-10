//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// TestHTTPEndpoints drives the real server binary over HTTP. One server (with
// its own throwaway SQLite DB) is shared across the subtests; they are ordered
// so they do not collide on the unique account name constraint.
func TestHTTPEndpoints(t *testing.T) {
	ts := startServer(t)

	t.Run("health", func(t *testing.T) {
		status, _ := ts.get(t, "/health")
		if status != http.StatusOK {
			t.Fatalf("GET /health: want 200, got %d", status)
		}
	})

	t.Run("list vendors seeded by bootstrap", func(t *testing.T) {
		status, body := ts.get(t, "/vendors")
		if status != http.StatusOK {
			t.Fatalf("GET /vendors: want 200, got %d: %s", status, body)
		}
		var vendors []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(body, &vendors); err != nil {
			t.Fatalf("decode vendors: %v (body=%s)", err, body)
		}
		if len(vendors) == 0 {
			t.Fatalf("expected bootstrap-seeded vendors, got none (body=%s)", body)
		}
	})

	t.Run("create account", func(t *testing.T) {
		status, body := ts.postJSON(t, "/accounts", map[string]any{"name": "Brokerage"})
		if status != http.StatusCreated {
			t.Fatalf("POST /accounts: want 201, got %d: %s", status, body)
		}
		var out struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("decode create response: %v (body=%s)", err, body)
		}
		if _, err := uuid.Parse(out.ID); err != nil {
			t.Fatalf("expected a UUID id, got %q", out.ID)
		}
	})

	t.Run("create account validation error", func(t *testing.T) {
		status, body := ts.postJSON(t, "/accounts", map[string]any{"name": "   "})
		if status != http.StatusBadRequest {
			t.Fatalf("POST /accounts with blank name: want 400, got %d: %s", status, body)
		}
	})

	// The daily-priced instrument catalogue is static and costs no provider request, so it
	// answers on a fresh database. The blocked rows are the point: an instrument with no euro
	// source is offered unselectable with its reason rather than left out.
	t.Run("search daily-priced quotes", func(t *testing.T) {
		status, body := ts.get(t, "/marketdata/quotes")
		if status != http.StatusOK {
			t.Fatalf("GET /marketdata/quotes: want 200, got %d: %s", status, body)
		}
		var quotes []struct {
			Symbol     string `json:"symbol"`
			Currency   string `json:"currency"`
			Selectable bool   `json:"selectable"`
			Reason     string `json:"reason"`
		}
		if err := json.Unmarshal(body, &quotes); err != nil {
			t.Fatalf("decode quotes: %v (body=%s)", err, body)
		}

		var selectable, blocked int
		for _, quote := range quotes {
			if quote.Selectable {
				selectable++
				if quote.Currency != "EUR" {
					t.Errorf("selectable quote %s is quoted in %s, not EUR", quote.Symbol, quote.Currency)
				}
				continue
			}
			blocked++
			if quote.Reason == "" {
				t.Errorf("blocked quote %s carries no reason", quote.Symbol)
			}
		}
		if selectable == 0 {
			t.Errorf("expected at least one selectable instrument (body=%s)", body)
		}
		if blocked == 0 {
			t.Errorf("expected the metals to be offered blocked (body=%s)", body)
		}
	})
}
