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

// Purpose is what a transaction counts as towards the monthly wealth-growth goal: the
// income the goal is a share of, the money put towards wealth, or nothing at all. It is
// independent of the tag and of the ignored flag -- a transfer to a savings account is
// usually ignored for the cashflow totals and is still a contribution.
type Purpose string

const (
	// PurposeNone is the state of a transaction nobody has pointed at yet.
	PurposeNone Purpose = ""
	// PurposeIncome marks incoming money as part of the month's income.
	PurposeIncome Purpose = "income"
	// PurposeWealth marks outgoing money as a contribution towards wealth.
	PurposeWealth Purpose = "wealth"
)

// RequiredDirection reports the direction a purpose can be applied to. Only incoming money
// can be income and only outgoing money can be a contribution: a transfer between your own
// accounts leaves both sides in the ledger, and counting them both would double the month.
func (p Purpose) RequiredDirection() *CashFlowDirection {
	switch p {
	case PurposeIncome:
		direction := CashIn
		return &direction
	case PurposeWealth:
		direction := CashOut
		return &direction
	default:
		return nil
	}
}

// ParsePurpose reads a purpose from client input. "none" and the empty string both mean
// "not assigned", which is also what clearing a purpose writes.
func ParsePurpose(raw string) (Purpose, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "none":
		return PurposeNone, nil
	case string(PurposeIncome):
		return PurposeIncome, nil
	case string(PurposeWealth):
		return PurposeWealth, nil
	default:
		return "", ErrInvalidPurpose
	}
}

// SplitPurposes reads a comma-separated purpose filter, e.g. "income,none". An empty input
// yields no filter rather than a filter on "not assigned".
func SplitPurposes(raw string) ([]Purpose, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	out := make([]Purpose, 0, 3)
	seen := make(map[Purpose]struct{}, 3)
	for _, entry := range strings.Split(raw, ",") {
		if strings.TrimSpace(entry) == "" {
			continue
		}
		purpose, err := ParsePurpose(entry)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[purpose]; ok {
			continue
		}
		seen[purpose] = struct{}{}
		out = append(out, purpose)
	}
	return out, nil
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
	Purpose     Purpose           `db:"purpose"`
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
func NewTransaction(desc, note, source, tag string, direction CashFlowDirection, amount money.Price, date time.Time, rowNumber int, importID *uuid.UUID, accountType *AccountType, accID uuid.UUID) (*Transaction, error) {
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
	t.Checksum = t.generateChecksum()
	return t, nil
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
	moved.Checksum = moved.generateChecksum()
	return &moved
}

// generateChecksum creates a checksum for the transaction based on the fields
// description, note, source, amountCents, and date. It uses amountCents instead
// of amount to avoid floating-point precision issues.
func (t *Transaction) generateChecksum() string {
	// initialize fields to be used in checksum generation, these fields need to be
	// of type string
	desc := strings.TrimSpace(t.Description)
	note := strings.TrimSpace(t.Note)
	source := strings.TrimSpace(t.Source)
	direction := string(t.Direction)
	amountCents := fmt.Sprintf("%d", t.AmountCents)
	rowNumber := fmt.Sprintf("%d", t.RowNumber)
	date := t.Date.Format("20060102") // Standard date format
	accountID := ""
	if t.AccountID != uuid.Nil {
		accountID = t.AccountID.String()
	}
	// concatenate all fields to form the payload string to generate a checksum
	const sep = "\x1F" // Unit Separator character see -> https://www.ascii-code.com/character/%E2%90%9F
	payload := strings.Join([]string{desc, note, source, direction, amountCents, date, rowNumber, accountID}, sep)
	// digest the payload in byte format and encode it to hexadecimal string
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
