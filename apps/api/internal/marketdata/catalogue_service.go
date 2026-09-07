package marketdata

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Catalogue searches provider ticker catalogues and caches what it finds locally,
// so repeat searches cost no provider requests.
//
// It exists because a provider's search is a server-side substring match over its
// whole universe: searching "ASML" returns 31 rows while "ASM" returns 125, so a
// cache of past results can never be assumed complete for a new query. The
// catalogue therefore fills from two directions -- a bounded popularity seed, and
// an explicit per-query top-up -- and search over the cache never silently
// substitutes for asking the provider.
type Catalogue struct {
	cs CatalogueStore
	ts map[Source]TickerSearcher
}

// NewCatalogue creates the catalogue service using ticker searchers keyed by source.
func NewCatalogue(cs CatalogueStore, searchers map[Source]TickerSearcher) *Catalogue {
	if searchers == nil {
		searchers = map[Source]TickerSearcher{}
	}
	return &Catalogue{cs: cs, ts: searchers}
}

// RefreshResult reports what one metered top-up search returned. Truncated is true
// when the provider matched more rows than the single page fetched, which callers
// must surface rather than presenting the page as the complete answer.
type RefreshResult struct {
	Fetched       int
	Upserted      int
	UpstreamTotal int
	Truncated     bool
}

// Refresh runs one metered provider search for q and merges the results into the
// local catalogue. It deliberately fetches a single page: provider results are
// relevance-ranked, so the intended instrument is on the first page, and paging
// deeper multiplies request cost for rows nobody scrolls to.
func (c *Catalogue) Refresh(ctx context.Context, source Source, q string) (*RefreshResult, error) {
	query := strings.TrimSpace(q)
	if query == "" {
		return nil, ErrCatalogueQueryEmpty
	}
	searcher, ok := c.ts[source]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrCatalogueSourceUnsupported, source)
	}

	page, err := searcher.SearchTickers(ctx, query, CataloguePageSize, 0)
	if err != nil {
		return nil, fmt.Errorf("catalogue refresh: failed to search provider tickers: %w", err)
	}
	upserted, err := c.cs.UpsertProviderListings(ctx, page.Data)
	if err != nil {
		return nil, fmt.Errorf("catalogue refresh: failed to cache provider listings: %w", err)
	}

	return &RefreshResult{
		Fetched:       len(page.Data),
		Upserted:      upserted,
		UpstreamTotal: page.Total,
		Truncated:     page.Total > len(page.Data),
	}, nil
}

// StartSeed launches a bounded catalogue seed run in the background and returns the
// run record immediately. The run pages the provider's empty search, which comes
// back in popularity order, so a small page budget yields the instruments most
// people actually hold.
//
// It runs detached from the caller's context: seeding outlives the HTTP request
// that asks for it, and progress is observable through LatestSync.
func (c *Catalogue) StartSeed(ctx context.Context, source Source, pages int) (*CatalogueSync, error) {
	if pages <= 0 {
		pages = DefaultSeedPages
	}
	if pages > MaxSeedPages {
		return nil, fmt.Errorf("%w: %d exceeds the %d page cap", ErrCatalogueSeedPagesInvalid, pages, MaxSeedPages)
	}
	if _, ok := c.ts[source]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrCatalogueSourceUnsupported, source)
	}

	latest, err := c.cs.LatestCatalogueSync(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("catalogue seed: failed to read latest sync: %w", err)
	}
	if latest != nil && latest.Status == CatalogueSyncRunning {
		return nil, fmt.Errorf("%w: %s", ErrCatalogueSyncInProgress, source)
	}

	run := &CatalogueSync{
		ID:        uuid.New(),
		Source:    source,
		Status:    CatalogueSyncRunning,
		StartedAt: time.Now().UTC(),
	}
	if err := c.cs.CreateCatalogueSync(ctx, run); err != nil {
		return nil, fmt.Errorf("catalogue seed: failed to record sync: %w", err)
	}

	go c.runSeed(run, source, pages)
	return run, nil
}

// runSeed pages the provider's popularity-ordered catalogue and records progress on
// the run row. Each page is committed as it arrives, so a run that fails partway
// still leaves the catalogue better populated than before.
func (c *Catalogue) runSeed(run *CatalogueSync, source Source, pages int) {
	// Detached from the request context on purpose: the caller has already been
	// answered, and cancelling mid-page would leave the run row stuck at "running".
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pages)*time.Minute)
	defer cancel()

	searcher := c.ts[source]
	for page := 0; page < pages; page++ {
		result, err := searcher.SearchTickers(ctx, "", CataloguePageSize, page*CataloguePageSize)
		if err != nil {
			c.finishSeed(ctx, run, CatalogueSyncFailed, fmt.Errorf("page %d: %w", page, err))
			return
		}
		if run.UpstreamTotal == nil {
			total := result.Total
			run.UpstreamTotal = &total
		}
		if len(result.Data) == 0 {
			break
		}

		upserted, err := c.cs.UpsertProviderListings(ctx, result.Data)
		if err != nil {
			c.finishSeed(ctx, run, CatalogueSyncFailed, fmt.Errorf("page %d: %w", page, err))
			return
		}
		run.PagesFetched++
		run.RowsUpserted += upserted
		if err := c.cs.UpdateCatalogueSync(ctx, run); err != nil {
			c.finishSeed(ctx, run, CatalogueSyncFailed, fmt.Errorf("page %d: %w", page, err))
			return
		}

		// The provider reported fewer matches than we have already walked, so there
		// is nothing left to page through.
		if run.UpstreamTotal != nil && (page+1)*CataloguePageSize >= *run.UpstreamTotal {
			break
		}
	}
	c.finishSeed(ctx, run, CatalogueSyncCompleted, nil)
}

func (c *Catalogue) finishSeed(ctx context.Context, run *CatalogueSync, status CatalogueSyncStatus, cause error) {
	finished := time.Now().UTC()
	run.Status = status
	run.FinishedAt = &finished
	if cause != nil {
		msg := cause.Error()
		run.LastError = &msg
	}
	// The run row is the only place a background failure is visible, so persist it
	// even when the seed context is already spent.
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	_ = c.cs.UpdateCatalogueSync(writeCtx, run)
}
