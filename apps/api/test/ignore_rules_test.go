//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/storage"
)

// seedIgnoreRule stores a rule matching outgoing rows whose description contains text.
func seedIgnoreRule(t *testing.T, db *storage.DB, accountID uuid.UUID, name, text string) *cashflow.IgnoreRule {
	t.Helper()

	draft, err := cashflow.NewIgnoreRuleDraft(name, "description", text, "out", "", true)
	if err != nil {
		t.Fatalf("new ignore rule draft: %v", err)
	}
	rule := cashflow.NewIgnoreRule(accountID, draft)
	if err := storage.NewSQLXCashflowIgnoreRuleStore(db).CreateIgnoreRule(t.Context(), rule); err != nil {
		t.Fatalf("create ignore rule: %v", err)
	}
	return rule
}

// seedRuleTransaction stores one outgoing transaction with an explicit ignored state and
// a unique checksum, so several rows survive the bulk insert's conflict clause.
func seedRuleTransaction(
	t *testing.T,
	db *storage.DB,
	accountID uuid.UUID,
	description string,
	ignored, overridden bool,
	importID *uuid.UUID,
) uuid.UUID {
	t.Helper()

	tx := &cashflow.Transaction{
		ID:               uuid.New(),
		AccountID:        accountID,
		Description:      description,
		Source:           "ING",
		Direction:        cashflow.CashOut,
		AmountCents:      money.Price(1000),
		Date:             time.Now().UTC(),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
		Checksum:         uuid.NewString(),
		Ignored:          ignored,
		IgnoreOverridden: overridden,
		ImportID:         importID,
	}
	if _, err := storage.NewSQLXCashflowStore(db).CreateTransactions(t.Context(), []*cashflow.Transaction{tx}); err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	// The insert deliberately never writes ignore_overridden -- a new row is nobody's
	// decision yet -- so the flag is set here to stand for a row someone has since
	// ignored or restored by hand.
	if overridden {
		if _, err := db.ExecContext(t.Context(),
			db.Rebind(`UPDATE transactions SET ignore_overridden = ? WHERE id = ?`), true, tx.ID,
		); err != nil {
			t.Fatalf("mark transaction as decided by hand: %v", err)
		}
	}
	return tx.ID
}

