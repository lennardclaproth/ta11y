package account

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Account is the canonical account aggregate shared across domains.
type Account struct {
	ID uuid.UUID `db:"id"`
	// ExternalID in the account domain is an optional identifier that can come from an external system (i.e. Google or EntraID)
	ExternalID *string `db:"external_id"`
	Name       string  `db:"name"`
	// Email is the only personal data the application keeps about a person. It is
	// optional so accounts predating authentication remain valid, and is matched
	// case-insensitively against a LOWER(email) unique index.
	Email *string `db:"email"`
	// Admin marks accounts allowed to reach admin-only screens. The API is
	// unauthenticated today, so this records intent and does not enforce it.
	Admin     bool      `db:"admin"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// AccountOption mutates optional account fields during construction.
type AccountOption func(*Account)

// AsAdmin marks the account as an administrator.
func AsAdmin() AccountOption {
	return func(a *Account) { a.Admin = true }
}

// WithEmail sets the account's email, normalized to lowercase so it matches the
// case-insensitive unique index. An empty email leaves the account without one.
func WithEmail(email string) AccountOption {
	return func(a *Account) {
		normalized := strings.ToLower(strings.TrimSpace(email))
		if normalized == "" {
			return
		}
		a.Email = &normalized
	}
}

// NewAccount constructs a validated account instance with generated identity and timestamps.
func NewAccount(name string, id *uuid.UUID, externalID *string, options ...AccountOption) (*Account, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrAccountNameRequired
	}

	resolvedID := uuid.New()
	if id != nil && *id != uuid.Nil {
		resolvedID = *id
	}

	now := time.Now().UTC()

	acc := &Account{
		ID:         resolvedID,
		ExternalID: externalID,
		Name:       trimmed,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	for _, option := range options {
		option(acc)
	}
	return acc, nil
}
