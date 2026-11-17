package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

// ImportRecord is an import with the name of the vendor it ran for.
type ImportRecord struct {
	importer.Import
	VendorName string `db:"vendor_name"`
}

// ItemRecord is an asset item with the listing it is priced from, if any.
type ItemRecord struct {
	assets.Asset
	ListingSymbol *string            `db:"listing_symbol"`
	ListingSource *marketdata.Source `db:"listing_source"`
}

// LinkRecord is one ledger row linked to a recurring item.
type LinkRecord struct {
	ItemID        uuid.UUID `db:"item_id"`
	TransactionID uuid.UUID `db:"transaction_id"`
}

// ListingRecord is a listing that holds end-of-day prices.
type ListingRecord struct {
	ID     uuid.UUID         `db:"id"`
	Symbol string            `db:"symbol"`
	Source marketdata.Source `db:"source"`
	Name   string            `db:"name"`
}

// AccountReader reads an account's source records, everything that is not derived.
type AccountReader interface {
	// AccountImports returns the account's imports of one type, with their vendor.
	AccountImports(ctx context.Context, accID uuid.UUID, importType importer.ImportType) ([]ImportRecord, error)
	CashflowTransactions(ctx context.Context, accID uuid.UUID) ([]*cashflow.Transaction, error)
	IgnoreRules(ctx context.Context, accID uuid.UUID) ([]*cashflow.IgnoreRule, error)
	RecurringItems(ctx context.Context, accID uuid.UUID) ([]*cashflow.RecurringItem, error)
	RecurringLinks(ctx context.Context, accID uuid.UUID) ([]LinkRecord, error)
	RecurringDismissals(ctx context.Context, accID uuid.UUID) ([]cashflow.DismissedSuggestion, error)
	PortfolioTransactions(ctx context.Context, accID uuid.UUID) ([]*portfolio.Transaction, error)
	// ManualAssetClasses leaves out the class the portfolio sync owns.
	ManualAssetClasses(ctx context.Context, accID uuid.UUID) ([]*assets.Class, error)
	AssetItems(ctx context.Context, accID uuid.UUID) ([]ItemRecord, error)
	// ManualAssetMutations leaves out the mutations of daily-priced items, which the
	// holdings sync derives from purchases and prices.
	ManualAssetMutations(ctx context.Context, accID uuid.UUID) ([]*assets.Mutation, error)
	AssetPurchases(ctx context.Context, accID uuid.UUID) ([]*assets.Purchase, error)
}

// EODReader reads the instance's end-of-day prices.
type EODReader interface {
	ListingsWithEODs(ctx context.Context) ([]ListingRecord, error)
	ListingEODs(ctx context.Context, listingID uuid.UUID) ([]*marketdata.EOD, error)
}

// Queries writes backup files.
type Queries struct {
	accounts AccountReader
	eods     EODReader
	now      func() time.Time
}

// NewQueries creates backup queries.
func NewQueries(accounts AccountReader, eods EODReader) *Queries {
	return &Queries{accounts: accounts, eods: eods, now: time.Now}
}

// ExportCashflow returns the account's ledger with its ignore rules and recurring items.
func (q *Queries) ExportCashflow(ctx context.Context, accID uuid.UUID) (*CashflowFile, error) {
	r := q.accounts
	imports, err := r.AccountImports(ctx, accID, importer.ImportTypeCashflow)
	if err != nil {
		return nil, fmt.Errorf("backup export: cashflow imports: %w", err)
	}
	txs, err := r.CashflowTransactions(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: cashflow transactions: %w", err)
	}
	rules, err := r.IgnoreRules(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: ignore rules: %w", err)
	}
	items, err := r.RecurringItems(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: recurring items: %w", err)
	}
	links, err := r.RecurringLinks(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: recurring links: %w", err)
	}
	dismissals, err := r.RecurringDismissals(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: recurring dismissals: %w", err)
	}
	return &CashflowFile{
		Header:          newHeader(KindCashflow, q.now()),
		SourceAccountID: accID,
		Imports:         exportImports(imports),
		Transactions:    exportCashflowTransactions(txs),
		IgnoreRules:     exportIgnoreRules(rules),
		RecurringItems:  exportRecurringItems(items, links),
		Dismissals:      exportDismissals(dismissals),
	}, nil
}

