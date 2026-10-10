package marketdata

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// identityStore answers only the two exact lookups ListingByIdentity is allowed to
// make; every other QueryStore method panics so an accidental fallback to a fuzzy
// search fails the test loudly.
type identityStore struct {
	QueryStore
	byISIN   map[string]*Listing
	bySymbol map[string]*Listing
}

func (s *identityStore) GetByISIN(_ context.Context, isin string) (*Listing, error) {
	return s.byISIN[isin], nil
}

func (s *identityStore) GetBySymbol(_ context.Context, symbol string) (*Listing, error) {
	return s.bySymbol[symbol], nil
}

func newIdentityQueries() (*Queries, *Listing, *Listing) {
	byISIN := &Listing{ID: uuid.New(), Symbol: "EXAMPL"}
	bySymbol := &Listing{ID: uuid.New(), Symbol: "DEMO"}
	store := &identityStore{
		byISIN:   map[string]*Listing{"NLTEST0001": byISIN},
		bySymbol: map[string]*Listing{"DEMO": bySymbol},
	}
	return NewQueries(store, nil), byISIN, bySymbol
}

func TestListingByIdentity_MatchesOnExactISIN(t *testing.T) {
	queries, byISIN, _ := newIdentityQueries()

	listing, err := queries.ListingByIdentity(context.Background(), " NLTEST0001 ", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if listing == nil || listing.ID != byISIN.ID {
		t.Fatalf("expected the listing with that exact ISIN")
	}
}

func TestListingByIdentity_UnknownISINDoesNotFallBackToSymbol(t *testing.T) {
	// A holding that carries an ISIN we do not track must stay unlinked: falling back
	// to the symbol is how a short ticker ends up attached to somebody else's product.
	queries, _, _ := newIdentityQueries()

	listing, err := queries.ListingByIdentity(context.Background(), "NLUNKNOWN1", "DEMO")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if listing != nil {
		t.Fatalf("expected no match for an untracked ISIN, got %s", listing.Symbol)
	}
}

func TestListingByIdentity_FallsBackToSymbolWithoutISIN(t *testing.T) {
	queries, _, bySymbol := newIdentityQueries()

	listing, err := queries.ListingByIdentity(context.Background(), "", "DEMO")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if listing == nil || listing.ID != bySymbol.ID {
		t.Fatalf("expected the listing with that exact symbol")
	}
}

func TestListingByIdentity_NoIdentityIsNoMatch(t *testing.T) {
	queries, _, _ := newIdentityQueries()

	listing, err := queries.ListingByIdentity(context.Background(), "", "  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if listing != nil {
		t.Fatalf("expected no match without an identity")
	}
}
