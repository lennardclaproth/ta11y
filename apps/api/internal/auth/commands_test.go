package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/auth"
)

// memoryAccountStore is an in-memory account store standing in for internal/storage.
type memoryAccountStore struct {
	accounts map[uuid.UUID]*account.Account
}

func newMemoryAccountStore() *memoryAccountStore {
	return &memoryAccountStore{accounts: make(map[uuid.UUID]*account.Account)}
}

func (m *memoryAccountStore) Create(_ context.Context, acc *account.Account) error {
	for _, existing := range m.accounts {
		if existing.Email != nil && acc.Email != nil && strings.EqualFold(*existing.Email, *acc.Email) {
			return account.ErrAccountAlreadyExists
		}
	}
	copied := *acc
	m.accounts[acc.ID] = &copied
	return nil
}

func (m *memoryAccountStore) SetEmail(_ context.Context, id uuid.UUID, email string) error {
	acc, ok := m.accounts[id]
	if !ok {
		return account.ErrAccountNotFound
	}
	acc.Email = &email
	return nil
}

func (m *memoryAccountStore) GetByID(_ context.Context, id uuid.UUID) (*account.Account, error) {
	if acc, ok := m.accounts[id]; ok {
		return acc, nil
	}
	return nil, account.ErrAccountNotFound
}

func (m *memoryAccountStore) GetByEmail(_ context.Context, email string) (*account.Account, error) {
	for _, acc := range m.accounts {
		if acc.Email != nil && strings.EqualFold(*acc.Email, email) {
			return acc, nil
		}
	}
	return nil, account.ErrAccountNotFound
}

func (m *memoryAccountStore) List(_ context.Context) ([]*account.Account, error) {
	out := make([]*account.Account, 0, len(m.accounts))
	for _, acc := range m.accounts {
		out = append(out, acc)
	}
	return out, nil
}

func (m *memoryAccountStore) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.accounts[id]
	return ok, nil
}

// memoryAuthStore is an in-memory identity and session store.
type memoryAuthStore struct {
	identities []*auth.Identity
	sessions   []*auth.Session
}

func (m *memoryAuthStore) GetByIssuerSubject(_ context.Context, issuer, subject string) (*auth.Identity, error) {
	for _, identity := range m.identities {
		if identity.Issuer == issuer && identity.Subject == subject {
			return identity, nil
		}
	}
	return nil, auth.ErrIdentityNotFound
}

func (m *memoryAuthStore) CreateIdentity(_ context.Context, identity *auth.Identity) error {
	for _, existing := range m.identities {
		if existing.Issuer == identity.Issuer && existing.Subject == identity.Subject {
			return auth.ErrIdentityExists
		}
	}
	m.identities = append(m.identities, identity)
	return nil
}

func (m *memoryAuthStore) TouchLastLogin(_ context.Context, id uuid.UUID, at time.Time) error {
	for _, identity := range m.identities {
		if identity.ID == id {
			identity.LastLoginAt = &at
		}
	}
	return nil
}

func (m *memoryAuthStore) CreateSession(_ context.Context, session *auth.Session) error {
	m.sessions = append(m.sessions, session)
	return nil
}

func (m *memoryAuthStore) GetByTokenHash(_ context.Context, tokenHash string) (*auth.Session, error) {
	for _, session := range m.sessions {
		if session.TokenHash == tokenHash {
			return session, nil
		}
	}
	return nil, auth.ErrSessionNotFound
}

func (m *memoryAuthStore) Revoke(_ context.Context, tokenHash string, at time.Time) error {
	for _, session := range m.sessions {
		if session.TokenHash == tokenHash && session.RevokedAt == nil {
			session.RevokedAt = &at
		}
	}
	return nil
}

func (m *memoryAuthStore) RevokeAllForAccount(_ context.Context, accountID uuid.UUID, at time.Time) error {
	for _, session := range m.sessions {
		if session.AccountID == accountID && session.RevokedAt == nil {
			session.RevokedAt = &at
		}
	}
	return nil
}

func (m *memoryAuthStore) TouchLastSeen(_ context.Context, id uuid.UUID, at time.Time) error {
	for _, session := range m.sessions {
		if session.ID == id {
			session.LastSeenAt = at
		}
	}
	return nil
}

type harness struct {
	commands *auth.Commands
	queries  *auth.Queries
	authDB   *memoryAuthStore
	accounts *memoryAccountStore
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	return newHarnessAllowing(t, nil)
}

// newHarnessAllowing builds a harness whose sign-in allowlist holds exactly the given
// addresses. A nil list admits everyone.
func newHarnessAllowing(t *testing.T, allowed []string) *harness {
	t.Helper()
	accounts := newMemoryAccountStore()
	authDB := &memoryAuthStore{}
	accountQueries := account.NewQueries(accounts)
	accountCommands := account.NewCommands(accounts, nil)

	var allow auth.EmailAllowFunc
	if allowed != nil {
		allow = func(email string) bool {
			for _, candidate := range allowed {
				if strings.EqualFold(candidate, email) {
					return true
				}
			}
			return false
		}
	}

	return &harness{
		commands: auth.NewCommands(authDB, authDB, accountQueries, accountCommands, time.Hour, allow),
		queries:  auth.NewQueries(authDB, accountQueries),
		authDB:   authDB,
		accounts: accounts,
	}
}

