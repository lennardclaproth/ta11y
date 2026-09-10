package marketdata

import (
	"errors"
)

// Sync errors
var (
	ErrReleaseSyncLock = errors.New("failed to release lock for the listing")
	ErrSyncInProgress  = errors.New("sync is already in progress for this listing")
)

// Accumulate errors
var (
	ErrShouldAccumulateFailed = errors.New("failed to set should accumulate")
)

// Listing errors
var (
	// ErrListingInUse refuses a delete that a portfolio still depends on. The listing FKs
	// cascade: deleting one takes its position snapshots -- the account's valuation history
	// for that instrument -- with it, and orphans any open position by nulling its listing.
	// That is not recoverable from the UI, so the delete is refused rather than confirmed.
	ErrListingInUse = errors.New("listing is referenced by a portfolio")
)

// Catalogue errors
var (
	ErrCatalogueQueryEmpty        = errors.New("catalogue search query cannot be empty")
	ErrCatalogueSourceUnsupported = errors.New("source does not support catalogue search")
	ErrCatalogueSyncInProgress    = errors.New("a catalogue sync is already running for this source")
	ErrCatalogueScopeInvalid      = errors.New("invalid catalogue scope")
	ErrCatalogueSeedPagesInvalid  = errors.New("catalogue seed pages out of range")
)

// Provider credential errors
var (
	ErrNoCredentialFieldsToUpdate = errors.New("no credential fields to update")
	ErrCredentialsNotConfigurable = errors.New("manual providers have no credentials to configure")
)
