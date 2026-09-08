//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/auth"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// seedAccount creates an account the auth tables can reference, since both
// identities and sessions are FK-bound to it.
func seedAccount(t *testing.T, db *storage.DB, name, email string) uuid.UUID {
	t.Helper()

	store := storage.NewSQLXAccountStore(db)
	acc, err := account.NewAccount(name, nil, nil, account.WithEmail(email))
	if err != nil {
		t.Fatalf("build account: %v", err)
	}
	if err := store.Create(t.Context(), acc); err != nil {
		t.Fatalf("create account: %v", err)
	}
	return acc.ID
}

func TestAccountStoreFindsByEmailCaseInsensitively(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "person@example.com", "Person@Example.com")
		store := storage.NewSQLXAccountStore(db)

		for _, lookup := range []string{"person@example.com", "PERSON@EXAMPLE.COM", " Person@Example.com "} {
			got, err := store.GetByEmail(t.Context(), lookup)
			if err != nil {
				t.Fatalf("GetByEmail(%q): %v", lookup, err)
			}
			if got.ID != accountID {
				t.Fatalf("GetByEmail(%q) returned the wrong account", lookup)
			}
		}

		if _, err := store.GetByEmail(t.Context(), "nobody@example.com"); !errors.Is(err, account.ErrAccountNotFound) {
			t.Fatalf("GetByEmail(unknown) = %v, want ErrAccountNotFound", err)
		}
	})
}

// The unique index is on LOWER(email), so two accounts differing only in case
// must collide.
func TestAccountStoreRejectsADuplicateEmail(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		seedAccount(t, db, "first", "person@example.com")

		store := storage.NewSQLXAccountStore(db)
		duplicate, err := account.NewAccount("second", nil, nil, account.WithEmail("PERSON@example.com"))
		if err != nil {
			t.Fatalf("build account: %v", err)
		}
		if err := store.Create(t.Context(), duplicate); !errors.Is(err, account.ErrAccountAlreadyExists) {
			t.Fatalf("Create(duplicate email) = %v, want ErrAccountAlreadyExists", err)
		}
	})
}

// Accounts predating authentication have no email, and a unique index must not
// treat several NULLs as duplicates of each other.
func TestAccountStoreAllowsManyAccountsWithoutAnEmail(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		store := storage.NewSQLXAccountStore(db)

		for _, name := range []string{"first", "second"} {
			acc, err := account.NewAccount(name, nil, nil)
			if err != nil {
				t.Fatalf("build account: %v", err)
			}
			if err := store.Create(t.Context(), acc); err != nil {
				t.Fatalf("create %s: %v", name, err)
			}
		}
	})
}

func TestAuthStoreIdentityRoundTrip(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "person@example.com", "person@example.com")
		store := storage.NewSQLXAuthStore(db)

		const issuer, subject = "https://accounts.google.com", "google-sub-1"

		if _, err := store.GetByIssuerSubject(t.Context(), issuer, subject); !errors.Is(err, auth.ErrIdentityNotFound) {
			t.Fatalf("GetByIssuerSubject(unknown) = %v, want ErrIdentityNotFound", err)
		}

		now := time.Now().UTC().Truncate(time.Second)
		identity := &auth.Identity{
			ID:        uuid.New(),
			AccountID: accountID,
			Issuer:    issuer,
			Subject:   subject,
			CreatedAt: now,
		}
		if err := store.CreateIdentity(t.Context(), identity); err != nil {
			t.Fatalf("create identity: %v", err)
		}

		got, err := store.GetByIssuerSubject(t.Context(), issuer, subject)
		if err != nil {
			t.Fatalf("GetByIssuerSubject: %v", err)
		}
		if got.AccountID != accountID || got.ID != identity.ID {
			t.Fatalf("identity = %+v, want the one just created", got)
		}
		if got.LastLoginAt != nil {
			t.Fatal("a freshly created identity has not logged in yet")
		}

		loginAt := now.Add(time.Minute)
		if err := store.TouchLastLogin(t.Context(), identity.ID, loginAt); err != nil {
			t.Fatalf("touch last login: %v", err)
		}
		got, err = store.GetByIssuerSubject(t.Context(), issuer, subject)
		if err != nil {
			t.Fatalf("GetByIssuerSubject after touch: %v", err)
		}
		if got.LastLoginAt == nil || !got.LastLoginAt.UTC().Equal(loginAt) {
			t.Fatalf("last login = %v, want %v", got.LastLoginAt, loginAt)
		}
	})
}

// The (issuer, subject) unique constraint is what makes a returning login resolve
// to the same account instead of provisioning a second one.
func TestAuthStoreRejectsADuplicateIdentity(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "person@example.com", "person@example.com")
		store := storage.NewSQLXAuthStore(db)

		const issuer, subject = "https://accounts.google.com", "google-sub-1"
		first := &auth.Identity{ID: uuid.New(), AccountID: accountID, Issuer: issuer, Subject: subject, CreatedAt: time.Now().UTC()}
		if err := store.CreateIdentity(t.Context(), first); err != nil {
			t.Fatalf("create first identity: %v", err)
		}

		second := &auth.Identity{ID: uuid.New(), AccountID: accountID, Issuer: issuer, Subject: subject, CreatedAt: time.Now().UTC()}
		if err := store.CreateIdentity(t.Context(), second); !errors.Is(err, auth.ErrIdentityExists) {
			t.Fatalf("CreateIdentity(duplicate) = %v, want ErrIdentityExists", err)
		}
	})
}

