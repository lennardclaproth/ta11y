package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/my-finances-tracker/internal/auth"
)

func TestNewSessionStoresOnlyTheTokenHash(t *testing.T) {
	issued, err := auth.NewSession(uuid.New(), time.Hour)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	if issued.Token == "" {
		t.Fatal("expected a raw token to hand to the client")
	}
	if issued.Session.TokenHash == issued.Token {
		t.Fatal("the raw token must never be what is persisted")
	}
	if got, want := issued.Session.TokenHash, auth.HashToken(issued.Token); got != want {
		t.Fatalf("token hash = %q, want %q", got, want)
	}
	if len(issued.Session.TokenHash) != 64 {
		t.Fatalf("token hash length = %d, want 64 hex characters", len(issued.Session.TokenHash))
	}
}

func TestNewSessionTokensAreUnique(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		issued, err := auth.NewSession(uuid.New(), time.Hour)
		if err != nil {
			t.Fatalf("new session: %v", err)
		}
		if _, duplicate := seen[issued.Token]; duplicate {
			t.Fatal("session tokens must not repeat")
		}
		seen[issued.Token] = struct{}{}
	}
}

func TestNewSessionFallsBackToDefaultTTL(t *testing.T) {
	issued, err := auth.NewSession(uuid.New(), 0)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	got := issued.Session.ExpiresAt.Sub(issued.Session.IssuedAt)
	if got != auth.DefaultSessionTTL {
		t.Fatalf("ttl = %v, want %v", got, auth.DefaultSessionTTL)
	}
}

func TestSessionActive(t *testing.T) {
	now := time.Now().UTC()
	revoked := now.Add(-time.Minute)

	cases := []struct {
		name    string
		session auth.Session
		want    bool
	}{
		{"live", auth.Session{ExpiresAt: now.Add(time.Hour)}, true},
		{"expired", auth.Session{ExpiresAt: now.Add(-time.Hour)}, false},
		{"revoked", auth.Session{ExpiresAt: now.Add(time.Hour), RevokedAt: &revoked}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.session.Active(now); got != tc.want {
				t.Fatalf("Active() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	cases := map[string]string{
		"  Person@Example.COM ": "person@example.com",
		"person@example.com":    "person@example.com",
		"   ":                   "",
	}
	for input, want := range cases {
		if got := auth.NormalizeEmail(input); got != want {
			t.Fatalf("NormalizeEmail(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPrincipalRoundTripsThroughContext(t *testing.T) {
	want := &auth.Principal{AccountID: uuid.New(), Email: "person@example.com", Admin: true}
	ctx := auth.WithPrincipal(t.Context(), want)

	got, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		t.Fatal("expected a principal on the context")
	}
	if got.AccountID != want.AccountID || got.Email != want.Email || !got.Admin {
		t.Fatalf("principal = %+v, want %+v", got, want)
	}

	accountID, ok := auth.AccountFromContext(ctx)
	if !ok || accountID != want.AccountID {
		t.Fatalf("AccountFromContext() = %v, %v; want %v, true", accountID, ok, want.AccountID)
	}
}

func TestAccountFromUnauthenticatedContext(t *testing.T) {
	if _, ok := auth.AccountFromContext(t.Context()); ok {
		t.Fatal("an unauthenticated context must not yield an account")
	}
	if _, ok := auth.PrincipalFromContext(t.Context()); ok {
		t.Fatal("an unauthenticated context must not yield a principal")
	}
}
