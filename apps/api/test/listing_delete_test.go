//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/storage"
)

// newListing persists a minimal listing and returns it.
func newListing(t *testing.T, store *storage.SQLXMarketDataStore, ctx context.Context, symbol string) *marketdata.Listing {
	t.Helper()
	listing := &marketdata.Listing{
		ID:     uuid.New(),
		Symbol: symbol,
		Name:   symbol + " Test Instrument",
		Source: marketdata.SourceMarketStack,
	}
	if err := store.Create(ctx, listing); err != nil {
		t.Fatalf("create listing %s: %v", symbol, err)
	}
	return listing
}

// seedPosition attaches an open position to a listing, which is what makes the
// listing undeletable. Written with raw SQL rather than through the portfolio
// feature because the only thing under test is that the reference is counted.
func seedPosition(t *testing.T, db *storage.DB, listingID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	accountID := seedAccount(t, db, "holder", "holder@example.com")

	exec := db.GetExecutor(ctx)
	if _, err := exec.ExecContext(ctx,
		db.Rebind(`INSERT INTO portfolio_accounts (id, account_id) VALUES (?, ?)`),
		uuid.New(), accountID,
	); err != nil {
		t.Fatalf("seed portfolio account: %v", err)
	}
	if _, err := exec.ExecContext(ctx,
		db.Rebind(`INSERT INTO positions (id, account_id, listing_id, symbol, open_date, quantity)
		           VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, 10)`),
		uuid.New(), accountID, listingID, "HELD",
	); err != nil {
		t.Fatalf("seed position: %v", err)
	}
}

// TestDeleteListingRemovesUnusedListing covers the ordinary case: a listing added by
// mistake, referenced by nothing, is removed.
func TestDeleteListingRemovesUnusedListing(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXMarketDataStore(db)
		commands := marketdata.NewCommands(store, nil)

		listing := newListing(t, store, ctx, "UNUSED")

		used, err := store.CountPortfolioUsage(ctx, listing.ID)
		if err != nil {
			t.Fatalf("count portfolio usage: %v", err)
		}
		if used != 0 {
			t.Fatalf("unused listing reported %d portfolio references, want 0", used)
		}

		if err := commands.DeleteListing(ctx, listing.ID); err != nil {
			t.Fatalf("delete unused listing: %v", err)
		}

		got, err := store.Get(ctx, listing.ID)
		if err != nil {
			t.Fatalf("get after delete: %v", err)
		}
		if got != nil {
			t.Fatalf("listing still present after delete")
		}
	})
}

// TestDeleteListingRefusesWhenPortfolioUsesIt is the guard that matters. The schema
// cascades a listing delete into position snapshots and nulls open positions, so a
// listing a portfolio depends on must be refused rather than deleted -- otherwise the
// account silently loses its valuation history for that instrument.
func TestDeleteListingRefusesWhenPortfolioUsesIt(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXMarketDataStore(db)
		commands := marketdata.NewCommands(store, nil)

		listing := newListing(t, store, ctx, "HELD")
		seedPosition(t, db, listing.ID)

		used, err := store.CountPortfolioUsage(ctx, listing.ID)
		if err != nil {
			t.Fatalf("count portfolio usage: %v", err)
		}
		if used != 1 {
			t.Fatalf("held listing reported %d portfolio references, want 1", used)
		}

		err = commands.DeleteListing(ctx, listing.ID)
		if !errors.Is(err, marketdata.ErrListingInUse) {
			t.Fatalf("delete held listing returned %v, want ErrListingInUse", err)
		}

		// The refusal has to leave the listing alone, not half-delete it.
		got, err := store.Get(ctx, listing.ID)
		if err != nil {
			t.Fatalf("get after refused delete: %v", err)
		}
		if got == nil {
			t.Fatalf("listing was deleted despite being in use")
		}
	})
}

// TestDeleteListingRejectsUnknownListing keeps the not-found path distinct from the
// in-use one, so the handler can answer 404 rather than 409.
func TestDeleteListingRejectsUnknownListing(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		commands := marketdata.NewCommands(storage.NewSQLXMarketDataStore(db), nil)

		err := commands.DeleteListing(ctx, uuid.New())
		if !errors.Is(err, marketdata.ErrListingNotFound) {
			t.Fatalf("delete unknown listing returned %v, want ErrListingNotFound", err)
		}
	})
}