func googleClaims(subject, email string, verified bool) auth.Claims {
	return auth.Claims{
		Issuer:        "https://accounts.google.com",
		Subject:       subject,
		Email:         email,
		EmailVerified: verified,
	}
}

func TestAuthenticateProvisionsAnAccountOnFirstLogin(t *testing.T) {
	h := newHarness(t)

	issued, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "person@example.com", true))
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	acc, err := h.accounts.GetByEmail(t.Context(), "person@example.com")
	if err != nil {
		t.Fatalf("expected an account to have been provisioned: %v", err)
	}
	if acc.ID != issued.Session.AccountID {
		t.Fatal("the session must belong to the provisioned account")
	}
	// The display name starts as the email because accounts.name is UNIQUE.
	if acc.Name != "person@example.com" {
		t.Fatalf("account name = %q, want the email", acc.Name)
	}
	if len(h.authDB.identities) != 1 {
		t.Fatalf("identities = %d, want 1", len(h.authDB.identities))
	}
}

func TestAuthenticateReusesTheAccountOnReturningLogin(t *testing.T) {
	h := newHarness(t)
	claims := googleClaims("sub-1", "person@example.com", true)

	first, err := h.commands.Authenticate(t.Context(), claims)
	if err != nil {
		t.Fatalf("first login: %v", err)
	}
	second, err := h.commands.Authenticate(t.Context(), claims)
	if err != nil {
		t.Fatalf("second login: %v", err)
	}

	if first.Session.AccountID != second.Session.AccountID {
		t.Fatal("a returning login must resolve to the same account")
	}
	if first.Token == second.Token {
		t.Fatal("each login must mint a fresh session token")
	}
	if len(h.authDB.identities) != 1 {
		t.Fatalf("identities = %d, want the existing one to be reused", len(h.authDB.identities))
	}
	if len(h.accounts.accounts) != 1 {
		t.Fatalf("accounts = %d, want 1", len(h.accounts.accounts))
	}
}

func TestAuthenticateLinksASecondProviderByVerifiedEmail(t *testing.T) {
	h := newHarness(t)

	viaGoogle, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "person@example.com", true))
	if err != nil {
		t.Fatalf("google login: %v", err)
	}

	viaEntra, err := h.commands.Authenticate(t.Context(), auth.Claims{
		Issuer:        "https://login.microsoftonline.com/tenant/v2.0",
		Subject:       "entra-sub",
		Email:         "Person@Example.com",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("entra login: %v", err)
	}

	if viaGoogle.Session.AccountID != viaEntra.Session.AccountID {
		t.Fatal("a second provider with the same verified email must link to the same account")
	}
	if len(h.authDB.identities) != 2 {
		t.Fatalf("identities = %d, want one per provider", len(h.authDB.identities))
	}
	if len(h.accounts.accounts) != 1 {
		t.Fatalf("accounts = %d, want the existing account to be reused", len(h.accounts.accounts))
	}
}

// An identity provider that lets a user assert an arbitrary unverified address must not
// be able to claim an account that already belongs to that address.
func TestAuthenticateRefusesToLinkAnUnverifiedEmail(t *testing.T) {
	h := newHarness(t)

	if _, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "person@example.com", true)); err != nil {
		t.Fatalf("google login: %v", err)
	}

	_, err := h.commands.Authenticate(t.Context(), auth.Claims{
		Issuer:        "https://sketchy.example",
		Subject:       "attacker",
		Email:         "person@example.com",
		EmailVerified: false,
	})
	if !errors.Is(err, auth.ErrEmailUnverified) {
		t.Fatalf("error = %v, want ErrEmailUnverified", err)
	}
	if len(h.authDB.identities) != 1 {
		t.Fatal("a refused login must not leave an identity behind")
	}
	if len(h.authDB.sessions) != 1 {
		t.Fatal("a refused login must not mint a session")
	}
}

func TestAuthenticateRequiresAnEmail(t *testing.T) {
	h := newHarness(t)

	_, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "   ", true))
	if !errors.Is(err, auth.ErrEmailMissing) {
		t.Fatalf("error = %v, want ErrEmailMissing", err)
	}
	if len(h.accounts.accounts) != 0 {
		t.Fatal("no account should be provisioned without an email")
	}
}

