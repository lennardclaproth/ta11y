package backup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

// Count is how many records of one kind a file held that were new to the target, and
// how many it already had.
type Count struct {
	New      int
	Existing int
}

// Missing names what a file refers to that this instance does not have.
type Missing struct {
	Vendors  []string
	Listings []ListingRef
}

func (m Missing) empty() bool { return len(m.Vendors) == 0 && len(m.Listings) == 0 }

// CashflowResult is what a cashflow import did, or on a dry run what it would do.
type CashflowResult struct {
	DryRun         bool
	Imports        Count
	Transactions   Count
	IgnoreRules    Count
	RecurringItems Count
	RecurringLinks Count
	Dismissals     Count
	Missing        Missing
}

// PortfolioResult is what a portfolio import did, or on a dry run what it would do.
type PortfolioResult struct {
	DryRun       bool
	Imports      Count
	Transactions Count
	Missing      Missing
}

// AssetsResult is what an assets import did, or on a dry run what it would do.
type AssetsResult struct {
	DryRun    bool
	Classes   Count
	Items     Count
	Mutations Count
	Purchases Count
	Missing   Missing
}

// EODResult is what an EOD import did, or on a dry run what it would do. A listing
// counts as new when the import created it.
type EODResult struct {
	DryRun   bool
	Listings Count
	EODs     Count
	Missing  Missing
}

// UnitOfWork runs fn in one database transaction.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// ReferenceResolver finds the shared records a file names by what two instances agree on.
type ReferenceResolver interface {
	VendorIDs(ctx context.Context) (map[string]uuid.UUID, error)
	ListingIDs(ctx context.Context) (map[ListingRef]uuid.UUID, error)
}

// AccountWriter merges an account's files in. Every insert writes all of a record's
// columns; the existing-record lookups are what make the merge skip what is there.
type AccountWriter interface {
	UnitOfWork
	ReferenceResolver
	// CashflowChecksums and PortfolioChecksums return the id holding each checksum
	// that exists. Checksums are unique across accounts, as the tables enforce.
	CashflowChecksums(ctx context.Context, checksums []string) (map[string]uuid.UUID, error)
	PortfolioChecksums(ctx context.Context, checksums []string) (map[string]uuid.UUID, error)
	IgnoreRules(ctx context.Context, accID uuid.UUID) ([]*cashflow.IgnoreRule, error)
	RecurringItems(ctx context.Context, accID uuid.UUID) ([]*cashflow.RecurringItem, error)
	// AssetClasses includes the portfolio-sourced class: names are unique across both.
	AssetClasses(ctx context.Context, accID uuid.UUID) ([]*assets.Class, error)
	AssetItems(ctx context.Context, accID uuid.UUID) ([]ItemRecord, error)

	InsertImports(ctx context.Context, imports []*importer.Import) error
	InsertCashflowTransactions(ctx context.Context, txs []*cashflow.Transaction) error
	InsertIgnoreRules(ctx context.Context, rules []*cashflow.IgnoreRule) error
	InsertRecurringItems(ctx context.Context, items []*cashflow.RecurringItem) error
	// InsertRecurringLinks skips a transaction that is linked already and returns how
	// many links it added.
	InsertRecurringLinks(ctx context.Context, links []LinkRecord) (int, error)
	// InsertDismissals skips a dismissal the account has and returns how many it added.
	InsertDismissals(ctx context.Context, accID uuid.UUID, dismissals []cashflow.DismissedSuggestion) (int, error)
	InsertPortfolioTransactions(ctx context.Context, txs []*portfolio.Transaction) error
	InsertAssetClasses(ctx context.Context, classes []*assets.Class) error
	InsertAssetItems(ctx context.Context, items []*assets.Asset) error
	InsertAssetMutations(ctx context.Context, mutations []*assets.Mutation) error
	InsertAssetPurchases(ctx context.Context, purchases []*assets.Purchase) error
}

// EODWriter merges an EOD file in.
type EODWriter interface {
	UnitOfWork
	ReferenceResolver
	// ProviderIDs returns, per provider name, the provider a listing of that source
	// is linked to on this instance. A name without a provider is absent.
	ProviderIDs(ctx context.Context) (map[marketdata.ProviderName]uuid.UUID, error)
	// InsertListings writes new listings with every column, their provider included.
	InsertListings(ctx context.Context, listings []*marketdata.Listing) error
	// InsertEODs skips a day the listing has and returns how many days it added.
	InsertEODs(ctx context.Context, eods []*marketdata.EOD) (int, error)
}

