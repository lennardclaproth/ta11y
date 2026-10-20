package cashflow

import (
	"time"

	"github.com/google/uuid"
)

// Account is the cashflow-domain projection for a base account. Every cashflow
// transaction references it, so it must exist before anything can be imported or
// entered for the account.
type Account struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// NewAccount builds the cashflow projection for the given base account.
func NewAccount(accountID uuid.UUID) *Account {
	now := time.Now().UTC()
	return &Account{
		ID:        uuid.New(),
		AccountID: accountID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