// ExportPortfolio returns the account's portfolio transactions.
func (q *Queries) ExportPortfolio(ctx context.Context, accID uuid.UUID) (*PortfolioFile, error) {
	imports, err := q.accounts.AccountImports(ctx, accID, importer.ImportTypePortfolio)
	if err != nil {
		return nil, fmt.Errorf("backup export: portfolio imports: %w", err)
	}
	txs, err := q.accounts.PortfolioTransactions(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: portfolio transactions: %w", err)
	}
	return &PortfolioFile{
		Header:          newHeader(KindPortfolio, q.now()),
		SourceAccountID: accID,
		Imports:         exportImports(imports),
		Transactions:    exportPortfolioTransactions(txs),
	}, nil
}

// ExportAssets returns the account's hand-kept asset classes with their items.
func (q *Queries) ExportAssets(ctx context.Context, accID uuid.UUID) (*AssetsFile, error) {
	r := q.accounts
	classes, err := r.ManualAssetClasses(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: asset classes: %w", err)
	}
	items, err := r.AssetItems(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: asset items: %w", err)
	}
	mutations, err := r.ManualAssetMutations(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: asset mutations: %w", err)
	}
	purchases, err := r.AssetPurchases(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("backup export: asset purchases: %w", err)
	}
	return &AssetsFile{
		Header:  newHeader(KindAssets, q.now()),
		Classes: exportAssetClasses(classes, items, mutations, purchases),
	}, nil
}

// ExportEODs returns the end-of-day prices of every listing that has any.
func (q *Queries) ExportEODs(ctx context.Context) (*EODFile, error) {
	listings, err := q.eods.ListingsWithEODs(ctx)
	if err != nil {
		return nil, fmt.Errorf("backup export: listings: %w", err)
	}
	file := &EODFile{Header: newHeader(KindEOD, q.now()), Listings: make([]EODListing, 0, len(listings))}
	for _, l := range listings {
		eods, err := q.eods.ListingEODs(ctx, l.ID)
		if err != nil {
			return nil, fmt.Errorf("backup export: eods of %s: %w", l.Symbol, err)
		}
		out := EODListing{ListingRef: ListingRef{Symbol: l.Symbol, Source: l.Source}, Name: l.Name, EODs: make([]EOD, 0, len(eods))}
		for _, e := range eods {
			out.EODs = append(out.EODs, EOD{
				Date: dayOf(e.Date), Open: e.Open, High: e.High, Low: e.Low, Close: e.Close,
				Volume: e.Volume, SplitFactor: e.SplitFactor,
			})
		}
		file.Listings = append(file.Listings, out)
	}
	return file, nil
}

func ref(id uuid.UUID) string { return id.String() }

func refPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	r := ref(*id)
	return &r
}

func exportImports(records []ImportRecord) []Import {
	out := make([]Import, 0, len(records))
	for _, r := range records {
		out = append(out, Import{
			Ref: ref(r.ID), Vendor: r.VendorName, Type: r.Type, Source: r.Source,
			Status: r.Status, StatusMsg: r.StatusMsg, TotalRows: r.TotalRows, Imported: r.Imported,
			Duplicates: r.Duplicates, Failed: r.Failed, AutoIgnored: r.AutoIgnored, CreatedAt: r.CreatedAt,
		})
	}
	return out
}

func exportCashflowTransactions(txs []*cashflow.Transaction) []CashflowTransaction {
	out := make([]CashflowTransaction, 0, len(txs))
	for _, t := range txs {
		out = append(out, CashflowTransaction{
			Ref: ref(t.ID), ImportRef: refPtr(t.ImportID), Description: t.Description, Note: t.Note,
			Source: t.Source, AmountCents: t.AmountCents, Direction: t.Direction, Date: dayOf(t.Date),
			Tag: t.Tag, AccountType: t.AccountType, Ignored: t.Ignored,
			IgnoredByRuleRef: refPtr(t.IgnoredByRuleID), IgnoreOverridden: t.IgnoreOverridden,
			RowNumber: t.RowNumber, Checksum: t.Checksum, CreatedAt: t.CreatedAt,
		})
	}
	return out
}

