package http

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/auth"
	"github.com/lennardclaproth/ta11y/internal/logging"
)

// PrincipalResolver turns a session token into the authenticated caller. It is
// satisfied by auth.Queries.
type PrincipalResolver interface {
	ResolveSession(ctx context.Context, token string) (*auth.Principal, error)
}

// WithAuthentication resolves the session cookie and puts the caller on the request
// context, answering 401 when there is no live session. Routes are registered through
// this by default, so a new route is protected unless it is deliberately made public.
//
// fallback, when non-nil, authenticates every request that presents no session. It
// exists only for running with authentication switched off, where the API behaves as
// the single bootstrapped account exactly as it did before sign-in existed. Config
// refuses that combination outside development.
func WithAuthentication(resolver PrincipalResolver, fallback *auth.Principal, log logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := SessionToken(r)

			if token == "" && fallback != nil {
				next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), fallback)))
				return
			}

			principal, err := resolver.ResolveSession(r.Context(), token)
			if err != nil {
				// An expired or revoked session is worth telling apart from never having
				// signed in: the client should clear its state and sign in again.
				if errors.Is(err, auth.ErrSessionExpired) {
					writeUnauthorized(w, "session expired")
					return
				}
				if !errors.Is(err, auth.ErrSessionNotFound) {
					log.Error(r.Context(), "authentication: failed to resolve session", err)
				}
				writeUnauthorized(w, "not signed in")
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
		})
	}
}

// RequireAdmin rejects callers whose account is not marked admin. It layers on top of
// WithAuthentication and answers 403, since the caller is authenticated but not
// permitted.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			writeUnauthorized(w, "not signed in")
			return
		}
		if !principal.Admin {
			_ = JSONEncode(w, http.StatusForbidden, map[string]string{"error": "administrator access is required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WithOriginCheck rejects state-changing requests whose Origin is not one this API
// serves. Combined with a SameSite=Lax session cookie it is the CSRF defence: Lax
// already withholds the cookie from cross-site subrequests, and this closes the gap for
// clients that send an Origin the browser would otherwise allow.
//
// Requests without an Origin header pass. Browsers always send one on cross-origin and
// on every non-GET fetch, so an absent Origin means a non-browser caller, which carries
// no ambient cookie authority to abuse.
func WithOriginCheck(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			allowAll = true
		}
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isSafeMethod(r.Method) || allowAll {
				next.ServeHTTP(w, r)
				return
			}

			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := allowed[origin]; ok {
				next.ServeHTTP(w, r)
				return
			}
			// The API's own origin is always acceptable: the sign-in endpoints are
			// reached from the browser's address bar, not from the frontend.
			if sameHost(origin, r.Host) {
				next.ServeHTTP(w, r)
				return
			}

			_ = JSONEncode(w, http.StatusForbidden, map[string]string{"error": "cross-origin request rejected"})
		})
	}
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func sameHost(origin, host string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, host)
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	_ = JSONEncode(w, http.StatusUnauthorized, map[string]string{"error": message})
}

// AccountID returns the authenticated caller's account, writing a 401 and reporting
// false when the request carried no principal. Handlers use it in place of decoding an
// account id from transport input, which is what keeps one account's data out of
// another's requests.
func AccountID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	accountID, ok := auth.AccountFromContext(r.Context())
	if !ok {
		writeUnauthorized(w, "not signed in")
		return uuid.Nil, false
	}
	return accountID, true
}
