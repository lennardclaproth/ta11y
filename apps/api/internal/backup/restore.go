package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

// Each restore merges one file inside one transaction. It first works out what is new —
// so that a file naming missing vendors or listings is refused before anything is
// written — and then writes only that.

// importsRestore recreates the record of every import that brings at least one new row,
// so restored rows keep their origin and their import's review page. The upload itself
// is gone after processing, here as on the instance the file came from.
type importsRestore struct {
	store   AccountWriter
	accID   uuid.UUID
	now     time.Time
	imports []Import
	needed  map[string]struct{}
}

func newImportsRestore(store AccountWriter, accID uuid.UUID, imports []Import, now time.Time) *importsRestore {
	return &importsRestore{store: store, accID: accID, now: now, imports: imports, needed: make(map[string]struct{})}
}

// need marks the import a new row came from.
func (r *importsRestore) need(ref *string) {
	if ref != nil {
		r.needed[*ref] = struct{}{}
	}
}

// missingVendors returns the vendors of the needed imports that this instance lacks.
func (r *importsRestore) missingVendors(vendorIDs map[string]uuid.UUID) []string {
	missing := make(map[string]struct{})
	for _, i := range r.imports {
		if _, needed := r.needed[i.Ref]; !needed {
			continue
		}
		if _, ok := vendorIDs[i.Vendor]; !ok {
			missing[i.Vendor] = struct{}{}
		}
	}
	var out []string
	for v := range missing {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}

func (r *importsRestore) write(ctx context.Context, vendorIDs map[string]uuid.UUID, count *Count) (map[string]uuid.UUID, error) {
	ids := make(map[string]uuid.UUID, len(r.needed))
	records := make([]*importer.Import, 0, len(r.needed))
	for _, i := range r.imports {
		if _, needed := r.needed[i.Ref]; !needed {
			count.Existing++
			continue
		}
		id := uuid.New()
		ids[i.Ref] = id
		accID := r.accID
		status, msg := i.Status, i.StatusMsg
		if status != importer.ImportStatusCompleted && status != importer.ImportStatusFailed {
			status, msg = importer.ImportStatusFailed, "interrupted before the backup was made"
		}
		records = append(records, &importer.Import{
			ID: id, Type: i.Type, CreatedAt: i.CreatedAt, UpdatedAt: r.now, VendorID: vendorIDs[i.Vendor],
			AccountID: &accID, Source: i.Source, Status: status, StatusMsg: msg,
			Duplicates: i.Duplicates, TotalRows: i.TotalRows, Imported: i.Imported, Failed: i.Failed,
			AutoIgnored: i.AutoIgnored,
		})
		count.New++
	}
	if err := r.store.InsertImports(ctx, records); err != nil {
		return nil, fmt.Errorf("backup import: imports: %w", err)
	}
	return ids, nil
}

// resolveVendors refuses the file when a needed import's vendor is missing here.
func (r *importsRestore) resolveVendors(ctx context.Context, missing *Missing) (map[string]uuid.UUID, error) {
	vendorIDs, err := r.store.VendorIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("backup import: vendors: %w", err)
	}
	missing.Vendors = r.missingVendors(vendorIDs)
	if !missing.empty() {
		return nil, ErrMissingReferences
	}
	return vendorIDs, nil
}

// --- Cashflow -----------------------------------------------------------------

type cashflowRestore struct {
	store   AccountWriter
	accID   uuid.UUID
	file    *CashflowFile
	now     time.Time
	result  *CashflowResult
	imports *importsRestore

	// ids maps every transaction ref to the row that holds it in the account, whether
	// it was there already or is added now.
	ids   map[string]uuid.UUID
	added []pendingCashflow
}

type pendingCashflow struct {
	file     *CashflowTransaction
	id       uuid.UUID
	checksum string
}

func newCashflowRestore(store AccountWriter, accID uuid.UUID, file *CashflowFile, result *CashflowResult, now time.Time) *cashflowRestore {
	return &cashflowRestore{
		store: store, accID: accID, file: file, now: now, result: result,
		imports: newImportsRestore(store, accID, file.Imports, now),
		ids:     make(map[string]uuid.UUID, len(file.Transactions)),
	}
}

