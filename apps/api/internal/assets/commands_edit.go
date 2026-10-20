package assets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// MutationEdit is a corrected worth change. Its kind stays what it was: a SET keeps
// carrying the new worth, an ADJUST the amount added or withdrawn.
type MutationEdit struct {
	Amount        money.Price
	Direction     *ChangeDirection
	EffectiveDate time.Time
	Note          *string
}

// editableAsset loads an item for an edit and checks it belongs to the account and
// sits in a class the user owns the history of. The portfolio class is written by
// the portfolio sync, so nothing in it is edited by hand.
func (c *Commands) editableAsset(ctx context.Context, accID, assetID uuid.UUID) (*Asset, error) {
	asset, err := c.cg.Asset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	if asset.Class == nil {
		return nil, fmt.Errorf("asset is not populated from database, this error should not occur")
	}
	if asset.Class.AccountID != accID {
		return nil, ErrClassAccountMismatch
	}
	if asset.Class.Source == ClassSourcePortfolio {
		return nil, ErrClassReserved
	}
	if asset.Class.Source != ClassSourceManual {
		return nil, ErrClassNotManual
	}
	return asset, nil
}

// RenameAsset changes an item's name. The name carries no worth, so nothing is rebuilt.
func (c *Commands) RenameAsset(ctx context.Context, accID, assetID uuid.UUID, name string) error {
	asset, err := c.editableAsset(ctx, accID, assetID)
	if err != nil {
		return fmt.Errorf("rename asset: %w", err)
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("rename asset: %w", ErrAssetNameEmpty)
	}
	asset.Name = trimmed
	if err := c.cs.UpdateAsset(ctx, asset); err != nil {
		return fmt.Errorf("rename asset: %w", err)
	}
	return nil
}

// DeleteAsset removes an item with its worth history and purchases, and replays the
// class's history without it.
func (c *Commands) DeleteAsset(ctx context.Context, accID, assetID uuid.UUID) error {
	asset, err := c.editableAsset(ctx, accID, assetID)
	if err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}
	if err := c.uow.Do(ctx, func(txCtx context.Context) error {
		if err := c.cs.DeleteAsset(txCtx, asset.ID); err != nil {
			return err
		}
		return c.resequence(txCtx, asset.ClassID)
	}); err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}
	c.publishSnapshotsRebuildRequested(ctx, accID)
	return nil
}

// UpdateMutation corrects one worth change of an item whose worth is set by hand and
// replays the class's history, so a changed date or amount carries through to every
// later change.
func (c *Commands) UpdateMutation(ctx context.Context, accID, assetID, mutationID uuid.UUID, edit MutationEdit) error {
	if edit.Amount < 0 {
		return fmt.Errorf("update worth change: %w", ErrMutationAmount)
	}
	day := date.StartOfDayUTC(edit.EffectiveDate)
	if day.After(date.StartOfDayUTC(time.Now().UTC())) {
		return fmt.Errorf("update worth change: %w", ErrMutationInFuture)
	}
	// Resolved inside the transaction, so the change cannot move between the check and the write.
	if err := c.uow.Do(ctx, func(txCtx context.Context) error {
		asset, mutation, _, err := c.editableMutation(txCtx, accID, assetID, mutationID)
		if err != nil {
			return err
		}
		if mutation.ChangeType == ChangeTypeAdjust && edit.Direction == nil {
			return ErrMutationDirection
		}
		mutation.Amount = edit.Amount
		mutation.EffectiveDate = day
		mutation.Note = edit.Note
		if mutation.ChangeType == ChangeTypeAdjust {
			mutation.Direction = edit.Direction
		}
		// The replay recomputes the figures, but it reads them first to recognise a row
		// written before NewMutation was corrected. A row left holding the old worth
		// next to the new amount can match that signature by accident, and the user's
		// amount would be silently replaced; keeping the worth in step rules that out.
		mutation.NewWorth = worthAfter(mutation.ChangeType, mutation.Direction, mutation.Amount, mutation.PreviousWorth)
		if err := c.cs.UpdateMutation(txCtx, mutation); err != nil {
			return err
		}
		return c.resequence(txCtx, asset.ClassID)
	}); err != nil {
		return fmt.Errorf("update worth change: %w", err)
	}
	c.publishSnapshotsRebuildRequested(ctx, accID)
	return nil
}

