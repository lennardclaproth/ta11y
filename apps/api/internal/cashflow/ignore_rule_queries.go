package cashflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/sorting"
)

// ignoreRulePreviewSampleSize is how many matching transactions the preview shows. It
// is a sample, not a page: the count above it is what says how wide the rule is.
const ignoreRulePreviewSampleSize = 4

// IgnoreRulePreview reports what a rule would catch, before it is allowed to catch
// anything. A rule that is too wide has to be visible here rather than as a surprise
// in next month's totals.
type IgnoreRulePreview struct {
	// Matching is every transaction in the account the rule matches.
	Matching int
	// NotYetIgnored is what applying the rule to the existing ledger would ignore:
	// matching, not ignored yet, and not decided by hand.
	NotYetIgnored int
	// Scanned is the size of the ledger the rule was held against.
	Scanned int
	// Sample is the most recent matching transactions.
	Sample []*Transaction
}

// IgnoredRuleGroup is the harvest of one ignore rule within one import: the rule, how
// many rows it ignored there, and the first of them. Grouping by rule is what lets a
// rule be judged on what it caught instead of row by row.
type IgnoredRuleGroup struct {
	Rule         *IgnoreRule
	Total        int
	Transactions []*Transaction
}

// IgnoreRules returns the account's ignore rules, newest first.
func (q *Queries) IgnoreRules(ctx context.Context, accountID uuid.UUID) ([]*IgnoreRule, error) {
	rules, err := q.irqs.ListIgnoreRules(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("cashflow ignore rules: %w", err)
	}
	return rules, nil
}

// IgnoreRule returns one of the account's ignore rules.
func (q *Queries) IgnoreRule(ctx context.Context, accountID, id uuid.UUID) (*IgnoreRule, error) {
	rule, err := q.irqs.GetIgnoreRule(ctx, accountID, id)
	if err != nil {
		return nil, fmt.Errorf("cashflow ignore rule: %w", err)
	}
	if rule == nil {
		return nil, ErrIgnoreRuleNotFound
	}
	return rule, nil
}

// PreviewIgnoreRule reports what a draft rule matches in the account's ledger today.
func (q *Queries) PreviewIgnoreRule(ctx context.Context, accountID uuid.UUID, draft IgnoreRuleDraft) (*IgnoreRulePreview, error) {
	filters := draft.Filters(accountID)

	matching, err := q.qs.CountByFilter(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("preview ignore rule: count matching: %w", err)
	}
	notYetIgnored, err := q.irqs.CountIgnoreRuleTargets(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("preview ignore rule: count targets: %w", err)
	}
	scanned, err := q.qs.CountByFilter(ctx, TransactionFilters{AccountID: accountID})
	if err != nil {
		return nil, fmt.Errorf("preview ignore rule: count scanned: %w", err)
	}

	preview := &IgnoreRulePreview{
		Matching:      matching,
		NotYetIgnored: notYetIgnored,
		Scanned:       scanned,
		Sample:        []*Transaction{},
	}
	if matching == 0 {
		return preview, nil
	}

	sample, err := q.qs.ListTransactions(ctx, TransactionListQuery{
		AccountID:   accountID,
		Limit:       ignoreRulePreviewSampleSize,
		Sort:        sorting.Sort{Field: TransactionSortFieldDate, Direction: sorting.DESC},
		Description: filters.Description,
		Note:        filters.Note,
		SourceExact: filters.SourceExact,
		Direction:   filters.Direction,
	})
	if err != nil {
		return nil, fmt.Errorf("preview ignore rule: sample: %w", err)
	}
	if sample != nil && sample.Transactions != nil {
		preview.Sample = sample.Transactions
	}
	return preview, nil
}

// ignoredByRuleGroupSize is how many rows each group on the import review carries. The
// rest of a group is read through the ledger, filtered on that import and that rule.
const ignoredByRuleGroupSize = 5

// IgnoredByRule returns what each rule ignored in one import, grouped by rule.
func (q *Queries) IgnoredByRule(ctx context.Context, accountID, importID uuid.UUID) ([]IgnoredRuleGroup, error) {
	groups, err := q.irqs.ListIgnoredByRule(ctx, accountID, importID, ignoredByRuleGroupSize)
	if err != nil {
		return nil, fmt.Errorf("cashflow ignored by rule: %w", err)
	}
	return groups, nil
}
