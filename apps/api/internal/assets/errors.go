package assets

import "errors"

// Asset errors.
var (
	ErrAssetNotFound   = errors.New("asset not found")
	ErrAssetNotInClass = errors.New("asset does not belong to the specified class")
)

// Daily-priced holding errors
var (
	ErrAssetNotDailyPriced      = errors.New("asset is not linked to a daily price")
	ErrAssetDailyPriced         = errors.New("asset worth follows a daily price and cannot be set by hand")
	ErrPurchaseQuantityInvalid  = errors.New("purchase quantity must be greater than zero")
	ErrPurchaseUnitPriceInvalid = errors.New("purchase unit price cannot be negative")
	ErrPurchaseCostOutOfRange   = errors.New("purchase quantity times unit price is out of range")
	ErrPurchaseDateInFuture     = errors.New("purchase date cannot be in the future")
	ErrPurchaseNotFound         = errors.New("purchase not found")
	ErrLastPurchase             = errors.New("a daily-priced item keeps at least one purchase; delete the item instead")
)

// Worth change errors
var (
	ErrMutationNotFound  = errors.New("worth change not found")
	ErrLastMutation      = errors.New("an item keeps at least one worth change; delete the item instead")
	ErrMutationAmount    = errors.New("worth change amount cannot be negative")
	ErrMutationDirection = errors.New("an add or withdraw needs a direction")
	ErrMutationInFuture  = errors.New("worth change date cannot be in the future")
	// ErrOpeningNotSet refuses a history whose first change for an item adds or withdraws:
	// there is no worth yet to add to.
	ErrOpeningNotSet = errors.New("an item's first change must set its worth; date this change after the opening worth")
)

// Account errors
var (
	ErrAccountNotFound = errors.New("account not found")
)

// Class errors
var (
	ErrClassNotFound        = errors.New("class not found")
	ErrClassNotManual       = errors.New("asset class must be manual")
	ErrClassReserved        = errors.New("asset class cannot use a reserved name")
	ErrClassNameEmpty       = errors.New("asset class name cannot be empty")
	ErrClassAlreadyExists   = errors.New("asset class already exists")
	ErrClassAccountMismatch = errors.New("asset class does not belong to the specified account")
)

// Syncer errors
var (
	ErrSyncInProgress  = errors.New("sync already in progress")
	ErrReleaseSyncLock = errors.New("failed to release lock for the listing")
)
