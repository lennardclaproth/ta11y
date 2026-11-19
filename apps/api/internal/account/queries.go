package account

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Queries struct {
	qs QueryStore
}

// QueryStore reads account records for query use cases.
type QueryStore interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	List(ctx context.Context) ([]*Account, error)
}

// NewQueries creates account read-side use cases.
func NewQueries(qs QueryStore) *Queries {
	return &Queries{
		qs: qs,
	}
}

func (q *Queries) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return q.qs.Exists(ctx, id)
}

func (q *Queries) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	return q.qs.GetByID(ctx, id)
}

// GetByEmail returns the account owning the given email, matched case-insensitively,
// or ErrAccountNotFound. It is how a new identity provider is linked to an existing
// account.
func (q *Queries) GetByEmail(ctx context.Context, email string) (*Account, error) {
	return q.qs.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

// List returns all accounts.
func (q *Queries) List(ctx context.Context) ([]*Account, error) {
	return q.qs.List(ctx)
}
