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

	// ImportID narrows to the rows one import brought in. Duplicates keep the import
	// they first arrived with, so this is exactly "new in that import".
	ImportID *uuid.UUID
	// IgnoredByRuleID narrows to the rows one ignore rule ignored.
	IgnoredByRuleID *uuid.UUID

	From *time.Time
	To   *time.Time
}
