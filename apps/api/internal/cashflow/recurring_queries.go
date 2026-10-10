package cashflow

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// RecurringQueries exposes the read side of recurring items.
type RecurringQueries struct {
	qs RecurringQueryStore
}

// RecurringQueryStore reads recurring items, their links and the transactions
// suggestions are derived from.
type RecurringQueryStore interface {
	ListRecurringItems(ctx context.Context, accountID uuid.UUID) ([]*RecurringItem, error)
	GetRecurringItem(ctx context.Context, accountID, id uuid.UUID) (*RecurringItem, error)
	ListRecurringLinks(ctx context.Context, accountID uuid.UUID) ([]RecurringLink, error)
	ListRecurringLinksForItem(ctx context.Context, accountID, itemID uuid.UUID) ([]RecurringLink, error)
	ListUnlinkedTransactions(ctx context.Context, accountID uuid.UUID) ([]*Transaction, error)
	ListDismissedSuggestions(ctx context.Context, accountID uuid.UUID) ([]DismissedSuggestion, error)
	ListTransactionsByIDs(ctx context.Context, accountID uuid.UUID, ids []uuid.UUID) ([]*Transaction, error)
	ListImportedTransactions(ctx context.Context, accountID, importID uuid.UUID) ([]*Transaction, error)
}

// NewRecurringQueries creates the recurring read-side use cases.
func NewRecurringQueries(qs RecurringQueryStore) *RecurringQueries {
	return &RecurringQueries{qs: qs}
}

// RecurringAmountPoint is one observed amount and the day it was charged or received.
type RecurringAmountPoint struct {
	Date        time.Time
	AmountCents money.Price
}

// RecurringItemView is a recurring item with everything the overview reads off the
// transactions linked to it. Nothing here is stored: the amounts are the amounts.
type RecurringItemView struct {
	Item *RecurringItem
	// History holds the observed amounts oldest to newest. ta11y shows them as they
	// are and calls none of them expensive.
	History         []RecurringAmountPoint
	LastAmountCents money.Price
	LastSeen        *time.Time
	NextExpected    *time.Time
	LinkedCount     int
}

// RecurringMonthPoint is the monthly-equivalent total of the running items in one month.
type RecurringMonthPoint struct {
	Month        time.Time
	ExpenseCents int64
	IncomeCents  int64
}

// RecurringOverview is the whole recurring page in one read: the three groups, the
// two monthly totals and the series behind the chart.
type RecurringOverview struct {
	Expenses            []RecurringItemView
	Income              []RecurringItemView
	Ended               []RecurringItemView
	MonthlyExpenseCents int64
	MonthlyIncomeCents  int64
	Series              []RecurringMonthPoint
}

// RecurringSuggestion is a pattern found in transactions the user has not ignored
// and has not linked. It counts for nothing until it is confirmed.
type RecurringSuggestion struct {
	MatchKey    string
	Name        string
	Direction   CashFlowDirection
	Rhythm      Rhythm
	AmountCents money.Price
	Matches     int
	Since       time.Time
	// Sample is the statement text the suggested name was read off, so the guess
	// can be checked rather than taken on trust.
	Sample string
}

// recurringSeriesMonths is how far back the chart on the overview looks.
const recurringSeriesMonths = 12

// Overview returns the recurring items grouped into running expenses, running
// income and ended items, with the monthly totals and the trend behind them.
// A non-empty search narrows the three groups by name.
func (q *RecurringQueries) Overview(ctx context.Context, accountID uuid.UUID, search string) (*RecurringOverview, error) {
	views, err := q.views(ctx, accountID)
	if err != nil {
		return nil, err
	}

	overview := &RecurringOverview{
		Expenses: []RecurringItemView{},
		Income:   []RecurringItemView{},
		Ended:    []RecurringItemView{},
	}

	needle := strings.ToLower(strings.TrimSpace(search))
	running := make([]RecurringItemView, 0, len(views))
	for _, view := range views {
		// The totals and the trend describe the account, not the search: filtering
		// them too would make a search for one item claim it is all you pay.
		if !view.Item.IsEnded() {
			running = append(running, view)
		}
		if needle != "" && !strings.Contains(strings.ToLower(view.Item.Name), needle) {
			continue
		}
		switch {
		case view.Item.IsEnded():
			overview.Ended = append(overview.Ended, view)
		case view.Item.Direction == CashIn:
			overview.Income = append(overview.Income, view)
		default:
			overview.Expenses = append(overview.Expenses, view)
		}
	}

	sortByNextExpected(overview.Expenses)
	sortByNextExpected(overview.Income)
	sortByEndedFrom(overview.Ended)

	for _, view := range running {
		monthly := MonthlyEquivalentCents(view.LastAmountCents, view.Item.Rhythm)
		if view.Item.Direction == CashIn {
			overview.MonthlyIncomeCents += monthly
		} else {
			overview.MonthlyExpenseCents += monthly
		}
	}
	overview.Series = monthlySeries(running, time.Now().UTC(), recurringSeriesMonths)

	return overview, nil
}