func (r *cashflowRestore) run(ctx context.Context) error {
	if err := r.plan(ctx); err != nil {
		return err
	}
	vendorIDs, err := r.imports.resolveVendors(ctx, &r.result.Missing)
	if err != nil {
		return err
	}
	importIDs, err := r.imports.write(ctx, vendorIDs, &r.result.Imports)
	if err != nil {
		return err
	}
	ruleIDs, err := r.writeIgnoreRules(ctx)
	if err != nil {
		return err
	}
	if err := r.writeTransactions(ctx, importIDs, ruleIDs); err != nil {
		return err
	}
	return r.writeRecurring(ctx)
}

func (r *cashflowRestore) plan(ctx context.Context) error {
	txs := r.file.Transactions
	checksums := make([]string, len(txs))
	for i := range txs {
		checksums[i] = cashflowChecksum(&txs[i], r.file.SourceAccountID, r.accID)
	}
	existing, err := r.store.CashflowChecksums(ctx, checksums)
	if err != nil {
		return fmt.Errorf("backup import: cashflow checksums: %w", err)
	}

	added := make(map[string]uuid.UUID)
	for i := range txs {
		t, sum := &txs[i], checksums[i]
		id, there := existing[sum]
		if !there {
			// Two rows of the file that land on one checksum are one row to the ledger.
			id, there = added[sum]
		}
		if there {
			r.ids[t.Ref] = id
			r.result.Transactions.Existing++
			continue
		}
		id = uuid.New()
		added[sum] = id
		r.ids[t.Ref] = id
		r.added = append(r.added, pendingCashflow{file: t, id: id, checksum: sum})
		r.result.Transactions.New++
		r.imports.need(t.ImportRef)
	}
	return nil
}

type ruleKey struct {
	field     cashflow.MatchField
	contains  string
	direction string
	source    string
}

func keyOf(field cashflow.MatchField, contains string, direction *cashflow.CashFlowDirection, source string) ruleKey {
	k := ruleKey{field: field, contains: contains, source: source}
	if direction != nil {
		k.direction = string(*direction)
	}
	return k
}

// writeIgnoreRules adds the rules the account lacks. A rule is what it matches, not
// what it is called, so one that catches the same rows is the same rule.
func (r *cashflowRestore) writeIgnoreRules(ctx context.Context) (map[string]uuid.UUID, error) {
	held, err := r.store.IgnoreRules(ctx, r.accID)
	if err != nil {
		return nil, fmt.Errorf("backup import: ignore rules: %w", err)
	}
	byKey := make(map[ruleKey]uuid.UUID, len(held))
	for _, h := range held {
		byKey[keyOf(h.MatchField, h.Contains, h.Direction, h.Source)] = h.ID
	}

	ids := make(map[string]uuid.UUID, len(r.file.IgnoreRules))
	var added []*cashflow.IgnoreRule
	for _, fr := range r.file.IgnoreRules {
		key := keyOf(fr.MatchField, fr.Contains, fr.Direction, fr.Source)
		if id, ok := byKey[key]; ok {
			ids[fr.Ref] = id
			r.result.IgnoreRules.Existing++
			continue
		}
		rule := &cashflow.IgnoreRule{
			ID: uuid.New(), AccountID: r.accID, Name: fr.Name, MatchField: fr.MatchField,
			Contains: fr.Contains, Direction: fr.Direction, Source: fr.Source, Enabled: fr.Enabled,
			IgnoredTotal: fr.IgnoredTotal, LastAppliedAt: fr.LastAppliedAt, CreatedAt: fr.CreatedAt, UpdatedAt: r.now,
		}
		byKey[key] = rule.ID
		ids[fr.Ref] = rule.ID
		added = append(added, rule)
		r.result.IgnoreRules.New++
	}
	if err := r.store.InsertIgnoreRules(ctx, added); err != nil {
		return nil, fmt.Errorf("backup import: ignore rules: %w", err)
	}
	return ids, nil
}

