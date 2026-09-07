package marketdata

import (
	"errors"
)

// Sync errors
var (
	ErrReleaseSyncLock = errors.New("Failed to release lock for the listing")
	ErrSyncInProgress  = errors.New("sync is already in progress for this listing")
)

// Accumulate errors
var (
	ErrShouldAccumulateFailed = errors.New("Failed to set should accumulate")
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
