package assets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/auth"
)

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
