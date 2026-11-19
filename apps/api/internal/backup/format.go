// Package backup writes an account's records to JSON files, one per domain — cashflow,
// portfolio, assets — and merges such files back in, and does the same for the shared
// end-of-day market data. It is what lets the work put into a ledger survive a broken
// database or a move to another instance.
//
// A file carries source records only. What ta11y derives — positions, snapshots, the
// portfolio-sourced asset class and the mutations of daily-priced items — is rebuilt
// after a restore rather than copied, so it can never disagree with its sources.
package backup

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

const (
	// FormatName identifies a ta11y backup file.
	FormatName = "ta11y.backup"
	// FormatVersion is the version this build writes and the only one it reads.
	FormatVersion = 1
)

// Kind says what a backup file holds.
type Kind string

const (
	// KindCashflow is an account's ledger with its ignore rules and recurring items.
	KindCashflow Kind = "cashflow"
	// KindPortfolio is an account's portfolio transactions.
	KindPortfolio Kind = "portfolio"
	// KindAssets is an account's hand-kept asset classes.
	KindAssets Kind = "assets"
	// KindEOD is the instance's end-of-day market data.
	KindEOD Kind = "eods"
)

// Header opens every backup file.
type Header struct {
	Format     string    `json:"format"`
	Version    int       `json:"version"`
	Kind       Kind      `json:"kind"`
	ExportedAt time.Time `json:"exported_at"`
}

func newHeader(kind Kind, now time.Time) Header {
	return Header{Format: FormatName, Version: FormatVersion, Kind: kind, ExportedAt: now.UTC()}
}

func (h Header) check(kind Kind) error {
	if h.Format != FormatName {
		return fmt.Errorf("%w: not a ta11y backup", ErrInvalidFile)
	}
	if h.Version != FormatVersion {
		return fmt.Errorf("%w: version %d", ErrUnsupportedVersion, h.Version)
	}
	if h.Kind != kind {
		return fmt.Errorf("%w: holds %q, expected %q", ErrWrongKind, h.Kind, kind)
	}
	return nil
}

// Day is a calendar date, written as YYYY-MM-DD.
type Day struct{ time.Time }

const dayLayout = "2006-01-02"

// MarshalJSON writes the date without a time or zone.
func (d Day) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.UTC().Format(dayLayout))
}

// UnmarshalJSON reads a YYYY-MM-DD date as midnight UTC.
func (d *Day) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	t, err := time.Parse(dayLayout, raw)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func dayOf(t time.Time) Day { return Day{t.UTC()} }

func dayPtr(t *time.Time) *Day {
	if t == nil {
		return nil
	}
	d := dayOf(*t)
	return &d
}

func (d *Day) timePtr() *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time
	return &t
}

// ListingRef names a listing the way two instances agree on it: ids differ, the
// symbol and its source do not.
type ListingRef struct {
	Symbol string            `json:"symbol"`
	Source marketdata.Source `json:"source"`
}

// CashflowFile is an account's ledger: every transaction with its tag and ignored
// state, the ignore rules and the recurring items. Records point at each other with
// file-local refs, never with ids of the instance they came from.
type CashflowFile struct {
	Header
	// SourceAccountID is the account the file was exported from. Dedup checksums
	// digest the account, so it is what lets a restore into another account re-derive
	// them.
	SourceAccountID uuid.UUID             `json:"source_account_id"`
	Imports         []Import              `json:"imports"`
	Transactions    []CashflowTransaction `json:"transactions"`
	IgnoreRules     []IgnoreRule          `json:"ignore_rules"`
	RecurringItems  []RecurringItem       `json:"recurring_items"`
	Dismissals      []Dismissal           `json:"recurring_dismissals"`
}

// PortfolioFile is an account's portfolio transactions; positions and snapshots are
// rebuilt from them.
type PortfolioFile struct {
	Header
	// SourceAccountID is the account the file was exported from; see CashflowFile.
	SourceAccountID uuid.UUID              `json:"source_account_id"`
	Imports         []Import               `json:"imports"`
	Transactions    []PortfolioTransaction `json:"transactions"`
}

// AssetsFile is an account's hand-kept asset classes; the portfolio-sourced class is
// rebuilt from the portfolio.
type AssetsFile struct {
	Header
	Classes []AssetClass `json:"classes"`
}

// Import is the record of a CSV import that brought rows in. The uploaded file is not
// kept after processing, so neither is it here.
type Import struct {
	Ref         string                `json:"ref"`
	Vendor      string                `json:"vendor"`
	Type        importer.ImportType   `json:"type"`
	Source      string                `json:"source"`
	Status      importer.ImportStatus `json:"status"`
	StatusMsg   string                `json:"status_msg,omitempty"`
	TotalRows   int                   `json:"total_rows"`
	Imported    int                   `json:"imported"`
	Duplicates  int                   `json:"duplicates"`
	Failed      int                   `json:"failed"`
	AutoIgnored int                   `json:"auto_ignored"`
	CreatedAt   time.Time             `json:"created_at"`
}

