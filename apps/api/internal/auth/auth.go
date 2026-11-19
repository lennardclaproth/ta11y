// Package auth authenticates people through OpenID Connect and turns a verified
// identity into a server-side session.
//
// The package stores as little as it can: an account is an id and an email, and each
// identity provider contributes only a pseudonymous (issuer, subject) pair. Tokens
// issued by the provider are verified once at the callback and then discarded -- the
// session is the only credential that outlives the login.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

// sessionTokenBytes is the entropy of a session token. 32 bytes puts guessing far
// beyond reach while keeping the cookie short.
const sessionTokenBytes = 32

// DefaultSessionTTL is how long a session stays valid when configuration says nothing.
const DefaultSessionTTL = 30 * 24 * time.Hour

// Claims is the subset of an OIDC ID token this application consumes. Everything a
// provider offers beyond these fields is deliberately ignored rather than stored.
type Claims struct {
	// Issuer identifies the provider, taken from the verified token rather than config.
	Issuer string
	// Subject is the provider's stable, opaque identifier for the person.
	Subject string
	// Email is the only personal data retained.
	Email string
	// EmailVerified reports whether the provider vouches for the address. Account
	// linking depends on it; see ErrEmailUnverified.
	EmailVerified bool
}

// Identity links one identity provider's view of a person to a local account.
type Identity struct {
	ID          uuid.UUID  `db:"id"`
	AccountID   uuid.UUID  `db:"account_id"`
	Issuer      string     `db:"issuer"`
	Subject     string     `db:"subject"`
	CreatedAt   time.Time  `db:"created_at"`
	LastLoginAt *time.Time `db:"last_login_at"`
}

// Session is a server-side login. The token handed to the client is not stored; only
// TokenHash is, so a leak of this table yields nothing that can be presented as a
// credential.
type Session struct {
	ID         uuid.UUID  `db:"id"`
	AccountID  uuid.UUID  `db:"account_id"`
	TokenHash  string     `db:"token_hash"`
	IssuedAt   time.Time  `db:"issued_at"`
	ExpiresAt  time.Time  `db:"expires_at"`
	LastSeenAt time.Time  `db:"last_seen_at"`
	RevokedAt  *time.Time `db:"revoked_at"`
}

// Active reports whether the session may still authenticate a request at the given time.
func (s *Session) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}

// Principal is the authenticated caller as the rest of the application sees it. It is
// what middleware puts on the request context.
type Principal struct {
	AccountID uuid.UUID
	Email     string
	Admin     bool
}

// IssuedSession pairs a persisted session with the one-time raw token that belongs in
// the client's cookie. The raw token exists only for the length of this value.
type IssuedSession struct {
	Session *Session
	Token   string
}

// NewSession mints a session and its raw token. The caller is responsible for handing
// Token to the client and never persisting it.
func NewSession(accountID uuid.UUID, ttl time.Duration) (*IssuedSession, error) {
	token, err := newSessionToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}

	return &IssuedSession{
		Session: &Session{
			ID:         uuid.New(),
			AccountID:  accountID,
			TokenHash:  HashToken(token),
			IssuedAt:   now,
			ExpiresAt:  now.Add(ttl),
			LastSeenAt: now,
		},
		Token: token,
	}, nil
}

// HashToken derives the stored form of a session token. Sessions are looked up by this
// value, never by the token itself.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NormalizeEmail lowercases and trims an email so lookups match the database's
// LOWER(email) unique index.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func newSessionToken() (string, error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

type principalContextKey struct{}

var principalKey principalContextKey

// WithPrincipal returns a context carrying the authenticated caller.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// PrincipalFromContext returns the authenticated caller, or false when the request was
// not authenticated.
func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey).(*Principal)
	return p, ok && p != nil
}

// AccountFromContext returns the authenticated caller's account id. Handlers use it
// instead of reading an account id from transport input.
func AccountFromContext(ctx context.Context) (uuid.UUID, bool) {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}
	return p.AccountID, true
}
