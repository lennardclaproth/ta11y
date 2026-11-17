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
	// SourceExact narrows to one bank by name rather than by text it contains. It is
	// what a chosen scope means: a substring match could only ever catch more than was
	// chosen, and silently.
	SourceExact string

	Direction   *CashFlowDirection
	Tags        []string
	Untagged    *bool
	HideIgnored *bool
	// Purposes OR-matches the goal purpose; PurposeNone matches the rows nobody has
	// pointed at yet. Empty means no filter on purpose at all.
	Purposes []Purpose

	// ImportID narrows to the rows one import brought in. Duplicates keep the import
	// they first arrived with, so this is exactly "new in that import".
	ImportID *uuid.UUID
	// IgnoredByRuleID narrows to the rows one ignore rule ignored.
	IgnoredByRuleID *uuid.UUID

	From *time.Time
	To   *time.Time
}
