package portfolio

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewTransaction_CashEncodesDirectionInQuantity(t *testing.T) {
	importID := uuid.New()
	tx, err := NewTransaction(
		TransactionData{
			Source:     "TEST",
			OccurredAt: time.Now().UTC(),
			ISIN:       ptrString("NLTEST0001"),
			Type:       TxCash,
			Amount:     -50,
		},
		1,
		importID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.AmountCents <= 0 {
		t.Fatalf("expected absolute positive amount for cash transaction, got %d", tx.AmountCents)
	}
	if tx.Quantity != -1 {
		t.Fatalf("expected cash quantity sign marker -1, got %f", tx.Quantity)
	}
}

func TestNewTransaction_NonCashUsesAbsoluteAmount(t *testing.T) {
	importID := uuid.New()
	tx, err := NewTransaction(
		TransactionData{
			Source:     "TEST",
			OccurredAt: time.Now().UTC(),
			ISIN:       ptrString("NLTEST0002"),
			Type:       TxDividend,
			Amount:     -2,
		},
		1,
		importID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.AmountCents <= 0 {
		t.Fatalf("expected absolute positive amount for non-cash transaction, got %d", tx.AmountCents)
	}
	if tx.Quantity != 0 {
		t.Fatalf("expected non-cash quantity to remain unchanged, got %f", tx.Quantity)
	}
}

func TestNewTransaction_ChecksumIgnoresRowNumber(t *testing.T) {
	// The same trade sits on a different line in next month's export. Its checksum has
	// to stay the same, or a partly overlapping export imports it twice.
	importID := uuid.New()
	data := TransactionData{
		Source:      "DEGIRO",
		OccurredAt:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		ISIN:        ptrString("NLTEST0003"),
		Description: "Example Fund",
		Type:        TxBuy,
		Quantity:    2,
		Price:       10,
		Amount:      20,
		DedupSeq:    1,
	}

	data.RowNumber = 12
	first, err := NewTransaction(data, data.RowNumber, importID, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data.RowNumber = 47
	second, err := NewTransaction(data, data.RowNumber, importID, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if first.Checksum != second.Checksum {
		t.Fatalf("expected the same row to keep its checksum across exports")
	}
	if second.RowNumber != 47 {
		t.Fatalf("expected the source row number to be preserved, got %d", second.RowNumber)
	}
}

func TestNewTransaction_ChecksumSeparatesRepeatsWithinOneFile(t *testing.T) {
	importID := uuid.New()
	data := TransactionData{
		Source:      "DEGIRO",
		OccurredAt:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		ISIN:        ptrString("NLTEST0004"),
		Description: "Example Fund",
		Type:        TxBuy,
		Quantity:    2,
		Price:       10,
		Amount:      20,
		DedupSeq:    1,
	}

	first, err := NewTransaction(data, 1, importID, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data.DedupSeq = 2
	second, err := NewTransaction(data, 2, importID, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if first.Checksum == second.Checksum {
		t.Fatalf("expected two identical rows in one export to stay two transactions")
	}
}

func ptrString(v string) *string {
	return &v
}
