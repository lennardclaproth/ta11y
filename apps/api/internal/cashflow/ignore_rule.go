package cashflow

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// MatchField is the transaction field an ignore rule reads.
type MatchField string

const (
	// MatchFieldDescription matches text in the transaction description.
	MatchFieldDescription MatchField = "description"
	// MatchFieldNote matches text in the transaction note.
	MatchFieldNote MatchField = "note"
)

// ParseMatchField validates the field an ignore rule matches on.
func ParseMatchField(raw string) (MatchField, error) {
	switch MatchField(strings.ToLower(strings.TrimSpace(raw))) {
	case MatchFieldDescription:
		return MatchFieldDescription, nil
	case MatchFieldNote:
		return MatchFieldNote, nil
	default:
		return "", ErrIgnoreRuleInvalidMatchField
	}
}

// ignoreRuleNameMaxLength and ignoreRuleContainsMaxLength mirror the column widths.
const (
	ignoreRuleNameMaxLength     = 120
	ignoreRuleContainsMaxLength = 255
	// ignoreRuleContainsMinLength keeps a rule from being so wide that it would ignore
	// most of a statement. A single character matches almost everything, and the damage
	// is silent: the rows leave the monthly totals without anybody reading them.
	ignoreRuleContainsMinLength = 3
)

// IgnoreRule recognises transactions that should arrive ignored. It matches text in
// one field, optionally narrowed to a direction and to the bank the transaction was
// imported from, and never spans accounts.
type IgnoreRule struct {
	ID         uuid.UUID          `db:"id"`
	AccountID  uuid.UUID          `db:"account_id"`
	Name       string             `db:"name"`
	MatchField MatchField         `db:"match_field"`
	Contains   string             `db:"contains"`
	Direction  *CashFlowDirection `db:"direction"`
	// Source is the bank the rule is limited to, empty for every bank.
	Source  string `db:"source"`
	Enabled bool   `db:"enabled"`
	// IgnoredTotal is how many transactions this rule has ignored since it was made.
	IgnoredTotal  int        `db:"ignored_total"`
	LastAppliedAt *time.Time `db:"last_applied_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// IgnoreRuleDraft is validated ignore-rule input, shared by create, update, and the
// preview that shows what a rule would catch before it is saved.
type IgnoreRuleDraft struct {
	Name       string
	MatchField MatchField
	Contains   string
	Direction  *CashFlowDirection
	Source     string
	Enabled    bool
}

// NewIgnoreRuleDraft validates raw ignore-rule input.
func NewIgnoreRuleDraft(name, matchFieldRaw, contains, directionRaw, source string, enabled bool) (IgnoreRuleDraft, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return IgnoreRuleDraft{}, ErrIgnoreRuleNameRequired
	}
	if len(trimmedName) > ignoreRuleNameMaxLength {
		return IgnoreRuleDraft{}, ErrIgnoreRuleNameTooLong
	}

	matchField, err := ParseMatchField(matchFieldRaw)
	if err != nil {
		return IgnoreRuleDraft{}, err
	}

	trimmedContains := strings.TrimSpace(contains)
	// The text goes into a LIKE unescaped, so % and _ would keep their wildcard meaning
	// and walk straight past the minimum length: "a%z" is three characters and matches
	// almost every statement. A backslash escapes the next character on Postgres and
	// means nothing on SQLite, so a rule carrying one would catch different rows in
	// each dialect. "Contains" should mean contains, and a refusal says so where
	// quietly stripping the characters would not.
	if strings.ContainsAny(trimmedContains, `%_\`) {
		return IgnoreRuleDraft{}, ErrIgnoreRuleContainsWildcard
	}
	if len([]rune(trimmedContains)) < ignoreRuleContainsMinLength {
		return IgnoreRuleDraft{}, ErrIgnoreRuleContainsTooShort
	}
	if len(trimmedContains) > ignoreRuleContainsMaxLength {
		return IgnoreRuleDraft{}, ErrIgnoreRuleContainsTooLong
	}

	direction, err := ParseDirection(directionRaw)
	if err != nil {
		return IgnoreRuleDraft{}, ErrIgnoreRuleInvalidDirection
	}

	return IgnoreRuleDraft{
		Name:       trimmedName,
		MatchField: matchField,
		Contains:   trimmedContains,
		Direction:  direction,
		Source:     strings.TrimSpace(source),
		Enabled:    enabled,
	}, nil
}

// NewIgnoreRule creates a rule for an account from a validated draft.
func NewIgnoreRule(accountID uuid.UUID, draft IgnoreRuleDraft) *IgnoreRule {
	now := time.Now().UTC()
	rule := &IgnoreRule{
		ID:        uuid.New(),
		AccountID: accountID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	rule.Apply(draft)
	return rule
}

// Apply overwrites the rule's matching fields with the draft's. Counters and
// timestamps are left alone: editing a rule says what it catches from now on, not
// what it has caught.
func (r *IgnoreRule) Apply(draft IgnoreRuleDraft) {
	r.Name = draft.Name
	r.MatchField = draft.MatchField
	r.Contains = draft.Contains
	r.Direction = draft.Direction
	r.Source = draft.Source
	r.Enabled = draft.Enabled
	r.UpdatedAt = time.Now().UTC()
}

// Filters returns the transaction filters that select what a draft matches, scoped to
// one account. The rule reuses the ledger's own filters, so what it catches is
// previewed with the query the ledger itself runs. The bank is the exception: it is
// chosen from a list of names rather than typed, so it matches the whole name.
func (d IgnoreRuleDraft) Filters(accountID uuid.UUID) TransactionFilters {
	filters := TransactionFilters{
		AccountID:   accountID,
		Direction:   d.Direction,
		SourceExact: d.Source,
	}
	switch d.MatchField {
	case MatchFieldNote:
		filters.Note = d.Contains
	default:
		filters.Description = d.Contains
	}
	return filters
}

// Draft returns the rule's matching fields as a draft.
func (r *IgnoreRule) Draft() IgnoreRuleDraft {
	return IgnoreRuleDraft{
		Name:       r.Name,
		MatchField: r.MatchField,
		Contains:   r.Contains,
		Direction:  r.Direction,
		Source:     r.Source,
		Enabled:    r.Enabled,
	}
}

// Filters returns the transaction filters that select what the rule matches.
func (r *IgnoreRule) Filters() TransactionFilters {
	return r.Draft().Filters(r.AccountID)
}