func (r *cashflowRestore) writeTransactions(ctx context.Context, importIDs, ruleIDs map[string]uuid.UUID) error {
	txs := make([]*cashflow.Transaction, 0, len(r.added))
	for _, p := range r.added {
		t := p.file
		tx := &cashflow.Transaction{
			ID: p.id, AccountID: r.accID, Description: t.Description, Note: t.Note, Source: t.Source,
			AmountCents: t.AmountCents, Direction: t.Direction, Date: t.Date.Time, Checksum: p.checksum,
			CreatedAt: t.CreatedAt, UpdatedAt: r.now, Tag: t.Tag, RowNumber: t.RowNumber, Ignored: t.Ignored,
			AccountType: t.AccountType, IgnoreOverridden: t.IgnoreOverridden,
		}
		if t.ImportRef != nil {
			id := importIDs[*t.ImportRef]
			tx.ImportID = &id
		}
		if t.IgnoredByRuleRef != nil {
			id := ruleIDs[*t.IgnoredByRuleRef]
			tx.IgnoredByRuleID = &id
		}
		txs = append(txs, tx)
	}
	if err := r.store.InsertCashflowTransactions(ctx, txs); err != nil {
		return fmt.Errorf("backup import: cashflow transactions: %w", err)
	}
	return nil
}

func (r *cashflowRestore) writeRecurring(ctx context.Context) error {
	held, err := r.store.RecurringItems(ctx, r.accID)
	if err != nil {
		return fmt.Errorf("backup import: recurring items: %w", err)
	}
	byName := make(map[string]uuid.UUID, len(held))
	for _, h := range held {
		byName[h.Name] = h.ID
	}

	var added []*cashflow.RecurringItem
	var links []LinkRecord
	for _, fi := range r.file.RecurringItems {
		id, ok := byName[fi.Name]
		if ok {
			r.result.RecurringItems.Existing++
		} else {
			id = uuid.New()
			added = append(added, &cashflow.RecurringItem{
				ID: id, AccountID: r.accID, Name: fi.Name, Direction: fi.Direction, Rhythm: fi.Rhythm,
				MatchKey: fi.MatchKey, EndedFrom: fi.EndedFrom.timePtr(), CreatedAt: fi.CreatedAt, UpdatedAt: r.now,
			})
			r.result.RecurringItems.New++
		}
		for _, ref := range fi.Transactions {
			links = append(links, LinkRecord{ItemID: id, TransactionID: r.ids[ref]})
		}
	}
	if err := r.store.InsertRecurringItems(ctx, added); err != nil {
		return fmt.Errorf("backup import: recurring items: %w", err)
	}

	linked, err := r.store.InsertRecurringLinks(ctx, links)
	if err != nil {
		return fmt.Errorf("backup import: recurring links: %w", err)
	}
	r.result.RecurringLinks = Count{New: linked, Existing: len(links) - linked}

	dismissals := make([]cashflow.DismissedSuggestion, 0, len(r.file.Dismissals))
	for _, d := range r.file.Dismissals {
		dismissals = append(dismissals, cashflow.DismissedSuggestion{MatchKey: d.MatchKey, Direction: d.Direction})
	}
	dismissed, err := r.store.InsertDismissals(ctx, r.accID, dismissals)
	if err != nil {
		return fmt.Errorf("backup import: recurring dismissals: %w", err)
	}
	r.result.Dismissals = Count{New: dismissed, Existing: len(dismissals) - dismissed}
	return nil
}

// --- Portfolio ----------------------------------------------------------------

type portfolioRestore struct {
	store   AccountWriter
	accID   uuid.UUID
	file    *PortfolioFile
	now     time.Time
	result  *PortfolioResult
	imports *importsRestore
	added   []pendingPortfolio
}

type pendingPortfolio struct {
	file     *PortfolioTransaction
	checksum string
}

func newPortfolioRestore(store AccountWriter, accID uuid.UUID, file *PortfolioFile, result *PortfolioResult, now time.Time) *portfolioRestore {
	return &portfolioRestore{
		store: store, accID: accID, file: file, now: now, result: result,
		imports: newImportsRestore(store, accID, file.Imports, now),
	}
}

func (r *portfolioRestore) run(ctx context.Context) error {
	if err := r.plan(ctx); err != nil {
		return err
	}
	vendorIDs, err := r.imports.resolveVendors(ctx, &r.result.Missing)
	if err != nil {
		return err
	}
	importIDs, err := r.imports.write(ctx, vendorIDs, &r.result.Imports)
	if err != nil {
		return err
	}
	return r.writeTransactions(ctx, importIDs)
}

