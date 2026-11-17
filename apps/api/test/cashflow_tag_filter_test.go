//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/sorting"
	"github.com/lennardclaproth/ta11y/internal/storage"
)

// seedTaggedTransaction inserts one row with its own checksum: seedTransaction leaves it
// empty, so a second row in the same account would be skipped as a duplicate.
func seedTaggedTransaction(t *testing.T, store *storage.SQLXCashflowStore, accountID uuid.UUID, description, tag string) {
	t.Helper()
	now := time.Now().UTC()
	tx := &cashflow.Transaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		Description: description,
		Source:      "manual",
		Direction:   cashflow.CashOut,
		AmountCents: money.Price(1000),
		Tag:         tag,
		Checksum:    uuid.NewString(),
		Date:        now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := store.CreateTransactions(t.Context(), []*cashflow.Transaction{tx}); err != nil {
		t.Fatalf("create transaction: %v", err)
	}
}

// The tag list is the account's own, counts each tag once per transaction, and has no
// entry for untagged rows.
func TestCashflowTagsListTheAccountsTagsMostUsedFirst(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		store := storage.NewSQLXCashflowStore(db)
		seedTaggedTransaction(t, store, mine, "albert heijn", "groceries")
		seedTaggedTransaction(t, store, mine, "jumbo", "groceries")
		seedTaggedTransaction(t, store, mine, "rent", "housing")
		seedTaggedTransaction(t, store, mine, "bike repair", "")
		seedTaggedTransaction(t, store, theirs, "their gym", "sport")

		tags, err := store.ListTags(t.Context(), mine)
		if err != nil {
			t.Fatalf("list tags: %v", err)
		}
		want := []cashflow.TagUsage{{Tag: "groceries", Count: 2}, {Tag: "housing", Count: 1}}
		if len(tags) != len(want) {
			t.Fatalf("tags = %+v, want %+v", tags, want)
		}
		for i := range want {
			if tags[i] != want[i] {
				t.Fatalf("tags = %+v, want %+v", tags, want)
			}
		}
	})
}

// Untagged is one more choice in the ledger's tag filter, so picking it next to a tag
// widens the list. It used to narrow it to rows both untagged and tagged: always none.
func TestCashflowTagFilterTreatsUntaggedAsOneMoreTag(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		account := seedCashflowAccount(t, db, "mine@example.com")
		store := storage.NewSQLXCashflowStore(db)
		seedTaggedTransaction(t, store, account, "bike repair", "")
		seedTaggedTransaction(t, store, account, "albert heijn", "groceries")
		seedTaggedTransaction(t, store, account, "rent", "housing")

		cases := []struct {
			name     string
			tags     []string
			untagged bool
			want     int
		}{
			{"untagged only", nil, true, 1},
			{"untagged or groceries", []string{"groceries"}, true, 2},
			{"groceries only", []string{"groceries"}, false, 1},
		}
		for _, tc := range cases {
			result, err := store.ListTransactions(t.Context(), cashflow.TransactionListQuery{
				AccountID: account,
				Tags:      tc.tags,
				Untagged:  tc.untagged,
				Sort:      sorting.Sort{Field: "date", Direction: sorting.DESC},
			})
			if err != nil {
				t.Fatalf("%s: list transactions: %v", tc.name, err)
			}
			if result.Total != tc.want {
				t.Fatalf("%s: total = %d, want %d", tc.name, result.Total, tc.want)
			}
		}
	})
}
