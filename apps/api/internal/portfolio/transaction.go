package portfolio

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"iter"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

type TransactionType string
type TransactionOrigin string

const (
	TxBuy      TransactionType = "BUY"
	TxSell     TransactionType = "SELL"
	TxDividend TransactionType = "DIVIDEND"
	TxTax      TransactionType = "TAX"
	TxFee      TransactionType = "FEE"
	TxCash     TransactionType = "CASH"
	// TxSplit is a synthetic event, never a stored row: the rebuild injects one per
	// share split so the split lands in the same chronological stream as the trades
	// it affects. Quantity carries the multiplier (4 means one share became four).
	// It is deliberately absent from the transactions table's type constraint --
	// nothing persists it, and it carries no PositionID so the transaction-to-position
	// mapping never sees it.
	TxSplit TransactionType = "SPLIT"

	TransactionOriginImport TransactionOrigin = "IMPORT"
	TransactionOriginManual TransactionOrigin = "MANUAL"
)

type Transaction struct {
	ID          uuid.UUID         `db:"id"`
	AccountID   *uuid.UUID        `db:"account_id"`
	ImportID    *uuid.UUID        `db:"import_id"`
	Origin      TransactionOrigin `db:"origin"`
	Source      string            `db:"source"` // "degiro", "ibkr", ...
	OccurredAt  time.Time         `db:"occurred_at"`
	PositionID  *uuid.UUID        `db:"position_id"` // links to the position this transaction belongs to (can be null for cash transactions or if position mapping failed)
	ISIN        *string           `db:"isin"`
	Symbol      *string           `db:"symbol"`
	Description string            `db:"description"` // optional raw description from the broker, can be used for debugging or more complex parsing if needed
	Type        TransactionType   `db:"type"`
	Quantity    float64           `db:"quantity"`
	UnitPrice   money.Price       `db:"unit_price"`   // per-unit for BUY/SELL
	AmountCents money.Price       `db:"amount_cents"` // absolute amount for cash-impact rows
	Checksum    string            `db:"checksum"`
	RowNumber   int               `db:"row_number"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at"`
}

type TransactionWithListingID struct {
	Transaction
	ListingID *uuid.UUID `db:"listing_id"`
}

type CsvParser interface {
	ParseAll(rc io.ReadCloser) (iter.Seq2[int, TransactionData], error)
}

type TransactionData struct {
	Source      string
	OccurredAt  time.Time
	ValueDate   *time.Time
	ISIN        *string
	Symbol      *string
	Description string
	Type        TransactionType
	Quantity    float64
	Price       float64
	Amount      float64
	// RowNumber is the source row number (e.g. CSV line). It orders rows that share a
	// day; it deliberately no longer feeds the dedup checksum, because an overlapping
	// export moves every row to a different line.
	RowNumber int
	// DedupSeq distinguishes rows with identical content inside one import file and
	// feeds the dedup checksum in RowNumber's place. Importers set it from
	// importer.DedupSequencer; manual entries leave it zero.
	DedupSeq int
}

// checksumDateLayout is the day precision both the checksum and the dedup key work at.
const checksumDateLayout = "20060102"

// DedupKey is the content generateChecksum digests, normalised exactly as newTransaction
// normalises it, minus the values that are constant within one import (account, position,
// origin) and the sequence the key is used to produce.
//
// The importer sequences rows on this key, so the two have to stay in step. A key that
// separates rows the checksum cannot separate gives both of them sequence 1, so they end
// up with the same checksum and the bulk insert silently drops one as a duplicate.
func (d TransactionData) DedupKey() []string {
	unitPrice, _ := d.unitPrice()
	amount, _ := d.amountCents()
	return []string{
		strings.TrimSpace(d.Source),
		d.OccurredAt.Format(checksumDateLayout),
		trimOptional(d.ISIN),
		trimOptional(d.Symbol),
		string(d.Type),
		fmt.Sprintf("%.8f", d.quantity()),
		fmt.Sprintf("%d", unitPrice),
		fmt.Sprintf("%d", amount),
	}
}

// quantity is the quantity the stored transaction carries. A cash row has no units, so it
// records only the direction of its amount.
func (d TransactionData) quantity() float64 {
	if d.Type != TxCash {
		return d.Quantity
	}
	switch {
	case d.Amount < 0:
		return -1
	case d.Amount > 0:
		return 1
	default:
		return 0
	}
}

func (d TransactionData) unitPrice() (money.Price, error) {
	return money.NewPrice(math.Abs(d.Price))
}

func (d TransactionData) amountCents() (money.Price, error) {
	return money.NewPrice(math.Abs(d.Amount))
}