// seedImport stores the vendor and import row transactions can be attributed to.
func seedImport(t *testing.T, db *storage.DB, accountID uuid.UUID, name string) uuid.UUID {
	t.Helper()

	vendorID := uuid.New()
	now := time.Now().UTC()
	if _, err := db.ExecContext(t.Context(),
		db.Rebind(`INSERT INTO vendors (id, name, type, active, import_disabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`),
		vendorID, name, "bank", true, false, now, now,
	); err != nil {
		t.Fatalf("create vendor: %v", err)
	}

	importID := uuid.New()
	if _, err := db.ExecContext(t.Context(),
		db.Rebind(`INSERT INTO imports (id, vendor_id, account_id, type, source, path, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		importID, vendorID, accountID, "cashflow", name, fmt.Sprintf("/tmp/%s.csv", name), "completed", now, now,
	); err != nil {
		t.Fatalf("create import: %v", err)
	}
	return importID
}

func fetchTransaction(t *testing.T, db *storage.DB, accountID, id uuid.UUID) *cashflow.Transaction {
	t.Helper()

	tx, err := storage.NewSQLXCashflowStore(db).GetTransaction(t.Context(), accountID, id)
	if err != nil {
		t.Fatalf("fetch transaction: %v", err)
	}
	if tx == nil {
		t.Fatalf("transaction %v is gone", id)
	}
	return tx
}

// A rule may only touch rows that are not ignored yet and whose ignored state nobody
// decided by hand. The by-hand case is the one that matters: a row put back by hand must
// stay back, or restoring is pointless the moment the rule runs again.
func TestApplyIgnoreRuleTouchesOnlyWhatItMay(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedCashflowAccount(t, db, "mine@example.com")
		rule := seedIgnoreRule(t, db, accountID, "Credit card payment", "Credit card")

		fresh := seedRuleTransaction(t, db, accountID, "Credit card payment", false, false, nil)
		alreadyIgnored := seedRuleTransaction(t, db, accountID, "Credit card payment June", true, false, nil)
		restoredByHand := seedRuleTransaction(t, db, accountID, "Credit card annual fee", false, true, nil)
		unrelated := seedRuleTransaction(t, db, accountID, "Groceries week 26", false, false, nil)

		store := storage.NewSQLXCashflowIgnoreRuleStore(db)
		ignored, err := store.ApplyIgnoreRule(t.Context(), rule, nil)
		if err != nil {
			t.Fatalf("apply ignore rule: %v", err)
		}
		if ignored != 1 {
			t.Fatalf("ignored = %d, want only the one row the rule may touch", ignored)
		}

		if tx := fetchTransaction(t, db, accountID, fresh); !tx.Ignored || tx.IgnoredByRuleID == nil || *tx.IgnoredByRuleID != rule.ID {
			t.Fatalf("the matching row was not ignored by the rule: ignored=%v by=%v", tx.Ignored, tx.IgnoredByRuleID)
		}
		if tx := fetchTransaction(t, db, accountID, alreadyIgnored); tx.IgnoredByRuleID != nil {
			t.Fatal("a row that was already ignored was re-attributed to the rule")
		}
		if tx := fetchTransaction(t, db, accountID, restoredByHand); tx.Ignored {
			t.Fatal("a row restored by hand was ignored again by the rule")
		}
		if tx := fetchTransaction(t, db, accountID, unrelated); tx.Ignored {
			t.Fatal("a row the rule does not match was ignored")
		}

		// The rule's own total is the count the rules page reports, so it has to land in
		// the same breath as the rows it describes.
		stored, err := store.GetIgnoreRule(t.Context(), accountID, rule.ID)
		if err != nil {
			t.Fatalf("fetch ignore rule: %v", err)
		}
		if stored.IgnoredTotal != 1 {
			t.Fatalf("ignored_total = %d, want 1", stored.IgnoredTotal)
		}
		if stored.LastAppliedAt == nil {
			t.Fatal("last_applied_at was not stamped")
		}
	})
}

// CountIgnoreRuleTargets is the number the confirmation before "apply to existing" names,
// so it has to agree with what applying actually does.
func TestCountIgnoreRuleTargetsMatchesWhatApplyingDoes(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedCashflowAccount(t, db, "mine@example.com")
		rule := seedIgnoreRule(t, db, accountID, "Credit card payment", "Credit card")

		seedRuleTransaction(t, db, accountID, "Credit card payment", false, false, nil)
		seedRuleTransaction(t, db, accountID, "Credit card payment May", false, false, nil)
		seedRuleTransaction(t, db, accountID, "Credit card annual fee", false, true, nil)

		store := storage.NewSQLXCashflowIgnoreRuleStore(db)
		counted, err := store.CountIgnoreRuleTargets(t.Context(), rule.Filters())
		if err != nil {
			t.Fatalf("count ignore rule targets: %v", err)
		}
		applied, err := store.ApplyIgnoreRule(t.Context(), rule, nil)
		if err != nil {
			t.Fatalf("apply ignore rule: %v", err)
		}
		if counted != applied {
			t.Fatalf("counted %d targets but ignored %d", counted, applied)
		}
		if counted != 2 {
			t.Fatalf("counted = %d, want 2", counted)
		}
	})
}

// Scoped to an import, a rule may only see the rows that import inserted. That is what
// keeps a re-imported statement from handing the rules rows the person already reviewed.
func TestApplyIgnoreRuleScopedToAnImportLeavesTheRestAlone(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedCashflowAccount(t, db, "mine@example.com")
		rule := seedIgnoreRule(t, db, accountID, "Credit card payment", "Credit card")
		importID := seedImport(t, db, accountID, "ING")

		inImport := seedRuleTransaction(t, db, accountID, "Credit card payment", false, false, &importID)
		older := seedRuleTransaction(t, db, accountID, "Credit card payment May", false, false, nil)

		ignored, err := storage.NewSQLXCashflowIgnoreRuleStore(db).ApplyIgnoreRule(t.Context(), rule, &importID)
		if err != nil {
			t.Fatalf("apply ignore rule: %v", err)
		}
		if ignored != 1 {
			t.Fatalf("ignored = %d, want only the row this import brought in", ignored)
		}
		if tx := fetchTransaction(t, db, accountID, inImport); !tx.Ignored {
			t.Fatal("the imported row was not ignored")
		}
		if tx := fetchTransaction(t, db, accountID, older); tx.Ignored {
			t.Fatal("a row from outside the import was ignored")
		}
	})
}

func TestApplyIgnoreRuleCannotCrossAccounts(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		rule := seedIgnoreRule(t, db, mine, "Credit card payment", "Credit card")
		theirTx := seedRuleTransaction(t, db, theirs, "Credit card payment", false, false, nil)

		ignored, err := storage.NewSQLXCashflowIgnoreRuleStore(db).ApplyIgnoreRule(t.Context(), rule, nil)
		if err != nil {
			t.Fatalf("apply ignore rule: %v", err)
		}
		if ignored != 0 {
			t.Fatalf("ignored = %d, want 0 -- the rule reached into another account", ignored)
		}
		if tx := fetchTransaction(t, db, theirs, theirTx); tx.Ignored {
			t.Fatal("another account's transaction was ignored")
		}
	})
}

// Reading, rewriting and deleting a rule all carry the account predicate, so a known id
// from elsewhere is simply not found rather than acted on.
func TestIgnoreRuleReadsAndWritesAreScopedToTheAccount(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		mine := seedCashflowAccount(t, db, "mine@example.com")
		theirs := seedCashflowAccount(t, db, "theirs@example.com")
		theirRule := seedIgnoreRule(t, db, theirs, "Their rule", "Their transfer")

		store := storage.NewSQLXCashflowIgnoreRuleStore(db)

		found, err := store.GetIgnoreRule(t.Context(), mine, theirRule.ID)
		if err != nil {
			t.Fatalf("fetch ignore rule: %v", err)
		}
		if found != nil {
			t.Fatal("another account's rule was readable")
		}

		hijacked := *theirRule
		hijacked.AccountID = mine
		hijacked.Name = "hijacked"
		updated, err := store.UpdateIgnoreRule(t.Context(), &hijacked)
		if err != nil {
			t.Fatalf("update ignore rule: %v", err)
		}
		if updated != 0 {
			t.Fatalf("updated = %d, want 0 -- another account's rule was rewritten", updated)
		}

		deleted, err := store.DeleteIgnoreRule(t.Context(), mine, theirRule.ID)
		if err != nil {
			t.Fatalf("delete ignore rule: %v", err)
		}
		if deleted != 0 {
			t.Fatalf("deleted = %d, want 0 -- another account's rule was removed", deleted)
		}

		mineList, err := store.ListIgnoreRules(t.Context(), mine)
		if err != nil {
			t.Fatalf("list ignore rules: %v", err)
		}
		if len(mineList) != 0 {
			t.Fatalf("listed %d rules, want none of another account's", len(mineList))
		}
	})
}

// The import review reads its groups from here: one per rule that caught something in
// that import, carrying the full total alongside the capped sample.
func TestListIgnoredByRuleGroupsOneImport(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		accountID := seedCashflowAccount(t, db, "mine@example.com")
		card := seedIgnoreRule(t, db, accountID, "Credit card payment", "Credit card")
		savings := seedIgnoreRule(t, db, accountID, "Transfer to savings", "Transfer to savings")
		thisImport := seedImport(t, db, accountID, "ING")
		otherImport := seedImport(t, db, accountID, "N26")

		for i := range 3 {
			seedRuleTransaction(t, db, accountID, fmt.Sprintf("Credit card payment %d", i), false, false, &thisImport)
		}
		seedRuleTransaction(t, db, accountID, "Transfer to savings", false, false, &thisImport)
		seedRuleTransaction(t, db, accountID, "Credit card payment elsewhere", false, false, &otherImport)

		store := storage.NewSQLXCashflowIgnoreRuleStore(db)
		for _, rule := range []*cashflow.IgnoreRule{card, savings} {
			if _, err := store.ApplyIgnoreRule(t.Context(), rule, &thisImport); err != nil {
				t.Fatalf("apply ignore rule: %v", err)
			}
		}

		groups, err := store.ListIgnoredByRule(t.Context(), accountID, thisImport, 2)
		if err != nil {
			t.Fatalf("list ignored by rule: %v", err)
		}
		if len(groups) != 2 {
			t.Fatalf("got %d groups, want one per rule that caught something", len(groups))
		}

		totals := map[uuid.UUID]int{}
		for _, group := range groups {
			totals[group.Rule.ID] = group.Total
			if len(group.Transactions) > 2 {
				t.Fatalf("group %q returned %d rows, want at most the cap of 2", group.Rule.Name, len(group.Transactions))
			}
		}
		// The row in the other import must not be counted here, or the review would
		// report rows the person never imported in this run.
		if totals[card.ID] != 3 {
			t.Fatalf("card group total = %d, want 3", totals[card.ID])
		}
		if totals[savings.ID] != 1 {
			t.Fatalf("savings group total = %d, want 1", totals[savings.ID])
		}
	})
}