// RecurringItemDetail is one item with the transactions behind it, newest first.
type RecurringItemDetail struct {
	RecurringItemView
	Transactions []RecurringLink
}

// Item returns one recurring item with its amount history and linked transactions.
func (q *RecurringQueries) Item(ctx context.Context, accountID, id uuid.UUID) (*RecurringItemDetail, error) {
	item, err := q.qs.GetRecurringItem(ctx, accountID, id)
	if err != nil {
		return nil, fmt.Errorf("recurring item: %w", err)
	}
	if item == nil {
		return nil, fmt.Errorf("recurring item: %w", ErrRecurringItemNotFound)
	}

	own, err := q.qs.ListRecurringLinksForItem(ctx, accountID, item.ID)
	if err != nil {
		return nil, fmt.Errorf("recurring item links: %w", err)
	}
	sort.Slice(own, func(i, j int) bool { return own[i].Date.After(own[j].Date) })

	return &RecurringItemDetail{
		RecurringItemView: buildItemView(item, own),
		Transactions:      own,
	}, nil
}

// Suggestions returns the patterns ta11y found in transactions that are neither
// ignored nor already linked, leaving out every pattern that was dismissed or
// already belongs to an item.
func (q *RecurringQueries) Suggestions(ctx context.Context, accountID uuid.UUID) ([]RecurringSuggestion, error) {
	transactions, err := q.qs.ListUnlinkedTransactions(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("recurring suggestions: %w", err)
	}
	items, err := q.qs.ListRecurringItems(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("recurring suggestions: known items: %w", err)
	}
	dismissed, err := q.qs.ListDismissedSuggestions(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("recurring suggestions: dismissed: %w", err)
	}

	known := make(map[string]struct{}, len(items)+len(dismissed))
	for _, item := range items {
		known[suggestionKey(item.MatchKey, item.Direction)] = struct{}{}
	}
	for _, entry := range dismissed {
		known[suggestionKey(entry.MatchKey, entry.Direction)] = struct{}{}
	}

	return suggestFromTransactions(transactions, known), nil
}

// views assembles every item with the transactions linked to it.
func (q *RecurringQueries) views(ctx context.Context, accountID uuid.UUID) ([]RecurringItemView, error) {
	items, err := q.qs.ListRecurringItems(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("recurring items: %w", err)
	}
	links, err := q.qs.ListRecurringLinks(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("recurring item links: %w", err)
	}

	byItem := make(map[uuid.UUID][]RecurringLink, len(items))
	for _, link := range links {
		byItem[link.ItemID] = append(byItem[link.ItemID], link)
	}

	views := make([]RecurringItemView, 0, len(items))
	for _, item := range items {
		views = append(views, buildItemView(item, byItem[item.ID]))
	}
	return views, nil
}

// buildItemView reads an item's amounts, last sighting and expectation off its links.
func buildItemView(item *RecurringItem, links []RecurringLink) RecurringItemView {
	ordered := make([]RecurringLink, len(links))
	copy(ordered, links)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Date.Before(ordered[j].Date) })

	history := make([]RecurringAmountPoint, 0, len(ordered))
	for _, link := range ordered {
		history = append(history, RecurringAmountPoint{Date: link.Date, AmountCents: link.AmountCents})
	}

	view := RecurringItemView{
		Item:        item,
		History:     history,
		LinkedCount: len(ordered),
	}
	if len(ordered) > 0 {
		last := ordered[len(ordered)-1]
		lastSeen := last.Date
		view.LastAmountCents = last.AmountCents
		view.LastSeen = &lastSeen
		view.NextExpected = item.NextExpectedAfter(&lastSeen)
	}
	return view
}

// monthlySeries totals the monthly equivalents of the running items for each of the
// last months. It looks backwards only: a month carries the most recent amount
// observed up to and including it, and an item that had not arrived yet counts as
// nothing rather than as a prediction.
func monthlySeries(running []RecurringItemView, now time.Time, months int) []RecurringMonthPoint {
	points := make([]RecurringMonthPoint, 0, months)
	current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	for offset := months - 1; offset >= 0; offset-- {
		month := current.AddDate(0, -offset, 0)
		end := month.AddDate(0, 1, 0)
		point := RecurringMonthPoint{Month: month}

		for _, view := range running {
			amount, seen := amountAsOf(view.History, end)
			if !seen {
				continue
			}
			monthly := MonthlyEquivalentCents(amount, view.Item.Rhythm)
			if view.Item.Direction == CashIn {
				point.IncomeCents += monthly
				continue
			}
			point.ExpenseCents += monthly
		}
		points = append(points, point)
	}
	return points
}

// amountAsOf returns the most recent amount observed strictly before the cutoff.
func amountAsOf(history []RecurringAmountPoint, cutoff time.Time) (money.Price, bool) {
	var (
		amount money.Price
		seen   bool
	)
	for _, point := range history {
		if !point.Date.Before(cutoff) {
			break
		}
		amount = point.AmountCents
		seen = true
	}
	return amount, seen
}

