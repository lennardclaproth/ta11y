package assets

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

func euros(units int64) money.Price { return money.Price(units * money.Scale) }

func onDay(s string) time.Time {
	d, _ := time.Parse(time.DateOnly, s)
	return d
}

// A deposit entered after a later revaluation, but dated before it, has to land in
// between: the revaluation is still the worth that holds today, and the class total
// on each day sums every item as it stood then.
func TestResequenceClassReplaysChangesInDateOrder(t *testing.T) {
	savings, cash := uuid.New(), uuid.New()
	increase := ChangeDirectionIncrease
	created := time.Now()
	at := func(offset int) time.Time { return created.Add(time.Duration(offset) * time.Second) }

	mutations := []Mutation{
		{ID: uuid.New(), AssetID: savings, ChangeType: ChangeTypeSet, Amount: euros(1_000), EffectiveDate: onDay("2026-01-01"), CreatedAt: at(0)},
		{ID: uuid.New(), AssetID: cash, ChangeType: ChangeTypeSet, Amount: euros(200), EffectiveDate: onDay("2026-02-01"), CreatedAt: at(1)},
		{ID: uuid.New(), AssetID: savings, ChangeType: ChangeTypeSet, Amount: euros(1_350), EffectiveDate: onDay("2026-04-01"), CreatedAt: at(2)},
		// Entered last, dated before the revaluation above.
		{ID: uuid.New(), AssetID: savings, ChangeType: ChangeTypeAdjust, Direction: &increase, Amount: euros(300), EffectiveDate: onDay("2026-03-01"), CreatedAt: at(3)},
	}
	items := []Asset{{ID: savings}, {ID: cash}}

	changed, worth, err := resequenceClass(items, mutations)
	if err != nil {
		t.Fatalf("resequence: %v", err)
	}

	byID := make(map[uuid.UUID]*Mutation, len(changed))
	for _, m := range changed {
		byID[m.ID] = m
	}
	deposit := byID[mutations[3].ID]
	if deposit == nil || deposit.PreviousWorth != euros(1_000) || deposit.NewWorth != euros(1_300) || deposit.ClassTotalWorth != euros(1_500) {
		t.Fatalf("deposit: want 1000 -> 1300 with class total 1500, got %+v", deposit)
	}
	revaluation := byID[mutations[2].ID]
	if revaluation == nil || revaluation.PreviousWorth != euros(1_300) || revaluation.NewWorth != euros(1_350) || revaluation.ClassTotalWorth != euros(1_550) {
		t.Fatalf("revaluation: want 1300 -> 1350 with class total 1550, got %+v", revaluation)
	}
	if worth[savings] != euros(1_350) || worth[cash] != euros(200) {
		t.Fatalf("current worth: want savings 1350 and cash 200, got %v and %v", worth[savings], worth[cash])
	}
}

