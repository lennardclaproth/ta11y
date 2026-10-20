//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

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

	// The note is optional; leaving it out used to dereference a nil pointer and drop the
	// connection instead of creating the item.
	t.Run("create manual asset without a note", func(t *testing.T) {
		status, body := ts.postJSON(t, "/assets/classes", map[string]any{"name": "Savings"})
		if status != http.StatusCreated {
			t.Fatalf("POST /assets/classes: want 201, got %d: %s", status, body)
		}
		var class struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(body, &class); err != nil {
			t.Fatalf("decode class: %v (body=%s)", err, body)
		}

		status, body = ts.postJSON(t, "/assets", map[string]any{
			"class_id":       class.ID,
			"name":           "Cash",
			"initial_worth":  "1000.000000",
			"effective_date": time.Now().UTC().Format(time.DateOnly),
		})
		if status != http.StatusCreated {
			t.Fatalf("POST /assets without note: want 201, got %d: %s", status, body)
		}
	})

	// Editing a hand-valued item's history: a back-dated deposit and an edited date both
	// replay the class in date order, so the worth that holds today stays the latest one.
	t.Run("edit a hand-valued item's history", func(t *testing.T) {
		_, body := ts.postJSON(t, "/assets/classes", map[string]any{"name": "Deposits"})
		var class struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &class)
		_, body = ts.postJSON(t, "/assets", map[string]any{
			"class_id": class.ID, "name": "Savings", "initial_worth": "1000", "effective_date": "2026-01-01",
		})
		var asset struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &asset)
		base := "/assets/" + asset.ID

		if status, body := ts.send(t, http.MethodPut, base+"/worth", map[string]any{
			"worth": "1350", "effective_date": "2026-04-01",
		}); status != http.StatusNoContent {
			t.Fatalf("set worth: want 204, got %d: %s", status, body)
		}
		// Entered after the revaluation, dated before it.
		if status, body := ts.send(t, http.MethodPut, base+"/adjust", map[string]any{
			"direction": "INCREASE", "amount": "300", "effective_date": "2026-03-01",
		}); status != http.StatusNoContent {
			t.Fatalf("adjust: want 204, got %d: %s", status, body)
		}

		type mutation struct {
			ID            string `json:"id"`
			ChangeType    string `json:"change_type"`
			NewWorth      string `json:"new_worth"`
			EffectiveDate string `json:"effective_date"`
		}
		details := func() (worth string, mutations []mutation) {
			status, body := ts.get(t, "/assets/classes/"+class.ID)
			if status != http.StatusOK {
				t.Fatalf("class details: want 200, got %d: %s", status, body)
			}
			var out struct {
				Assets []struct {
					Name         string `json:"name"`
					CurrentWorth string `json:"current_worth"`
				} `json:"assets"`
				Mutations []mutation `json:"mutations"`
			}
			if err := json.Unmarshal(body, &out); err != nil {
				t.Fatalf("decode class details: %v (body=%s)", err, body)
			}
			if len(out.Assets) == 1 {
				worth = out.Assets[0].CurrentWorth
			}
			return worth, out.Mutations
		}

		worth, mutations := details()
		if worth != "1350.000000" {
			t.Fatalf("after a back-dated deposit the later revaluation holds: want 1350, got %s", worth)
		}

		var deposit mutation
		for _, m := range mutations {
			if m.ChangeType == "ADJUST" {
				deposit = m
			}
		}
		// Moving the deposit after the revaluation makes it the latest change.
		if status, body := ts.send(t, http.MethodPut, base+"/mutations/"+deposit.ID, map[string]any{
			"amount": "300", "direction": "INCREASE", "effective_date": "2026-05-01",
		}); status != http.StatusNoContent {
			t.Fatalf("edit worth change: want 204, got %d: %s", status, body)
		}
		if worth, _ = details(); worth != "1650.000000" {
			t.Fatalf("after moving the deposit last: want 1650, got %s", worth)
		}

		if status, body := ts.send(t, http.MethodDelete, base+"/mutations/"+deposit.ID, nil); status != http.StatusNoContent {
			t.Fatalf("delete worth change: want 204, got %d: %s", status, body)
		}
		worth, mutations = details()
		if worth != "1350.000000" || len(mutations) != 2 {
			t.Fatalf("after deleting the deposit: want 1350 over 2 changes, got %s over %d", worth, len(mutations))
		}

		if status, body := ts.send(t, http.MethodPatch, base, map[string]any{"name": "Savings account"}); status != http.StatusNoContent {
			t.Fatalf("rename: want 204, got %d: %s", status, body)
		}
		if status, body := ts.send(t, http.MethodDelete, base, nil); status != http.StatusNoContent {
			t.Fatalf("delete item: want 204, got %d: %s", status, body)
		}
		if _, mutations = details(); len(mutations) != 0 {
			t.Fatalf("deleting the item removes its history: got %d changes", len(mutations))
		}
	})

	t.Run("an add before the opening worth is refused", func(t *testing.T) {
		_, body := ts.postJSON(t, "/assets/classes", map[string]any{"name": "Openings"})
		var class struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &class)
		_, body = ts.postJSON(t, "/assets", map[string]any{
			"class_id": class.ID, "name": "Jar", "initial_worth": "100", "effective_date": "2026-03-01",
		})
		var asset struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &asset)

		status, body := ts.send(t, http.MethodPut, "/assets/"+asset.ID+"/adjust", map[string]any{
			"direction": "INCREASE", "amount": "50", "effective_date": "2026-02-01",
		})
		if status != http.StatusBadRequest {
			t.Fatalf("adjust before the opening: want 400, got %d: %s", status, body)
		}
	})

	t.Run("edit and delete purchases", func(t *testing.T) {
		_, body := ts.postJSON(t, "/assets/classes", map[string]any{"name": "Coins"})
		var class struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &class)
		status, body := ts.postJSON(t, "/assets/holdings", map[string]any{
			"class_id": class.ID, "name": "Bitcoin", "symbol": "BTC/EUR",
			"purchase": map[string]any{"date": "2026-02-01", "quantity": "0.01", "unit_price": "40000"},
		})
		if status != http.StatusCreated {
			t.Fatalf("create holding: want 201, got %d: %s", status, body)
		}
		var holding struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &holding)
		base := "/assets/" + holding.ID

		status, body = ts.postJSON(t, base+"/purchases", map[string]any{
			"date": "2026-03-01", "quantity": "0.02", "unit_price": "45000",
		})
		if status != http.StatusCreated {
			t.Fatalf("add purchase: want 201, got %d: %s", status, body)
		}
		var second struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(body, &second)

		if status, body := ts.send(t, http.MethodPut, base+"/purchases/"+second.ID, map[string]any{
			"date": "2026-03-02", "quantity": "0.03", "unit_price": "44000",
		}); status != http.StatusNoContent {
			t.Fatalf("edit purchase: want 204, got %d: %s", status, body)
		}

		status, body = ts.get(t, "/assets/holdings/"+holding.ID)
		if status != http.StatusOK {
			t.Fatalf("get holding: want 200, got %d: %s", status, body)
		}
		var detail struct {
			Purchases []struct {
				ID       string `json:"id"`
				Date     string `json:"date"`
				Quantity string `json:"quantity"`
			} `json:"purchases"`
		}
		_ = json.Unmarshal(body, &detail)
		var edited bool
		for _, p := range detail.Purchases {
			if p.ID == second.ID && p.Date == "2026-03-02" && p.Quantity == "0.03" {
				edited = true
			}
		}
		if !edited {
			t.Fatalf("edited purchase not stored: %s", body)
		}

		if status, body := ts.send(t, http.MethodDelete, base+"/purchases/"+second.ID, nil); status != http.StatusNoContent {
			t.Fatalf("delete purchase: want 204, got %d: %s", status, body)
		}
		first := detail.Purchases[0].ID
		if first == second.ID {
			first = detail.Purchases[1].ID
		}
		if status, body := ts.send(t, http.MethodDelete, base+"/purchases/"+first, nil); status != http.StatusConflict {
			t.Fatalf("delete the last purchase: want 409, got %d: %s", status, body)
		}
	})
}
