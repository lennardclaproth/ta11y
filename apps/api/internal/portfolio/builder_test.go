package portfolio

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
)

func positionSnap(positionID uuid.UUID, at time.Time, marketValue, costBasis, realized money.Price) *PositionSnapshot {
	return &PositionSnapshot{
		ID:          uuid.New(),
		PositionID:  positionID,
		OccurredAt:  at,
		MarketValue: marketValue,
		CostBasis:   costBasis,
		RealizedPnL: realized,
		TotalPnL:    realized,
	}
}

func TestAggregatePortfolioSnapshotsSumsPositionsGroupedPerPosition(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	// The builder hands snapshots over grouped per position, not per day.
	snaps := []*PositionSnapshot{
		positionSnap(a, day(2026, 1, 1), 100, 90, 0),
		positionSnap(a, day(2026, 1, 2), 110, 90, 0),
		positionSnap(a, day(2026, 1, 3), 120, 90, 0),
		positionSnap(b, day(2026, 1, 1), 1000, 900, 0),
		positionSnap(b, day(2026, 1, 2), 1100, 900, 0),
		positionSnap(b, day(2026, 1, 3), 1200, 900, 0),
	}

	got := aggregatePortfolioSnapshots(uuid.New(), snaps, day(2026, 1, 4))

	want := []money.Price{1100, 1210, 1320}
	if len(got) != len(want) {
		t.Fatalf("want %d daily snapshots, got %d", len(want), len(got))
	}
	for i, mv := range want {
		if got[i].MarketValue != mv {
			t.Errorf("day %d: want market value %d, got %d", i+1, mv, got[i].MarketValue)
		}
		if got[i].CostBasis != 990 {
			t.Errorf("day %d: want cost basis 990, got %d", i+1, got[i].CostBasis)
		}
	}
}

func TestAggregatePortfolioSnapshotsKeepsAClosedPositionsRealizedResult(t *testing.T) {
	closed, open := uuid.New(), uuid.New()
	snaps := []*PositionSnapshot{
		positionSnap(closed, day(2026, 1, 1), 100, 90, 0),
		positionSnap(closed, day(2026, 1, 2), 0, 0, 25),
		positionSnap(open, day(2026, 1, 1), 1000, 900, 0),
		positionSnap(open, day(2026, 1, 2), 1000, 900, 0),
		positionSnap(open, day(2026, 1, 3), 1000, 900, 0),
		positionSnap(open, day(2026, 1, 4), 1000, 900, 0),
	}

	got := aggregatePortfolioSnapshots(uuid.New(), snaps, day(2026, 1, 5))

	last := got[len(got)-1]
	if last.MarketValue != 1000 || last.CostBasis != 900 {
		t.Errorf("want market value 1000 and cost basis 900 after the close, got %d and %d", last.MarketValue, last.CostBasis)
	}
	if last.RealizedPnL != 25 {
		t.Errorf("want the closed position's realized 25 carried forward, got %d", last.RealizedPnL)
	}
	if last.NetCashflow != 0 {
		t.Errorf("want no cash flow on a quiet day after the close, got %d", last.NetCashflow)
	}
}

func TestAggregatePortfolioSnapshotsCarriesForwardAndStartsAtTheEarliestDay(t *testing.T) {
	late, early := uuid.New(), uuid.New()
	snaps := []*PositionSnapshot{
		positionSnap(late, day(2026, 1, 3), 50, 50, 0),
		positionSnap(early, day(2026, 1, 1), 200, 180, 0),
	}

	got := aggregatePortfolioSnapshots(uuid.New(), snaps, day(2026, 1, 5))

	if len(got) != 4 || !got[0].OccurredAt.Equal(day(2026, 1, 1)) {
		t.Fatalf("want 4 daily snapshots from 2026-01-01, got %d starting %v", len(got), got[0].OccurredAt)
	}
	want := []money.Price{200, 200, 250, 250}
	for i, mv := range want {
		if got[i].MarketValue != mv {
			t.Errorf("day %d: want market value %d, got %d", i+1, mv, got[i].MarketValue)
		}
	}
}

func TestLinkToListingGivesAnISINOnlyPositionTheListingSymbol(t *testing.T) {
	listing := &marketdata.Listing{ID: uuid.New(), Symbol: "TDT.AS"}
	pos := &Position{ISIN: ptrString("NL0009272749")}

	pos.linkToListing(listing)

	if pos.ListingID == nil || *pos.ListingID != listing.ID {
		t.Fatal("expected the position to be linked to the listing")
	}
	if pos.Symbol == nil || *pos.Symbol != "TDT.AS" {
		t.Fatalf("expected the listing symbol TDT.AS, got %v", pos.Symbol)
	}
}

func TestLinkToListingKeepsASymbolThePositionAlreadyHas(t *testing.T) {
	pos := &Position{Symbol: ptrString("NWIF")}

	pos.linkToListing(&marketdata.Listing{ID: uuid.New(), Symbol: "NWIF.AS"})

	if *pos.Symbol != "NWIF" {
		t.Fatalf("expected the entered symbol to stay, got %s", *pos.Symbol)
	}
}