func (r *portfolioRestore) plan(ctx context.Context) error {
	txs := r.file.Transactions
	checksums := make([]string, len(txs))
	for i := range txs {
		checksums[i] = portfolioChecksum(&txs[i], r.file.SourceAccountID, r.accID)
	}
	existing, err := r.store.PortfolioChecksums(ctx, checksums)
	if err != nil {
		return fmt.Errorf("backup import: portfolio checksums: %w", err)
	}

	added := make(map[string]struct{})
	for i := range txs {
		t, sum := &txs[i], checksums[i]
		_, there := existing[sum]
		_, twice := added[sum]
		if there || twice {
			r.result.Transactions.Existing++
			continue
		}
		added[sum] = struct{}{}
		r.added = append(r.added, pendingPortfolio{file: t, checksum: sum})
		r.result.Transactions.New++
		r.imports.need(t.ImportRef)
	}
	return nil
}

func (r *portfolioRestore) writeTransactions(ctx context.Context, importIDs map[string]uuid.UUID) error {
	txs := make([]*portfolio.Transaction, 0, len(r.added))
	for _, p := range r.added {
		t := p.file
		accID := r.accID
		tx := &portfolio.Transaction{
			ID: uuid.New(), AccountID: &accID, Origin: t.Origin, Source: t.Source, OccurredAt: t.OccurredAt.Time,
			ISIN: t.ISIN, Symbol: t.Symbol, Description: t.Description, Type: t.Type, Quantity: t.Quantity,
			UnitPrice: t.UnitPrice, AmountCents: t.AmountCents, Checksum: p.checksum, RowNumber: t.RowNumber,
			CreatedAt: t.CreatedAt, UpdatedAt: r.now,
		}
		if t.ImportRef != nil {
			id := importIDs[*t.ImportRef]
			tx.ImportID = &id
		}
		txs = append(txs, tx)
	}
	if err := r.store.InsertPortfolioTransactions(ctx, txs); err != nil {
		return fmt.Errorf("backup import: portfolio transactions: %w", err)
	}
	return nil
}

// --- Checksums ----------------------------------------------------------------

// cashflowChecksum is the checksum the row carries in the target account. Into the
// account it came from it keeps its own; into another one it is re-derived, so a later
// CSV import of the same statement there still recognises the row.
func cashflowChecksum(t *CashflowTransaction, source, target uuid.UUID) string {
	if source == target {
		return t.Checksum
	}
	probe := cashflow.Transaction{
		AccountID: source, Description: t.Description, Note: t.Note, Source: t.Source,
		AmountCents: t.AmountCents, Direction: t.Direction, Date: t.Date.Time, RowNumber: t.RowNumber,
		Checksum: t.Checksum,
	}
	if sum, ok := probe.RekeyedChecksum(target); ok {
		return sum
	}
	return fallbackChecksum(t.Checksum, target)
}

func portfolioChecksum(t *PortfolioTransaction, source, target uuid.UUID) string {
	if source == target {
		return t.Checksum
	}
	probe := portfolio.Transaction{
		AccountID: &source, Origin: t.Origin, Source: t.Source, OccurredAt: t.OccurredAt.Time,
		ISIN: t.ISIN, Symbol: t.Symbol, Type: t.Type, Quantity: t.Quantity, UnitPrice: t.UnitPrice,
		AmountCents: t.AmountCents, RowNumber: t.RowNumber, Checksum: t.Checksum,
	}
	if sum, ok := probe.RekeyedChecksum(target); ok {
		return sum
	}
	return fallbackChecksum(t.Checksum, target)
}

// fallbackChecksum keys a row whose dedup sequence cannot be recovered — a field the
// checksum digests was edited after import — to the target account. Restoring the same
// file twice still finds it; only a later CSV import of that statement will not.
func fallbackChecksum(checksum string, accID uuid.UUID) string {
	sum := sha256.Sum256([]byte(checksum + "\x1F" + accID.String()))
	return hex.EncodeToString(sum[:])
}

// sortListingRefs orders listing refs by source, then symbol, for stable reporting.
func sortListingRefs(refs []ListingRef) {
	slices.SortFunc(refs, func(a, b ListingRef) int {
		return strings.Compare(string(a.Source)+"\x1F"+a.Symbol, string(b.Source)+"\x1F"+b.Symbol)
	})
}
