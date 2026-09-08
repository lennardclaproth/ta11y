package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lennardclaproth/ta11y/internal/account"
)

// touchInterval is how stale a session's last-seen timestamp may get before it is
// refreshed. Writing on every request would turn each authenticated read into a write.
const touchInterval = time.Hour

// Queries exposes authentication read-side use cases.
type Queries struct {
	sessions SessionStore
	accounts *account.Queries
}

// NewQueries creates authentication read-side use cases.
func NewQueries(sessions SessionStore, accounts *account.Queries) *Queries {
	return &Queries{sessions: sessions, accounts: accounts}
}

// ResolveSession turns a raw session token into the authenticated caller. It returns
// ErrSessionNotFound for an unknown token and ErrSessionExpired for one that has lapsed
// or been revoked, so callers can distinguish "never logged in" from "log in again".
func (q *Queries) ResolveSession(ctx context.Context, token string) (*Principal, error) {
	if token == "" {
		return nil, ErrSessionNotFound
	}

	session, err := q.sessions.GetByTokenHash(ctx, HashToken(token))
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if !session.Active(now) {
		return nil, ErrSessionExpired
	}

	acc, err := q.accounts.GetByID(ctx, session.AccountID)
	if err != nil {
		// A session outliving its account is a deleted account, not a server fault.
		if errors.Is(err, account.ErrAccountNotFound) {
			return nil, ErrSessionExpired
		}
		return nil, fmt.Errorf("resolve session: load account: %w", err)
	}

	if now.Sub(session.LastSeenAt) > touchInterval {
		if err := q.sessions.TouchLastSeen(ctx, session.ID, now); err != nil {
			return nil, fmt.Errorf("resolve session: record activity: %w", err)
		}
	}

	email := ""
	if acc.Email != nil {
		email = *acc.Email
	}
	return &Principal{AccountID: acc.ID, Email: email, Admin: acc.Admin}, nil
}
