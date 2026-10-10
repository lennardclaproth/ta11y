package parsers

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/lennardclaproth/ta11y/internal/portfolio"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

func TestDegiroParser_ParseAll_HappyPath(t *testing.T) {
	input := `Date,Time,Value date,Product,ISIN,Description,FX,Change,,Balance,,Order Id
19-01-2026,11:30,19-01-2026,VANGUARD S&P 500 UCITS ETF USD DIS,IE00B3XXRP09,"Koop 3 @ 111,962 EUR",,EUR,"-335,89",EUR,"792,56",47509942-dc4e-4d27-8806-903428c8f88a
19-01-2026,11:30,19-01-2026,VANECK AEX UCITS ETF,NL0009272749,DEGIRO Transactiekosten en/of kosten van derden,,EUR,"-3,00",EUR,"487,43",6c44d01d-ca3a-4786-a432-b1fe0f19be4b
`
	p := NewDegiroParser()
	seq, err := p.ParseAll(io.NopCloser(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("ParseAll returned error: %v", err)
	}

	var got []portfolio.TransactionData
	for _, txd := range seq {
		got = append(got, txd)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
	if got[0].Type != portfolio.TxBuy {
		t.Fatalf("expected first row to be BUY, got %s", got[0].Type)
	}
	if got[0].Quantity != 3 {
		t.Fatalf("expected quantity 3, got %f", got[0].Quantity)
	}
	if got[0].Price != 111.962 {
		t.Fatalf("expected price 111.962, got %f", got[0].Price)
	}
	if got[1].Type != portfolio.TxFee {
		t.Fatalf("expected second row to be FEE, got %s", got[1].Type)
	}
}

func TestDegiroParser_ParseAll_WrongHeadersAreRecognisableAsSuch(t *testing.T) {
	// The importer classifies on this sentinel: it is what separates "that is not the
	// DEGIRO export" from a read failure the user should simply retry.
	cases := map[string]string{
		"other export": "Date,Name / Description,Amount (EUR)\n01-09-2026,Groceries,\"-12,00\"\n",
		"empty file":   "",
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := NewDegiroParser().ParseAll(io.NopCloser(strings.NewReader(input)))
			if !errors.Is(err, ErrMissingHeader) {
				t.Fatalf("expected ErrMissingHeader, got %v", err)
			}
		})
	}
}

func TestDegiroParser_ParseAll_ReadErrorIsNotAHeaderMismatch(t *testing.T) {
	_, err := NewDegiroParser().ParseAll(io.NopCloser(failingReader{}))
	if err == nil {
		t.Fatalf("expected an error")
	}
	if errors.Is(err, ErrMissingHeader) {
		t.Fatalf("expected a read failure not to be reported as the wrong file, got %v", err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("disk is having a day")
}

func TestCreateCsvParser_Degiro(t *testing.T) {
	p, err := CreateCsvParser(vendor.VendorDEGIRO)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatalf("expected parser instance")
	}
}
