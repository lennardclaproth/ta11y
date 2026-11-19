package parsers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// IngParser parses ING CSV files and constructs Transactions.
type IngParser struct {
	headerToColumn map[string]int
	accountType    *cashflow.AccountType
}

func NewIngParser() *IngParser {
	acct := cashflow.AccountTypeChecking
	return &IngParser{
		headerToColumn: make(map[string]int),
		accountType:    &acct,
	}
}

func (p *IngParser) ParseAll(rc io.ReadCloser) (iter.Seq2[cashflow.TransactionData, error], error) {
	raw, err := io.ReadAll(rc)
	closeErr := rc.Close()
	if err != nil {
		if closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	// A UTF-8 byte order mark has to go before the CSV reader sees it: in front of a
	// quoted header it stops the quote from opening the field, and "Date" stays unmatched.
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	csvReader := csv.NewReader(bytes.NewReader(raw))
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	header, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%w: the file has no header row", ErrMissingHeader)
		}
		return nil, err
	}
	if err := p.parseHeader(header); err != nil {
		return nil, err
	}
	seq := func(yield func(cashflow.TransactionData, error) bool) {
		rowNumber := 1 // first data row after header
		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				return
			}
			var td cashflow.TransactionData
			if err == nil {
				td, err = p.ParseRow(record, rowNumber, uuid.Nil)
			}
			td.RowNumber = rowNumber
			if !yield(td, err) {
				return
			}
			rowNumber++
		}
	}

	return seq, nil
}

// ingRequiredHeaders are the columns of ING's "Download transactions" CSV a row is read
// from. Notifications is optional: an empty note is still a valid transaction.
var ingRequiredHeaders = []string{"Date", "Name / Description", "Debit/credit", "Amount (EUR)"}

// parseHeader maps the header row to column indexes and refuses a file that lacks any
// required column, which is what tells a wrong export apart from a damaged one.
func (p *IngParser) parseHeader(headers []string) error {
	p.headerToColumn = make(map[string]int, len(headers))
	for i, h := range headers {
		p.headerToColumn[strings.TrimSpace(h)] = i
	}
	for _, key := range ingRequiredHeaders {
		if _, ok := p.headerToColumn[key]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingHeader, key)
		}
	}
	return nil
}

// ParseRow parses a single ING CSV row into a TransactionData.
func (p *IngParser) ParseRow(record []string, rowNumber int, importId uuid.UUID) (cashflow.TransactionData, error) {
	if len(p.headerToColumn) == 0 {
		return cashflow.TransactionData{}, fmt.Errorf("header map not initialized")
	}
	// Extract fields
	dateStr := p.value(record, "Date")
	desc := p.value(record, "Name / Description")
	note := p.value(record, "Notifications")
	source := "ING"
	amountStr := p.value(record, "Amount (EUR)")
	directionRaw := p.value(record, "Debit/credit")

	// Parse amount (replace comma with dot)
	amountStr = strings.ReplaceAll(amountStr, ",", ".")
	var amount float64
	if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
	}

	// Parse direction
	var direction cashflow.CashFlowDirection
	if strings.EqualFold(directionRaw, "Debit") {
		direction = cashflow.CashOut
	} else if strings.EqualFold(directionRaw, "Credit") {
		direction = cashflow.CashIn
	} else {
		return cashflow.TransactionData{}, fmt.Errorf("invalid direction: %s", directionRaw)
	}

	// Parse date
	parsedDate, err := time.Parse("20060102", dateStr)
	if err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid date: %w", err)
	}
	price, err := money.NewPrice(math.Abs(amount))
	if err != nil {
		return cashflow.TransactionData{}, fmt.Errorf("invalid amount: %w", err)
	}
	return cashflow.TransactionData{
		Description: desc,
		Note:        note,
		Source:      source,
		Direction:   direction,
		Amount:      price,
		Date:        parsedDate,
		AccountType: p.accountType, // default to checking for ING
	}, nil
}

// value reads one column of a row, empty when the column is absent. ParseAll never hands
// it a short row — the CSV reader reports that as a field-count error — but ParseRow can
// also be called directly.
func (p *IngParser) value(record []string, key string) string {
	idx, ok := p.headerToColumn[key]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}
