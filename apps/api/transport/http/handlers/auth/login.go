package auth

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/auth"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// Login starts an OpenID Connect login and redirects the browser to the provider.
//
// @Summary Start sign-in
// @Description Redirects to the identity provider's authorization endpoint. Browser-facing; not callable from XHR.
// @Tags auth
// @Param provider path string true "Identity provider slug (e.g. google)"
// @Success 302 {string} string "Redirect to the identity provider"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/{provider}/login [get]
func Login(log logging.Logger, registry *auth.Registry, cookies httpx.CookieSettings, frontendURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("provider")
		provider, err := registry.Get(slug)
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "unknown identity provider"})
			return
		}

		flow, err := auth.NewFlow(provider.Slug(), frontendURL)
		if err != nil {
			log.Error(r.Context(), "login: failed to start authentication flow", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to start sign-in"})
			return
		}

		encoded, err := flow.Encode()
		if err != nil {
			log.Error(r.Context(), "login: failed to encode authentication flow", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to start sign-in"})
			return
		}

		http.SetCookie(w, cookies.FlowCookie(encoded))
		http.Redirect(w, r, provider.AuthCodeURL(flow), http.StatusFound)
	})
}

// Providers lists the configured identity provider slugs so the sign-in page can render
// one button per provider instead of hard-coding them.
//
// @Summary List identity providers
// @Description Returns the configured OpenID Connect provider slugs.
// @Tags auth
// @Produce json
// @Success 200 {object} ProvidersResponse
// @Router /auth/providers [get]
func Providers(registry *auth.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.JSONEncode(w, http.StatusOK, ProvidersResponse{Providers: registry.Slugs()})
	})
}

// ProvidersResponse lists the identity providers a client may sign in with.
type ProvidersResponse struct {
	Providers []string `json:"providers"`
}
