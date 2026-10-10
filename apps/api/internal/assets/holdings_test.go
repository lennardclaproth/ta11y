package assets

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

func day(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		t.Fatalf("parse day %q: %v", value, err)
	}
	return parsed
}

// TestPlanHoldingMutationsDerivesWorthPerDay covers the rule the whole feature
// rests on: one mutation per day from the first purchase onwards, valued at the
// quantity held that day times the price of that day.
func TestPlanHoldingMutationsDerivesWorthPerDay(t *testing.T) {
	accID, classID, assetID := uuid.New(), uuid.New(), uuid.New()
	holding := &Asset{ID: assetID, ClassID: classID, AccountID: accID, ListingID: &[]uuid.UUID{uuid.New()}[0]}

	purchases := []*Purchase{
		{AccountID: accID, AssetID: assetID, PurchasedOn: day(t, "2026-01-01"), Quantity: 2, UnitPrice: money.Price(10_00)},
		{AccountID: accID, AssetID: assetID, PurchasedOn: day(t, "2026-01-03"), Quantity: 1, UnitPrice: money.Price(30_00)},
	}
	prices := map[uuid.UUID]map[string]money.Price{
		assetID: {
			"2026-01-01": money.Price(10_00),
			"2026-01-02": money.Price(20_00),
			// 2026-01-03 has no price: the day before is carried forward.
			"2026-01-04": money.Price(40_00),
		},
	}

	plan := planHoldingMutations(accID, []*Asset{holding}, purchases, nil, prices, day(t, "2026-01-04"))

	if len(plan.mutations) != 4 {
		t.Fatalf("expected one mutation per day, got %d", len(plan.mutations))
	}
	want := []money.Price{
		money.Price(20_00), // 2 units at 10.00
		money.Price(40_00), // 2 units at 20.00
		money.Price(60_00), // 3 units at the carried-forward 20.00
		money.Price(120_00),
	}
	for i, mutation := range plan.mutations {
		if mutation.NewWorth != want[i] {
			t.Errorf("day %d: expected worth %s, got %s", i, want[i], mutation.NewWorth)
		}
		if mutation.ChangeType != ChangeTypeSet {
			t.Errorf("day %d: expected a SET mutation, got %s", i, mutation.ChangeType)
		}
		if mutation.ClassTotalWorth != want[i] {
			t.Errorf("day %d: expected class total %s, got %s", i, want[i], mutation.ClassTotalWorth)
		}
	}
	if got := plan.currentWorth[assetID]; got != money.Price(120_00) {
		t.Errorf("expected current worth 120.00, got %s", got)
	}
}

// TestPlanHoldingMutationsCarriesManualItemsInTheClassTotal guards the class
// total a snapshot reads: a price move must not erase the hand-set items beside
// the holding in the same class.
func TestPlanHoldingMutationsCarriesManualItemsInTheClassTotal(t *testing.T) {
	accID, classID, assetID, manualID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	holding := &Asset{ID: assetID, ClassID: classID, AccountID: accID, ListingID: &[]uuid.UUID{uuid.New()}[0]}

	purchases := []*Purchase{
		{AccountID: accID, AssetID: assetID, PurchasedOn: day(t, "2026-01-01"), Quantity: 1, UnitPrice: money.Price(10_00)},
	}
	manual := []*Mutation{
		{AccountID: accID, ClassID: classID, AssetID: manualID, NewWorth: money.Price(100_00), EffectiveDate: day(t, "2026-01-01")},
	}
	prices := map[uuid.UUID]map[string]money.Price{
		assetID: {"2026-01-01": money.Price(10_00), "2026-01-02": money.Price(30_00)},
	}

	plan := planHoldingMutations(accID, []*Asset{holding}, purchases, manual, prices, day(t, "2026-01-02"))

	if len(plan.mutations) != 2 {
		t.Fatalf("expected two mutations, got %d", len(plan.mutations))
	}
	if got := plan.mutations[0].ClassTotalWorth; got != money.Price(110_00) {
		t.Errorf("day 1: expected class total 110.00, got %s", got)
	}
	if got := plan.mutations[1].ClassTotalWorth; got != money.Price(130_00) {
		t.Errorf("day 2: expected class total 130.00, got %s", got)
	}
}

