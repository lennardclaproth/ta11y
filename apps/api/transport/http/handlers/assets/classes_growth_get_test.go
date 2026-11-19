package assets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/auth"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// fakeClassGrowthStore serves ClassesForAccount and MutationsForClasses from memory;
// the embedded interface panics on any other read.
type fakeClassGrowthStore struct {
	assets.QueryStore
	class *assets.Class
}

func (f *fakeClassGrowthStore) ClassesForAccount(context.Context, uuid.UUID, bool) (map[uuid.UUID]*assets.Class, error) {
	return map[uuid.UUID]*assets.Class{f.class.ID: {ID: f.class.ID, Name: f.class.Name}}, nil
}

func (f *fakeClassGrowthStore) MutationsForClasses(context.Context, []uuid.UUID) (map[uuid.UUID][]assets.Mutation, error) {
	return map[uuid.UUID][]assets.Mutation{f.class.ID: f.class.Mutations}, nil
}

// TestGetClassesGrowthAppliesQueryWindow verifies that from and to reach the query, so
// only points inside the window plus the opening point on from are returned.
func TestGetClassesGrowthAppliesQueryWindow(t *testing.T) {
	mutation := func(day string, worth money.Price) assets.Mutation {
		d, err := time.Parse(time.DateOnly, day)
		if err != nil {
			t.Fatalf("parse %q: %v", day, err)
		}
		return assets.Mutation{EffectiveDate: d, ClassTotalWorth: worth}
	}
	class := &assets.Class{ID: uuid.New(), Name: "Cash", Mutations: []assets.Mutation{
		mutation("2026-01-01", money.Price(100_00)),
		mutation("2026-02-10", money.Price(200_00)),
		mutation("2026-04-01", money.Price(300_00)),
	}}
	queries := assets.NewQueries(&fakeClassGrowthStore{class: class}, nil)

	req := httptest.NewRequest(http.MethodGet, "/assets/classes/growth?from=2026-02-01&to=2026-03-01", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), &auth.Principal{AccountID: uuid.New()}))
	rec := httptest.NewRecorder()

	GetClassesGrowth(nil, *queries).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	var res []ClassGrowthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("series = %+v, want one class", res)
	}
	growth := res[0].Growth
	if len(growth) != 2 || growth[0].Date != "2026-02-01" || growth[1].Date != "2026-02-10" {
		t.Fatalf("growth = %+v, want the opening point on 2026-02-01 and the 2026-02-10 point", growth)
	}
}

// TestGetClassesGrowthRejectsInvalidRange verifies that a malformed date or a reversed
// range is answered with 400 before the query runs.
func TestGetClassesGrowthRejectsInvalidRange(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "malformed from", query: "from=2026-13-01"},
		{name: "malformed to", query: "to=01-01-2026"},
		{name: "from after to", query: "from=2026-03-01&to=2026-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/assets/classes/growth?"+tt.query, nil)
			req = req.WithContext(auth.WithPrincipal(req.Context(), &auth.Principal{AccountID: uuid.New()}))
			rec := httptest.NewRecorder()

			GetClassesGrowth(nil, assets.Queries{}).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}