// Commands merges backup files in.
type Commands struct {
	accounts AccountWriter
	eods     EODWriter
	bus      eventbus.Bus
	now      func() time.Time
}

// NewCommands creates backup commands. The bus may be nil, in which case nothing is
// rebuilt after a restore.
func NewCommands(accounts AccountWriter, eods EODWriter, bus eventbus.Bus) *Commands {
	return &Commands{accounts: accounts, eods: eods, bus: bus, now: time.Now}
}

// errRollback ends a dry run's transaction after everything was written, so the
// counts it reports come from the same path a real import takes.
var errRollback = errors.New("backup: dry run")

// merge runs fn in one transaction: the whole file lands or nothing does, and a dry run
// is rolled back after fn wrote everything.
func merge(ctx context.Context, uow UnitOfWork, dryRun bool, fn func(ctx context.Context) error) error {
	err := uow.Do(ctx, func(ctx context.Context) error {
		if err := fn(ctx); err != nil {
			return err
		}
		if dryRun {
			return errRollback
		}
		return nil
	})
	if errors.Is(err, errRollback) {
		return nil
	}
	return err
}

// settle hands back what a merge produced: the result alone on success, the result with
// ErrMissingReferences when the file named what is not here, and nothing on failure.
func settle[R any](result *R, err error) (*R, error) {
	switch {
	case err == nil:
		return result, nil
	case errors.Is(err, ErrMissingReferences):
		return result, err
	default:
		return nil, err
	}
}

func (c *Commands) publish(ctx context.Context, topic string, accID uuid.UUID) error {
	if c.bus == nil {
		return nil
	}
	if err := c.bus.Publish(ctx, topic, Restored{AccID: accID}); err != nil {
		return fmt.Errorf("backup import: publish %s: %w", topic, err)
	}
	return nil
}

// ImportCashflow merges a cashflow file into accID: rows the account has are skipped,
// new ones are added. A dry run reports the same counts and writes nothing. When the
// file names vendors this instance lacks, nothing is written and Missing names them.
func (c *Commands) ImportCashflow(ctx context.Context, accID uuid.UUID, file *CashflowFile, dryRun bool) (*CashflowResult, error) {
	if err := file.check(KindCashflow); err != nil {
		return nil, err
	}
	if err := file.validate(); err != nil {
		return nil, err
	}
	result := &CashflowResult{DryRun: dryRun}
	err := merge(ctx, c.accounts, dryRun, func(ctx context.Context) error {
		return newCashflowRestore(c.accounts, accID, file, result, c.now().UTC()).run(ctx)
	})
	return settle(result, err)
}

// ImportPortfolio merges a portfolio file into accID, then has the portfolio rebuilt
// from what it holds. A dry run reports the same counts and writes nothing.
func (c *Commands) ImportPortfolio(ctx context.Context, accID uuid.UUID, file *PortfolioFile, dryRun bool) (*PortfolioResult, error) {
	if err := file.check(KindPortfolio); err != nil {
		return nil, err
	}
	if err := file.validate(); err != nil {
		return nil, err
	}
	result := &PortfolioResult{DryRun: dryRun}
	err := merge(ctx, c.accounts, dryRun, func(ctx context.Context) error {
		return newPortfolioRestore(c.accounts, accID, file, result, c.now().UTC()).run(ctx)
	})
	if err != nil {
		return settle(result, err)
	}
	if !dryRun && result.Transactions.New > 0 {
		if err := c.publish(ctx, TopicPortfolioRestored, accID); err != nil {
			return result, err
		}
	}
	return result, nil
}

// ImportAssets merges an assets file into accID, then has the asset snapshots rebuilt.
// An item the account already holds keeps its own history. A dry run reports the same
// counts and writes nothing.
func (c *Commands) ImportAssets(ctx context.Context, accID uuid.UUID, file *AssetsFile, dryRun bool) (*AssetsResult, error) {
	if err := file.check(KindAssets); err != nil {
		return nil, err
	}
	if err := file.validate(); err != nil {
		return nil, err
	}
	result := &AssetsResult{DryRun: dryRun}
	err := merge(ctx, c.accounts, dryRun, func(ctx context.Context) error {
		return newAssetsRestore(c.accounts, accID, file, result, c.now().UTC()).run(ctx)
	})
	if err != nil {
		return settle(result, err)
	}
	if !dryRun && result.Items.New > 0 {
		if err := c.publish(ctx, TopicAssetsRestored, accID); err != nil {
			return result, err
		}
	}
	return result, nil
}

