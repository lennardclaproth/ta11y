package assets

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// fakeClassGrowthStore serves ClassesForAccount and Class from memory; the
// embedded interface panics on any other read, which ClassGrowth must not make.
type fakeClassGrowthStore struct {
	QueryStore
	classes         map[uuid.UUID]*Class
	includeArchived *bool
}

func (f *fakeClassGrowthStore) ClassesForAccount(_ context.Context, _ uuid.UUID, includeArchived bool) (map[uuid.UUID]*Class, error) {
	f.includeArchived = &includeArchived
	out := make(map[uuid.UUID]*Class, len(f.classes))
	for id, class := range f.classes {
		out[id] = &Class{ID: class.ID, AccountID: class.AccountID, Name: class.Name}
	}
	return out, nil
}

func (f *fakeClassGrowthStore) Class(_ context.Context, id uuid.UUID) (*Class, error) {
	return f.classes[id], nil
}

func TestClassGrowthReturnsSortedFilteredSeries(t *testing.T) {
	accID := uuid.New()
	cash := &Class{ID: uuid.New(), AccountID: accID, Name: "Cash & savings", Mutations: []Mutation{
		{EffectiveDate: day(t, "2026-03-01"), ClassTotalWorth: money.Price(300_00)},
		{EffectiveDate: day(t, "2026-02-01").Add(2 * time.Hour), ClassTotalWorth: money.Price(250_00)},
		{EffectiveDate: day(t, "2026-02-01"), ClassTotalWorth: money.Price(200_00)},
		{EffectiveDate: day(t, "2026-01-01"), ClassTotalWorth: money.Price(100_00)},
	}}
	empty := &Class{ID: uuid.New(), AccountID: accID, Name: "Art"}
	store := &fakeClassGrowthStore{classes: map[uuid.UUID]*Class{cash.ID: cash, empty.ID: empty}}

	from, to := day(t, "2026-02-01"), day(t, "2026-02-28")
	got, err := NewQueries(store, nil).ClassGrowth(context.Background(), accID, &from, &to)
	if err != nil {
		t.Fatalf("ClassGrowth: %v", err)
	}
	if store.includeArchived == nil || *store.includeArchived {
		t.Fatalf("ClassesForAccount includeArchived = %v, want false", store.includeArchived)
	}
	if len(got) != 2 || got[0].Name != "Art" || got[1].Name != "Cash & savings" {
		t.Fatalf("series = %+v, want Art then Cash & savings", got)
	}
	if got[0].Growth == nil || len(got[0].Growth) != 0 {
		t.Fatalf("empty class growth = %#v, want empty non-nil slice", got[0].Growth)
	}
	if got[1].ClassID != cash.ID {
		t.Fatalf("class id = %s, want %s", got[1].ClassID, cash.ID)
	}
	growth := got[1].Growth
	if len(growth) != 1 || !growth[0].Date.Equal(from) || growth[0].TotalWorth != money.Price(250_00) {
		t.Fatalf("growth = %+v, want one 2026-02-01 point worth the day's latest total 250.00", growth)
	}
}

func TestClassGrowthWithoutBoundsKeepsEveryDay(t *testing.T) {
	accID := uuid.New()
	class := &Class{ID: uuid.New(), AccountID: accID, Name: "Cash", Mutations: []Mutation{
		{EffectiveDate: day(t, "2026-02-01"), ClassTotalWorth: money.Price(200_00)},
		{EffectiveDate: day(t, "2026-01-01"), ClassTotalWorth: money.Price(100_00)},
	}}
	store := &fakeClassGrowthStore{classes: map[uuid.UUID]*Class{class.ID: class}}

	got, err := NewQueries(store, nil).ClassGrowth(context.Background(), accID, nil, nil)
	if err != nil {
		t.Fatalf("ClassGrowth: %v", err)
	}
	if len(got) != 1 || len(got[0].Growth) != 2 {
		t.Fatalf("series = %+v, want one class with two points", got)
	}
	if !got[0].Growth[0].Date.Equal(day(t, "2026-01-01")) || !got[0].Growth[1].Date.Equal(day(t, "2026-02-01")) {
		t.Fatalf("growth = %+v, want ascending by date", got[0].Growth)
	}
}

func TestClassGrowthCarriesWorthIntoTheWindow(t *testing.T) {
	accID := uuid.New()
	moved := &Class{ID: uuid.New(), AccountID: accID, Name: "Cash", Mutations: []Mutation{
		{EffectiveDate: day(t, "2026-02-01"), ClassTotalWorth: money.Price(250_00)},
		{EffectiveDate: day(t, "2026-01-01"), ClassTotalWorth: money.Price(100_00)},
	}}
	still := &Class{ID: uuid.New(), AccountID: accID, Name: "House", Mutations: []Mutation{
		{EffectiveDate: day(t, "2025-06-01"), ClassTotalWorth: money.Price(900_00)},
	}}
	store := &fakeClassGrowthStore{classes: map[uuid.UUID]*Class{moved.ID: moved, still.ID: still}}

	from, to := day(t, "2026-01-15"), day(t, "2026-03-01")
	got, err := NewQueries(store, nil).ClassGrowth(context.Background(), accID, &from, &to)
	if err != nil {
		t.Fatalf("ClassGrowth: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("series = %+v, want two classes", got)
	}
	cash := got[0].Growth
	if len(cash) != 2 || !cash[0].Date.Equal(from) || cash[0].TotalWorth != money.Price(100_00) ||
		cash[1].TotalWorth != money.Price(250_00) {
		t.Fatalf("cash growth = %+v, want 100.00 on the window start then 250.00", cash)
	}
	house := got[1].Growth
	if len(house) != 1 || !house[0].Date.Equal(from) || house[0].TotalWorth != money.Price(900_00) {
		t.Fatalf("house growth = %+v, want its 900.00 carried to the window start", house)
	}
}
