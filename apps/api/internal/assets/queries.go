package assets

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
)

type Queries struct {
	qs  QueryStore
	mdq *marketdata.Queries
}

// NewQueries creates assets read-side use cases. The market-data queries may be
// nil, in which case daily-priced items are reported without their instrument
// metadata and price date.
func NewQueries(qs QueryStore, mdq *marketdata.Queries) *Queries {
	return &Queries{qs: qs, mdq: mdq}
}

// QueryStore reads asset classes, bounds, purchases, and snapshots for the read models.
type QueryStore interface {
	ClassesForAccount(ctx context.Context, accID uuid.UUID, includeArchived bool) (map[uuid.UUID]*Class, error)
	Class(ctx context.Context, classID uuid.UUID) (*Class, error)
	ClassBounds(ctx context.Context, classIDs []uuid.UUID) (map[uuid.UUID]*ClassBounds, error)
	Snapshots(ctx context.Context, accID uuid.UUID, from, to *time.Time) ([]*Snapshot, error)
	Asset(ctx context.Context, assetID uuid.UUID) (*Asset, error)
	PurchasesForClass(ctx context.Context, classID uuid.UUID) ([]*Purchase, error)
	PurchasesForAsset(ctx context.Context, assetID uuid.UUID) ([]*Purchase, error)
	MutationsForAsset(ctx context.Context, assetID uuid.UUID) ([]*Mutation, error)
}

type ClassBounds struct {
	ClassID uuid.UUID
	First   *Mutation
	Last    *Mutation
}

type ClassSummary struct {
	ID           uuid.UUID
	Name         string
	Source       ClassSource
	Archived     bool
	CurrentWorth money.Price
	LastChangeAt *time.Time
	GrowthPct    *float64
	UpdatedAt    time.Time
}

// ListClasses returns table rows for account classes.
func (q *Queries) ListClasses(ctx context.Context, accountID uuid.UUID, includeArchived bool) ([]ClassSummary, error) {
	// Get classes
	classes, err := q.qs.ClassesForAccount(ctx, accountID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("list classes: failed to get classes for account %s: %w", accountID, err)
	}
	// Make list of class ID's, initialize empty and allocate memory
	// for len(classes)
	ids := make([]uuid.UUID, 0, len(classes))
	for id := range classes {
		ids = append(ids, id)
	}
	// Get bounds of classes (first mutation, last mutation)
	bounds, err := q.qs.ClassBounds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list classes: failed to get bounds: %w", err)
	}
	summaries := make([]ClassSummary, 0, len(ids))
	// Determine growth
	for _, id := range ids {
		class, ok := classes[id]
		if !ok {
			return nil, fmt.Errorf("list classes: class not in map")
		}
		summary := ClassSummary{
			ID:       class.ID,
			Name:     class.Name,
			Archived: class.Archived,
		}
		bound, ok := bounds[id]
		// if bound.First is nil we can assume that there have been
		// no mutations. When bound.First is not null then we assume bound.Last
		// is always filled, if a mutation is done and it is the first mutation then
		// it is automatically also the last mutation.
		if !ok {
			summaries = append(summaries, summary)
			continue
		}
		// This might seem redundant but it clarifies the code
		if bound.First == nil {
			summaries = append(summaries, summary)
			continue
		}
		growth := growthPctFromBouds(*bound)
		summary.GrowthPct = &growth
		if bound.Last != nil {
			summary.CurrentWorth = bound.Last.ClassTotalWorth
			summary.LastChangeAt = &bound.Last.EffectiveDate
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

type AssetSummary struct {
	ID           uuid.UUID
	Name         string
	CurrentWorth money.Price
	Archived     bool
	UpdatedAt    time.Time
}

type GrowthPoint struct {
	Date       time.Time
	TotalWorth money.Price
}

type ClassDetails struct {
	Class  ClassSummary
	Assets []AssetSummary
	// Holdings are the class's daily-priced items. They are reported separately
	// from Assets because they carry a quantity, a price and a purchase history
	// that an item with a hand-set worth has no equivalent of.
	Holdings  []HoldingSummary
	Growth    []GrowthPoint
	Mutations []Mutation
}

// ClassDetails returns class assets, holdings, growth points, and mutations.
func (q *Queries) ClassDetails(ctx context.Context, classID, accID uuid.UUID) (*ClassDetails, error) {
	// Get class including assets and mutations (mutations sorted DESC)
	class, err := q.qs.Class(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("class details: failed to get class: %w", err)
	}
	if class == nil {
		return nil, fmt.Errorf("class details: %w", ErrClassNotFound)
	}
	if class.AccountID != accID {
		return nil, fmt.Errorf("class details: %w", ErrClassAccountMismatch)
	}

	details := &ClassDetails{
		Class: ClassSummary{
			ID:        class.ID,
			Name:      class.Name,
			Source:    class.Source,
			Archived:  class.Archived,
			UpdatedAt: class.UpdatedAt,
		},
		Assets:    make([]AssetSummary, 0, len(class.Assets)),
		Mutations: make([]Mutation, 0, len(class.Mutations)),
	}

	// A class whose items have no recorded worth yet -- a brand-new class, or one
	// holding a daily-priced item whose first rebuild has not run -- has no bounds
	// to derive a total or a growth percentage from.
	if len(class.Mutations) > 0 {
		// Mutations arrive newest first, so the latest is the first entry and the
		// earliest the last.
		bound := ClassBounds{
			ClassID: classID,
			Last:    &class.Mutations[0],
			First:   &class.Mutations[len(class.Mutations)-1],
		}
		growthPct := growthPctFromBouds(bound)
		details.Class.CurrentWorth = bound.Last.ClassTotalWorth
		details.Class.LastChangeAt = &bound.Last.EffectiveDate
		details.Class.GrowthPct = &growthPct
		details.Growth = growthPointsFromMutations(class.Mutations)
	}

	dailyPriced := make(map[uuid.UUID]*Asset, len(class.Assets))
	for i := range class.Assets {
		asset := &class.Assets[i]
		if asset.IsDailyPriced() {
			dailyPriced[asset.ID] = asset
			continue
		}
		details.Assets = append(details.Assets, AssetSummary{
			ID:           asset.ID,
			Name:         asset.Name,
			CurrentWorth: asset.CurrentWorth,
			Archived:     asset.Archived,
			UpdatedAt:    asset.UpdatedAt,
		})
	}

	// The derived mutations of a daily-priced item are a price feed, not a record
	// of something the user did, so they stay out of the change list. Their class
	// totals are still what Growth above is built from.
	for _, mutation := range class.Mutations {
		if _, derived := dailyPriced[mutation.AssetID]; derived {
			continue
		}
		details.Mutations = append(details.Mutations, mutation)
	}

	if len(dailyPriced) > 0 {
		purchases, err := q.qs.PurchasesForClass(ctx, classID)
		if err != nil {
			return nil, fmt.Errorf("class details: failed to get purchases: %w", err)
		}
		byAsset := purchasesByAsset(purchases)
		worthByAsset := worthSeriesByAsset(class.Mutations)
		for i := range class.Assets {
			asset := &class.Assets[i]
			if !asset.IsDailyPriced() {
				continue
			}
			summary, err := q.holdingSummary(ctx, asset, byAsset[asset.ID], worthByAsset[asset.ID])
			if err != nil {
				return nil, err
			}
			details.Holdings = append(details.Holdings, *summary)
		}
	}

	return details, nil
}

// HoldingDetails returns one daily-priced item with the purchases behind it.
func (q *Queries) HoldingDetails(ctx context.Context, assetID, accID uuid.UUID) (*HoldingDetails, error) {
	asset, err := q.qs.Asset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("holding details: failed to get asset: %w", err)
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	if asset.AccountID != accID {
		return nil, ErrClassAccountMismatch
	}
	if !asset.IsDailyPriced() {
		return nil, ErrAssetNotDailyPriced
	}

	purchases, err := q.qs.PurchasesForAsset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("holding details: failed to get purchases: %w", err)
	}
	mutations, err := q.qs.MutationsForAsset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("holding details: failed to get worth history: %w", err)
	}
	worth := make([]Mutation, 0, len(mutations))
	for _, mutation := range mutations {
		if mutation != nil {
			worth = append(worth, *mutation)
		}
	}

	summary, err := q.holdingSummary(ctx, asset, purchases, worthSeriesByAsset(worth)[assetID])
	if err != nil {
		return nil, err
	}
	details := &HoldingDetails{HoldingSummary: *summary, Purchases: make([]Purchase, 0, len(purchases))}
	if asset.Class != nil {
		details.ClassName = asset.Class.Name
	}
	for _, purchase := range purchases {
		if purchase != nil {
			details.Purchases = append(details.Purchases, *purchase)
		}
	}
	return details, nil
}

