//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/storage"
)

// TestDeleteClassRemovesItsChildren guards the worth of a deleted class. The
// snapshot builder reads mutations per account rather than per existing class, so
// an item or a worth change left behind keeps counting towards net worth with no
// way back to it from the app. SQLite runs without foreign-key enforcement, so the
// declared cascades do not clean up here.
func TestDeleteClassRemovesItsChildren(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXAssetsStore(db)
		accID := uuid.New()

		if err := store.CreateAccount(ctx, assets.NewAccount(accID)); err != nil {
			t.Fatalf("create account projection: %v", err)
		}
		class, err := assets.NewClass(accID, nil, "Coins")
		if err != nil {
			t.Fatalf("new class: %v", err)
		}
		if err := store.CreateClass(ctx, class); err != nil {
			t.Fatalf("create class: %v", err)
		}
		item, err := assets.NewAsset(accID, class.ID, "Bitcoin", 0, "")
		if err != nil {
			t.Fatalf("new asset: %v", err)
		}
		listingID := uuid.New()
		item.ListingID = &listingID
		if err := store.CreateAsset(ctx, item); err != nil {
			t.Fatalf("create asset: %v", err)
		}
		mutation, err := assets.NewMutation(
			accID, class.ID, item.ID,
			assets.ChangeTypeSet, nil,
			money.Price(100_000_000), 0, money.Price(100_000_000),
			time.Now().UTC(), nil,
		)
		if err != nil {
			t.Fatalf("new mutation: %v", err)
		}
		if err := store.CreateMutation(ctx, mutation); err != nil {
			t.Fatalf("create mutation: %v", err)
		}
		purchase, err := assets.NewPurchase(accID, item.ID, time.Now().UTC().AddDate(0, 0, -1), 0.5, money.Price(100_000_000))
		if err != nil {
			t.Fatalf("new purchase: %v", err)
		}
		if err := store.CreatePurchase(ctx, purchase); err != nil {
			t.Fatalf("create purchase: %v", err)
		}

		if err := store.DeleteClass(ctx, class.ID); err != nil {
			t.Fatalf("delete class: %v", err)
		}

		if got, err := store.Class(ctx, class.ID); err != nil || got != nil {
			t.Fatalf("class after delete: got %v, err %v", got, err)
		}
		if got, err := store.Asset(ctx, item.ID); err != nil || got != nil {
			t.Fatalf("item after delete: got %v, err %v", got, err)
		}
		mutations, err := store.Mutations(ctx, accID, nil, nil, nil)
		if err != nil {
			t.Fatalf("read mutations: %v", err)
		}
		if len(mutations) != 0 {
			t.Errorf("expected no worth changes left, got %d", len(mutations))
		}
		purchases, err := store.PurchasesForAccount(ctx, accID)
		if err != nil {
			t.Fatalf("read purchases: %v", err)
		}
		if len(purchases) != 0 {
			t.Errorf("expected no purchases left, got %d", len(purchases))
		}
	})
}

// TestMutationsForClassesGroupsNewestFirst guards the batched read behind class growth:
// every requested class gets its own mutations, newest first, and other classes are left out.
func TestMutationsForClassesGroupsNewestFirst(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXAssetsStore(db)
		accID := uuid.New()

		if err := store.CreateAccount(ctx, assets.NewAccount(accID)); err != nil {
			t.Fatalf("create account projection: %v", err)
		}
		newClassWithItem := func(name string) (*assets.Class, *assets.Asset) {
			class, err := assets.NewClass(accID, nil, name)
			if err != nil {
				t.Fatalf("new class: %v", err)
			}
			if err := store.CreateClass(ctx, class); err != nil {
				t.Fatalf("create class: %v", err)
			}
			item, err := assets.NewAsset(accID, class.ID, name+" item", 0, "")
			if err != nil {
				t.Fatalf("new asset: %v", err)
			}
			if err := store.CreateAsset(ctx, item); err != nil {
				t.Fatalf("create asset: %v", err)
			}
			return class, item
		}
		addMutation := func(class *assets.Class, item *assets.Asset, worth money.Price, at time.Time) {
			m, err := assets.NewMutation(accID, class.ID, item.ID, assets.ChangeTypeSet, nil, worth, 0, worth, at, nil)
			if err != nil {
				t.Fatalf("new mutation: %v", err)
			}
			if err := store.CreateMutation(ctx, m); err != nil {
				t.Fatalf("create mutation: %v", err)
			}
		}

		cash, cashItem := newClassWithItem("Cash")
		house, houseItem := newClassWithItem("House")
		other, otherItem := newClassWithItem("Other")
		jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		addMutation(cash, cashItem, money.Price(100_00), jan)
		addMutation(cash, cashItem, money.Price(200_00), jan.AddDate(0, 1, 0))
		addMutation(house, houseItem, money.Price(900_00), jan)
		addMutation(other, otherItem, money.Price(1_00), jan)

		got, err := store.MutationsForClasses(ctx, []uuid.UUID{cash.ID, house.ID})
		if err != nil {
			t.Fatalf("mutations for classes: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("classes = %d, want 2", len(got))
		}
		if c := got[cash.ID]; len(c) != 2 || c[0].ClassTotalWorth != money.Price(200_00) || c[1].ClassTotalWorth != money.Price(100_00) {
			t.Errorf("cash mutations = %+v, want 200.00 then 100.00", c)
		}
		if h := got[house.ID]; len(h) != 1 {
			t.Errorf("house mutations = %d, want 1", len(h))
		}
	})
}
