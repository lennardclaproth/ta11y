package auth

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/auth"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// MeResponse is the authenticated caller as clients see it. It is intentionally the
// whole of what the application knows about a person.
type MeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	// Admin gates the admin-only screens.
	Admin bool `json:"admin"`
}

// Me returns the authenticated caller, or 401 when the request carries no live session.
// It is how the frontend bootstraps its session and learns its account id.
//
// @Summary Current session
// @Description Returns the signed-in account, or 401 when not signed in.
// @Tags auth
// @Produce json
// @Success 200 {object} MeResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func Me(log logging.Logger, queries *auth.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := queries.ResolveSession(r.Context(), httpx.SessionToken(r))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusUnauthorized, map[string]string{"error": "not signed in"})
			return
		}
		_ = httpx.JSONEncode(w, http.StatusOK, MeResponse{
			ID:    principal.AccountID.String(),
			Email: principal.Email,
			Admin: principal.Admin,
		})
	})
}

// Logout revokes the caller's session and clears the cookie. It succeeds even when no
// session was presented, so a client can always reach a signed-out state.
//
// @Summary Sign out
// @Description Revokes the current session and clears the session cookie.
// @Tags auth
// @Success 204 {string} string "Signed out"
// @Failure 500 {object} map[string]string
// @Router /auth/logout [post]
func Logout(log logging.Logger, commands *auth.Commands, cookies httpx.CookieSettings) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := commands.Logout(r.Context(), httpx.SessionToken(r)); err != nil {
			log.Error(r.Context(), "logout: failed to revoke session", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to sign out"})
			return
		}
		http.SetCookie(w, cookies.Clear(httpx.SessionCookieName, "/"))
		w.WriteHeader(http.StatusNoContent)
	})
}
