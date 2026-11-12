package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeSite(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"index.html":                  "<html>app</html>",
		"robots.txt":                  "robots",
		"_app/immutable/entry/app.js": "js",
	}
	for name, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func get(t *testing.T, h http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestSiteHandler(t *testing.T) {
	h := SiteHandler(writeSite(t))

	cases := []struct {
		name, target, body, cache string
	}{
		{name: "existing file", target: "/robots.txt", body: "robots"},
		{name: "immutable asset", target: "/_app/immutable/entry/app.js", body: "js", cache: "public, max-age=31536000, immutable"},
		{name: "client route falls back", target: "/cashflow/imports/123", body: "<html>app</html>", cache: "no-cache"},
		{name: "root", target: "/", body: "<html>app</html>", cache: "no-cache"},
		{name: "directory falls back", target: "/_app/immutable/", body: "<html>app</html>", cache: "no-cache"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := get(t, h, http.MethodGet, tc.target)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}
			if body, _ := io.ReadAll(rec.Body); string(body) != tc.body {
				t.Fatalf("body = %q, want %q", body, tc.body)
			}
			if got := rec.Header().Get("Cache-Control"); got != tc.cache {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.cache)
			}
		})
	}

	// A path that climbs out of the site is refused rather than resolved.
	if rec := get(t, h, http.MethodGet, "/../../etc/passwd"); rec.Code != http.StatusBadRequest {
		t.Fatalf("traversal status = %d, want 400", rec.Code)
	}
	if rec := get(t, h, http.MethodPost, "/cashflow"); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", rec.Code)
	}
}

func TestWithSiteRoutesTheAPIPrefixToTheAPI(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "api:"+r.URL.Path)
	})
	site := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "site:"+r.URL.Path)
	})
	s := &Server{}
	WithSite("/api/", site)(s)
	h := s.withSite(api)

	for target, want := range map[string]string{
		"/api/cashflow/transactions": "api:/cashflow/transactions",
		"/api/health":                "api:/health",
		"/cashflow":                  "site:/cashflow",
		"/apiary":                    "site:/apiary",
	} {
		if body, _ := io.ReadAll(get(t, h, http.MethodGet, target).Body); string(body) != want {
			t.Errorf("%s -> %q, want %q", target, body, want)
		}
	}
}

func TestFlowCookieCarriesThePathPrefix(t *testing.T) {
	if got := (CookieSettings{}).FlowCookie("v").Path; got != "/auth" {
		t.Fatalf("unprefixed flow cookie path = %q", got)
	}
	if got := (CookieSettings{PathPrefix: "/api"}).FlowCookie("v").Path; got != "/api/auth" {
		t.Fatalf("prefixed flow cookie path = %q", got)
	}
}
