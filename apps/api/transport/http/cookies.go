package http

import (
	"net/http"
	"time"

	"github.com/lennardclaproth/ta11y/internal/auth"
)

const (
	// SessionCookieName carries the opaque session token.
	SessionCookieName = "mft_session"
	// FlowCookieName carries the in-progress login's state, nonce and PKCE verifier.
	FlowCookieName = "mft_auth_flow"
	// FlowCookiePath scopes the flow cookie to the sign-in endpoints.
	FlowCookiePath = "/auth"
)

// CookieSettings describes how auth cookies are written. They are always HttpOnly and
// SameSite=Lax: Lax still accompanies the provider's top-level redirect back to the
// callback, while keeping the cookie off cross-site subrequests.
type CookieSettings struct {
	// Domain scopes the cookie. Empty leaves it host-only, which is what you want
	// whenever the API and frontend share a host.
	Domain string
	// Secure adds the Secure attribute. It must be true anywhere served over HTTPS.
	Secure bool
}

// SessionCookie builds the cookie carrying a freshly issued session token.
func (c CookieSettings) SessionCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Domain:   c.Domain,
		Expires:  expires,
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// FlowCookie builds the short-lived cookie carrying an in-progress login's state.
func (c CookieSettings) FlowCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     FlowCookieName,
		Value:    value,
		Path:     FlowCookiePath,
		Domain:   c.Domain,
		MaxAge:   int(auth.FlowTTL.Seconds()),
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// Clear expires a cookie by name. Path and Domain must match the original for the
// browser to replace rather than add a cookie.
func (c CookieSettings) Clear(name, path string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   c.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// SessionToken returns the session token presented by a request, if any.
func SessionToken(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
