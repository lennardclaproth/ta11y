package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/account"
)

// IdentityStore persists the link between an identity provider's view of a person and
// a local account.
type IdentityStore interface {
	GetByIssuerSubject(ctx context.Context, issuer, subject string) (*Identity, error)
	CreateIdentity(ctx context.Context, identity *Identity) error
	TouchLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error
}

// SessionStore persists server-side sessions.
type SessionStore interface {
	CreateSession(ctx context.Context, session *Session) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	Revoke(ctx context.Context, tokenHash string, at time.Time) error
	RevokeAllForAccount(ctx context.Context, accountID uuid.UUID, at time.Time) error
	TouchLastSeen(ctx context.Context, id uuid.UUID, at time.Time) error
}

// EmailAllowFunc reports whether an address may sign in. It is the admission policy:
// authenticating with an identity provider proves who someone is, not that they are
// welcome here.
type EmailAllowFunc func(email string) bool

// Commands exposes authentication write-side use cases.
type Commands struct {
	identities IdentityStore
	sessions   SessionStore
	accounts   *account.Queries
	provision  *account.Commands
	sessionTTL time.Duration
	allowEmail EmailAllowFunc
}

// NewCommands creates authentication write-side use cases. A non-positive ttl falls
// back to DefaultSessionTTL.
func NewCommands(
	identities IdentityStore,
	sessions SessionStore,
	accounts *account.Queries,
	provision *account.Commands,
	sessionTTL time.Duration,
	allowEmail EmailAllowFunc,
) *Commands {
	if sessionTTL <= 0 {
		sessionTTL = DefaultSessionTTL
	}
	if allowEmail == nil {
		allowEmail = func(string) bool { return true }
	}
	return &Commands{
		identities: identities,
		sessions:   sessions,
		accounts:   accounts,
		provision:  provision,
		sessionTTL: sessionTTL,
		allowEmail: allowEmail,
	}
}

// Authenticate turns verified ID token claims into a session, provisioning or linking
// an account as needed. It is the single place where a login becomes an account.
//
// Resolution runs in three steps: a known (issuer, subject) authenticates its account
// directly; an unknown identity whose verified email matches an existing account is
// linked to it; anything else provisions a new account, which publishes
// account.Created so the portfolio, cashflow and assets projections are built.
//
// The steps are not wrapped in one transaction because provisioning publishes an event
// whose handlers run asynchronously and would race an uncommitted account row. A
// failure between provisioning and linking leaves an account that the next login
// adopts by email, so the flow is self-healing.
func (c *Commands) Authenticate(ctx context.Context, claims Claims) (*IssuedSession, error) {
	// Checked before resolving the account, and on every sign-in rather than only the
	// first, so removing someone from the allowlist locks them out at their next login.
	if !c.allowEmail(NormalizeEmail(claims.Email)) {
		return nil, ErrEmailNotAllowed
	}

	accountID, err := c.resolveAccount(ctx, claims)
	if err != nil {
		return nil, err
	}

	issued, err := NewSession(accountID, c.sessionTTL)
	if err != nil {
		return nil, fmt.Errorf("authenticate: mint session: %w", err)
	}
	if err := c.sessions.CreateSession(ctx, issued.Session); err != nil {
		return nil, fmt.Errorf("authenticate: persist session: %w", err)
	}
	return issued, nil
}

// Logout revokes the session behind the presented token. Revoking an unknown or
// already-revoked token is not an error -- the caller ends up logged out either way.
func (c *Commands) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := c.sessions.Revoke(ctx, HashToken(token), time.Now().UTC()); err != nil {
		return fmt.Errorf("logout: revoke session: %w", err)
	}
	return nil
}

// RevokeAll ends every session for an account, for example after the account's access
// is withdrawn.
func (c *Commands) RevokeAll(ctx context.Context, accountID uuid.UUID) error {
	if err := c.sessions.RevokeAllForAccount(ctx, accountID, time.Now().UTC()); err != nil {
		return fmt.Errorf("revoke all: %w", err)
	}
	return nil
}

func (c *Commands) resolveAccount(ctx context.Context, claims Claims) (uuid.UUID, error) {
	identity, err := c.identities.GetByIssuerSubject(ctx, claims.Issuer, claims.Subject)
	if err == nil {
		if touchErr := c.identities.TouchLastLogin(ctx, identity.ID, time.Now().UTC()); touchErr != nil {
			return uuid.Nil, fmt.Errorf("authenticate: record login: %w", touchErr)
		}
		return identity.AccountID, nil
	}
	if !errors.Is(err, ErrIdentityNotFound) {
		return uuid.Nil, fmt.Errorf("authenticate: look up identity: %w", err)
	}

	email := NormalizeEmail(claims.Email)
	if email == "" {
		return uuid.Nil, ErrEmailMissing
	}
	// Linking a provider to an existing account on an unverified address would let any
	// provider that accepts arbitrary emails claim someone else's data.
	if !claims.EmailVerified {
		return uuid.Nil, ErrEmailUnverified
	}

	accountID, err := c.findOrProvisionAccount(ctx, email)
	if err != nil {
		return uuid.Nil, err
	}

	now := time.Now().UTC()
	link := &Identity{
		ID:          uuid.New(),
		AccountID:   accountID,
		Issuer:      claims.Issuer,
		Subject:     claims.Subject,
		CreatedAt:   now,
		LastLoginAt: &now,
	}
	if err := c.identities.CreateIdentity(ctx, link); err != nil {
		// Two first logins racing each other: whichever lost re-reads the winner's row
		// rather than failing a legitimate sign-in.
		if errors.Is(err, ErrIdentityExists) {
			existing, getErr := c.identities.GetByIssuerSubject(ctx, claims.Issuer, claims.Subject)
			if getErr != nil {
				return uuid.Nil, fmt.Errorf("authenticate: reload raced identity: %w", getErr)
			}
			return existing.AccountID, nil
		}
		return uuid.Nil, fmt.Errorf("authenticate: link identity: %w", err)
	}
	return accountID, nil
}

func (c *Commands) findOrProvisionAccount(ctx context.Context, email string) (uuid.UUID, error) {
	acc, err := c.accounts.GetByEmail(ctx, email)
	if err == nil {
		return acc.ID, nil
	}
	if !errors.Is(err, account.ErrAccountNotFound) {
		return uuid.Nil, fmt.Errorf("authenticate: look up account by email: %w", err)
	}

	// The account's display name starts as the email: accounts.name is UNIQUE in both
	// dialects, and the email already is, so the two constraints coincide.
	id, err := c.provision.Create(ctx, nil, nil, email, account.WithEmail(email))
	if err != nil {
		return uuid.Nil, fmt.Errorf("authenticate: provision account: %w", err)
	}
	return *id, nil
}
