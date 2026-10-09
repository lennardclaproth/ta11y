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
	// Purposes OR-matches the goal purpose; PurposeNone matches the rows nobody has
	// pointed at yet. Empty means no filter on purpose at all.
	Purposes []Purpose

	From *time.Time
	To   *time.Time
}