func exportIgnoreRules(rules []*cashflow.IgnoreRule) []IgnoreRule {
	out := make([]IgnoreRule, 0, len(rules))
	for _, r := range rules {
		out = append(out, IgnoreRule{
			Ref: ref(r.ID), Name: r.Name, MatchField: r.MatchField, Contains: r.Contains,
			Direction: r.Direction, Source: r.Source, Enabled: r.Enabled, IgnoredTotal: r.IgnoredTotal,
			LastAppliedAt: r.LastAppliedAt, CreatedAt: r.CreatedAt,
		})
	}
	return out
}

func exportRecurringItems(items []*cashflow.RecurringItem, links []LinkRecord) []RecurringItem {
	linked := make(map[uuid.UUID][]string, len(items))
	for _, l := range links {
		linked[l.ItemID] = append(linked[l.ItemID], ref(l.TransactionID))
	}
	out := make([]RecurringItem, 0, len(items))
	for _, i := range items {
		refs := linked[i.ID]
		if refs == nil {
			refs = []string{}
		}
		out = append(out, RecurringItem{
			Name: i.Name, Direction: i.Direction, Rhythm: i.Rhythm, MatchKey: i.MatchKey,
			EndedFrom: dayPtr(i.EndedFrom), CreatedAt: i.CreatedAt, Transactions: refs,
		})
	}
	return out
}

func exportDismissals(dismissals []cashflow.DismissedSuggestion) []Dismissal {
	out := make([]Dismissal, 0, len(dismissals))
	for _, d := range dismissals {
		out = append(out, Dismissal{MatchKey: d.MatchKey, Direction: d.Direction})
	}
	return out
}

func exportPortfolioTransactions(txs []*portfolio.Transaction) []PortfolioTransaction {
	out := make([]PortfolioTransaction, 0, len(txs))
	for _, t := range txs {
		out = append(out, PortfolioTransaction{
			Ref: ref(t.ID), ImportRef: refPtr(t.ImportID), Origin: t.Origin, Source: t.Source,
			OccurredAt: dayOf(t.OccurredAt), ISIN: t.ISIN, Symbol: t.Symbol, Description: t.Description,
			Type: t.Type, Quantity: t.Quantity, UnitPrice: t.UnitPrice, AmountCents: t.AmountCents,
			RowNumber: t.RowNumber, Checksum: t.Checksum, CreatedAt: t.CreatedAt,
		})
	}
	return out
}

func exportAssetClasses(classes []*assets.Class, items []ItemRecord, mutations []*assets.Mutation, purchases []*assets.Purchase) []AssetClass {
	mutationsOf := make(map[uuid.UUID][]AssetMutation)
	for _, m := range mutations {
		mutationsOf[m.AssetID] = append(mutationsOf[m.AssetID], AssetMutation{
			ChangeType: m.ChangeType, Direction: m.Direction, Amount: m.Amount,
			PreviousWorth: m.PreviousWorth, NewWorth: m.NewWorth, ClassTotalWorth: m.ClassTotalWorth,
			EffectiveDate: dayOf(m.EffectiveDate), Note: m.Note, CreatedAt: m.CreatedAt,
		})
	}
	purchasesOf := make(map[uuid.UUID][]AssetPurchase)
	for _, p := range purchases {
		purchasesOf[p.AssetID] = append(purchasesOf[p.AssetID], AssetPurchase{
			PurchasedOn: dayOf(p.PurchasedOn), Quantity: p.Quantity, UnitPrice: p.UnitPrice, CreatedAt: p.CreatedAt,
		})
	}
	itemsOf := make(map[uuid.UUID][]AssetItem)
	for _, i := range items {
		var listing *ListingRef
		if i.ListingSymbol != nil && i.ListingSource != nil {
			listing = &ListingRef{Symbol: *i.ListingSymbol, Source: *i.ListingSource}
		}
		itemsOf[i.ClassID] = append(itemsOf[i.ClassID], AssetItem{
			Name: i.Name, Listing: listing, CurrentWorth: i.CurrentWorth, Archived: i.Archived,
			CreatedAt: i.CreatedAt, Mutations: orEmpty(mutationsOf[i.ID]), Purchases: orEmpty(purchasesOf[i.ID]),
		})
	}
	out := make([]AssetClass, 0, len(classes))
	for _, c := range classes {
		out = append(out, AssetClass{
			Name: c.Name, Archived: c.Archived, CreatedAt: c.CreatedAt, Items: orEmpty(itemsOf[c.ID]),
		})
	}
	return out
}

// orEmpty keeps an empty list an empty JSON array rather than null.
func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