// One account may hold several identities -- that is what signing in through a
// second provider produces.
func TestAuthStoreAllowsManyIdentitiesPerAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "person@example.com", "person@example.com")
		store := storage.NewSQLXAuthStore(db)

		issuers := []string{"https://accounts.google.com", "https://login.microsoftonline.com/tenant/v2.0"}
		for i, issuer := range issuers {
			identity := &auth.Identity{
				ID:        uuid.New(),
				AccountID: accountID,
				Issuer:    issuer,
				Subject:   "subject",
				CreatedAt: time.Now().UTC(),
			}
			if err := store.CreateIdentity(t.Context(), identity); err != nil {
				t.Fatalf("create identity %d: %v", i, err)
			}
		}

		for _, issuer := range issuers {
			got, err := store.GetByIssuerSubject(t.Context(), issuer, "subject")
			if err != nil {
				t.Fatalf("GetByIssuerSubject(%s): %v", issuer, err)
			}
			if got.AccountID != accountID {
				t.Fatalf("identity for %s resolved to the wrong account", issuer)
			}
		}
	})
}

func TestAuthStoreSessionLifecycle(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "person@example.com", "person@example.com")
		store := storage.NewSQLXAuthStore(db)

		issued, err := auth.NewSession(accountID, time.Hour)
		if err != nil {
			t.Fatalf("mint session: %v", err)
		}
		if err := store.CreateSession(t.Context(), issued.Session); err != nil {
			t.Fatalf("create session: %v", err)
		}

		got, err := store.GetByTokenHash(t.Context(), auth.HashToken(issued.Token))
		if err != nil {
			t.Fatalf("GetByTokenHash: %v", err)
		}
		if got.AccountID != accountID {
			t.Fatal("session resolved to the wrong account")
		}
		if !got.Active(time.Now().UTC()) {
			t.Fatal("a freshly issued session must be active")
		}

		// The raw token must not be a lookup key -- only its hash is stored.
		if _, err := store.GetByTokenHash(t.Context(), issued.Token); !errors.Is(err, auth.ErrSessionNotFound) {
			t.Fatalf("GetByTokenHash(raw token) = %v, want ErrSessionNotFound", err)
		}

		revokedAt := time.Now().UTC().Truncate(time.Second)
		if err := store.Revoke(t.Context(), auth.HashToken(issued.Token), revokedAt); err != nil {
			t.Fatalf("revoke: %v", err)
		}
		got, err = store.GetByTokenHash(t.Context(), auth.HashToken(issued.Token))
		if err != nil {
			t.Fatalf("GetByTokenHash after revoke: %v", err)
		}
		if got.Active(time.Now().UTC()) {
			t.Fatal("a revoked session must not be active")
		}
	})
}

func TestAuthStoreRevokeAllForAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "person@example.com", "person@example.com")
		other := seedAccount(t, db, "other@example.com", "other@example.com")
		store := storage.NewSQLXAuthStore(db)

		tokens := make([]string, 0, 2)
		for i := 0; i < 2; i++ {
			issued, err := auth.NewSession(accountID, time.Hour)
			if err != nil {
				t.Fatalf("mint session: %v", err)
			}
			if err := store.CreateSession(t.Context(), issued.Session); err != nil {
				t.Fatalf("create session: %v", err)
			}
			tokens = append(tokens, issued.Token)
		}

		untouched, err := auth.NewSession(other, time.Hour)
		if err != nil {
			t.Fatalf("mint other session: %v", err)
		}
		if err := store.CreateSession(t.Context(), untouched.Session); err != nil {
			t.Fatalf("create other session: %v", err)
		}

		if err := store.RevokeAllForAccount(t.Context(), accountID, time.Now().UTC()); err != nil {
			t.Fatalf("revoke all: %v", err)
		}

		for i, token := range tokens {
			got, err := store.GetByTokenHash(t.Context(), auth.HashToken(token))
			if err != nil {
				t.Fatalf("GetByTokenHash(%d): %v", i, err)
			}
			if got.Active(time.Now().UTC()) {
				t.Fatalf("session %d should have been revoked", i)
			}
		}

		// Revoking one account's sessions must not touch another's.
		got, err := store.GetByTokenHash(t.Context(), auth.HashToken(untouched.Token))
		if err != nil {
			t.Fatalf("GetByTokenHash(other): %v", err)
		}
		if !got.Active(time.Now().UTC()) {
			t.Fatal("another account's session must survive")
		}
	})
}

func TestAuthStoreDeleteExpiredSessions(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedAccount(t, db, "person@example.com", "person@example.com")
		store := storage.NewSQLXAuthStore(db)

		// A non-positive TTL means "use the default", so an already-lapsed session is
		// built by moving its expiry back rather than by asking for a negative TTL.
		expired, err := auth.NewSession(accountID, time.Hour)
		if err != nil {
			t.Fatalf("mint expired session: %v", err)
		}
		expired.Session.ExpiresAt = time.Now().UTC().Add(-time.Hour)
		if err := store.CreateSession(t.Context(), expired.Session); err != nil {
			t.Fatalf("create expired session: %v", err)
		}
		live, err := auth.NewSession(accountID, time.Hour)
		if err != nil {
			t.Fatalf("mint live session: %v", err)
		}
		if err := store.CreateSession(t.Context(), live.Session); err != nil {
			t.Fatalf("create live session: %v", err)
		}

		removed, err := store.DeleteExpiredSessions(t.Context(), time.Now().UTC())
		if err != nil {
			t.Fatalf("delete expired: %v", err)
		}
		if removed != 1 {
			t.Fatalf("removed = %d, want 1", removed)
		}

		if _, err := store.GetByTokenHash(t.Context(), auth.HashToken(expired.Token)); !errors.Is(err, auth.ErrSessionNotFound) {
			t.Fatalf("expired session = %v, want ErrSessionNotFound", err)
		}
		if _, err := store.GetByTokenHash(t.Context(), auth.HashToken(live.Token)); err != nil {
			t.Fatalf("live session should survive: %v", err)
		}
	})
}
