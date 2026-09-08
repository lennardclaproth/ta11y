package cashflow

import (
	"time"

	"github.com/google/uuid"
)

type TransactionFilters struct {
	// AccountID scopes every filtered read and bulk mutation to one account.
	AccountID   uuid.UUID
	Query       string
	Description string
	Note        string
	Source      string

	Direction   *CashFlowDirection
	Tags        []string
	Untagged    *bool
	HideIgnored *bool

	From *time.Time
	To   *time.Time
}
