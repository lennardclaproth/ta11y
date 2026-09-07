package marketdata

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

const (
	// CataloguePageSize is how many rows one provider ticker-search call requests.
	// Provider search is relevance-ranked (a search for "ASM" returns ASML first out
	// of 125 matches), so a single page reliably contains the intended instrument.
	// The catalogue therefore never pages deeper on a user's behalf; a narrower
	// query is both cheaper and more precise than a second page.
	CataloguePageSize = 100

	// DefaultSeedPages bounds a catalogue seed sync. The provider returns its
	// empty search in popularity order, so the first pages are the most-traded
	// instruments worldwide. Twenty pages is 2000 rows for 20 provider requests.
	DefaultSeedPages = 20

	// MaxSeedPages caps how much of the provider catalogue one seed run may pull,
	// so a mistyped request cannot drain the monthly request budget. Mirroring the
	// provider's full ~683k-ticker universe would cost thousands of requests and
	// several hours, and is deliberately not offered.
	MaxSeedPages = 100
)

// ProviderListing is one cached entry from a provider's ticker catalogue. It is
// deliberately not a Listing: it records what the provider knows about a tradable
// symbol, not an instrument the user has chosen to track.
type ProviderListing struct {
	ID           uuid.UUID `db:"id"`
	Source       Source    `db:"source"`
	Symbol       string    `db:"symbol"`
	Name         *string   `db:"name"`
	ExchangeName *string   `db:"exchange_name"`
	ExchangeMIC  *string   `db:"exchange_mic"`
	HasEOD       bool      `db:"has_eod"`
	HasIntraday  bool      `db:"has_intraday"`
	FetchedAt    time.Time `db:"fetched_at"`
}

// Adoptable reports whether this catalogue entry can become a tracked Listing,
// and if not, a reason fit for display. Provider catalogues are uneven: some
// symbols carry no name at all (listings.name is NOT NULL and NewListing rejects
// an empty name), and some have no end-of-day history, so tracking them would
// never yield prices.
func (p *ProviderListing) Adoptable() (bool, string) {
	if p == nil {
		return false, "this entry is unavailable"
	}
	if p.Name == nil || strings.TrimSpace(*p.Name) == "" {
		return false, "the provider reports no name for this symbol"
	}
	if !p.HasEOD {
		return false, "the provider has no end-of-day history for this symbol"
	}
	return true, ""
}

// CatalogueScope selects which side of a combined listing search to return.
type CatalogueScope string

const (
	// ScopeTracked returns only instruments the user tracks. It is the default so
	// that existing callers of listing search are unaffected by the catalogue.
	ScopeTracked CatalogueScope = "tracked"
	// ScopeCatalogue returns only cached provider entries that are not yet tracked.
	ScopeCatalogue CatalogueScope = "catalogue"
	// ScopeAll returns both, tracked instruments first.
	ScopeAll CatalogueScope = "all"
)

// IsValid reports whether the scope is one of the known values.
func (s CatalogueScope) IsValid() bool {
	switch s {
	case ScopeTracked, ScopeCatalogue, ScopeAll:
		return true
	default:
		return false
	}
}

// CatalogueSearchResult is one row of a combined listing search: either an
// instrument the user tracks or an untracked provider-catalogue entry. ID is the
// listing id for tracked rows and the cached provider-listing id otherwise, so
// every row carries a stable identity regardless of which side it came from.
//
// Every field a tracked listing has is present and populated exactly as listing
// search has always returned it. The fields provider ticker search cannot supply
// -- description, region, currency, ISIN, ticker, type and the listing timestamps
// -- are nil on catalogue rows rather than zero-valued, so an untracked row never
// claims metadata the provider did not give us.
type CatalogueSearchResult struct {
	ID          uuid.UUID       `db:"id"`
	Symbol      string          `db:"symbol"`
	Name        *string         `db:"name"`
	Source      Source          `db:"source"`
	Description *string         `db:"description"`
	Exchange    *string         `db:"exchange"`
	ExchangeMIC *string         `db:"exchange_mic"`
	Region      *string         `db:"region"`
	Currency    *money.Currency `db:"currency"`
	ISIN        *string         `db:"isin"`
	Ticker      *string         `db:"ticker"`
	Type        *string         `db:"type"`
	HasEOD      bool            `db:"has_eod"`
	Tracked     bool            `db:"tracked"`
	CreatedAt   *time.Time      `db:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at"`
}

// Adoptable reports whether this row can be turned into a tracked Listing, and if
// not, a reason fit for display. It mirrors ProviderListing.Adoptable for rows that
// came from the catalogue side; a row that is already tracked is never adoptable
// again.
func (r *CatalogueSearchResult) Adoptable() (bool, string) {
	if r == nil {
		return false, "this entry is unavailable"
	}
	if r.Tracked {
		return false, "this instrument is already in your listings"
	}
	entry := ProviderListing{Name: r.Name, HasEOD: r.HasEOD}
	return entry.Adoptable()
}

// CatalogueSyncStatus is the lifecycle state of a catalogue seed run.
type CatalogueSyncStatus string

const (
	CatalogueSyncRunning   CatalogueSyncStatus = "running"
	CatalogueSyncCompleted CatalogueSyncStatus = "completed"
	CatalogueSyncFailed    CatalogueSyncStatus = "failed"
)

// CatalogueSync is the durable record of one bounded catalogue seed run. It holds
// no resume offset: a run is capped at MaxSeedPages, which is cheap enough to
// simply repeat rather than resume.
type CatalogueSync struct {
	ID            uuid.UUID           `db:"id"`
	Source        Source              `db:"source"`
	Status        CatalogueSyncStatus `db:"status"`
	PagesFetched  int                 `db:"pages_fetched"`
	RowsUpserted  int                 `db:"rows_upserted"`
	UpstreamTotal *int                `db:"upstream_total"`
	LastError     *string             `db:"last_error"`
	StartedAt     time.Time           `db:"started_at"`
	FinishedAt    *time.Time          `db:"finished_at"`
}

// TickerPage is one page of provider ticker-search results. Total is the provider's
// full match count, which is normally far larger than the returned page.
type TickerPage struct {
	Data   []ProviderListing
	Limit  int
	Offset int
	Total  int
}

// TickerSearcher searches a provider's ticker catalogue. Every call costs one
// provider request, so callers must treat it as a metered resource and never
// invoke it per keystroke. Returned entries carry no ID or FetchedAt; the store
// assigns those on upsert.
type TickerSearcher interface {
	SearchTickers(ctx context.Context, query string, limit, offset int) (TickerPage, error)
}

// CatalogueStore caches provider catalogue entries and records seed runs.
type CatalogueStore interface {
	UpsertProviderListings(ctx context.Context, entries []ProviderListing) (int, error)
	SearchCatalogue(ctx context.Context, q string, scope CatalogueScope, limit, offset int) ([]*CatalogueSearchResult, int, error)
	CountProviderListings(ctx context.Context, source Source) (int, error)
	CreateCatalogueSync(ctx context.Context, sync *CatalogueSync) error
	UpdateCatalogueSync(ctx context.Context, sync *CatalogueSync) error
	LatestCatalogueSync(ctx context.Context, source Source) (*CatalogueSync, error)
}

// CatalogueStatus summarises what the local catalogue holds for one source.
type CatalogueStatus struct {
	Source     Source
	Entries    int
	LatestSync *CatalogueSync
}
