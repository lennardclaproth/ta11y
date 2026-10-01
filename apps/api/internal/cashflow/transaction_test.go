package cashflow

import (
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