// A deposit written before NewMutation was corrected holds the running total where
// the change belongs. Replaying it from that total would add 1300 to the 1000 it was
// already part of; the replay has to read the change from where it was written and
// rewrite the row with it, so the next replay finds a correct one.
func TestResequenceClassCorrectsLegacyAdjustRows(t *testing.T) {
	savings := uuid.New()
	increase := ChangeDirectionIncrease

	mutations := []Mutation{
		{ID: uuid.New(), AssetID: savings, ChangeType: ChangeTypeSet, Amount: euros(1_000), PreviousWorth: 0, NewWorth: euros(1_000), ClassTotalWorth: euros(1_000), EffectiveDate: onDay("2026-01-01")},
		// previous 1000 + 300 deposited, stored with the total in Amount and the
		// change in NewWorth.
		{ID: uuid.New(), AssetID: savings, ChangeType: ChangeTypeAdjust, Direction: &increase, Amount: euros(1_300), PreviousWorth: euros(1_000), NewWorth: euros(300), ClassTotalWorth: euros(300), EffectiveDate: onDay("2026-02-01")},
	}

	changed, worth, err := resequenceClass([]Asset{{ID: savings}}, mutations)
	if err != nil {
		t.Fatalf("resequence: %v", err)
	}

	if len(changed) != 1 {
		t.Fatalf("want only the deposit rewritten, got %d rows", len(changed))
	}
	deposit := changed[0]
	if deposit.Amount != euros(300) || deposit.PreviousWorth != euros(1_000) || deposit.NewWorth != euros(1_300) {
		t.Fatalf("deposit: want 300 taking 1000 -> 1300, got %+v", deposit)
	}
	if worth[savings] != euros(1_300) {
		t.Fatalf("current worth: want 1300, got %v", worth[savings])
	}

	// The rewritten row is a correct one, so replaying it again leaves it alone.
	again, _, err := resequenceClass([]Asset{{ID: savings}}, []Mutation{mutations[0], *deposit})
	if err != nil {
		t.Fatalf("second resequence: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("want a stable row on the second replay, got %d rewritten", len(again))
	}
}

// Rows that already hold the right figures are left alone, so an edit only rewrites
// what it actually moved.
func TestResequenceClassLeavesConsistentRowsAlone(t *testing.T) {
	item := uuid.New()
	mutations := []Mutation{
		{ID: uuid.New(), AssetID: item, ChangeType: ChangeTypeSet, Amount: euros(500), NewWorth: euros(500), ClassTotalWorth: euros(500), EffectiveDate: onDay("2026-01-01")},
	}
	changed, _, err := resequenceClass([]Asset{{ID: item}}, mutations)
	if err != nil {
		t.Fatalf("resequence: %v", err)
	}
	if len(changed) != 0 {
		t.Fatalf("want no rewritten rows, got %d", len(changed))
	}
}

// A withdrawal takes from the worth it follows; an archived item stays out of the total.
func TestResequenceClassWithdrawalsAndArchivedItems(t *testing.T) {
	savings, closed := uuid.New(), uuid.New()
	decrease := ChangeDirectionDecrease
	mutations := []Mutation{
		{ID: uuid.New(), AssetID: savings, ChangeType: ChangeTypeSet, Amount: euros(1_000), EffectiveDate: onDay("2026-01-01")},
		{ID: uuid.New(), AssetID: closed, ChangeType: ChangeTypeSet, Amount: euros(400), EffectiveDate: onDay("2026-01-02")},
		{ID: uuid.New(), AssetID: savings, ChangeType: ChangeTypeAdjust, Direction: &decrease, Amount: euros(250), EffectiveDate: onDay("2026-02-01")},
	}
	changed, worth, err := resequenceClass([]Asset{{ID: savings}, {ID: closed, Archived: true}}, mutations)
	if err != nil {
		t.Fatalf("resequence: %v", err)
	}
	if worth[savings] != euros(750) {
		t.Fatalf("withdrawal: want 750, got %v", worth[savings])
	}
	for _, m := range changed {
		if m.ID == mutations[2].ID && m.ClassTotalWorth != euros(750) {
			t.Fatalf("archived item counted in the class total: got %v", m.ClassTotalWorth)
		}
	}
}

// An add or withdraw dated before an item's opening worth has nothing to apply to.
func TestResequenceClassRefusesAnAdjustBeforeTheOpening(t *testing.T) {
	item := uuid.New()
	increase := ChangeDirectionIncrease
	mutations := []Mutation{
		{ID: uuid.New(), AssetID: item, ChangeType: ChangeTypeSet, Amount: euros(1_000), EffectiveDate: onDay("2026-02-01")},
		{ID: uuid.New(), AssetID: item, ChangeType: ChangeTypeAdjust, Direction: &increase, Amount: euros(300), EffectiveDate: onDay("2026-01-01")},
	}
	if _, _, err := resequenceClass([]Asset{{ID: item}}, mutations); err != ErrOpeningNotSet {
		t.Fatalf("want ErrOpeningNotSet, got %v", err)
	}
}
