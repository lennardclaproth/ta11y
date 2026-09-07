//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/storage"
)

// seedCatalogue caches a few provider entries and returns the store under test.
func seedCatalogue(t *testing.T, db *storage.DB) (*storage.SQLXMarketDataStore, context.Context) {
	t.Helper()
	ctx := context.Background()
	store := storage.NewSQLXMarketDataStore(db)

	name := func(v string) *string { return &v }
	entries := []marketdata.ProviderListing{
		{Source: marketdata.SourceMarketStack, Symbol: "ASML", Name: name("ASML Holding NV"),
			ExchangeName: name("NASDAQ"), ExchangeMIC: name("XNAS"), HasEOD: true},
		{Source: marketdata.SourceMarketStack, Symbol: "ASML.XAMS", Name: name("ASML HOLDING"),
			ExchangeName: name("EURONEXT AMSTERDAM"), ExchangeMIC: name("XAMS"), HasEOD: true},
		// The provider returns nameless symbols; they must survive caching but
		// never be adoptable, because listings.name is NOT NULL.
		{Source: marketdata.SourceMarketStack, Symbol: "ASMCX", Name: nil,
			ExchangeMIC: name("NMFQS"), HasEOD: true},
		// Present in the catalogue but without price history.
		{Source: marketdata.SourceMarketStack, Symbol: "ASMB", Name: name("Assembly Biosciences Inc"),
			ExchangeMIC: name("XNAS"), HasEOD: false},
	}
	written, err := store.UpsertProviderListings(ctx, entries)
	if err != nil {
		t.Fatalf("upsert provider listings: %v", err)
	}
	if written != len(entries) {
		t.Fatalf("upsert wrote %d rows, want %d", written, len(entries))
	}
	return store, ctx
}

// TestCatalogueUpsertIsIdempotent re-caches the same symbols and expects the row
// count to stay flat: a repeated provider search must refresh rows, not duplicate
// them, or the catalogue would grow without bound.
func TestCatalogueUpsertIsIdempotent(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		store, ctx := seedCatalogue(t, db)

		before, err := store.CountProviderListings(ctx, marketdata.SourceMarketStack)
		if err != nil {
			t.Fatalf("count: %v", err)
		}

		renamed := "ASML Holding N.V."
		if _, err := store.UpsertProviderListings(ctx, []marketdata.ProviderListing{
			{Source: marketdata.SourceMarketStack, Symbol: "ASML", Name: &renamed, HasEOD: true},
		}); err != nil {
			t.Fatalf("re-upsert: %v", err)
		}

		after, err := store.CountProviderListings(ctx, marketdata.SourceMarketStack)
		if err != nil {
			t.Fatalf("count after: %v", err)
		}
		if after != before {
			t.Fatalf("row count changed on re-upsert: got %d, want %d", after, before)
		}

		results, _, err := store.SearchCatalogue(ctx, "ASML", marketdata.ScopeCatalogue, 25, 0)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		for _, row := range results {
			if row.Symbol == "ASML" {
				if row.Name == nil || *row.Name != renamed {
					t.Fatalf("re-upsert did not refresh the name: got %v, want %q", row.Name, renamed)
				}
				return
			}
		}
		t.Fatal("expected to find ASML in catalogue search results")
	})
}

// TestCatalogueSearchScopes checks the three scopes against the real schema. The
// combined query is a UNION with casted columns on both sides, so this is what
// proves it executes and maps at all.
func TestCatalogueSearchScopes(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		store, ctx := seedCatalogue(t, db)

		// One of the catalogue symbols is also tracked as a real listing.
		tracked, err := marketdata.NewListing("ASML.XAMS", "ASML Holding", marketdata.SourceMarketStack,
			marketdata.ListingWithExchange("XAMS"))
		if err != nil {
			t.Fatalf("new listing: %v", err)
		}
		if err := store.Create(ctx, tracked); err != nil {
			t.Fatalf("create listing: %v", err)
		}

		trackedRows, trackedTotal, err := store.SearchCatalogue(ctx, "asml", marketdata.ScopeTracked, 25, 0)
		if err != nil {
			t.Fatalf("search tracked: %v", err)
		}
		if trackedTotal != 1 || len(trackedRows) != 1 {
			t.Fatalf("tracked scope: got %d rows / total %d, want 1/1", len(trackedRows), trackedTotal)
		}
		if !trackedRows[0].Tracked || trackedRows[0].ID != tracked.ID {
			t.Fatalf("tracked row should be the listing itself: tracked=%v id=%s want id=%s",
				trackedRows[0].Tracked, trackedRows[0].ID, tracked.ID)
		}
		if trackedRows[0].CreatedAt == nil {
			t.Fatal("tracked row must keep its listing timestamps")
		}

		// The catalogue side must exclude the entry that is already tracked, so an
		// instrument never shows up twice in one result set.
		catalogueRows, _, err := store.SearchCatalogue(ctx, "asml", marketdata.ScopeCatalogue, 25, 0)
		if err != nil {
			t.Fatalf("search catalogue: %v", err)
		}
		for _, row := range catalogueRows {
			if row.Symbol == "ASML.XAMS" {
				t.Fatal("catalogue scope returned an already-tracked symbol")
			}
			if row.Tracked {
				t.Fatalf("catalogue scope returned a tracked row: %s", row.Symbol)
			}
			if row.CreatedAt != nil {
				t.Fatalf("catalogue row %s must not claim listing timestamps", row.Symbol)
			}
		}

		allRows, allTotal, err := store.SearchCatalogue(ctx, "asml", marketdata.ScopeAll, 25, 0)
		if err != nil {
			t.Fatalf("search all: %v", err)
		}
		if allTotal != trackedTotal+len(catalogueRows) {
			t.Fatalf("all scope total %d, want %d", allTotal, trackedTotal+len(catalogueRows))
		}
		if len(allRows) == 0 || !allRows[0].Tracked {
			t.Fatal("all scope must sort tracked rows first")
		}
	})
}

