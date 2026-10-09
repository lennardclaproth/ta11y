package cashflow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"iter"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

type CashFlowDirection string

func ParseDirection(raw string) (*CashFlowDirection, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return nil, nil
	case "in", "income":
		direction := CashIn
		return &direction, nil
	case "out", "expense":
		direction := CashOut
		return &direction, nil
	default:
		return nil, fmt.Errorf("direction must be either in or out")
	}
}

func SplitTags(tags string) []string {
	if strings.TrimSpace(tags) == "" {
		return nil
	}

	raw := strings.Split(tags, ",")
	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, entry := range raw {
		tag := strings.TrimSpace(entry)
		if tag == "" {
			continue
		}

		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tag)
	}
	return out
}

type AccountType string

type CsvParser interface {
	ParseAll(rc io.ReadCloser) (iter.Seq2[int, TransactionData], error)
}

const (
	AccountTypeChecking  AccountType = "checking"
	AccountTypeSavings   AccountType = "savings"
	AccountTypeCredit    AccountType = "credit"
	AccountTypeBrokerage AccountType = "brokerage"
)

const (
	CashIn  CashFlowDirection = "in"
	CashOut CashFlowDirection = "out"
)

// SourceManual is the source of a transaction entered by hand. Entries made against a
// named vendor carry "manual:<vendor>"; imported rows carry the vendor name alone.
const SourceManual = "manual"

type Transaction struct {
	ID          uuid.UUID         `db:"id"`
	AccountID   uuid.UUID         `db:"account_id"`
	Description string            `db:"description"`
	Note        string            `db:"note"`
	Source      string            `db:"source"`
	AmountCents money.Price       `db:"amount_cents"`
	Direction   CashFlowDirection `db:"direction"`
	Date        time.Time         `db:"date"`
	Checksum    string            `db:"checksum"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at"`
	Tag         string            `db:"tag"`
	RowNumber   int               `db:"row_number"`
	Ignored     bool              `db:"ignored"`
	ImportID    *uuid.UUID        `db:"import_id"`
	AccountType *AccountType      `db:"account_type"` // Allow nullable account type.
}

var (
	ErrDuplicateTransaction = fmt.Errorf("duplicate transaction")
	ErrUnsupportedDirection = fmt.Errorf("unsupported direction")
	ErrNoTransactionFound   = fmt.Errorf("no transaction found with the given ID")
)

// NewTransaction creates a new Transaction instance and generates its checksum.
// rowNumber records where the row sat in its source file; dedupSeq separates rows
// with identical content inside one file and is what the checksum uses.
func NewTransaction(desc, note, source, tag string, direction CashFlowDirection, amount money.Price, date time.Time, rowNumber, dedupSeq int, importID *uuid.UUID, accountType *AccountType, accID uuid.UUID) (*Transaction, error) {
	t := &Transaction{
		ID:          uuid.New(),
		AccountID:   accID,
		Description: desc,
		Note:        note,
		Source:      source,
		Direction:   direction,
		AmountCents: amount,
		Date:        date,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		RowNumber:   rowNumber,
		ImportID:    importID,
		AccountType: accountType,
		Tag:         tag,
	}
	t.Checksum = t.generateChecksum(dedupSeq)
	return t, nil
}

// checksumDateLayout is the day precision both the checksum and the dedup key work at.
const checksumDateLayout = "20060102"

// DedupKey is the content generateChecksum digests, normalised the same way and minus
// the account, which is constant within one import, and the sequence the key is used to
// produce.
//
// The importer sequences rows on this key, so the two have to stay in step. A key that
// separates rows the checksum cannot separate gives both of them sequence 1, so they end
// up with the same checksum and the bulk insert silently drops one as a duplicate.
func (d TransactionData) DedupKey() []string {
	return []string{
		strings.TrimSpace(d.Description),
		strings.TrimSpace(d.Note),
		strings.TrimSpace(d.Source),
		string(d.Direction),
		fmt.Sprintf("%d", d.Amount),
		d.Date.Format(checksumDateLayout),
	}
}

// IsManual reports whether the transaction was entered by hand rather than imported.
// Only manual transactions may be moved to another date: the checksum that recognises
// a re-imported row carries its date, so moving an imported row would make the next
// import of the same statement insert it again.
func (t *Transaction) IsManual() bool {
	return t.Source == SourceManual || strings.HasPrefix(t.Source, SourceManual+":")
}

// MovedTo returns a copy of the transaction dated date, with the checksum recomputed so
// its dedup identity follows the new date.
func (t *Transaction) MovedTo(date time.Time) *Transaction {
	moved := *t
	moved.Date = date.UTC()
	moved.UpdatedAt = time.Now().UTC()
	moved.Checksum = moved.generateChecksum(moved.RowNumber)
	return &moved
}

// generateChecksum creates a checksum for the transaction based on the fields
// description, note, source, amountCents, and date. It uses amountCents instead
// of amount to avoid floating-point precision issues.
//
// It digests what the row is, not where it sat in the file: dedupSeq separates
// identical rows within one import, so the same transaction in a partly overlapping
// export produces the same checksum and is recognised as already imported.
//
// TransactionData.DedupKey must digest the same fields; see its doc for what goes wrong
// when the two drift apart.
func (t *Transaction) generateChecksum(dedupSeq int) string {
	// initialize fields to be used in checksum generation, these fields need to be
	// of type string
	desc := strings.TrimSpace(t.Description)
	note := strings.TrimSpace(t.Note)
	source := strings.TrimSpace(t.Source)
	direction := string(t.Direction)
	amountCents := fmt.Sprintf("%d", t.AmountCents)
	sequence := fmt.Sprintf("%d", dedupSeq)
	date := t.Date.Format(checksumDateLayout)
	accountID := ""
	if t.AccountID != uuid.Nil {
		accountID = t.AccountID.String()
	}
	// concatenate all fields to form the payload string to generate a checksum
	const sep = "\x1F" // Unit Separator character see -> https://www.ascii-code.com/character/%E2%90%9F
	payload := strings.Join([]string{desc, note, source, direction, amountCents, date, sequence, accountID}, sep)
	// digest the payload in byte format and encode it to hexadecimal string
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
