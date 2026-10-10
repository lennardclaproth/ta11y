package cashflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// recurringStore is an in-memory RecurringCommandStore/RecurringQueryStore pair,
// enough to drive the recurring use cases without a database.
type recurringStore struct {
	items        []*RecurringItem
	links        []RecurringLink
	unlinked     []*Transaction
	imported     []*Transaction
	dismissed    []DismissedSuggestion
	transactions map[uuid.UUID]*Transaction

	linkedTo map[uuid.UUID][]uuid.UUID

	// linkErr makes linking fail, and broadLinksErr makes the account-wide link
	// read fail, so a test can tell the two reads apart.
	linkErr       error
	broadLinksErr error
}

func newRecurringStore() *recurringStore {
	return &recurringStore{
		transactions: map[uuid.UUID]*Transaction{},
		linkedTo:     map[uuid.UUID][]uuid.UUID{},
	}
}

func (s *recurringStore) CreateRecurringItem(_ context.Context, item *RecurringItem) error {
	for _, existing := range s.items {
		if existing.Name == item.Name {
			return ErrRecurringItemExists
		}
	}
	s.items = append(s.items, item)
	return nil
}

func (s *recurringStore) UpdateRecurringItem(_ context.Context, _ *RecurringItem) (int, error) {
	return 1, nil
}

func (s *recurringStore) EndRecurringItem(_ context.Context, _, id uuid.UUID, from time.Time) (int, error) {
	for _, item := range s.items {
		if item.ID == id {
			ended := from
			item.EndedFrom = &ended
			return 1, nil
		}
	}
	return 0, nil
}

func (s *recurringStore) LinkTransactions(_ context.Context, itemID uuid.UUID, ids []uuid.UUID) (int, error) {
	if s.linkErr != nil {
		return 0, s.linkErr
	}
	s.linkedTo[itemID] = append(s.linkedTo[itemID], ids...)
	return len(ids), nil
}

// Do runs fn and undoes the items and links it wrote when it fails, which is the
// part of a real transaction these use cases depend on.
func (s *recurringStore) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	items := append([]*RecurringItem(nil), s.items...)
	linked := make(map[uuid.UUID][]uuid.UUID, len(s.linkedTo))
	for id, ids := range s.linkedTo {
		linked[id] = append([]uuid.UUID(nil), ids...)
	}

	if err := fn(ctx); err != nil {
		s.items = items
		s.linkedTo = linked
		return err
	}
	return nil
}

func (s *recurringStore) UnlinkTransaction(_ context.Context, _, _, _ uuid.UUID) (int, error) {
	return 1, nil
}

func (s *recurringStore) DismissRecurringSuggestion(_ context.Context, _ uuid.UUID, matchKey string, direction CashFlowDirection) error {
	s.dismissed = append(s.dismissed, DismissedSuggestion{MatchKey: matchKey, Direction: direction})
	return nil
}

func (s *recurringStore) ListRecurringItems(_ context.Context, _ uuid.UUID) ([]*RecurringItem, error) {
	return s.items, nil
}

