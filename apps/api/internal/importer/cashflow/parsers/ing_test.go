package parsers

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

const ingHeader = `"Date";"Name / Description";"Account";"Counterparty";"Code";"Debit/credit";"Amount (EUR)";"Transaction type";"Notifications"`

func TestIngParser_ParseAll_HappyPath(t *testing.T) {
	input := "\ufeff" + ingHeader + `
"20250621";"Albert Heijn 1234";"NL00INGB0000000001";"";"BA";"Debit";"23,45";"Payment terminal";"Pasvolgnr: 001"
"20250625";"Werkgever BV";"NL00INGB0000000001";"NL00RABO0000000002";"GT";"Credit";"2500,00";"Online banking";"Salaris juni"
`

	seq, err := NewIngParser().ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}

	var got []cashflow.TransactionData
	for txd, rowErr := range seq {
		if rowErr != nil {
			t.Fatalf("row %d: %v", txd.RowNumber, rowErr)
		}
		got = append(got, txd)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 parsed rows, got %d", len(got))
	}
	if got[0].Description != "Albert Heijn 1234" || got[0].Direction != cashflow.CashOut {
		t.Fatalf("unexpected row 1: %+v", got[0])
	}
	wantOut, _ := money.NewPrice(23.45)
	if got[0].Amount != wantOut {
		t.Fatalf("expected row 1 amount %v, got %v", wantOut, got[0].Amount)
	}
	if got[0].Note != "Pasvolgnr: 001" {
		t.Fatalf("unexpected row 1 note: %s", got[0].Note)
	}
	if got[1].Direction != cashflow.CashIn || got[1].Date.Format("2006-01-02") != "2025-06-25" {
		t.Fatalf("unexpected row 2: %+v", got[1])
	}
	if got[0].RowNumber != 1 || got[1].RowNumber != 2 {
		t.Fatalf("expected row numbers 1 and 2, got %d and %d", got[0].RowNumber, got[1].RowNumber)
	}
}

func TestIngParser_ParseAll_MissingRequiredHeader(t *testing.T) {
	input := `"Date";"Name / Description";"Amount (EUR)"
"20250621";"Albert Heijn 1234";"23,45"
`

	_, err := NewIngParser().ParseAll(io.NopCloser(strings.NewReader(input)))
	if !errors.Is(err, ErrMissingHeader) {
		t.Fatalf("expected ErrMissingHeader, got %v", err)
	}
}

func TestIngParser_ParseAll_OtherExportIsNotRecognised(t *testing.T) {
	input := `Date,Time,Value date,Product,ISIN,Description,FX,Change,,Balance,,Order Id
21-06-2025,09:00,21-06-2025,,,Deposit,,EUR,100.00,EUR,100.00,
`

	_, err := NewIngParser().ParseAll(io.NopCloser(strings.NewReader(input)))
	if !errors.Is(err, ErrMissingHeader) {
		t.Fatalf("expected ErrMissingHeader, got %v", err)
	}
}

func TestIngParser_ParseAll_EmptyFileIsNotRecognised(t *testing.T) {
	_, err := NewIngParser().ParseAll(io.NopCloser(strings.NewReader("")))
	if !errors.Is(err, ErrMissingHeader) {
		t.Fatalf("expected ErrMissingHeader, got %v", err)
	}
}

func TestIngParser_ParseAll_UnreadableRowIsYieldedWithItsError(t *testing.T) {
	input := ingHeader + `
"20250621";"Albert Heijn 1234";"";"";"BA";"Debit";"23,45";"Payment terminal";""
"not-a-date";"Broken row";"";"";"BA";"Debit";"1,00";"Payment terminal";""
"20250623";"Jumbo";"";"";"BA";"Sideways";"4,00";"Payment terminal";""
"20250624";"Etos";"";"";"BA";"Debit";"5,10";"Payment terminal";""
`

	seq, err := NewIngParser().ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}

	var parsed, failed []int
	for txd, rowErr := range seq {
		if rowErr != nil {
			failed = append(failed, txd.RowNumber)
			continue
		}
		parsed = append(parsed, txd.RowNumber)
	}

	if len(parsed) != 2 || parsed[0] != 1 || parsed[1] != 4 {
		t.Fatalf("expected rows 1 and 4 parsed, got %v", parsed)
	}
	if len(failed) != 2 || failed[0] != 2 || failed[1] != 3 {
		t.Fatalf("expected rows 2 and 3 failed, got %v", failed)
	}
}

func TestCreateCsvParser_ING(t *testing.T) {
	p, err := CreateCsvParser(vendor.VendorING)
	if err != nil {
		t.Fatalf("CreateCsvParser returned error: %v", err)
	}
	if _, ok := p.(*IngParser); !ok {
		t.Fatalf("expected *IngParser, got %T", p)
	}
}