// TestPlanHoldingMutationsStartEachHoldingAtItsOwnFirstPurchase guards against a
// holding bought later being back-dated to the account's earliest purchase: those
// days are not zero worth, they are days the item did not exist.
func TestPlanHoldingMutationsStartEachHoldingAtItsOwnFirstPurchase(t *testing.T) {
	accID, classID := uuid.New(), uuid.New()
	oldID, newID := uuid.New(), uuid.New()
	listingID := uuid.New()
	older := &Asset{ID: oldID, ClassID: classID, AccountID: accID, ListingID: &listingID}
	newer := &Asset{ID: newID, ClassID: classID, AccountID: accID, ListingID: &listingID}

	purchases := []*Purchase{
		{AccountID: accID, AssetID: oldID, PurchasedOn: day(t, "2026-01-01"), Quantity: 1, UnitPrice: money.Price(10_00)},
		{AccountID: accID, AssetID: newID, PurchasedOn: day(t, "2026-01-03"), Quantity: 1, UnitPrice: money.Price(50_00)},
	}
	prices := map[uuid.UUID]map[string]money.Price{
		oldID: {"2026-01-01": money.Price(10_00)},
		newID: {"2026-01-03": money.Price(50_00)},
	}

	plan := planHoldingMutations(accID, []*Asset{older, newer}, purchases, nil, prices, day(t, "2026-01-03"))

	days := map[uuid.UUID][]string{}
	for _, mutation := range plan.mutations {
		days[mutation.AssetID] = append(days[mutation.AssetID], dayKey(mutation.EffectiveDate))
	}
	if got := len(days[oldID]); got != 3 {
		t.Errorf("expected the older item to span all three days, got %d: %v", got, days[oldID])
	}
	if got := days[newID]; len(got) != 1 || got[0] != "2026-01-03" {
		t.Errorf("expected the newer item to start on its own first purchase, got %v", got)
	}

	// The class total still has to account for both on the day they overlap.
	for _, mutation := range plan.mutations {
		if dayKey(mutation.EffectiveDate) == "2026-01-03" && mutation.ClassTotalWorth != money.Price(60_00) {
			t.Errorf("expected a class total of 60.00 on the overlapping day, got %s", mutation.ClassTotalWorth)
		}
	}
}

// TestPlanHoldingMutationsWithoutPurchases verifies there is nothing to derive
// before the first purchase, so an item linked today writes no history.
func TestPlanHoldingMutationsWithoutPurchases(t *testing.T) {
	accID, classID := uuid.New(), uuid.New()
	holding := &Asset{ID: uuid.New(), ClassID: classID, AccountID: accID, ListingID: &[]uuid.UUID{uuid.New()}[0]}

	plan := planHoldingMutations(accID, []*Asset{holding}, nil, nil, nil, day(t, "2026-01-04"))

	if len(plan.mutations) != 0 {
		t.Fatalf("expected no mutations without purchases, got %d", len(plan.mutations))
	}
}

// TestHoldingValueFallsBackToPaid covers the day a purchase is recorded before
// the provider has a price for it: the item reads as what it cost, not as zero.
func TestHoldingValueFallsBackToPaid(t *testing.T) {
	if got := holdingValue(2, money.Price(50_00), 0); got != money.Price(50_00) {
		t.Errorf("expected the paid amount when no price is known, got %s", got)
	}
	if got := holdingValue(0, money.Price(50_00), money.Price(10_00)); got != 0 {
		t.Errorf("expected nothing held to be worth nothing, got %s", got)
	}
}

// TestBuildHoldingSeriesPairsWorthWithPaid checks that a day only counts the
// purchases made up to and including it.
func TestBuildHoldingSeriesPairsWorthWithPaid(t *testing.T) {
	assetID := uuid.New()
	worth := []HoldingWorthPoint{
		{Date: day(t, "2026-01-01"), Worth: money.Price(20_00)},
		{Date: day(t, "2026-01-03"), Worth: money.Price(90_00)},
	}
	purchases := []*Purchase{
		{AssetID: assetID, PurchasedOn: day(t, "2026-01-01"), Quantity: 2, UnitPrice: money.Price(10_00)},
		{AssetID: assetID, PurchasedOn: day(t, "2026-01-03"), Quantity: 1, UnitPrice: money.Price(30_00)},
	}

	series := buildHoldingSeries(worth, purchases)

	if len(series) != 2 {
		t.Fatalf("expected two points, got %d", len(series))
	}
	if series[0].Paid != money.Price(20_00) {
		t.Errorf("expected 20.00 paid on the first day, got %s", series[0].Paid)
	}
	if series[1].Paid != money.Price(50_00) {
		t.Errorf("expected 50.00 paid by the third day, got %s", series[1].Paid)
	}
}
