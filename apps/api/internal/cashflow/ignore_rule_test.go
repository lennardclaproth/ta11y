package cashflow

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// ruleStore is an IgnoreRuleCommandStore/IgnoreRuleQueryStore pair holding rules in
// memory, enough to drive the rule use cases without a database.
type ruleStore struct {
	rules   []*IgnoreRule
	applied []appliedRule
	perRule int
	// failOn is the rule whose application fails, standing for a pass that breaks down
	// after earlier rules have already committed what they ignored.
	failOn *uuid.UUID
}

type appliedRule struct {
	ruleID   uuid.UUID
	importID *uuid.UUID
}

func (s *ruleStore) CreateIgnoreRule(_ context.Context, rule *IgnoreRule) error {
	s.rules = append(s.rules, rule)
	return nil
}

func (s *ruleStore) UpdateIgnoreRule(_ context.Context, rule *IgnoreRule) (int, error) {
	for i, existing := range s.rules {
		if existing.ID == rule.ID {
			s.rules[i] = rule
			return 1, nil
		}
	}
	return 0, nil
}

func (s *ruleStore) DeleteIgnoreRule(_ context.Context, _, id uuid.UUID) (int, error) {
	for i, rule := range s.rules {
		if rule.ID == id {
			s.rules = append(s.rules[:i], s.rules[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}

func (s *ruleStore) ApplyIgnoreRule(_ context.Context, rule *IgnoreRule, importID *uuid.UUID) (int, error) {
	if s.failOn != nil && *s.failOn == rule.ID {
		return 0, errors.New("store is unreachable")
	}
	s.applied = append(s.applied, appliedRule{ruleID: rule.ID, importID: importID})
	return s.perRule, nil
}

func (s *ruleStore) ListIgnoreRules(_ context.Context, _ uuid.UUID) ([]*IgnoreRule, error) {
	return s.rules, nil
}

func (s *ruleStore) GetIgnoreRule(_ context.Context, _, id uuid.UUID) (*IgnoreRule, error) {
	for _, rule := range s.rules {
		if rule.ID == id {
			return rule, nil
		}
	}
	return nil, nil
}

func (s *ruleStore) CountIgnoreRuleTargets(_ context.Context, _ TransactionFilters) (int, error) {
	return s.perRule, nil
}

func (s *ruleStore) ListIgnoredByRule(_ context.Context, _, _ uuid.UUID, _ int) ([]IgnoredRuleGroup, error) {
	return nil, nil
}

func draft(t *testing.T, contains string) IgnoreRuleDraft {
	t.Helper()
	d, err := NewIgnoreRuleDraft("Credit card payment", "description", contains, "out", "ING", true)
	if err != nil {
		t.Fatalf("new draft: %v", err)
	}
	return d
}

func TestNewIgnoreRuleDraftRefusesTooWideText(t *testing.T) {
	// A one- or two-character rule matches nearly every statement, and the damage is
	// silent: the rows leave the monthly totals without anybody reading them.
	if _, err := NewIgnoreRuleDraft("Wide", "description", "to", "", "", true); !errors.Is(err, ErrIgnoreRuleContainsTooShort) {
		t.Fatalf("expected ErrIgnoreRuleContainsTooShort, got %v", err)
	}
}

func TestNewIgnoreRuleDraftRefusesUnknownMatchField(t *testing.T) {
	if _, err := NewIgnoreRuleDraft("Tagged", "tag", "savings", "", "", true); !errors.Is(err, ErrIgnoreRuleInvalidMatchField) {
		t.Fatalf("expected ErrIgnoreRuleInvalidMatchField, got %v", err)
	}
}

func TestIgnoreRuleFiltersMatchTheChosenField(t *testing.T) {
	accID := uuid.New()

	byDescription := draft(t, "Credit card").Filters(accID)
	if byDescription.Description != "Credit card" || byDescription.Note != "" {
		t.Fatalf("expected the text on description only, got %+v", byDescription)
	}
	if byDescription.Direction == nil || *byDescription.Direction != CashOut {
		t.Fatalf("expected an outgoing filter, got %+v", byDescription.Direction)
	}
	if byDescription.AccountID != accID {
		t.Fatalf("expected the account scope to be carried, got %s", byDescription.AccountID)
	}
	// The bank is chosen from a list of names, so it narrows on the whole name. Reusing
	// the ledger's free-text source filter would let a rule scoped to one bank reach a
	// second one whose name contains the first.
	if byDescription.SourceExact != "ING" || byDescription.Source != "" {
		t.Fatalf("expected the bank to narrow exactly, got source=%q exact=%q", byDescription.Source, byDescription.SourceExact)
	}

	note, err := NewIgnoreRuleDraft("Own transfer", "note", "Own account", "", "", true)
	if err != nil {
		t.Fatalf("new draft: %v", err)
	}
	byNote := note.Filters(accID)
	if byNote.Note != "Own account" || byNote.Description != "" {
		t.Fatalf("expected the text on note only, got %+v", byNote)
	}
	if byNote.Direction != nil {
		t.Fatalf("expected no direction narrowing, got %+v", byNote.Direction)
	}
}

func TestApplyIgnoreRulesToImportSkipsDisabledRules(t *testing.T) {
	accID := uuid.New()
	importID := uuid.New()

	enabled := NewIgnoreRule(accID, draft(t, "Credit card"))
	disabled := NewIgnoreRule(accID, draft(t, "Round-up"))
	disabled.Enabled = false

	store := &ruleStore{rules: []*IgnoreRule{enabled, disabled}, perRule: 3}
	commands := NewCommands(nil, nil, store, store, nil)

	ignored, err := commands.ApplyIgnoreRulesToImport(context.Background(), accID, importID)
	if err != nil {
		t.Fatalf("apply ignore rules to import: %v", err)
	}
	if ignored != 3 {
		t.Fatalf("expected only the enabled rule to ignore rows, got %d", ignored)
	}
	if len(store.applied) != 1 || store.applied[0].ruleID != enabled.ID {
		t.Fatalf("expected one application by the enabled rule, got %+v", store.applied)
	}
	// Rules only ever see what this import brought in; the ledger as a whole is out
	// of scope until the person asks for it.
	if store.applied[0].importID == nil || *store.applied[0].importID != importID {
		t.Fatalf("expected the application to be scoped to the import, got %+v", store.applied[0].importID)
	}
}

func TestApplyIgnoreRulesToImportReportsWhatEarlierRulesIgnored(t *testing.T) {
	// Each rule commits on its own, so a pass that breaks down halfway leaves the rows
	// of the rules before it ignored. That count has to travel with the error: an import
	// reporting nothing was ignored while rows did leave the monthly totals is the one
	// thing the counter exists to prevent.
	accID := uuid.New()
	importID := uuid.New()

	first := NewIgnoreRule(accID, draft(t, "Credit card"))
	broken := NewIgnoreRule(accID, draft(t, "Round-up"))
	last := NewIgnoreRule(accID, draft(t, "Transfer to savings"))

	store := &ruleStore{rules: []*IgnoreRule{first, broken, last}, perRule: 4, failOn: &broken.ID}
	commands := NewCommands(nil, nil, store, store, nil)

	ignored, err := commands.ApplyIgnoreRulesToImport(context.Background(), accID, importID)
	if err == nil {
		t.Fatal("expected the failing rule to surface as an error")
	}
	if ignored != 4 {
		t.Fatalf("ignored = %d, want the 4 rows the rule before the failure ignored", ignored)
	}
	// The pass stops at the failure: the rules after it never ran, which is what makes
	// applying them to the ledger afterwards the way back.
	if len(store.applied) != 1 || store.applied[0].ruleID != first.ID {
		t.Fatalf("expected only the rule before the failure to have been applied, got %+v", store.applied)
	}
}

func TestApplyIgnoreRuleToExistingReachesTheWholeLedger(t *testing.T) {
	accID := uuid.New()
	rule := NewIgnoreRule(accID, draft(t, "Credit card"))
	store := &ruleStore{rules: []*IgnoreRule{rule}, perRule: 23}
	commands := NewCommands(nil, nil, store, store, nil)

	ignored, err := commands.ApplyIgnoreRuleToExisting(context.Background(), accID, rule.ID)
	if err != nil {
		t.Fatalf("apply ignore rule to existing: %v", err)
	}
	if ignored != 23 {
		t.Fatalf("expected 23 ignored, got %d", ignored)
	}
	if len(store.applied) != 1 || store.applied[0].importID != nil {
		t.Fatalf("expected one unscoped application, got %+v", store.applied)
	}
}

func TestApplyIgnoreRuleToExistingRefusesUnknownRule(t *testing.T) {
	store := &ruleStore{}
	commands := NewCommands(nil, nil, store, store, nil)

	_, err := commands.ApplyIgnoreRuleToExisting(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrIgnoreRuleNotFound) {
		t.Fatalf("expected ErrIgnoreRuleNotFound, got %v", err)
	}
}

func TestUpdateIgnoreRuleKeepsWhatItAlreadyIgnored(t *testing.T) {
	accID := uuid.New()
	rule := NewIgnoreRule(accID, draft(t, "Credit card"))
	rule.IgnoredTotal = 19
	store := &ruleStore{rules: []*IgnoreRule{rule}}
	commands := NewCommands(nil, nil, store, store, nil)

	updated, err := commands.UpdateIgnoreRule(context.Background(), accID, rule.ID, draft(t, "Card settlement"))
	if err != nil {
		t.Fatalf("update ignore rule: %v", err)
	}
	if updated.Contains != "Card settlement" {
		t.Fatalf("expected the rewritten text, got %q", updated.Contains)
	}
	// An edit says what the rule catches from now on; its harvest so far stands, and
	// nothing it ignored is reconsidered.
	if updated.IgnoredTotal != 19 {
		t.Fatalf("expected the total to survive the edit, got %d", updated.IgnoredTotal)
	}
	if len(store.applied) != 0 {
		t.Fatalf("expected an edit not to apply the rule, got %+v", store.applied)
	}
}