func sortByNextExpected(views []RecurringItemView) {
	// Soonest first: the page answers "what comes off next" before anything else.
	// An item that has never been seen has no expectation and reads last.
	sort.SliceStable(views, func(i, j int) bool {
		left, right := views[i].NextExpected, views[j].NextExpected
		switch {
		case left == nil && right == nil:
			return views[i].Item.Name < views[j].Item.Name
		case left == nil:
			return false
		case right == nil:
			return true
		case left.Equal(*right):
			return views[i].Item.Name < views[j].Item.Name
		default:
			return left.Before(*right)
		}
	})
}

func sortByEndedFrom(views []RecurringItemView) {
	sort.SliceStable(views, func(i, j int) bool {
		left, right := views[i].Item.EndedFrom, views[j].Item.EndedFrom
		if left == nil || right == nil || left.Equal(*right) {
			return views[i].Item.Name < views[j].Item.Name
		}
		return left.After(*right)
	})
}

func suggestionKey(matchKey string, direction CashFlowDirection) string {
	return string(direction) + "\x1F" + matchKey
}

// recurringMinOccurrences is how many transactions a pattern needs before it is
// offered. Two could be a coincidence; the third is what makes it a rhythm.
const recurringMinOccurrences = 3

// rhythmBands are the day ranges a gap between two charges may fall in, with
// enough slack for a direct debit that shifts a few days or a short month.
var rhythmBands = []struct {
	rhythm   Rhythm
	min, max int
}{
	{RhythmMonthly, 24, 38},
	{RhythmQuarterly, 80, 100},
	{RhythmYearly, 350, 380},
}

// suggestFromTransactions groups transactions by counterparty fingerprint and
// direction and keeps the groups that arrive on one of the three rhythms.
func suggestFromTransactions(transactions []*Transaction, skip map[string]struct{}) []RecurringSuggestion {
	type group struct {
		matchKey  string
		direction CashFlowDirection
		rows      []*Transaction
	}

	groups := make(map[string]*group)
	for _, tx := range transactions {
		key := MatchKeyFor(tx.Description)
		if key == "" {
			continue
		}
		id := suggestionKey(key, tx.Direction)
		if _, dropped := skip[id]; dropped {
			continue
		}
		if groups[id] == nil {
			groups[id] = &group{matchKey: key, direction: tx.Direction}
		}
		groups[id].rows = append(groups[id].rows, tx)
	}

	suggestions := make([]RecurringSuggestion, 0, len(groups))
	for _, grp := range groups {
		if len(grp.rows) < recurringMinOccurrences {
			continue
		}
		rows := grp.rows
		sort.Slice(rows, func(i, j int) bool { return rows[i].Date.Before(rows[j].Date) })

		rhythm, ok := detectRhythm(rows)
		if !ok {
			continue
		}
		newest := rows[len(rows)-1]
		suggestions = append(suggestions, RecurringSuggestion{
			MatchKey:    grp.matchKey,
			Name:        SuggestedNameFor(grp.matchKey),
			Direction:   grp.direction,
			Rhythm:      rhythm,
			AmountCents: newest.AmountCents,
			Matches:     len(rows),
			Since:       rows[0].Date,
			Sample:      newest.Description,
		})
	}

	// Most evidence first, so the pattern that is hardest to dispute reads first.
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Matches == suggestions[j].Matches {
			return suggestions[i].Name < suggestions[j].Name
		}
		return suggestions[i].Matches > suggestions[j].Matches
	})
	return suggestions
}

// detectRhythm reports the rhythm every gap in a date-ordered series fits, if any.
// One gap outside the band is enough to leave the series alone: an irregular
// series gets no expectation rather than a guessed one.
func detectRhythm(rows []*Transaction) (Rhythm, bool) {
	gaps := make([]int, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		gaps = append(gaps, int(rows[i].Date.Sub(rows[i-1].Date).Hours()/24))
	}
	if len(gaps) == 0 {
		return "", false
	}

	for _, band := range rhythmBands {
		fits := true
		for _, gap := range gaps {
			if gap < band.min || gap > band.max {
				fits = false
				break
			}
		}
		if fits {
			return band.rhythm, true
		}
	}
	return "", false
}

// matchesItem reports whether a transaction belongs to a running item: the same
// counterparty, the same direction, an amount close to the last one, and a date
// the item was still expected on.
func matchesItem(item *RecurringItem, lastAmount money.Price, tx *Transaction) bool {
	if tx.Ignored || tx.Direction != item.Direction {
		return false
	}
	if MatchKeyFor(tx.Description) != item.MatchKey {
		return false
	}
	if item.EndedFrom != nil && !date.StartOfDayUTC(tx.Date).Before(*item.EndedFrom) {
		return false
	}
	return amountWithinTolerance(tx.AmountCents, lastAmount)
}