// holdingSummary assembles the figures a daily-priced item is read by. The worth
// series is the item's own derived mutations, so the value it reports is exactly
// the one that fed the account's net worth.
func (q *Queries) holdingSummary(
	ctx context.Context,
	asset *Asset,
	purchases []*Purchase,
	worth []HoldingWorthPoint,
) (*HoldingSummary, error) {
	summary := &HoldingSummary{
		AssetID:   asset.ID,
		ClassID:   asset.ClassID,
		Name:      asset.Name,
		ListingID: *asset.ListingID,
		Value:     asset.CurrentWorth,
	}
	for _, purchase := range purchases {
		if purchase == nil {
			continue
		}
		summary.Quantity += purchase.Quantity
		summary.Paid += purchase.Paid()
	}
	if summary.Quantity > 0 {
		summary.AvgUnitPrice = money.Price(float64(summary.Paid) / summary.Quantity)
	}
	summary.Unrealized = summary.Value - summary.Paid
	if summary.Paid > 0 {
		pct := (float64(summary.Unrealized) / float64(summary.Paid)) * 100
		summary.UnrealizedPct = &pct
	}
	summary.Series = buildHoldingSeries(worth, purchases)

	if q.mdq != nil {
		listing, err := q.mdq.Listing(ctx, *asset.ListingID)
		if err != nil {
			return nil, fmt.Errorf("holding summary: failed to get listing: %w", err)
		}
		if listing != nil {
			summary.Symbol = listing.Symbol
			summary.InstrumentName = listing.Name
		}
		latest, err := q.mdq.LatestEOD(ctx, *asset.ListingID)
		if err != nil {
			return nil, fmt.Errorf("holding summary: failed to get latest price: %w", err)
		}
		if latest != nil {
			day := date.StartOfDayUTC(latest.Date)
			summary.Price = latest.Close
			summary.PriceDate = &day
			summary.PriceCarriedForward = day.Before(date.StartOfDayUTC(time.Now().UTC()))
		}
	}
	return summary, nil
}

// ListSnapshots returns account-level daily total snapshots.
func (q *Queries) ListSnapshots(ctx context.Context, accID uuid.UUID, from, to *time.Time) ([]GrowthPoint, error) {
	snapshots, err := q.qs.Snapshots(ctx, accID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: failed to get snapshots: %w", err)
	}
	out := make([]GrowthPoint, 0, len(snapshots))
	for _, row := range snapshots {
		if row == nil {
			continue
		}
		out = append(out, GrowthPoint{
			Date:       row.OccurredAt,
			TotalWorth: row.TotalWorth,
		})
	}
	return out, nil
}
