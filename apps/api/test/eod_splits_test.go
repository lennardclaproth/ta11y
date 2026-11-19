//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/storage"
)

// seedListing creates a listing the EOD rows can hang off.
func seedListing(t *testing.T, db *storage.DB, symbol string) uuid.UUID {
	t.Helper()

	store := storage.NewSQLXMarketDataStore(db)
	listing, err := marketdata.NewListing(symbol, symbol+" Test", marketdata.SourceMarketStack)
	if err != nil {
		t.Fatalf("build listing: %v", err)
	}
	if err := store.Create(t.Context(), listing); err != nil {
		t.Fatalf("create listing: %v", err)
	}
	return listing.ID
}

func seedEOD(t *testing.T, db *storage.DB, listingID uuid.UUID, symbol string, date time.Time, close, splitFactor float64) {
	t.Helper()

	eod, err := marketdata.NewEOD(symbol, date, close, close, close, close, 1000, splitFactor)
	if err != nil {
		t.Fatalf("build eod: %v", err)
	}
	eod.ListingID = listingID
	if err := storage.NewSQLXMarketDataStore(db).InsertEOD(t.Context(), &eod); err != nil {
		t.Fatalf("insert eod: %v", err)
	}
}

// The split factor has to survive the write/read round-trip, otherwise the rebuild
// replays no splits regardless of what the provider reported.
func TestEODSplitFactorRoundTrips(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		listingID := seedListing(t, db, "AAPL")
		seedEOD(t, db, listingID, "AAPL", time.Date(2020, time.August, 28, 0, 0, 0, 0, time.UTC), 499.23, 1)
		seedEOD(t, db, listingID, "AAPL", time.Date(2020, time.August, 31, 0, 0, 0, 0, time.UTC), 129.04, 4)

		store := storage.NewSQLXMarketDataStore(db)
		rows, err := store.GetEODForListing(t.Context(), listingID, nil, nil, nil, nil, "asc")
		if err != nil {
			t.Fatalf("get eods: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("rows = %d, want 2", len(rows))
		}
		if rows[0].SplitFactor != 1 {
			t.Fatalf("ordinary day factor = %v, want 1", rows[0].SplitFactor)
		}
		if rows[1].SplitFactor != 4 {
			t.Fatalf("split day factor = %v, want 4", rows[1].SplitFactor)
		}
	})
}

// SplitsForListing is what the rebuild reads. It must return only real corporate
// actions, oldest first, so replaying them in order reproduces the share count.
func TestSplitsForListingReturnsOnlyRealSplits(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		listingID := seedListing(t, db, "AAPL")
		seedEOD(t, db, listingID, "AAPL", time.Date(2020, time.August, 31, 0, 0, 0, 0, time.UTC), 129.04, 4)
		seedEOD(t, db, listingID, "AAPL", time.Date(2014, time.June, 9, 0, 0, 0, 0, time.UTC), 93.70, 7)
		seedEOD(t, db, listingID, "AAPL", time.Date(2021, time.March, 1, 0, 0, 0, 0, time.UTC), 127.79, 1)

		splits, err := storage.NewSQLXMarketDataStore(db).SplitsForListing(t.Context(), listingID)
		if err != nil {
			t.Fatalf("splits for listing: %v", err)
		}

		if len(splits) != 2 {
			t.Fatalf("splits = %d, want 2 -- ordinary days are not corporate actions", len(splits))
		}
		if !splits[0].Date.Before(splits[1].Date) {
			t.Fatal("splits must come back oldest first so they replay in order")
		}
		if splits[0].Factor != 7 || splits[1].Factor != 4 {
			t.Fatalf("factors = %v, %v; want 7 then 4", splits[0].Factor, splits[1].Factor)
		}
	})
}

// A listing with no corporate actions must produce an empty list rather than an error,
// which is the overwhelmingly common case on every rebuild.
func TestSplitsForListingWithNoSplits(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		listingID := seedListing(t, db, "VWRL")
		seedEOD(t, db, listingID, "VWRL", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), 118.42, 1)

		splits, err := storage.NewSQLXMarketDataStore(db).SplitsForListing(t.Context(), listingID)
		if err != nil {
			t.Fatalf("splits for listing: %v", err)
		}
		if len(splits) != 0 {
			t.Fatalf("splits = %d, want none", len(splits))
		}
	})
}
