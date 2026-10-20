package assets

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// resequenceClass replays a class's worth changes in date order and returns the
// hand-valued ones whose recorded worth or class total no longer matches, together
// with each hand-valued item's worth after its last change.
//
// A change is written against the worth as it was when it was entered, so one that
// is back-dated, edited or deleted leaves every later change in the class holding a
// stale previous worth and class total. Replaying from the start is what makes the
// history, and the net-worth snapshots read from it, right again.
//
// SET carries the new worth in Amount; ADJUST carries the change. Daily-priced items'
// derived changes are not rewritten here -- the holdings rebuild owns them -- but
// their worth counts towards the class total on the days they cover.
//
// Every hand-valued item has to open with a SET: an ADJUST first has no worth to
// apply to, which is refused with ErrOpeningNotSet.
func resequenceClass(items []Asset, mutations []Mutation) ([]*Mutation, map[uuid.UUID]money.Price, error) {
	derived := make(map[uuid.UUID]bool, len(items))
	counted := make(map[uuid.UUID]bool, len(items))
	for _, item := range items {
		derived[item.ID] = item.IsDailyPriced()
		// The class total has always summed the items that are not archived.
		counted[item.ID] = !item.Archived
	}

	ordered := make([]Mutation, len(mutations))
	copy(ordered, mutations)
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].EffectiveDate.Equal(ordered[j].EffectiveDate) {
			return ordered[i].EffectiveDate.Before(ordered[j].EffectiveDate)
		}
		if !ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
			return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
		}
		// Deterministic even for rows entered in the same instant.
		return ordered[i].ID.String() < ordered[j].ID.String()
	})

	running := make(map[uuid.UUID]money.Price, len(items))
	worth := make(map[uuid.UUID]money.Price, len(items))
	// The class total moves with the one item that just changed, so it is carried
	// forward rather than summed over every item again for every row. A class with
	// daily-priced items carries one derived row per item per day, which made the
	// walk cost items times rows.
	total := money.Price(0)
	changed := make([]*Mutation, 0)
	for i := range ordered {
		m := ordered[i]
		if derived[m.AssetID] {
			if counted[m.AssetID] {
				total += m.NewWorth - running[m.AssetID]
			}
			running[m.AssetID] = m.NewWorth
			continue
		}

		previous, opened := running[m.AssetID]
		if !opened && m.ChangeType != ChangeTypeSet {
			return nil, nil, ErrOpeningNotSet
		}
		next := m.Amount
		amountFixed := false
		if m.ChangeType == ChangeTypeAdjust {
			amount := m.Amount
			if delta, ok := legacyAdjustAmount(m); ok {
				// The row holds the running total where the change belongs, so it is
				// corrected as it is replayed rather than recomputed from a total.
				amount = delta
				m.Amount = delta
				amountFixed = true
			}
			next = worthAfter(m.ChangeType, m.Direction, amount, previous)
		}
		if counted[m.AssetID] {
			total += next - previous
		}
		running[m.AssetID] = next
		worth[m.AssetID] = next

		if amountFixed || m.PreviousWorth != previous || m.NewWorth != next || m.ClassTotalWorth != total {
			m.PreviousWorth = previous
			m.NewWorth = next
			m.ClassTotalWorth = total
			changed = append(changed, &m)
		}
	}
	return changed, worth, nil
}

// legacyAdjustAmount recognises an ADJUST that adds, written before NewMutation was
// corrected: it stored the running total in Amount and the change in NewWorth, the
// two the wrong way around. Replaying such a row from its Amount would add the total
// to the worth it was already part of, turning one wrong number into another and
// carrying that through every later change and the class total. The change is read
// from where it was actually written, and the row is rewritten with it.
//
// The signature is exact for a row that holds its own result: one written this way
// satisfies amount = previous_worth + new_worth, a correct one amount = new_worth -
// previous_worth, and those coincide only when previous_worth is zero, which is
// excluded. It says nothing about a row whose worth no longer matches its amount, so
// every writer keeps the two in step -- see worthAfter.
func legacyAdjustAmount(m Mutation) (money.Price, bool) {
	if m.ChangeType != ChangeTypeAdjust || m.Direction == nil || *m.Direction != ChangeDirectionIncrease {
		return 0, false
	}
	if m.PreviousWorth == 0 || m.Amount != m.PreviousWorth+m.NewWorth {
		return 0, false
	}
	return m.NewWorth, true
}

// resequence rewrites a manual class's history after one of its worth changes was
// added, edited or removed. It must run inside the caller's transaction.
func (c *Commands) resequence(ctx context.Context, classID uuid.UUID) error {
	class, err := c.cg.Class(ctx, classID)
	if err != nil {
		return fmt.Errorf("resequence: get class: %w", err)
	}
	if class == nil {
		return ErrClassNotFound
	}
	if class.Source != ClassSourceManual {
		// The portfolio class mirrors the portfolio sync and is rewritten by it.
		return nil
	}

	changed, worth, err := resequenceClass(class.Assets, class.Mutations)
	if err != nil {
		return err
	}
	for _, m := range changed {
		if err := c.cs.UpdateMutation(ctx, m); err != nil {
			return fmt.Errorf("resequence: update worth change: %w", err)
		}
	}
	for i := range class.Assets {
		item := &class.Assets[i]
		next, ok := worth[item.ID]
		if !ok || item.IsDailyPriced() || item.CurrentWorth == next {
			continue
		}
		item.CurrentWorth = next
		if err := c.cs.SetWorth(ctx, item); err != nil {
			return fmt.Errorf("resequence: set worth: %w", err)
		}
	}
	return nil
}
