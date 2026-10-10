package cashflow

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newImportedTransaction(t *testing.T, rowNumber, dedupSeq int) *Transaction {
	t.Helper()
	importID := uuid.New()
	tx, err := NewTransaction(
		"Example payment",
		"Example Product | OrderID:example-order",
		"DEGIRO",
		"",
		CashOut,
		1000,
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		rowNumber,
		dedupSeq,
		&importID,
		nil,
		uuid.Nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	return tx
}

func TestNewTransaction_ChecksumIgnoresRowNumber(t *testing.T) {
	// The same row sits on a different line in next month's export; recognising it as
	// already imported is the whole point of deduplicating on content.
	first := newImportedTransaction(t, 12, 1)
	second := newImportedTransaction(t, 47, 1)

	if first.Checksum != second.Checksum {
		t.Fatalf("expected the same row to keep its checksum across exports")
	}
	if second.RowNumber != 47 {
		t.Fatalf("expected the source row number to be preserved, got %d", second.RowNumber)
	}
}

func TestNewTransaction_ChecksumSeparatesRepeatsWithinOneFile(t *testing.T) {
	first := newImportedTransaction(t, 12, 1)
	second := newImportedTransaction(t, 13, 2)

	if first.Checksum == second.Checksum {
		t.Fatalf("expected two identical rows in one export to stay two transactions")
	}
}

func TestTransactionData_DedupKeyDigestsTheChecksumFields(t *testing.T) {
	// The importer sequences rows on DedupKey, so the key may not separate rows the
	// checksum cannot separate: those rows would both get sequence 1, share a checksum
	// and lose one of the two to the duplicate skip in the bulk insert.
	base := TransactionData{
		Description: "Example payment",
		Note:        "Example Product | OrderID:example-order",
		Source:      "DEGIRO",
		Direction:   CashOut,
		Amount:      1000,
		Date:        time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Tag:         "groceries",
	}

	cases := []struct {
		name   string
		mutate func(*TransactionData)
	}{
		{"description", func(d *TransactionData) { d.Description = "Other payment" }},
		{"note", func(d *TransactionData) { d.Note = "Example Product | OrderID:other-order" }},
		{"source", func(d *TransactionData) { d.Source = "ING" }},
		{"direction", func(d *TransactionData) { d.Direction = CashIn }},
		{"amount", func(d *TransactionData) { d.Amount = 1001 }},
		{"date", func(d *TransactionData) { d.Date = d.Date.AddDate(0, 0, 1) }},
		{"tag", func(d *TransactionData) { d.Tag = "travel" }},
		{"row number", func(d *TransactionData) { d.RowNumber = 47 }},
		{"untrimmed description", func(d *TransactionData) { d.Description = "  Example payment  " }},
		{"untrimmed note", func(d *TransactionData) { d.Note = " Example Product | OrderID:example-order " }},
	}

	accountID := uuid.New()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := base
			tc.mutate(&mutated)

			first := checksumOf(t, base, accountID)
			second := checksumOf(t, mutated, accountID)

			sameKey := strings.Join(base.DedupKey(), "\x1F") == strings.Join(mutated.DedupKey(), "\x1F")
			if sameKey != (first == second) {
				t.Fatalf("dedup key and checksum disagree: same key %t, same checksum %t", sameKey, first == second)
			}
		})
	}
}

func checksumOf(t *testing.T, data TransactionData, accountID uuid.UUID) string {
	t.Helper()
	importID := uuid.New()
	tx, err := NewTransaction(
		data.Description,
		data.Note,
		data.Source,
		data.Tag,
		data.Direction,
		data.Amount,
		data.Date,
		data.RowNumber,
		1,
		&importID,
		data.AccountType,
		accountID,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	return tx.Checksum
}
