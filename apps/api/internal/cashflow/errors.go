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

// Ignore rule Errors
var (
	ErrIgnoreRuleNameRequired      = fmt.Errorf("name is required")
	ErrIgnoreRuleNameTooLong       = fmt.Errorf("name must be at most 120 characters")
	ErrIgnoreRuleInvalidMatchField = fmt.Errorf("match_field must be either description or note")
	ErrIgnoreRuleInvalidDirection  = fmt.Errorf("direction must be empty, in, or out")
	ErrIgnoreRuleContainsTooShort  = fmt.Errorf("contains must be at least 3 characters")
	ErrIgnoreRuleContainsTooLong   = fmt.Errorf("contains must be at most 255 characters")
	// ErrIgnoreRuleContainsWildcard explains the refusal rather than silently bending
	// the text: % and _ are LIKE wildcards, so "a%z" would pass the minimum length and
	// still match nearly every statement.
	ErrIgnoreRuleContainsWildcard = fmt.Errorf("contains must not use %% or _, which match any text")
	ErrIgnoreRuleNotFound         = fmt.Errorf("ignore rule not found")
)

// Account Errors
var (
	ErrAccountNotFound = fmt.Errorf("account not found")
)
