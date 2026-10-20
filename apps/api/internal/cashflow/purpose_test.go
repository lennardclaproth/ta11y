package cashflow

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// purposeStore records what a purpose mutation asked the store for, without a database.
type purposeStore struct {
	dateStore

	matched        int
	purposeByIDs   Purpose
	idsMarked      []uuid.UUID
	purposeByFiltr Purpose
	filtersMarked  TransactionFilters
	updated        int
}

func (s *purposeStore) CountByFilter(_ context.Context, _ TransactionFilters) (int, error) {
	return s.matched, nil
}

func (s *purposeStore) UpdatePurposeByIDs(_ context.Context, _ uuid.UUID, ids []uuid.UUID, purpose Purpose) (int, error) {
	s.idsMarked = ids
	s.purposeByIDs = purpose
	return s.updated, nil
}

func (s *purposeStore) UpdatePurposeByFilter(_ context.Context, filters TransactionFilters, purpose Purpose) (int, error) {
	s.filtersMarked = filters
	s.purposeByFiltr = purpose
	return s.updated, nil
}

func TestParsePurpose(t *testing.T) {
	cases := map[string]Purpose{
		"":       PurposeNone,
		"none":   PurposeNone,
		" None ": PurposeNone,
		"income": PurposeIncome,
		"WEALTH": PurposeWealth,
	}
	for raw, want := range cases {
		got, err := ParsePurpose(raw)
		if err != nil {
			t.Errorf("ParsePurpose(%q) error = %v", raw, err)
			continue
		}
		if got != want {
			t.Errorf("ParsePurpose(%q) = %q, want %q", raw, got, want)
		}
	}
	if _, err := ParsePurpose("savings"); err != ErrInvalidPurpose {
		t.Errorf("ParsePurpose(\"savings\") error = %v, want %v", err, ErrInvalidPurpose)
	}
}

// The filter accepts several purposes at once and drops duplicates; an empty input is no
// filter rather than a filter on "not assigned".
func TestSplitPurposes(t *testing.T) {
	got, err := SplitPurposes("income, none ,income")
	if err != nil {
		t.Fatalf("SplitPurposes: %v", err)
	}
	if len(got) != 2 || got[0] != PurposeIncome || got[1] != PurposeNone {
		t.Errorf("SplitPurposes = %v, want [income none]", got)
	}
	if empty, err := SplitPurposes("  "); err != nil || empty != nil {
		t.Errorf("SplitPurposes(blank) = %v, %v, want nil, nil", empty, err)
	}
	if _, err := SplitPurposes("income,savings"); err != ErrInvalidPurpose {
		t.Errorf("SplitPurposes error = %v, want %v", err, ErrInvalidPurpose)
	}
}

// Only incoming money can be income and only outgoing money a contribution: counting both
// sides of a transfer between your own accounts would double the month.
func TestPurposeRequiredDirection(t *testing.T) {
	if direction := PurposeIncome.RequiredDirection(); direction == nil || *direction != CashIn {
		t.Errorf("income direction = %v, want in", direction)
	}
	if direction := PurposeWealth.RequiredDirection(); direction == nil || *direction != CashOut {
		t.Errorf("wealth direction = %v, want out", direction)
	}
	if direction := PurposeNone.RequiredDirection(); direction != nil {
		t.Errorf("clearing direction = %v, want nil", direction)
	}
}

// A selection may hold both directions, so the result separates what was pointed at from
// what the purpose could actually apply to.
func TestMarkPurposeByIDsReportsWhatThePurposeStuckTo(t *testing.T) {
	store := &purposeStore{updated: 2}
	commands := NewCommands(store, store, &alwaysExists{})
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	result, err := commands.MarkPurposeByIDs(context.Background(), uuid.New(), ids, PurposeIncome)
	if err != nil {
		t.Fatalf("mark purpose by ids: %v", err)
	}
	if result.Matched != 3 || result.Updated != 2 {
		t.Errorf("result = %+v, want matched 3 / updated 2", result)
	}
	if store.purposeByIDs != PurposeIncome {
		t.Errorf("purpose = %q, want income", store.purposeByIDs)
	}
}

// The account is authoritative over anything the filters carried in.
func TestMarkPurposeByFilterScopesToTheAccount(t *testing.T) {
	store := &purposeStore{matched: 8, updated: 5}
	commands := NewCommands(store, store, &alwaysExists{})
	accountID := uuid.New()

	result, err := commands.MarkPurposeByFilter(
		context.Background(),
		accountID,
		TransactionFilters{AccountID: uuid.New()},
		PurposeWealth,
	)
	if err != nil {
		t.Fatalf("mark purpose by filter: %v", err)
	}
	if result.Matched != 8 || result.Updated != 5 {
		t.Errorf("result = %+v, want matched 8 / updated 5", result)
	}
	if store.filtersMarked.AccountID != accountID {
		t.Errorf("filters account = %s, want %s", store.filtersMarked.AccountID, accountID)
	}
	if store.purposeByFiltr != PurposeWealth {
		t.Errorf("purpose = %q, want wealth", store.purposeByFiltr)
	}
}

type alwaysExists struct{}

func (alwaysExists) Exists(context.Context, uuid.UUID) (bool, error) { return true, nil }
