package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
)

// assetsRestore keeps what the account has: an item it already holds keeps its own
// history, because mutations carry running totals that two histories would corrupt.
type assetsRestore struct {
	store  AccountWriter
	accID  uuid.UUID
	file   *AssetsFile
	now    time.Time
	result *AssetsResult

	newClasses []pendingClass
	newItems   []pendingItem
}

type pendingClass struct {
	file *AssetClass
	id   uuid.UUID
}

type pendingItem struct {
	file    *AssetItem
	classID uuid.UUID
}

func newAssetsRestore(store AccountWriter, accID uuid.UUID, file *AssetsFile, result *AssetsResult, now time.Time) *assetsRestore {
	return &assetsRestore{store: store, accID: accID, file: file, now: now, result: result}
}

func (r *assetsRestore) run(ctx context.Context) error {
	if err := r.plan(ctx); err != nil {
		return err
	}
	listingIDs, err := r.resolveListings(ctx)
	if err != nil {
		return err
	}
	return r.write(ctx, listingIDs)
}

func (r *assetsRestore) plan(ctx context.Context) error {
	classes, err := r.store.AssetClasses(ctx, r.accID)
	if err != nil {
		return fmt.Errorf("backup import: asset classes: %w", err)
	}
	items, err := r.store.AssetItems(ctx, r.accID)
	if err != nil {
		return fmt.Errorf("backup import: asset items: %w", err)
	}
	classByName := make(map[string]*assets.Class, len(classes))
	for _, c := range classes {
		classByName[c.Name] = c
	}
	held := make(map[uuid.UUID]map[string]struct{})
	for _, i := range items {
		if held[i.ClassID] == nil {
			held[i.ClassID] = make(map[string]struct{})
		}
		held[i.ClassID][i.Name] = struct{}{}
	}

	for ci := range r.file.Classes {
		fc := &r.file.Classes[ci]
		target, exists := classByName[fc.Name]
		classID := uuid.New()
		if exists {
			classID = target.ID
			r.result.Classes.Existing++
		} else {
			r.newClasses = append(r.newClasses, pendingClass{file: fc, id: classID})
			r.result.Classes.New++
		}
		for ii := range fc.Items {
			fi := &fc.Items[ii]
			// The portfolio sync rewrites its own class, so nothing is added to it.
			_, holds := held[classID][fi.Name]
			if exists && (holds || target.Source == assets.ClassSourcePortfolio) {
				r.result.Items.Existing++
				r.result.Mutations.Existing += len(fi.Mutations)
				r.result.Purchases.Existing += len(fi.Purchases)
				continue
			}
			r.newItems = append(r.newItems, pendingItem{file: fi, classID: classID})
			r.result.Items.New++
			r.result.Mutations.New += len(fi.Mutations)
			r.result.Purchases.New += len(fi.Purchases)
		}
	}
	return nil
}

// resolveListings refuses the file when a new daily-priced item's listing is missing.
func (r *assetsRestore) resolveListings(ctx context.Context) (map[ListingRef]uuid.UUID, error) {
	listingIDs, err := r.store.ListingIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("backup import: listings: %w", err)
	}
	missing := make(map[ListingRef]struct{})
	for _, i := range r.newItems {
		if i.file.Listing == nil {
			continue
		}
		if _, ok := listingIDs[*i.file.Listing]; !ok {
			missing[*i.file.Listing] = struct{}{}
		}
	}
	for l := range missing {
		r.result.Missing.Listings = append(r.result.Missing.Listings, l)
	}
	sortListingRefs(r.result.Missing.Listings)
	if !r.result.Missing.empty() {
		return nil, ErrMissingReferences
	}
	return listingIDs, nil
}

func (r *assetsRestore) write(ctx context.Context, listingIDs map[ListingRef]uuid.UUID) error {
	classes := make([]*assets.Class, 0, len(r.newClasses))
	for _, c := range r.newClasses {
		classes = append(classes, &assets.Class{
			ID: c.id, AccountID: r.accID, Name: c.file.Name, Source: assets.ClassSourceManual,
			Archived: c.file.Archived, CreatedAt: c.file.CreatedAt, UpdatedAt: r.now,
		})
	}

	items := make([]*assets.Asset, 0, len(r.newItems))
	var mutations []*assets.Mutation
	var purchases []*assets.Purchase
	for _, p := range r.newItems {
		fi := p.file
		item := &assets.Asset{
			ID: uuid.New(), ClassID: p.classID, AccountID: r.accID, Name: fi.Name, CurrentWorth: fi.CurrentWorth,
			Archived: fi.Archived, CreatedAt: fi.CreatedAt, UpdatedAt: r.now,
		}
		if fi.Listing != nil {
			id := listingIDs[*fi.Listing]
			item.ListingID = &id
		}
		items = append(items, item)
		for _, m := range fi.Mutations {
			mutations = append(mutations, &assets.Mutation{
				ID: uuid.New(), AccountID: r.accID, ClassID: p.classID, AssetID: item.ID,
				ChangeType: m.ChangeType, Direction: m.Direction, Amount: m.Amount,
				PreviousWorth: m.PreviousWorth, NewWorth: m.NewWorth, ClassTotalWorth: m.ClassTotalWorth,
				EffectiveDate: m.EffectiveDate.Time, Note: m.Note, CreatedAt: m.CreatedAt,
			})
		}
		for _, pu := range fi.Purchases {
			purchases = append(purchases, &assets.Purchase{
				ID: uuid.New(), AccountID: r.accID, AssetID: item.ID, PurchasedOn: pu.PurchasedOn.Time,
				Quantity: pu.Quantity, UnitPrice: pu.UnitPrice, CreatedAt: pu.CreatedAt,
			})
		}
	}

	if err := r.store.InsertAssetClasses(ctx, classes); err != nil {
		return fmt.Errorf("backup import: asset classes: %w", err)
	}
	if err := r.store.InsertAssetItems(ctx, items); err != nil {
		return fmt.Errorf("backup import: asset items: %w", err)
	}
	if err := r.store.InsertAssetMutations(ctx, mutations); err != nil {
		return fmt.Errorf("backup import: asset mutations: %w", err)
	}
	if err := r.store.InsertAssetPurchases(ctx, purchases); err != nil {
		return fmt.Errorf("backup import: asset purchases: %w", err)
	}
	return nil
}