// CashflowTransaction is one ledger row.
type CashflowTransaction struct {
	Ref              string                     `json:"ref"`
	ImportRef        *string                    `json:"import_ref,omitempty"`
	Description      string                     `json:"description"`
	Note             string                     `json:"note"`
	Source           string                     `json:"source"`
	AmountCents      money.Price                `json:"amount_cents"`
	Direction        cashflow.CashFlowDirection `json:"direction"`
	Date             Day                        `json:"date" swaggertype:"string" format:"date"`
	Tag              string                     `json:"tag"`
	AccountType      *cashflow.AccountType      `json:"account_type,omitempty"`
	Ignored          bool                       `json:"ignored"`
	IgnoredByRuleRef *string                    `json:"ignored_by_rule_ref,omitempty"`
	IgnoreOverridden bool                       `json:"ignore_overridden"`
	RowNumber        int                        `json:"row_number"`
	Checksum         string                     `json:"checksum"`
	CreatedAt        time.Time                  `json:"created_at"`
}

// IgnoreRule is one rule, with what it has done so far.
type IgnoreRule struct {
	Ref           string                      `json:"ref"`
	Name          string                      `json:"name"`
	MatchField    cashflow.MatchField         `json:"match_field"`
	Contains      string                      `json:"contains"`
	Direction     *cashflow.CashFlowDirection `json:"direction,omitempty"`
	Source        string                      `json:"source"`
	Enabled       bool                        `json:"enabled"`
	IgnoredTotal  int                         `json:"ignored_total"`
	LastAppliedAt *time.Time                  `json:"last_applied_at,omitempty"`
	CreatedAt     time.Time                   `json:"created_at"`
}

// RecurringItem is one recurring item and the ledger rows linked to it.
type RecurringItem struct {
	Name         string                     `json:"name"`
	Direction    cashflow.CashFlowDirection `json:"direction"`
	Rhythm       cashflow.Rhythm            `json:"rhythm"`
	MatchKey     string                     `json:"match_key"`
	EndedFrom    *Day                       `json:"ended_from,omitempty" swaggertype:"string" format:"date"`
	CreatedAt    time.Time                  `json:"created_at"`
	Transactions []string                   `json:"transaction_refs"`
}

// Dismissal is a recurring suggestion the user refused.
type Dismissal struct {
	MatchKey  string                     `json:"match_key"`
	Direction cashflow.CashFlowDirection `json:"direction"`
}

// PortfolioTransaction is one trade, dividend, fee, tax or cash movement.
type PortfolioTransaction struct {
	Ref         string                      `json:"ref"`
	ImportRef   *string                     `json:"import_ref,omitempty"`
	Origin      portfolio.TransactionOrigin `json:"origin"`
	Source      string                      `json:"source"`
	OccurredAt  Day                         `json:"occurred_at" swaggertype:"string" format:"date"`
	ISIN        *string                     `json:"isin,omitempty"`
	Symbol      *string                     `json:"symbol,omitempty"`
	Description string                      `json:"description"`
	Type        portfolio.TransactionType   `json:"type"`
	Quantity    float64                     `json:"quantity"`
	UnitPrice   money.Price                 `json:"unit_price_cents"`
	AmountCents money.Price                 `json:"amount_cents"`
	RowNumber   int                         `json:"row_number"`
	Checksum    string                      `json:"checksum"`
	CreatedAt   time.Time                   `json:"created_at"`
}

// AssetClass is one class with its items.
type AssetClass struct {
	Name      string      `json:"name"`
	Archived  bool        `json:"archived"`
	CreatedAt time.Time   `json:"created_at"`
	Items     []AssetItem `json:"items"`
}

// AssetItem is one item with its worth history. A daily-priced item carries its
// listing and purchases; its mutations are derived from them and left out.
type AssetItem struct {
	Name         string          `json:"name"`
	Listing      *ListingRef     `json:"listing,omitempty"`
	CurrentWorth money.Price     `json:"current_worth_cents"`
	Archived     bool            `json:"archived"`
	CreatedAt    time.Time       `json:"created_at"`
	Mutations    []AssetMutation `json:"mutations"`
	Purchases    []AssetPurchase `json:"purchases"`
}

// AssetMutation is one change of an item's worth, with the running totals it was
// written with.
type AssetMutation struct {
	ChangeType      assets.ChangeType       `json:"change_type"`
	Direction       *assets.ChangeDirection `json:"direction,omitempty"`
	Amount          money.Price             `json:"amount_cents"`
	PreviousWorth   money.Price             `json:"previous_worth_cents"`
	NewWorth        money.Price             `json:"new_worth_cents"`
	ClassTotalWorth money.Price             `json:"class_total_worth_cents"`
	EffectiveDate   Day                     `json:"effective_date" swaggertype:"string" format:"date"`
	Note            *string                 `json:"note,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
}

// AssetPurchase is one recorded purchase of a daily-priced item.
type AssetPurchase struct {
	PurchasedOn Day         `json:"purchased_on" swaggertype:"string" format:"date"`
	Quantity    float64     `json:"quantity"`
	UnitPrice   money.Price `json:"unit_price_cents"`
	CreatedAt   time.Time   `json:"created_at"`
}

// EODFile is the instance's end-of-day prices, per listing.
type EODFile struct {
	Header
	Listings []EODListing `json:"listings"`
}

// EODListing is one listing's prices.
type EODListing struct {
	ListingRef
	Name string `json:"name"`
	EODs []EOD  `json:"eods"`
}

// EOD is one day's prices.
type EOD struct {
	Date        Day         `json:"date" swaggertype:"string" format:"date"`
	Open        money.Price `json:"open_cents"`
	High        money.Price `json:"high_cents"`
	Low         money.Price `json:"low_cents"`
	Close       money.Price `json:"close_cents"`
	Volume      int64       `json:"volume"`
	SplitFactor float64     `json:"split_factor"`
}
