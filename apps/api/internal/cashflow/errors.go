package cashflow

import "fmt"

// Transaction Errors
var (
	ErrTransactionsRequired              = fmt.Errorf("transactions are required")
	ErrTransactionLimitExceeded          = fmt.Errorf("transactions exceeds maximum batch size")
	ErrManualCashflowInvalidDate         = fmt.Errorf("date must be in YYYY-MM-DD format")
	ErrCashflowDateInFuture              = fmt.Errorf("date must be today or earlier")
	ErrCashflowDateNotEditable           = fmt.Errorf("only manually entered transactions can be moved to another date")
	ErrManualCashflowInvalidAmount       = fmt.Errorf("amount must be a positive decimal string with up to 6 decimals")
	ErrManualCashflowInvalidType         = fmt.Errorf("type must be one of: in, out, income, expense")
	ErrManualCashflowDescriptionRequired = fmt.Errorf("description is required")
	ErrManualCashflowNoteRequired        = fmt.Errorf("note is required")
	ErrManualCashflowTagRequired         = fmt.Errorf("tag is required")
)

// Recurring item Errors
var (
	ErrRecurringNameRequired         = fmt.Errorf("name is required")
	ErrRecurringInvalidRhythm        = fmt.Errorf("rhythm must be one of: monthly, quarterly, yearly")
	ErrRecurringInvalidMonth         = fmt.Errorf("month must be in YYYY-MM format")
	ErrRecurringItemNotFound         = fmt.Errorf("no recurring item found with the given ID")
	ErrRecurringItemExists           = fmt.Errorf("a recurring item with that name already exists")
	ErrRecurringItemEnded            = fmt.Errorf("an ended recurring item takes no new transactions")
	ErrRecurringTransactionsRequired = fmt.Errorf("at least one transaction is required")
	ErrRecurringTransactionNotFound  = fmt.Errorf("transaction not found in this account")
	ErrRecurringTransactionLinked    = fmt.Errorf("transaction already belongs to a recurring item")
	ErrRecurringSuggestionNotFound   = fmt.Errorf("no suggestion found for that pattern")
)

// Account Errors
var (
	ErrAccountNotFound = fmt.Errorf("account not found")
)