func (s *recurringStore) GetRecurringItem(_ context.Context, _, id uuid.UUID) (*RecurringItem, error) {
	for _, item := range s.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (s *recurringStore) ListRecurringLinks(_ context.Context, _ uuid.UUID) ([]RecurringLink, error) {
	if s.broadLinksErr != nil {
		return nil, s.broadLinksErr
	}
	return s.links, nil
}

func (s *recurringStore) ListRecurringLinksForItem(_ context.Context, _, itemID uuid.UUID) ([]RecurringLink, error) {
	own := []RecurringLink{}
	for _, link := range s.links {
		if link.ItemID == itemID {
			own = append(own, link)
		}
	}
	return own, nil
}

func (s *recurringStore) ListUnlinkedTransactions(_ context.Context, _ uuid.UUID) ([]*Transaction, error) {
	return s.unlinked, nil
}

func (s *recurringStore) ListDismissedSuggestions(_ context.Context, _ uuid.UUID) ([]DismissedSuggestion, error) {
	return s.dismissed, nil
}

func (s *recurringStore) ListTransactionsByIDs(_ context.Context, _ uuid.UUID, ids []uuid.UUID) ([]*Transaction, error) {
	out := []*Transaction{}
	for _, id := range ids {
		if tx, ok := s.transactions[id]; ok {
			out = append(out, tx)
		}
	}
	return out, nil
}

func (s *recurringStore) ListImportedTransactions(_ context.Context, _, _ uuid.UUID) ([]*Transaction, error) {
	return s.imported, nil
}

func day(year int, month time.Month, dayOfMonth int) time.Time {
	return time.Date(year, month, dayOfMonth, 0, 0, 0, 0, time.UTC)
}

func recurringTransaction(description string, direction CashFlowDirection, amount money.Price, on time.Time) *Transaction {
	return &Transaction{
		ID:          uuid.New(),
		Description: description,
		Direction:   direction,
		AmountCents: amount,
		Date:        on,
	}
}

func TestMatchKeyIgnoresReferencesAndStatementLabels(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"reference digits are dropped", "CLOUDLOCKER*SUB 0312", "cloudlocker sub"},
		{"the same counterparty on two statements", "PIXEL STREAM MONTHLY 2026-09", "pixel stream monthly"},
		{"statement labels are dropped", "Naam: Pixel Stream Omschrijving: abonnement", "pixel stream abonnement"},
		{"description without letters falls back", "4411 9988", "4411 9988"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := MatchKeyFor(tc.input); got != tc.want {
				t.Fatalf("MatchKeyFor(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestMatchKeyRecognisesTheSameCounterpartyAcrossStatements(t *testing.T) {
	first := MatchKeyFor("PIXEL STREAM MONTHLY 0918 MANDATE 44112")
	second := MatchKeyFor("PIXEL STREAM MONTHLY 1018 MANDATE 44119")
	if first != second {
		t.Fatalf("expected one key for two statements of the same payment, got %q and %q", first, second)
	}
}

func TestSuggestionsNeedARhythmAndThreeOccurrences(t *testing.T) {
	cases := []struct {
		name  string
		dates []time.Time
		want  bool
	}{
		{"monthly with a few days of drift", []time.Time{day(2026, 6, 18), day(2026, 7, 20), day(2026, 8, 17), day(2026, 9, 18)}, true},
		{"quarterly", []time.Time{day(2026, 1, 15), day(2026, 4, 15), day(2026, 7, 15)}, true},
		{"only twice", []time.Time{day(2026, 8, 18), day(2026, 9, 18)}, false},
		{"irregular", []time.Time{day(2026, 6, 2), day(2026, 6, 20), day(2026, 9, 11)}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows := make([]*Transaction, 0, len(tc.dates))
			for _, on := range tc.dates {
				rows = append(rows, recurringTransaction("PIXEL STREAM MONTHLY", CashOut, 1200, on))
			}

			got := suggestFromTransactions(rows, map[string]struct{}{})
			if (len(got) == 1) != tc.want {
				t.Fatalf("expected suggestion=%v, got %d suggestions", tc.want, len(got))
			}
		})
	}
}

func TestSuggestionsLeaveOutDismissedPatterns(t *testing.T) {
	rows := []*Transaction{
		recurringTransaction("PIXEL STREAM MONTHLY", CashOut, 1200, day(2026, 7, 18)),
		recurringTransaction("PIXEL STREAM MONTHLY", CashOut, 1200, day(2026, 8, 18)),
		recurringTransaction("PIXEL STREAM MONTHLY", CashOut, 1200, day(2026, 9, 18)),
	}
	skip := map[string]struct{}{suggestionKey(MatchKeyFor("PIXEL STREAM MONTHLY"), CashOut): {}}

	if got := suggestFromTransactions(rows, skip); len(got) != 0 {
		t.Fatalf("expected a dismissed pattern not to come back, got %d suggestions", len(got))
	}
}

func TestOverviewTotalsSpreadQuarterlyAndYearlyAmounts(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()

	monthly := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Fiber", Direction: CashOut, Rhythm: RhythmMonthly}
	quarterly := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Insurance", Direction: CashOut, Rhythm: RhythmQuarterly}
	salary := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Payroll", Direction: CashIn, Rhythm: RhythmMonthly}
	store.items = []*RecurringItem{monthly, quarterly, salary}
	store.links = []RecurringLink{
		{ItemID: monthly.ID, TransactionID: uuid.New(), Date: day(2026, 9, 24), AmountCents: 4500},
		{ItemID: quarterly.ID, TransactionID: uuid.New(), Date: day(2026, 9, 1), AmountCents: 9600},
		{ItemID: salary.ID, TransactionID: uuid.New(), Date: day(2026, 9, 25), AmountCents: 420000},
	}

	overview, err := NewRecurringQueries(store).Overview(context.Background(), accountID, "")
	if err != nil {
		t.Fatalf("overview: %v", err)
	}

	// 4500 monthly + 9600 quarterly counted as a third.
	if overview.MonthlyExpenseCents != 4500+3200 {
		t.Fatalf("monthly expenses = %d, want %d", overview.MonthlyExpenseCents, 4500+3200)
	}
	if overview.MonthlyIncomeCents != 420000 {
		t.Fatalf("monthly income = %d, want 420000", overview.MonthlyIncomeCents)
	}
	if len(overview.Expenses) != 2 || len(overview.Income) != 1 {
		t.Fatalf("groups = %d expenses / %d income, want 2 / 1", len(overview.Expenses), len(overview.Income))
	}
}

func TestOverviewTotalsIgnoreTheSearchAndEndedItems(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()

	ended := day(2026, 8, 1)
	running := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Fiber", Direction: CashOut, Rhythm: RhythmMonthly}
	stopped := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Magazine", Direction: CashOut, Rhythm: RhythmMonthly, EndedFrom: &ended}
	store.items = []*RecurringItem{running, stopped}
	store.links = []RecurringLink{
		{ItemID: running.ID, TransactionID: uuid.New(), Date: day(2026, 9, 24), AmountCents: 4500},
		{ItemID: stopped.ID, TransactionID: uuid.New(), Date: day(2026, 7, 14), AmountCents: 800},
	}

	overview, err := NewRecurringQueries(store).Overview(context.Background(), accountID, "magazine")
	if err != nil {
		t.Fatalf("overview: %v", err)
	}

	if len(overview.Expenses) != 0 || len(overview.Ended) != 1 {
		t.Fatalf("search returned %d expenses / %d ended, want 0 / 1", len(overview.Expenses), len(overview.Ended))
	}
	// The totals describe the account, not the search, and an ended item is not running.
	if overview.MonthlyExpenseCents != 4500 {
		t.Fatalf("monthly expenses = %d, want 4500", overview.MonthlyExpenseCents)
	}
}

func TestItemReportsNextExpectedAndAmountHistory(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()

	item := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Pixel Stream", Direction: CashOut, Rhythm: RhythmMonthly}
	store.items = []*RecurringItem{item}
	store.links = []RecurringLink{
		{ItemID: item.ID, TransactionID: uuid.New(), Date: day(2026, 8, 18), AmountCents: 1000},
		{ItemID: item.ID, TransactionID: uuid.New(), Date: day(2026, 9, 18), AmountCents: 1200},
		{ItemID: item.ID, TransactionID: uuid.New(), Date: day(2026, 7, 18), AmountCents: 900},
	}

	detail, err := NewRecurringQueries(store).Item(context.Background(), accountID, item.ID)
	if err != nil {
		t.Fatalf("item: %v", err)
	}

	if detail.LastAmountCents != 1200 {
		t.Fatalf("last amount = %d, want 1200", detail.LastAmountCents)
	}
	if detail.NextExpected == nil || !detail.NextExpected.Equal(day(2026, 10, 18)) {
		t.Fatalf("next expected = %v, want 2026-10-18", detail.NextExpected)
	}
	if len(detail.History) != 3 || detail.History[0].AmountCents != 900 {
		t.Fatalf("history = %+v, want three amounts oldest first", detail.History)
	}
	// Newest first in the drawer, so the most recent charge reads without scrolling.
	if !detail.Transactions[0].Date.Equal(day(2026, 9, 18)) {
		t.Fatalf("transactions start at %v, want 2026-09-18", detail.Transactions[0].Date)
	}
}

func TestEndedItemIsNoLongerExpected(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()
	item := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Magazine", Direction: CashOut, Rhythm: RhythmMonthly}
	store.items = []*RecurringItem{item}
	store.links = []RecurringLink{{ItemID: item.ID, TransactionID: uuid.New(), Date: day(2026, 7, 14), AmountCents: 800}}

	commands := NewRecurringCommands(store, store, store)
	if _, err := commands.End(context.Background(), accountID, item.ID, "2026-08"); err != nil {
		t.Fatalf("end: %v", err)
	}

	detail, err := NewRecurringQueries(store).Item(context.Background(), accountID, item.ID)
	if err != nil {
		t.Fatalf("item: %v", err)
	}
	if detail.NextExpected != nil {
		t.Fatalf("ended item still expects %v", detail.NextExpected)
	}
	if detail.LinkedCount != 1 {
		t.Fatalf("ending dropped history: %d transactions linked", detail.LinkedCount)
	}
}

func TestEndRejectsAMonthItCannotRead(t *testing.T) {
	store := newRecurringStore()
	item := &RecurringItem{ID: uuid.New(), Name: "Magazine", Direction: CashOut, Rhythm: RhythmMonthly}
	store.items = []*RecurringItem{item}

	_, err := NewRecurringCommands(store, store, store).End(context.Background(), uuid.New(), item.ID, "August")
	if !errors.Is(err, ErrRecurringInvalidMonth) {
		t.Fatalf("expected ErrRecurringInvalidMonth, got %v", err)
	}
}

func TestLinkImportedAttachesMatchingRowsToConfirmedItems(t *testing.T) {
	accountID := uuid.New()
	importID := uuid.New()
	store := newRecurringStore()

	item := &RecurringItem{
		ID: uuid.New(), AccountID: accountID, Name: "Pixel Stream",
		Direction: CashOut, Rhythm: RhythmMonthly, MatchKey: MatchKeyFor("PIXEL STREAM MONTHLY"),
	}
	store.items = []*RecurringItem{item}
	store.links = []RecurringLink{{ItemID: item.ID, TransactionID: uuid.New(), Date: day(2026, 9, 18), AmountCents: 1200}}

	match := recurringTransaction("PIXEL STREAM MONTHLY 1018", CashOut, 1200, day(2026, 10, 18))
	otherAmount := recurringTransaction("PIXEL STREAM MONTHLY 1018", CashOut, 9900, day(2026, 10, 18))
	otherParty := recurringTransaction("HARBOUR MARKET", CashOut, 1200, day(2026, 10, 19))
	ignored := recurringTransaction("PIXEL STREAM MONTHLY 1118", CashOut, 1200, day(2026, 11, 18))
	ignored.Ignored = true
	store.imported = []*Transaction{match, otherAmount, otherParty, ignored}

	linked, err := NewRecurringCommands(store, store, store).LinkImported(context.Background(), accountID, importID)
	if err != nil {
		t.Fatalf("link imported: %v", err)
	}
	if linked != 1 {
		t.Fatalf("linked %d rows, want 1", linked)
	}
	if got := store.linkedTo[item.ID]; len(got) != 1 || got[0] != match.ID {
		t.Fatalf("linked %v, want only the matching row", got)
	}
}

func TestLinkImportedLeavesEndedItemsAlone(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()

	ended := day(2026, 10, 1)
	item := &RecurringItem{
		ID: uuid.New(), AccountID: accountID, Name: "Pixel Stream", Direction: CashOut,
		Rhythm: RhythmMonthly, MatchKey: MatchKeyFor("PIXEL STREAM MONTHLY"), EndedFrom: &ended,
	}
	store.items = []*RecurringItem{item}
	store.links = []RecurringLink{{ItemID: item.ID, TransactionID: uuid.New(), Date: day(2026, 9, 18), AmountCents: 1200}}
	store.imported = []*Transaction{recurringTransaction("PIXEL STREAM MONTHLY 1018", CashOut, 1200, day(2026, 10, 18))}

	linked, err := NewRecurringCommands(store, store, store).LinkImported(context.Background(), accountID, uuid.New())
	if err != nil {
		t.Fatalf("link imported: %v", err)
	}
	if linked != 0 {
		t.Fatalf("linked %d rows to an ended item, want 0", linked)
	}
}

func TestConfirmSuggestionLinksEveryTransactionItWasFoundIn(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()
	store.unlinked = []*Transaction{
		recurringTransaction("CLOUDLOCKER*SUB 0306", CashOut, 400, day(2026, 7, 6)),
		recurringTransaction("CLOUDLOCKER*SUB 0406", CashOut, 400, day(2026, 8, 6)),
		recurringTransaction("CLOUDLOCKER*SUB 0506", CashOut, 400, day(2026, 9, 6)),
		recurringTransaction("HARBOUR MARKET", CashOut, 6200, day(2026, 9, 7)),
	}

	item, err := NewRecurringCommands(store, store, store).ConfirmSuggestion(context.Background(), accountID, ConfirmSuggestionInput{
		MatchKey:  MatchKeyFor("CLOUDLOCKER*SUB 0306"),
		Name:      "Cloud Locker",
		Direction: CashOut,
		Rhythm:    RhythmMonthly,
	})
	if err != nil {
		t.Fatalf("confirm suggestion: %v", err)
	}
	if item.Name != "Cloud Locker" {
		t.Fatalf("item name = %q, want the confirmed name", item.Name)
	}
	if got := store.linkedTo[item.ID]; len(got) != 3 {
		t.Fatalf("linked %d transactions, want the 3 the pattern was found in", len(got))
	}
}

func TestCreateRefusesTransactionsTheAccountDoesNotHold(t *testing.T) {
	store := newRecurringStore()

	_, err := NewRecurringCommands(store, store, store).Create(context.Background(), uuid.New(), CreateRecurringInput{
		Name:           "Pixel Stream",
		Direction:      CashOut,
		Rhythm:         RhythmMonthly,
		TransactionIDs: []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, ErrRecurringTransactionNotFound) {
		t.Fatalf("expected ErrRecurringTransactionNotFound, got %v", err)
	}
}

func TestCreateTakesItsFingerprintFromTheOldestTransaction(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()

	older := recurringTransaction("PIXEL STREAM MONTHLY 0818", CashOut, 1200, day(2026, 8, 18))
	newer := recurringTransaction("SEPA INCASSO PIXEL STREAM", CashOut, 1200, day(2026, 9, 18))
	store.transactions[older.ID] = older
	store.transactions[newer.ID] = newer

	item, err := NewRecurringCommands(store, store, store).Create(context.Background(), accountID, CreateRecurringInput{
		Name:           "Pixel Stream",
		Direction:      CashOut,
		Rhythm:         RhythmMonthly,
		TransactionIDs: []uuid.UUID{newer.ID, older.ID},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if item.MatchKey != MatchKeyFor(older.Description) {
		t.Fatalf("match key = %q, want the oldest transaction's %q", item.MatchKey, MatchKeyFor(older.Description))
	}
	if got := store.linkedTo[item.ID]; len(got) != 2 {
		t.Fatalf("linked %d transactions, want both", len(got))
	}
}

func TestCreatingAnItemLeavesNothingBehindWhenLinkingFails(t *testing.T) {
	accountID := uuid.New()
	transaction := recurringTransaction("PIXEL STREAM MONTHLY 0818", CashOut, 1200, day(2026, 8, 18))

	t.Run("marked", func(t *testing.T) {
		store := newRecurringStore()
		store.transactions[transaction.ID] = transaction
		store.linkErr = errors.New("link failed")

		_, err := NewRecurringCommands(store, store, store).Create(context.Background(), accountID, CreateRecurringInput{
			Name:           "Pixel Stream",
			Direction:      CashOut,
			Rhythm:         RhythmMonthly,
			TransactionIDs: []uuid.UUID{transaction.ID},
		})
		if err == nil {
			t.Fatal("expected create to fail when linking fails")
		}
		// An item without links can be neither removed nor linked to again.
		if len(store.items) != 0 {
			t.Fatalf("a failed link left %d items behind, want none", len(store.items))
		}
	})

	t.Run("confirmed from a suggestion", func(t *testing.T) {
		store := newRecurringStore()
		store.unlinked = []*Transaction{
			recurringTransaction("CLOUDLOCKER*SUB 0306", CashOut, 400, day(2026, 7, 6)),
			recurringTransaction("CLOUDLOCKER*SUB 0406", CashOut, 400, day(2026, 8, 6)),
		}
		store.linkErr = errors.New("link failed")

		_, err := NewRecurringCommands(store, store, store).ConfirmSuggestion(context.Background(), accountID, ConfirmSuggestionInput{
			MatchKey:  MatchKeyFor("CLOUDLOCKER*SUB 0306"),
			Name:      "Cloud Locker",
			Direction: CashOut,
			Rhythm:    RhythmMonthly,
		})
		if err == nil {
			t.Fatal("expected confirm to fail when linking fails")
		}
		if len(store.items) != 0 {
			t.Fatalf("a failed link left %d items behind, want none", len(store.items))
		}
	})
}

func TestItemReadsOnlyItsOwnLinks(t *testing.T) {
	accountID := uuid.New()
	store := newRecurringStore()

	item := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Pixel Stream", Direction: CashOut, Rhythm: RhythmMonthly}
	other := &RecurringItem{ID: uuid.New(), AccountID: accountID, Name: "Fiber", Direction: CashOut, Rhythm: RhythmMonthly}
	store.items = []*RecurringItem{item, other}
	store.links = []RecurringLink{
		{ItemID: item.ID, TransactionID: uuid.New(), Date: day(2026, 9, 18), AmountCents: 1200},
		{ItemID: other.ID, TransactionID: uuid.New(), Date: day(2026, 9, 24), AmountCents: 4500},
	}
	// The drawer opens per item, so the whole account's links stay out of this read.
	store.broadLinksErr = errors.New("the account-wide link read must stay out of the item detail")

	detail, err := NewRecurringQueries(store).Item(context.Background(), accountID, item.ID)
	if err != nil {
		t.Fatalf("item: %v", err)
	}
	if len(detail.Transactions) != 1 || detail.Transactions[0].ItemID != item.ID {
		t.Fatalf("item detail carries %+v, want only its own link", detail.Transactions)
	}
}

func TestAmountToleranceIsAQuarterOfTheLastAmount(t *testing.T) {
	cases := []struct {
		name      string
		candidate money.Price
		reference money.Price
		want      bool
	}{
		{"exactly a quarter above", 1250, 1000, true},
		{"exactly a quarter below", 750, 1000, true},
		{"just outside", 1251, 1000, false},
		{"an unrelated payment", 9900, 1200, false},
		{"nothing observed yet", 1200, 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := amountWithinTolerance(tc.candidate, tc.reference); got != tc.want {
				t.Fatalf("amountWithinTolerance(%d, %d) = %v, want %v", tc.candidate, tc.reference, got, tc.want)
			}
		})
	}
}

func TestMonthlySeriesLooksBackwardsOnly(t *testing.T) {
	item := &RecurringItem{ID: uuid.New(), Name: "Fiber", Direction: CashOut, Rhythm: RhythmMonthly}
	views := []RecurringItemView{{
		Item: item,
		History: []RecurringAmountPoint{
			{Date: day(2026, 8, 24), AmountCents: 4000},
			{Date: day(2026, 9, 24), AmountCents: 4500},
		},
	}}

	series := monthlySeries(views, day(2026, 10, 9), 4)
	if len(series) != 4 {
		t.Fatalf("series has %d months, want 4", len(series))
	}
	// July: nothing observed yet, so the item contributes nothing rather than a guess.
	if series[0].ExpenseCents != 0 {
		t.Fatalf("July total = %d, want 0 before the first charge", series[0].ExpenseCents)
	}
	if series[1].ExpenseCents != 4000 {
		t.Fatalf("August total = %d, want the amount charged then (4000)", series[1].ExpenseCents)
	}
	if series[2].ExpenseCents != 4500 {
		t.Fatalf("September total = %d, want the amount charged then (4500)", series[2].ExpenseCents)
	}
	// October has no charge yet, so it carries the last amount rather than dropping to zero.
	if series[3].ExpenseCents != 4500 {
		t.Fatalf("October total = %d, want 4500", series[3].ExpenseCents)
	}
}
