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
