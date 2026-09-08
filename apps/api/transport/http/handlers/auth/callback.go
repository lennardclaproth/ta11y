package auth

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/lennardclaproth/ta11y/internal/auth"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// Callback completes an OpenID Connect login: it validates the flow, exchanges the
// authorization code, verifies the ID token, and issues a session cookie.
//
// Failures redirect back to the frontend carrying a coarse reason rather than rendering
// JSON, because the caller here is a browser following the provider's redirect. The
// reason is deliberately coarse -- enough for the sign-in page to explain itself,
// not enough to tell an attacker which step failed.
//
// @Summary Complete sign-in
// @Description Handles the identity provider's redirect and issues a session cookie. Browser-facing.
// @Tags auth
// @Param provider path string true "Identity provider slug (e.g. google)"
// @Param code query string true "Authorization code"
// @Param state query string true "Opaque state issued at login"
// @Success 302 {string} string "Redirect to the frontend"
// @Router /auth/{provider}/callback [get]
func Callback(
	log logging.Logger,
	registry *auth.Registry,
	commands *auth.Commands,
	cookies httpx.CookieSettings,
	frontendURL string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The flow cookie has served its purpose either way; clear it before any exit.
		http.SetCookie(w, cookies.Clear(httpx.FlowCookieName, httpx.FlowCookiePath))

		slug := r.PathValue("provider")
		provider, err := registry.Get(slug)
		if err != nil {
			redirectFailure(w, r, frontendURL, "unknown_provider")
			return
		}

		// A provider that refuses (consent denied, for instance) reports it here rather
		// than sending a code.
		if providerErr := r.URL.Query().Get("error"); providerErr != "" {
			log.Info(r.Context(), "callback: identity provider returned an error",
				"provider", slug, "provider_error", providerErr)
			redirectFailure(w, r, frontendURL, "provider_error")
			return
		}

		flowCookie, err := r.Cookie(httpx.FlowCookieName)
		if err != nil {
			redirectFailure(w, r, frontendURL, "flow_missing")
			return
		}
		flow, err := auth.DecodeFlow(flowCookie.Value)
		if err != nil {
			redirectFailure(w, r, frontendURL, "flow_invalid")
			return
		}
		if err := flow.Validate(provider.Slug(), r.URL.Query().Get("state")); err != nil {
			if errors.Is(err, auth.ErrFlowExpired) {
				redirectFailure(w, r, frontendURL, "flow_expired")
				return
			}
			log.Warn(r.Context(), "callback: authentication state mismatch", "provider", slug)
			redirectFailure(w, r, frontendURL, "flow_invalid")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			redirectFailure(w, r, frontendURL, "code_missing")
			return
		}

		claims, err := provider.Exchange(r.Context(), code, flow)
		if err != nil {
			log.Error(r.Context(), "callback: failed to exchange authorization code", err, "provider", slug)
			redirectFailure(w, r, frontendURL, "exchange_failed")
			return
		}

		issued, err := commands.Authenticate(r.Context(), *claims)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrEmailMissing):
				redirectFailure(w, r, frontendURL, "email_missing")
			case errors.Is(err, auth.ErrEmailUnverified):
				redirectFailure(w, r, frontendURL, "email_unverified")
			case errors.Is(err, auth.ErrEmailNotAllowed):
				log.Warn(r.Context(), "callback: sign-in refused for an address outside the allowlist",
					"provider", slug, "email", claims.Email)
				redirectFailure(w, r, frontendURL, "not_allowed")
			default:
				log.Error(r.Context(), "callback: failed to authenticate", err, "provider", slug)
				redirectFailure(w, r, frontendURL, "authentication_failed")
			}
			return
		}

		http.SetCookie(w, cookies.SessionCookie(issued.Token, issued.Session.ExpiresAt))
		log.Info(r.Context(), "signed in", "provider", slug, "account_id", issued.Session.AccountID.String())

		target := flow.Redirect
		if target == "" {
			target = frontendURL
		}
		http.Redirect(w, r, target, http.StatusFound)
	})
}

// redirectFailure sends the browser back to the sign-in page with a reason it can show.
func redirectFailure(w http.ResponseWriter, r *http.Request, frontendURL, reason string) {
	target, err := url.Parse(frontendURL)
	if err != nil {
		http.Error(w, "sign-in failed", http.StatusBadRequest)
		return
	}
	target.Path = "/login"
	query := target.Query()
	query.Set("error", reason)
	target.RawQuery = query.Encode()
	http.Redirect(w, r, target.String(), http.StatusFound)
}
