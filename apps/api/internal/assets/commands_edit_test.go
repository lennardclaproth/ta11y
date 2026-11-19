package assets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// editGetter serves one class and one of its items, as the store would.
type editGetter struct {
	class     *assets.Class
	asset     *assets.Asset
	purchases []*assets.Purchase
}

func (g editGetter) Class(_ context.Context, _ uuid.UUID) (*assets.Class, error) { return g.class, nil }
func (g editGetter) Asset(_ context.Context, _ uuid.UUID) (*assets.Asset, error) { return g.asset, nil }
func (g editGetter) PurchasesForAsset(_ context.Context, _ uuid.UUID) ([]*assets.Purchase, error) {
	return g.purchases, nil
}

// inlineUoW runs the work without a database; persistence is not under test.
type inlineUoW struct{}

func (inlineUoW) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

func editFixture(owner uuid.UUID, source assets.ClassSource) (editGetter, *assets.Commands) {
	class := &assets.Class{ID: uuid.New(), AccountID: owner, Name: "Savings", Source: source}
	asset := &assets.Asset{ID: uuid.New(), ClassID: class.ID, AccountID: owner, Name: "Cash", Class: class}
	class.Assets = []assets.Asset{*asset}
	class.Mutations = []assets.Mutation{{
		ID: uuid.New(), AssetID: asset.ID, ClassID: class.ID, ChangeType: assets.ChangeTypeSet,
		EffectiveDate: time.Now().AddDate(0, -1, 0),
	}}
	getter := editGetter{class: class, asset: asset}
	aq := account.NewQueries(fakeAccountStore{exists: true})
	commands := assets.NewCommands(fakeAssetsStore{}, getter, *aq, inlineUoW{}, nil, nil, nil)
	return getter, commands
}

func TestCreateAssetRefusesAnotherAccountsClass(t *testing.T) {
	owner, intruder := uuid.New(), uuid.New()
	getter, commands := editFixture(owner, assets.ClassSourceManual)

	_, err := commands.CreateAsset(context.Background(), intruder, getter.class.ID, "Injected", 0, time.Now(), nil)
	if !errors.Is(err, assets.ErrClassAccountMismatch) {
		t.Fatalf("want ErrClassAccountMismatch, got %v", err)
	}
}

func TestEditsRefuseAnotherAccountsItem(t *testing.T) {
	owner, intruder := uuid.New(), uuid.New()
	getter, commands := editFixture(owner, assets.ClassSourceManual)

	err := commands.UpdateMutation(context.Background(), intruder, getter.asset.ID, getter.class.Mutations[0].ID,
		assets.MutationEdit{EffectiveDate: time.Now()})
	if !errors.Is(err, assets.ErrClassAccountMismatch) {
		t.Fatalf("update worth change: want ErrClassAccountMismatch, got %v", err)
	}
	if err := commands.DeleteAsset(context.Background(), intruder, getter.asset.ID); !errors.Is(err, assets.ErrClassAccountMismatch) {
		t.Fatalf("delete asset: want ErrClassAccountMismatch, got %v", err)
	}
}

func TestEditsRefuseThePortfolioClass(t *testing.T) {
	owner := uuid.New()
	getter, commands := editFixture(owner, assets.ClassSourcePortfolio)

	if err := commands.RenameAsset(context.Background(), owner, getter.asset.ID, "Renamed"); !errors.Is(err, assets.ErrClassReserved) {
		t.Fatalf("rename: want ErrClassReserved, got %v", err)
	}
	if err := commands.DeleteAsset(context.Background(), owner, getter.asset.ID); !errors.Is(err, assets.ErrClassReserved) {
		t.Fatalf("delete: want ErrClassReserved, got %v", err)
	}
}

// The refusal itself is counted and raised by the store, in the statement that
// removes the row; the command has to carry it out unchanged so the handler can
// still answer 409.
func TestDeleteMutationKeepsTheLastChange(t *testing.T) {
	owner := uuid.New()
	getter, _ := editFixture(owner, assets.ClassSourceManual)
	aq := account.NewQueries(fakeAccountStore{exists: true})
	store := fakeAssetsStore{lastRow: assets.ErrLastMutation}
	commands := assets.NewCommands(store, getter, *aq, inlineUoW{}, nil, nil, nil)

	err := commands.DeleteMutation(context.Background(), owner, getter.asset.ID, getter.class.Mutations[0].ID)
	if !errors.Is(err, assets.ErrLastMutation) {
		t.Fatalf("want ErrLastMutation, got %v", err)
	}
}

// recordingStore keeps the last row written per worth change, so a test can read
// back what the replay left behind.
type recordingStore struct {
	fakeAssetsStore
	rows map[uuid.UUID]assets.Mutation
}