func trimOptional(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

var (
	ErrDuplicateTransaction            = fmt.Errorf("duplicate transaction")
	ErrTransactionISINAndSymbolMissing = fmt.Errorf("both ISIN and Symbol are missing, cannot determine position ID")
	ErrInvalidTransactionOrigin        = fmt.Errorf("invalid transaction origin")
)

func NewTransaction(data TransactionData, rowNumber int, importID uuid.UUID, accountID, positionID *uuid.UUID) (*Transaction, error) {
	imp := importID
	return newTransaction(data, rowNumber, &imp, accountID, positionID, TransactionOriginImport)
}

func NewManualTransaction(data TransactionData, accountID uuid.UUID) (*Transaction, error) {
	acc := accountID
	return newTransaction(data, 0, nil, &acc, nil, TransactionOriginManual)
}

func newTransaction(
	data TransactionData,
	rowNumber int,
	importID *uuid.UUID,
	accountID, positionID *uuid.UUID,
	origin TransactionOrigin,
) (*Transaction, error) {
	if origin != TransactionOriginImport && origin != TransactionOriginManual {
		return nil, ErrInvalidTransactionOrigin
	}
	if origin == TransactionOriginImport && importID == nil {
		return nil, ErrInvalidTransactionOrigin
	}
	if origin == TransactionOriginManual && importID != nil {
		return nil, ErrInvalidTransactionOrigin
	}

	price, err := data.unitPrice()
	if err != nil {
		return nil, fmt.Errorf("portfolio.NewTransaction price: %w", err)
	}
	amount, err := data.amountCents()
	if err != nil {
		return nil, fmt.Errorf("portfolio.NewTransaction amount: %w", err)
	}
	quantity := data.quantity()
	if data.ISIN == nil && data.Symbol == nil && data.Type != TxCash {
		return nil, ErrTransactionISINAndSymbolMissing
	}
	if data.ISIN != nil {
		isin := strings.TrimSpace(*data.ISIN)
		data.ISIN = &isin
	}
	if data.Symbol != nil {
		symbol := strings.TrimSpace(*data.Symbol)
		data.Symbol = &symbol
	}
	now := time.Now().UTC()
	tx := &Transaction{
		ID:          uuid.New(),
		AccountID:   accountID,
		ImportID:    importID,
		Origin:      origin,
		Source:      strings.TrimSpace(data.Source),
		OccurredAt:  data.OccurredAt,
		PositionID:  positionID,
		ISIN:        data.ISIN,
		Symbol:      data.Symbol,
		Description: data.Description,
		Type:        data.Type,
		Quantity:    quantity,
		UnitPrice:   price,
		AmountCents: amount,
		RowNumber:   rowNumber,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	// Manual entries carry no sequence; their row number kept rows apart before and
	// still does, so nothing about manual deduplication changes.
	dedupSeq := data.DedupSeq
	if dedupSeq == 0 {
		dedupSeq = rowNumber
	}
	tx.Checksum = tx.generateChecksum(dedupSeq)
	return tx, nil
}

// IsManual reports whether the transaction was entered by hand rather than imported.
// Only manual transactions may be moved to another date: the checksum that recognises a
// re-imported row carries its date, so moving an imported row would make the next import
// of the same file insert it again.
func (t *Transaction) IsManual() bool {
	return t.Origin == TransactionOriginManual
}

// MovedTo returns a copy of the transaction occurring on day, with the dedup checksum
// recomputed so its identity follows the new date. The position mapping is left out of
// the copy: a rebuild writes position_id onto the row long after its checksum was
// generated, so the identity to stay comparable with is the one it was created under.
func (t *Transaction) MovedTo(day time.Time) *Transaction {
	moved := *t
	moved.PositionID = nil
	moved.OccurredAt = day.UTC()
	moved.UpdatedAt = time.Now().UTC()
	moved.Checksum = moved.generateChecksum(moved.RowNumber)
	return &moved
}

// GetID returns the identifier for the transaction, which is either the ISIN or Symbol. This is used for mapping transactions to positions. If both ISIN and Symbol are missing, it returns an error.
// ISIN takes precedence over Symbol for the ID.
func (t *Transaction) GetID() (string, error) {
	if t.ISIN != nil {
		return *t.ISIN, nil
	} else if t.Symbol != nil {
		return *t.Symbol, nil
	}
	return "", ErrTransactionISINAndSymbolMissing
}

// generateChecksum digests what the row is, not where it sat in the file: dedupSeq
// separates identical rows within one import, so the same transaction in a partly
// overlapping export produces the same checksum and is recognised as already imported.
//
// TransactionData.DedupKey must digest the same fields; see its doc for what goes wrong
// when the two drift apart.
func (t *Transaction) generateChecksum(dedupSeq int) string {
	accountID := ""
	if t.AccountID != nil {
		accountID = t.AccountID.String()
	}
	positionID := ""
	if t.PositionID != nil {
		positionID = t.PositionID.String()
	}
	const sep = "\x1F"
	payload := strings.Join([]string{
		strings.TrimSpace(t.Source),
		t.OccurredAt.Format(checksumDateLayout),
		trimOptional(t.ISIN),
		trimOptional(t.Symbol),
		string(t.Type),
		fmt.Sprintf("%.8f", t.Quantity),
		fmt.Sprintf("%d", t.UnitPrice),
		fmt.Sprintf("%d", t.AmountCents),
		fmt.Sprintf("%d", dedupSeq),
		accountID,
		positionID,
		string(t.Origin),
	}, sep)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
