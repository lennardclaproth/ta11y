package cashflow

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// Rhythm is how often a recurring item is expected. Only these three exist: an
// irregular series is not given an expectation at all rather than an invented one.
type Rhythm string

const (
	RhythmMonthly   Rhythm = "monthly"
	RhythmQuarterly Rhythm = "quarterly"
	RhythmYearly    Rhythm = "yearly"
)

// Months is the number of calendar months between two occurrences.
func (r Rhythm) Months() int {
	switch r {
	case RhythmQuarterly:
		return 3
	case RhythmYearly:
		return 12
	default:
		return 1
	}
}

// ParseRhythm reads a rhythm from client input.
func ParseRhythm(raw string) (Rhythm, error) {
	switch Rhythm(strings.ToLower(strings.TrimSpace(raw))) {
	case RhythmMonthly:
		return RhythmMonthly, nil
	case RhythmQuarterly:
		return RhythmQuarterly, nil
	case RhythmYearly:
		return RhythmYearly, nil
	default:
		return "", ErrRecurringInvalidRhythm
	}
}

// RecurringItem is a subscription, fixed cost or recurring income: the series of
// transactions for one counterparty, under a name the user chose or confirmed.
// It is deliberately unrelated to the tag and the ignored flag on a transaction.
type RecurringItem struct {
	ID        uuid.UUID         `db:"id"`
	AccountID uuid.UUID         `db:"account_id"`
	Name      string            `db:"name"`
	Direction CashFlowDirection `db:"direction"`
	Rhythm    Rhythm            `db:"rhythm"`
	// MatchKey is the normalised statement description the item was started from;
	// it is how a transaction from a later import is recognised as belonging here.
	MatchKey string `db:"match_key"`
	// EndedFrom is the first day of the month the item stops being expected.
	EndedFrom *time.Time `db:"ended_from"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

// IsEnded reports whether the item has been ended from a month.
func (i *RecurringItem) IsEnded() bool {
	return i != nil && i.EndedFrom != nil
}

// NewRecurringItem builds a validated recurring item.
func NewRecurringItem(accountID uuid.UUID, name string, direction CashFlowDirection, rhythm Rhythm, matchKey string) (*RecurringItem, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrRecurringNameRequired
	}
	if direction != CashIn && direction != CashOut {
		return nil, ErrUnsupportedDirection
	}
	now := time.Now().UTC()
	return &RecurringItem{
		ID:        uuid.New(),
		AccountID: accountID,
		Name:      trimmed,
		Direction: direction,
		Rhythm:    rhythm,
		MatchKey:  matchKey,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NextExpectedAfter returns the day the next transaction is expected after the
// most recent one, or nil once the item is ended or has nothing to count from.
func (i *RecurringItem) NextExpectedAfter(lastSeen *time.Time) *time.Time {
	if i == nil || lastSeen == nil || i.IsEnded() {
		return nil
	}
	next := lastSeen.UTC().AddDate(0, i.Rhythm.Months(), 0)
	return &next
}

// RecurringLink is one cashflow transaction linked to a recurring item, carrying
// the transaction fields the item's own screens read.
type RecurringLink struct {
	ItemID        uuid.UUID   `db:"item_id"`
	TransactionID uuid.UUID   `db:"transaction_id"`
	Date          time.Time   `db:"date"`
	AmountCents   money.Price `db:"amount_cents"`
	Description   string      `db:"description"`
	Source        string      `db:"source"`
}

// DismissedSuggestion records a suggestion the user refused, so it is not offered again.
type DismissedSuggestion struct {
	MatchKey  string            `db:"match_key"`
	Direction CashFlowDirection `db:"direction"`
}

// recurringMatchKeyTokens is how many words of a description make up its
// fingerprint. Bank descriptions append references, dates and mandate numbers
// that differ every month; the leading words are the part that repeats.
const recurringMatchKeyTokens = 4

// recurringAmountTolerance is how far a new transaction's amount may sit from the
// item's most recent amount and still be linked automatically. Energy and other
// variable bills move between charges, so an exact match would link almost
// nothing; a quarter is wide enough for those and narrow enough that an unrelated
// payment to the same counterparty is left in Cashflow for the user to decide on.
const recurringAmountTolerance = 0.25

// recurringDescriptionNoise are the labels Dutch bank statements wrap the actual
// counterparty in. They repeat on every row, so leaving them in would make every
// fingerprint from one bank start with the same words.
var recurringDescriptionNoise = map[string]struct{}{
	"naam": {}, "name": {}, "omschrijving": {}, "description": {},
	"iban": {}, "bic": {}, "ibanbic": {}, "kenmerk": {}, "machtiging": {},
	"incasso": {}, "sepa": {}, "id": {}, "ref": {}, "referentie": {},
	"doorlopende": {}, "eenmalige": {}, "overboeking": {}, "betaalautomaat": {},
	"valutadatum": {}, "transactie": {}, "pasvolgnr": {}, "van": {}, "naar": {},
}

// MatchKeyFor reduces a statement description to the fingerprint that recognises
// the same counterparty across statements: lowercased words, with digits, symbols
// and statement boilerplate dropped. It is a heuristic for grouping and linking —
// the name a recurring item carries is always the user's, never this.
func MatchKeyFor(description string) string {
	fields := strings.FieldsFunc(strings.ToLower(description), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	tokens := make([]string, 0, recurringMatchKeyTokens)
	for _, field := range fields {
		if len(tokens) == recurringMatchKeyTokens {
			break
		}
		if len(field) < 2 || strings.IndexFunc(field, unicode.IsLetter) == -1 {
			continue
		}
		if _, noise := recurringDescriptionNoise[field]; noise {
			continue
		}
		tokens = append(tokens, field)
	}
	if len(tokens) == 0 {
		return strings.ToLower(strings.TrimSpace(description))
	}
	return strings.Join(tokens, " ")
}

// SuggestedNameFor is the name ta11y offers for a fingerprint: its words, capitalised.
// It is only ever a suggestion — a suggestion counts for nothing until confirmed.
func SuggestedNameFor(matchKey string) string {
	words := strings.Fields(matchKey)
	for i, word := range words {
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// MonthlyEquivalentCents spreads an amount over a month: a quarterly charge counts
// as a third, a yearly one as a twelfth. The screens that show it say so.
func MonthlyEquivalentCents(amount money.Price, rhythm Rhythm) int64 {
	return int64(amount) / int64(rhythm.Months())
}

// ParseEndMonth reads a "YYYY-MM" month and returns its first day in UTC.
func ParseEndMonth(raw string) (time.Time, error) {
	parsed, err := time.Parse("2006-01", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q", ErrRecurringInvalidMonth, raw)
	}
	return parsed.UTC(), nil
}

// amountWithinTolerance reports whether candidate sits close enough to reference
// to be linked without asking.
func amountWithinTolerance(candidate, reference money.Price) bool {
	if reference == 0 {
		return candidate == 0
	}
	diff := float64(candidate) - float64(reference)
	if diff < 0 {
		diff = -diff
	}
	return diff/float64(reference) <= recurringAmountTolerance
}