func (r *recordingStore) UpdateMutation(_ context.Context, m *assets.Mutation) error {
	r.rows[m.ID] = *m
	return nil
}

// A correction whose amount happens to equal the row's previous worth plus its worth
// so far carries the signature of a change written before NewMutation was corrected.
// The replay must not read it as one: it would take the stored worth for the amount
// and shrink the user's correction without saying so.
func TestUpdateMutationKeepsAnAmountThatLooksLikeALegacyRow(t *testing.T) {
	owner := uuid.New()
	euros := func(units int64) money.Price { return money.Price(units * money.Scale) }
	onDay := func(value string) time.Time {
		parsed, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
		if err != nil {
			t.Fatalf("parse day %q: %v", value, err)
		}
		return parsed
	}
	increase := assets.ChangeDirectionIncrease

	class := &assets.Class{ID: uuid.New(), AccountID: owner, Name: "Savings", Source: assets.ClassSourceManual}
	asset := &assets.Asset{ID: uuid.New(), ClassID: class.ID, AccountID: owner, Name: "Cash", Class: class}
	class.Assets = []assets.Asset{*asset}
	deposit := assets.Mutation{
		ID: uuid.New(), AssetID: asset.ID, ClassID: class.ID, ChangeType: assets.ChangeTypeAdjust,
		Direction: &increase, Amount: euros(300), PreviousWorth: euros(1_000), NewWorth: euros(1_300),
		ClassTotalWorth: euros(1_300), EffectiveDate: onDay("2026-02-01"),
	}
	class.Mutations = []assets.Mutation{
		{
			ID: uuid.New(), AssetID: asset.ID, ClassID: class.ID, ChangeType: assets.ChangeTypeSet,
			Amount: euros(1_000), NewWorth: euros(1_000), ClassTotalWorth: euros(1_000),
			EffectiveDate: onDay("2026-01-01"),
		},
		deposit,
	}

	store := &recordingStore{rows: make(map[uuid.UUID]assets.Mutation)}
	aq := account.NewQueries(fakeAccountStore{exists: true})
	commands := assets.NewCommands(store, editGetter{class: class, asset: asset}, *aq, inlineUoW{}, nil, nil, nil)

	// 2300 is exactly the previous worth plus the worth the row carried.
	edit := assets.MutationEdit{Amount: euros(2_300), Direction: &increase, EffectiveDate: onDay("2026-02-01")}
	if err := commands.UpdateMutation(context.Background(), owner, asset.ID, deposit.ID, edit); err != nil {
		t.Fatalf("update worth change: %v", err)
	}

	written, ok := store.rows[deposit.ID]
	if !ok {
		t.Fatal("the corrected worth change was never written")
	}
	if written.Amount != euros(2_300) || written.PreviousWorth != euros(1_000) || written.NewWorth != euros(3_300) {
		t.Fatalf("want 2300 taking 1000 -> 3300, got %v taking %v -> %v",
			written.Amount, written.PreviousWorth, written.NewWorth)
	}
}

// orderedStore records the store calls a command makes, in order.
type orderedStore struct {
	fakeAssetsStore
	calls []string
}

func (o *orderedStore) record(name string) error {
	o.calls = append(o.calls, name)
	return nil
}

func (o *orderedStore) LockClass(_ context.Context, _ uuid.UUID) error { return o.record("LockClass") }
func (o *orderedStore) CreateAsset(_ context.Context, _ *assets.Asset) error {
	return o.record("CreateAsset")
}
func (o *orderedStore) SetWorth(_ context.Context, _ *assets.Asset) error {
	return o.record("SetWorth")
}
func (o *orderedStore) CreateMutation(_ context.Context, _ *assets.Mutation) error {
	return o.record("CreateMutation")
}
func (o *orderedStore) UpdateMutation(_ context.Context, _ *assets.Mutation) error {
	return o.record("UpdateMutation")
}
func (o *orderedStore) DeleteClass(_ context.Context, _ uuid.UUID) error {
	return o.record("DeleteClass")
}
func (o *orderedStore) DeleteAsset(_ context.Context, _ uuid.UUID) error {
	return o.record("DeleteAsset")
}
func (o *orderedStore) DeleteMutation(_ context.Context, _, _ uuid.UUID) error {
	return o.record("DeleteMutation")
}
func (o *orderedStore) UpdatePurchase(_ context.Context, _ *assets.Purchase) error {
	return o.record("UpdatePurchase")
}
func (o *orderedStore) DeletePurchase(_ context.Context, _, _ uuid.UUID) error {
	return o.record("DeletePurchase")
}

// fixedAggregator stands in for the class total the write side reads.
type fixedAggregator struct{}

