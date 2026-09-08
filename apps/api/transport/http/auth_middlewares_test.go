package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/lennardclaproth/my-finances-tracker/internal/auth"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	httpx "github.com/lennardclaproth/my-finances-tracker/transport/http"
)

// stubResolver resolves exactly one token, and reports every other as unknown.
type stubResolver struct {
	token     string
	principal *auth.Principal
	err       error
}

func (s stubResolver) ResolveSession(_ context.Context, token string) (*auth.Principal, error) {
	if s.err != nil {
		return nil, s.err
	}
	if token != s.token || token == "" {
		return nil, auth.ErrSessionNotFound
	}
	return s.principal, nil
}

func testLogger() logging.Logger { return logging.NewSlogLogger(nil) }

// reached records whether the wrapped handler ran, which is what "was the request
// allowed through" means for these tests.
func reached(flag *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*flag = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestWithAuthenticationRejectsAMissingSession(t *testing.T) {
	var ran bool
	handler := httpx.WithAuthentication(stubResolver{}, nil, testLogger())(reached(&ran))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/cashflow/transactions", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if ran {
		t.Fatal("an unauthenticated request must not reach the handler")
	}
}

func TestWithAuthenticationPutsThePrincipalOnTheContext(t *testing.T) {
	want := &auth.Principal{AccountID: uuid.New(), Email: "person@example.com"}
	resolver := stubResolver{token: "good-token", principal: want}

	var got uuid.UUID
	handler := httpx.WithAuthentication(resolver, nil, testLogger())(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountID, ok := httpx.AccountID(w, r)
			if !ok {
				t.Error("expected an account on the request context")
				return
			}
			got = accountID
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/cashflow/transactions", nil)
	req.AddCookie(&http.Cookie{Name: httpx.SessionCookieName, Value: "good-token"})
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != want.AccountID {
		t.Fatalf("account = %v, want %v", got, want.AccountID)
	}
}

func TestWithAuthenticationRejectsAnUnknownToken(t *testing.T) {
	var ran bool
	resolver := stubResolver{token: "good-token", principal: &auth.Principal{AccountID: uuid.New()}}
	handler := httpx.WithAuthentication(resolver, nil, testLogger())(reached(&ran))

	req := httptest.NewRequest(http.MethodGet, "/cashflow/transactions", nil)
	req.AddCookie(&http.Cookie{Name: httpx.SessionCookieName, Value: "forged-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if ran {
		t.Fatal("a forged token must not reach the handler")
	}
}

func TestWithAuthenticationReportsAnExpiredSessionDistinctly(t *testing.T) {
	resolver := stubResolver{err: auth.ErrSessionExpired}
	handler := httpx.WithAuthentication(resolver, nil, testLogger())(reached(new(bool)))

	req := httptest.NewRequest(http.MethodGet, "/cashflow/transactions", nil)
	req.AddCookie(&http.Cookie{Name: httpx.SessionCookieName, Value: "stale"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "session expired") {
		t.Fatalf("body = %q, want it to name the expired session", body)
	}
}

// With sign-in switched off the fallback stands in, so the API stays usable exactly as
// it was before authentication existed.
func TestWithAuthenticationUsesTheFallbackWhenThereIsNoSession(t *testing.T) {
	fallback := &auth.Principal{AccountID: uuid.New(), Admin: true}

	var got uuid.UUID
	handler := httpx.WithAuthentication(stubResolver{}, fallback, testLogger())(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, _ = httpx.AccountID(w, r)
		}),
	)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/assets/snapshots", nil))

	if got != fallback.AccountID {
		t.Fatalf("account = %v, want the fallback %v", got, fallback.AccountID)
	}
}

func TestRequireAdmin(t *testing.T) {
	cases := []struct {
		name      string
		principal *auth.Principal
		want      int
	}{
		{"admin passes", &auth.Principal{AccountID: uuid.New(), Admin: true}, http.StatusOK},
		{"non-admin is forbidden", &auth.Principal{AccountID: uuid.New()}, http.StatusForbidden},
		{"unauthenticated is unauthorized", nil, http.StatusUnauthorized},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ran bool
			handler := httpx.RequireAdmin(reached(&ran))

			req := httptest.NewRequest(http.MethodGet, "/marketdata/providers", nil)
			if tc.principal != nil {
				req = req.WithContext(auth.WithPrincipal(req.Context(), tc.principal))
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
			if ran != (tc.want == http.StatusOK) {
				t.Fatalf("handler reached = %v, want %v", ran, tc.want == http.StatusOK)
			}
		})
	}
}

func TestWithOriginCheck(t *testing.T) {
	allowed := []string{"http://localhost:5199"}

	cases := []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{"allowed origin mutating", http.MethodPost, "http://localhost:5199", http.StatusOK},
		{"foreign origin mutating", http.MethodPost, "http://evil.example", http.StatusForbidden},
		{"foreign origin reading", http.MethodGet, "http://evil.example", http.StatusOK},
		{"no origin header", http.MethodPost, "", http.StatusOK},
		{"api's own origin", http.MethodPost, "http://api.example", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := httpx.WithOriginCheck(allowed)(reached(new(bool)))

			req := httptest.NewRequest(tc.method, "http://api.example/cashflow/transactions/tag", nil)
			req.Host = "api.example"
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
