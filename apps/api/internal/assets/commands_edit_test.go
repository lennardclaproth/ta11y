package assets_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/assets"
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

func TestDeleteMutationKeepsTheLastChange(t *testing.T) {
	owner := uuid.New()
	getter, commands := editFixture(owner, assets.ClassSourceManual)

	err := commands.DeleteMutation(context.Background(), owner, getter.asset.ID, getter.class.Mutations[0].ID)
	if !errors.Is(err, assets.ErrLastMutation) {
		t.Fatalf("want ErrLastMutation, got %v", err)
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
	commands := assets.NewCommands(fakeAssetsStore{}, getter, *aq, inlineUoW{}, nil, nil, nil)

	err := commands.DeletePurchase(context.Background(), owner, getter.asset.ID, purchase.ID)
	if !errors.Is(err, assets.ErrLastPurchase) {
		t.Fatalf("want ErrLastPurchase, got %v", err)
	}
}