func (fixedAggregator) AggregateValue(_ context.Context, _, _ uuid.UUID) (money.Price, error) {
	return 0, nil
}

// Every write to a class's history replays the whole class, so each of these commands
// has to claim the class row before it writes anything: a lock taken per item is an
// order another path can take the other way around, and two of those deadlock.
func TestHistoryWritesClaimTheClassFirst(t *testing.T) {
	owner := uuid.New()

	fixture := func(dailyPriced bool) (editGetter, *orderedStore, *assets.Commands) {
		class := &assets.Class{ID: uuid.New(), AccountID: owner, Name: "Savings", Source: assets.ClassSourceManual}
		asset := &assets.Asset{ID: uuid.New(), ClassID: class.ID, AccountID: owner, Name: "Cash", Class: class}
		class.Mutations = []assets.Mutation{{
			ID: uuid.New(), AssetID: asset.ID, ClassID: class.ID, ChangeType: assets.ChangeTypeSet,
			EffectiveDate: time.Now().AddDate(0, -1, 0),
		}}
		getter := editGetter{class: class, asset: asset}
		if dailyPriced {
			listing := uuid.New()
			asset.ListingID = &listing
			getter.purchases = []*assets.Purchase{
				{ID: uuid.New(), AssetID: asset.ID},
				{ID: uuid.New(), AssetID: asset.ID},
			}
		}
		class.Assets = []assets.Asset{*asset}
		store := &orderedStore{}
		aq := account.NewQueries(fakeAccountStore{exists: true})
		return getter, store, assets.NewCommands(store, getter, *aq, inlineUoW{}, fixedAggregator{}, nil, nil)
	}

	cases := []struct {
		name        string
		dailyPriced bool
		run         func(*assets.Commands, editGetter) error
	}{
		{"create item", false, func(c *assets.Commands, g editGetter) error {
			_, err := c.CreateAsset(context.Background(), owner, g.class.ID, "Savings account", 0, time.Now(), nil)
			return err
		}},
		{"set worth", false, func(c *assets.Commands, g editGetter) error {
			return c.UpdateAssetWorth(context.Background(), owner, g.asset.ID, 0, assets.ChangeTypeSet, nil, time.Now(), nil)
		}},
		{"edit worth change", false, func(c *assets.Commands, g editGetter) error {
			return c.UpdateMutation(context.Background(), owner, g.asset.ID, g.class.Mutations[0].ID,
				assets.MutationEdit{EffectiveDate: time.Now()})
		}},
		{"delete worth change", false, func(c *assets.Commands, g editGetter) error {
			return c.DeleteMutation(context.Background(), owner, g.asset.ID, g.class.Mutations[0].ID)
		}},
		{"delete item", false, func(c *assets.Commands, g editGetter) error {
			return c.DeleteAsset(context.Background(), owner, g.asset.ID)
		}},
		{"delete class", false, func(c *assets.Commands, g editGetter) error {
			return c.DeleteClass(context.Background(), owner, g.class.ID)
		}},
		{"edit purchase", true, func(c *assets.Commands, g editGetter) error {
			return c.UpdatePurchase(context.Background(), owner, g.asset.ID, g.purchases[0].ID,
				assets.PurchaseInput{Date: time.Now(), Quantity: 2, UnitPrice: money.Price(money.Scale)})
		}},
		{"delete purchase", true, func(c *assets.Commands, g editGetter) error {
			return c.DeletePurchase(context.Background(), owner, g.asset.ID, g.purchases[0].ID)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getter, store, commands := fixture(tc.dailyPriced)
			if err := tc.run(commands, getter); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if len(store.calls) < 2 {
				t.Fatalf("expected the class lock and a write, got %v", store.calls)
			}
			if store.calls[0] != "LockClass" {
				t.Fatalf("expected the class to be claimed first, got %v", store.calls)
			}
		})
	}
}

func TestDeletePurchaseKeepsTheLastPurchase(t *testing.T) {
	owner := uuid.New()
	getter, _ := editFixture(owner, assets.ClassSourceManual)
	listing := uuid.New()
	getter.asset.ListingID = &listing
	purchase := &assets.Purchase{ID: uuid.New(), AssetID: getter.asset.ID}
	getter.purchases = []*assets.Purchase{purchase}
	aq := account.NewQueries(fakeAccountStore{exists: true})
	store := fakeAssetsStore{lastRow: assets.ErrLastPurchase}
	commands := assets.NewCommands(store, getter, *aq, inlineUoW{}, nil, nil, nil)

	err := commands.DeletePurchase(context.Background(), owner, getter.asset.ID, purchase.ID)
	if !errors.Is(err, assets.ErrLastPurchase) {
		t.Fatalf("want ErrLastPurchase, got %v", err)
	}
}
