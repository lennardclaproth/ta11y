package cashflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// IgnoreRuleCommandStore persists ignore rules and applies them to transactions.
type IgnoreRuleCommandStore interface {
	CreateIgnoreRule(ctx context.Context, rule *IgnoreRule) error
	UpdateIgnoreRule(ctx context.Context, rule *IgnoreRule) (int, error)
	DeleteIgnoreRule(ctx context.Context, accountID, id uuid.UUID) (int, error)
	// ApplyIgnoreRule ignores the transactions the rule matches that are not ignored
	// yet and whose ignored state was not decided by hand, and adds the number it
	// ignored to the rule's own total. A non-nil importID narrows it to one import.
	ApplyIgnoreRule(ctx context.Context, rule *IgnoreRule, importID *uuid.UUID) (int, error)
}

// IgnoreRuleQueryStore reads ignore rules.
type IgnoreRuleQueryStore interface {
	ListIgnoreRules(ctx context.Context, accountID uuid.UUID) ([]*IgnoreRule, error)
	GetIgnoreRule(ctx context.Context, accountID, id uuid.UUID) (*IgnoreRule, error)
	// CountIgnoreRuleTargets counts what ApplyIgnoreRule would actually ignore, which
	// is the number the confirmation before "apply to existing" has to name.
	CountIgnoreRuleTargets(ctx context.Context, filters TransactionFilters) (int, error)
	// ListIgnoredByRule returns, per rule, the transactions that rule ignored in one
	// import, newest first, capped at perRule rows alongside the full group total.
	ListIgnoredByRule(ctx context.Context, accountID, importID uuid.UUID, perRule int) ([]IgnoredRuleGroup, error)
}

// CreateIgnoreRule stores a new ignore rule. It takes effect on what is imported from
// then on; applying it to transactions already in the ledger is a separate, confirmed
// action.
func (c *Commands) CreateIgnoreRule(ctx context.Context, accountID uuid.UUID, draft IgnoreRuleDraft) (*IgnoreRule, error) {
	rule := NewIgnoreRule(accountID, draft)
	if err := c.irs.CreateIgnoreRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("create ignore rule: %w", err)
	}
	return rule, nil
}

// UpdateIgnoreRule rewrites what an existing rule matches on. Transactions it already
// ignored keep their state — the edit says what the rule catches from now on — and are
// still restorable one by one.
func (c *Commands) UpdateIgnoreRule(ctx context.Context, accountID, id uuid.UUID, draft IgnoreRuleDraft) (*IgnoreRule, error) {
	rule, err := c.irqs.GetIgnoreRule(ctx, accountID, id)
	if err != nil {
		return nil, fmt.Errorf("update ignore rule: %w", err)
	}
	if rule == nil {
		return nil, ErrIgnoreRuleNotFound
	}

	rule.Apply(draft)
	updated, err := c.irs.UpdateIgnoreRule(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("update ignore rule: %w", err)
	}
	if updated == 0 {
		return nil, ErrIgnoreRuleNotFound
	}
	return rule, nil
}

// DeleteIgnoreRule removes a rule. Nothing it ignored is restored: the transactions
// stay as they are and are put back one by one, the same as any other ignored row.
func (c *Commands) DeleteIgnoreRule(ctx context.Context, accountID, id uuid.UUID) error {
	deleted, err := c.irs.DeleteIgnoreRule(ctx, accountID, id)
	if err != nil {
		return fmt.Errorf("delete ignore rule: %w", err)
	}
	if deleted == 0 {
		return ErrIgnoreRuleNotFound
	}
	return nil
}

// ApplyIgnoreRuleToExisting reaches back over the ledger and ignores what one rule
// matches. It is the only path that changes transactions that are already there, so it
// is never implied by saving a rule.
func (c *Commands) ApplyIgnoreRuleToExisting(ctx context.Context, accountID, id uuid.UUID) (int, error) {
	rule, err := c.irqs.GetIgnoreRule(ctx, accountID, id)
	if err != nil {
		return 0, fmt.Errorf("apply ignore rule: %w", err)
	}
	if rule == nil {
		return 0, ErrIgnoreRuleNotFound
	}

	ignored, err := c.irs.ApplyIgnoreRule(ctx, rule, nil)
	if err != nil {
		return 0, fmt.Errorf("apply ignore rule: %w", err)
	}
	return ignored, nil
}

// ApplyIgnoreRulesToImport runs every enabled rule of the account over the rows one
// import brought in, and reports how many were ignored. Duplicates keep the import
// they first arrived with, so a re-imported statement offers a rule nothing to catch.
func (c *Commands) ApplyIgnoreRulesToImport(ctx context.Context, accountID, importID uuid.UUID) (int, error) {
	rules, err := c.irqs.ListIgnoreRules(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("apply ignore rules to import: %w", err)
	}

	ignored := 0
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		count, err := c.irs.ApplyIgnoreRule(ctx, rule, &importID)
		if err != nil {
			return ignored, fmt.Errorf("apply ignore rules to import: rule %s: %w", rule.ID, err)
		}
		ignored += count
	}
	return ignored, nil
}