func TestResolveSession(t *testing.T) {
	h := newHarness(t)
	issued, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "person@example.com", true))
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	t.Run("resolves a live session", func(t *testing.T) {
		principal, err := h.queries.ResolveSession(t.Context(), issued.Token)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if principal.AccountID != issued.Session.AccountID {
			t.Fatal("resolved the wrong account")
		}
		if principal.Email != "person@example.com" {
			t.Fatalf("email = %q, want person@example.com", principal.Email)
		}
	})

	t.Run("rejects an unknown token", func(t *testing.T) {
		if _, err := h.queries.ResolveSession(t.Context(), "not-a-real-token"); !errors.Is(err, auth.ErrSessionNotFound) {
			t.Fatalf("error = %v, want ErrSessionNotFound", err)
		}
	})

	t.Run("rejects an empty token", func(t *testing.T) {
		if _, err := h.queries.ResolveSession(t.Context(), ""); !errors.Is(err, auth.ErrSessionNotFound) {
			t.Fatalf("error = %v, want ErrSessionNotFound", err)
		}
	})

	t.Run("rejects a session after logout", func(t *testing.T) {
		if err := h.commands.Logout(t.Context(), issued.Token); err != nil {
			t.Fatalf("logout: %v", err)
		}
		if _, err := h.queries.ResolveSession(t.Context(), issued.Token); !errors.Is(err, auth.ErrSessionExpired) {
			t.Fatalf("error = %v, want ErrSessionExpired", err)
		}
	})
}

func TestResolveSessionRejectsAnExpiredSession(t *testing.T) {
	h := newHarness(t)
	issued, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "person@example.com", true))
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	h.authDB.sessions[0].ExpiresAt = time.Now().UTC().Add(-time.Minute)

	if _, err := h.queries.ResolveSession(t.Context(), issued.Token); !errors.Is(err, auth.ErrSessionExpired) {
		t.Fatalf("error = %v, want ErrSessionExpired", err)
	}
}

func TestRevokeAllEndsEverySessionForAnAccount(t *testing.T) {
	h := newHarness(t)
	claims := googleClaims("sub-1", "person@example.com", true)

	first, err := h.commands.Authenticate(t.Context(), claims)
	if err != nil {
		t.Fatalf("first login: %v", err)
	}
	second, err := h.commands.Authenticate(t.Context(), claims)
	if err != nil {
		t.Fatalf("second login: %v", err)
	}

	if err := h.commands.RevokeAll(t.Context(), first.Session.AccountID); err != nil {
		t.Fatalf("revoke all: %v", err)
	}

	for name, token := range map[string]string{"first": first.Token, "second": second.Token} {
		if _, err := h.queries.ResolveSession(t.Context(), token); !errors.Is(err, auth.ErrSessionExpired) {
			t.Fatalf("%s session error = %v, want ErrSessionExpired", name, err)
		}
	}
}

func TestLogoutWithoutASessionIsNotAnError(t *testing.T) {
	h := newHarness(t)
	if err := h.commands.Logout(t.Context(), ""); err != nil {
		t.Fatalf("logout without a session = %v, want nil", err)
	}
}

// Authenticating with the provider proves identity; the allowlist decides admission.
func TestAuthenticateRefusesAnAddressOutsideTheAllowlist(t *testing.T) {
	h := newHarnessAllowing(t, []string{"owner@example.com"})

	_, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "stranger@example.com", true))
	if !errors.Is(err, auth.ErrEmailNotAllowed) {
		t.Fatalf("error = %v, want ErrEmailNotAllowed", err)
	}
	if len(h.accounts.accounts) != 0 {
		t.Fatal("a refused sign-in must not provision an account")
	}
	if len(h.authDB.sessions) != 0 {
		t.Fatal("a refused sign-in must not mint a session")
	}
}

func TestAuthenticateAdmitsAnAllowlistedAddress(t *testing.T) {
	h := newHarnessAllowing(t, []string{"Owner@Example.com"})

	if _, err := h.commands.Authenticate(t.Context(), googleClaims("sub-1", "owner@example.com", true)); err != nil {
		t.Fatalf("allowlisted sign-in = %v, want nil", err)
	}
}

// Removing someone from the allowlist must lock them out at their next sign-in, not
// only block first-time provisioning.
func TestAllowlistIsEnforcedOnReturningLogins(t *testing.T) {
	accounts := newMemoryAccountStore()
	authDB := &memoryAuthStore{}
	accountQueries := account.NewQueries(accounts)
	accountCommands := account.NewCommands(accounts, nil)

	permitted := true
	commands := auth.NewCommands(authDB, authDB, accountQueries, accountCommands, time.Hour,
		func(string) bool { return permitted })

	claims := googleClaims("sub-1", "owner@example.com", true)
	if _, err := commands.Authenticate(t.Context(), claims); err != nil {
		t.Fatalf("first sign-in: %v", err)
	}

	permitted = false
	if _, err := commands.Authenticate(t.Context(), claims); !errors.Is(err, auth.ErrEmailNotAllowed) {
		t.Fatalf("returning sign-in = %v, want ErrEmailNotAllowed", err)
	}
}