// ImportEODs merges an EOD file in: listings this instance lacks are created first,
// then days a listing has are skipped and new ones added. A listing that exists keeps
// its own data. A listing the file carries without instrument data — a file exported
// before listings travelled — must exist here; when one does not, nothing is written
// and Missing names them. A split factor travels with its day.
func (c *Commands) ImportEODs(ctx context.Context, file *EODFile, dryRun bool) (*EODResult, error) {
	if err := file.check(KindEOD); err != nil {
		return nil, err
	}
	if err := file.validate(); err != nil {
		return nil, err
	}

	result := &EODResult{DryRun: dryRun}
	err := merge(ctx, c.eods, dryRun, func(ctx context.Context) error {
		listingIDs, err := c.eods.ListingIDs(ctx)
		if err != nil {
			return fmt.Errorf("backup import: listings: %w", err)
		}
		var create []EODListing
		for _, l := range file.Listings {
			switch _, ok := listingIDs[l.ListingRef]; {
			case ok:
				result.Listings.Existing++
			case l.Instrument == nil:
				result.Missing.Listings = append(result.Missing.Listings, l.ListingRef)
			default:
				create = append(create, l)
			}
		}
		if !result.Missing.empty() {
			sortListingRefs(result.Missing.Listings)
			return ErrMissingReferences
		}

		now := c.now().UTC()
		if err := c.createListings(ctx, create, listingIDs, now); err != nil {
			return err
		}
		result.Listings.New = len(create)

		for _, l := range file.Listings {
			listingID := listingIDs[l.ListingRef]
			eods := make([]*marketdata.EOD, 0, len(l.EODs))
			for _, e := range l.EODs {
				eods = append(eods, &marketdata.EOD{
					ID: uuid.New(), ListingID: listingID, Symbol: l.Symbol, Date: e.Date.Time,
					Open: e.Open, High: e.High, Low: e.Low, Close: e.Close, Volume: e.Volume,
					SplitFactor: e.SplitFactor, CreatedAt: now, UpdatedAt: now,
				})
			}
			added, err := c.eods.InsertEODs(ctx, eods)
			if err != nil {
				return fmt.Errorf("backup import: eods of %s: %w", l.Symbol, err)
			}
			result.EODs.New += added
			result.EODs.Existing += len(eods) - added
		}
		return nil
	})
	return settle(result, err)
}

// createListings writes the file's listings this instance lacks and adds their ids to
// listingIDs. Each is linked to this instance's provider for its source, when there is
// one. Its accumulated range spans the file's first and last day, so for an API source
// the first end-of-day read only syncs the gap after the file; a listing the file
// brings without prices has no range and syncs its full history.
func (c *Commands) createListings(ctx context.Context, create []EODListing, listingIDs map[ListingRef]uuid.UUID, now time.Time) error {
	if len(create) == 0 {
		return nil
	}
	providerIDs, err := c.eods.ProviderIDs(ctx)
	if err != nil {
		return fmt.Errorf("backup import: providers: %w", err)
	}
	listings := make([]*marketdata.Listing, 0, len(create))
	for _, l := range create {
		listing, err := marketdata.NewListing(l.Symbol, l.Name, l.Source)
		if err != nil {
			return fmt.Errorf("backup import: listing %s: %w", l.Symbol, err)
		}
		in := l.Instrument
		listing.Type, listing.Ticker, listing.ISIN = in.Type, in.Ticker, in.ISIN
		listing.Description, listing.Exchange, listing.Region = in.Description, in.Exchange, in.Region
		listing.Currency, listing.Active = in.Currency, in.Active
		listing.CreatedAt, listing.UpdatedAt = now, now
		for _, e := range l.EODs {
			d := e.Date.Time
			if listing.AccumulatedStart == nil || d.Before(*listing.AccumulatedStart) {
				listing.AccumulatedStart = &d
			}
			if listing.AccumulatedEnd == nil || d.After(*listing.AccumulatedEnd) {
				listing.AccumulatedEnd = &d
			}
		}
		if name, err := marketdata.ProviderNameFromSource(l.Source); err == nil {
			if id, ok := providerIDs[name]; ok {
				listing.ProviderID = &id
			}
		}
		listings = append(listings, listing)
		listingIDs[l.ListingRef] = listing.ID
	}
	if err := c.eods.InsertListings(ctx, listings); err != nil {
		return fmt.Errorf("backup import: listings: %w", err)
	}
	return nil
}