// TestCatalogueSearchMatchesNamelessSymbols covers the provider's uneven data: a
// symbol with no name is still findable by symbol, and is reported as not
// adoptable rather than failing later on listing creation.
func TestCatalogueSearchMatchesNamelessSymbols(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		store, ctx := seedCatalogue(t, db)

		rows, _, err := store.SearchCatalogue(ctx, "asmcx", marketdata.ScopeAll, 25, 0)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("got %d rows for a nameless symbol, want 1", len(rows))
		}
		if adoptable, reason := rows[0].Adoptable(); adoptable || reason == "" {
			t.Fatalf("nameless symbol should not be adoptable, got adoptable=%v reason=%q", adoptable, reason)
		}

		noHistory, _, err := store.SearchCatalogue(ctx, "asmb", marketdata.ScopeAll, 25, 0)
		if err != nil {
			t.Fatalf("search asmb: %v", err)
		}
		if len(noHistory) != 1 {
			t.Fatalf("got %d rows for ASMB, want 1", len(noHistory))
		}
		if adoptable, _ := noHistory[0].Adoptable(); adoptable {
			t.Fatal("a symbol without end-of-day history should not be adoptable")
		}
	})
}

// TestCatalogueSearchPaginates checks that limit/offset walk the combined result
// set without repeating or dropping rows, since paging happens inside the UNION.
func TestCatalogueSearchPaginates(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		store, ctx := seedCatalogue(t, db)

		_, total, err := store.SearchCatalogue(ctx, "asm", marketdata.ScopeAll, 25, 0)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if total < 4 {
			t.Fatalf("expected at least the 4 seeded ASM* rows, got %d", total)
		}

		seen := map[string]bool{}
		for offset := 0; offset < total; offset += 2 {
			page, pageTotal, err := store.SearchCatalogue(ctx, "asm", marketdata.ScopeAll, 2, offset)
			if err != nil {
				t.Fatalf("search offset %d: %v", offset, err)
			}
			if pageTotal != total {
				t.Fatalf("total changed while paging: got %d, want %d", pageTotal, total)
			}
			for _, row := range page {
				if seen[row.Symbol] {
					t.Fatalf("symbol %s returned on more than one page", row.Symbol)
				}
				seen[row.Symbol] = true
			}
		}
		if len(seen) != total {
			t.Fatalf("paging covered %d rows, want %d", len(seen), total)
		}
	})
}

// TestCatalogueSyncLifecycle round-trips a seed run so a background failure is
// still observable: the run row is the only place that error surfaces.
func TestCatalogueSyncLifecycle(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXMarketDataStore(db)

		if latest, err := store.LatestCatalogueSync(ctx, marketdata.SourceMarketStack); err != nil {
			t.Fatalf("latest on empty: %v", err)
		} else if latest != nil {
			t.Fatal("expected no sync before one is recorded")
		}

		run := &marketdata.CatalogueSync{
			ID:     uuid.New(),
			Source: marketdata.SourceMarketStack,
			Status: marketdata.CatalogueSyncRunning,
		}
		if err := store.CreateCatalogueSync(ctx, run); err != nil {
			t.Fatalf("create sync: %v", err)
		}

		boom := "page 3: provider unavailable"
		run.Status = marketdata.CatalogueSyncFailed
		run.PagesFetched = 3
		run.RowsUpserted = 300
		run.LastError = &boom
		if err := store.UpdateCatalogueSync(ctx, run); err != nil {
			t.Fatalf("update sync: %v", err)
		}

		latest, err := store.LatestCatalogueSync(ctx, marketdata.SourceMarketStack)
		if err != nil {
			t.Fatalf("latest: %v", err)
		}
		if latest == nil || latest.Status != marketdata.CatalogueSyncFailed {
			t.Fatalf("expected a failed run, got %+v", latest)
		}
		if latest.RowsUpserted != 300 || latest.LastError == nil || *latest.LastError != boom {
			t.Fatalf("progress and error were not persisted: %+v", latest)
		}
	})
}