// DeleteMutation removes one worth change and replays the class's history. An item
// keeps at least one change: without it the item has no worth to start from.
func (c *Commands) DeleteMutation(ctx context.Context, accID, assetID, mutationID uuid.UUID) error {
	if err := c.uow.Do(ctx, func(txCtx context.Context) error {
		asset, mutation, _, err := c.editableMutation(txCtx, accID, assetID, mutationID)
		if err != nil {
			return err
		}
		// The store counts in the same statement that removes the row, so two deletes
		// cannot each see a second change left and together empty the item's history.
		if err := c.cs.DeleteMutation(txCtx, asset.ID, mutation.ID); err != nil {
			return err
		}
		return c.resequence(txCtx, asset.ClassID)
	}); err != nil {
		return fmt.Errorf("delete worth change: %w", err)
	}
	c.publishSnapshotsRebuildRequested(ctx, accID)
	return nil
}

// editableMutation resolves a worth change of a hand-valued item, with the number
// of changes that item has.
func (c *Commands) editableMutation(ctx context.Context, accID, assetID, mutationID uuid.UUID) (*Asset, *Mutation, int, error) {
	asset, err := c.editableAsset(ctx, accID, assetID)
	if err != nil {
		return nil, nil, 0, err
	}
	// A daily-priced item's changes are derived from its purchases and rewritten on
	// every rebuild; its purchases are what is edited instead.
	if asset.IsDailyPriced() {
		return nil, nil, 0, ErrAssetDailyPriced
	}
	class, err := c.cg.Class(ctx, asset.ClassID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get class: %w", err)
	}
	if class == nil {
		return nil, nil, 0, ErrClassNotFound
	}
	var found *Mutation
	count := 0
	for i := range class.Mutations {
		if class.Mutations[i].AssetID != asset.ID {
			continue
		}
		count++
		if class.Mutations[i].ID == mutationID {
			found = &class.Mutations[i]
		}
	}
	if found == nil {
		return nil, nil, 0, ErrMutationNotFound
	}
	return asset, found, count, nil
}

// UpdatePurchase corrects one purchase of a daily-priced item. Its worth history is
// derived from the purchases, so the rebuild that follows carries the correction.
func (c *Commands) UpdatePurchase(ctx context.Context, accID, assetID, purchaseID uuid.UUID, input PurchaseInput) error {
	if err := c.uow.Do(ctx, func(txCtx context.Context) error {
		asset, purchase, _, err := c.editablePurchase(txCtx, accID, assetID, purchaseID)
		if err != nil {
			return err
		}
		corrected, err := NewPurchase(accID, asset.ID, input.Date, input.Quantity, input.UnitPrice)
		if err != nil {
			return err
		}
		corrected.ID = purchase.ID
		corrected.CreatedAt = purchase.CreatedAt
		return c.cs.UpdatePurchase(txCtx, corrected)
	}); err != nil {
		return fmt.Errorf("update purchase: %w", err)
	}
	c.publishSnapshotsRebuildRequested(ctx, accID)
	return nil
}

// DeletePurchase removes one purchase of a daily-priced item. The item keeps at least
// one: a holding with nothing bought has no worth to follow.
func (c *Commands) DeletePurchase(ctx context.Context, accID, assetID, purchaseID uuid.UUID) error {
	if err := c.uow.Do(ctx, func(txCtx context.Context) error {
		asset, purchase, _, err := c.editablePurchase(txCtx, accID, assetID, purchaseID)
		if err != nil {
			return err
		}
		// The store counts in the same statement that removes the row; see DeleteMutation.
		return c.cs.DeletePurchase(txCtx, asset.ID, purchase.ID)
	}); err != nil {
		return fmt.Errorf("delete purchase: %w", err)
	}
	c.publishSnapshotsRebuildRequested(ctx, accID)
	return nil
}

// editablePurchase resolves a purchase of a daily-priced item, with the number of
// purchases that item has.
func (c *Commands) editablePurchase(ctx context.Context, accID, assetID, purchaseID uuid.UUID) (*Asset, *Purchase, int, error) {
	asset, err := c.editableAsset(ctx, accID, assetID)
	if err != nil {
		return nil, nil, 0, err
	}
	if !asset.IsDailyPriced() {
		return nil, nil, 0, ErrAssetNotDailyPriced
	}
	purchases, err := c.cg.PurchasesForAsset(ctx, asset.ID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get purchases: %w", err)
	}
	for _, purchase := range purchases {
		if purchase.ID == purchaseID {
			return asset, purchase, len(purchases), nil
		}
	}
	return nil, nil, 0, ErrPurchaseNotFound
}
