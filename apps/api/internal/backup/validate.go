package backup

import (
	"fmt"
	"strings"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidFile, fmt.Sprintf(format, args...))
}

// validateImports checks a file's import records, which must all be of importType, and
// returns their refs.
func validateImports(imports []Import, importType importer.ImportType) (map[string]struct{}, error) {
	refs := make(map[string]struct{}, len(imports))
	for _, i := range imports {
		if _, dup := refs[i.Ref]; dup || i.Ref == "" {
			return nil, invalid("import ref %q is empty or repeated", i.Ref)
		}
		if i.Type != importType {
			return nil, invalid("import %s has type %q, expected %q", i.Ref, i.Type, importType)
		}
		if strings.TrimSpace(i.Vendor) == "" {
			return nil, invalid("import %s names no vendor", i.Ref)
		}
		refs[i.Ref] = struct{}{}
	}
	return refs, nil
}

// The validate methods check what the database would otherwise refuse halfway through a
// restore: refs that point nowhere, and values outside the sets the tables allow.

func (f *CashflowFile) validate() error {
	imports, err := validateImports(f.Imports, importer.ImportTypeCashflow)
	if err != nil {
		return err
	}

	rules := make(map[string]struct{}, len(f.IgnoreRules))
	for _, r := range f.IgnoreRules {
		if _, dup := rules[r.Ref]; dup || r.Ref == "" {
			return invalid("ignore rule ref %q is empty or repeated", r.Ref)
		}
		if r.MatchField != cashflow.MatchFieldDescription && r.MatchField != cashflow.MatchFieldNote {
			return invalid("ignore rule %s matches field %q", r.Ref, r.MatchField)
		}
		if r.Direction != nil && !validDirection(*r.Direction) {
			return invalid("ignore rule %s has direction %q", r.Ref, *r.Direction)
		}
		rules[r.Ref] = struct{}{}
	}

	txs := make(map[string]struct{}, len(f.Transactions))
	for _, t := range f.Transactions {
		if _, dup := txs[t.Ref]; dup || t.Ref == "" {
			return invalid("cashflow transaction ref %q is empty or repeated", t.Ref)
		}
		if !validDirection(t.Direction) {
			return invalid("cashflow transaction %s has direction %q", t.Ref, t.Direction)
		}
		if t.Checksum == "" {
			return invalid("cashflow transaction %s has no checksum", t.Ref)
		}
		if t.ImportRef != nil {
			if _, ok := imports[*t.ImportRef]; !ok {
				return invalid("cashflow transaction %s points at import %q, which is not in the file", t.Ref, *t.ImportRef)
			}
		}
		if t.IgnoredByRuleRef != nil {
			if _, ok := rules[*t.IgnoredByRuleRef]; !ok {
				return invalid("cashflow transaction %s points at ignore rule %q, which is not in the file", t.Ref, *t.IgnoredByRuleRef)
			}
		}
		txs[t.Ref] = struct{}{}
	}

	names := make(map[string]struct{}, len(f.RecurringItems))
	for _, i := range f.RecurringItems {
		if strings.TrimSpace(i.Name) == "" {
			return invalid("a recurring item has no name")
		}
		if _, dup := names[i.Name]; dup {
			return invalid("recurring item %q is repeated", i.Name)
		}
		names[i.Name] = struct{}{}
		if !validDirection(i.Direction) {
			return invalid("recurring item %q has direction %q", i.Name, i.Direction)
		}
		if i.Rhythm != cashflow.RhythmMonthly && i.Rhythm != cashflow.RhythmQuarterly && i.Rhythm != cashflow.RhythmYearly {
			return invalid("recurring item %q has rhythm %q", i.Name, i.Rhythm)
		}
		for _, ref := range i.Transactions {
			if _, ok := txs[ref]; !ok {
				return invalid("recurring item %q links transaction %q, which is not in the file", i.Name, ref)
			}
		}
	}
	for _, d := range f.Dismissals {
		if !validDirection(d.Direction) {
			return invalid("recurring dismissal %q has direction %q", d.MatchKey, d.Direction)
		}
	}
	return nil
}

func (f *PortfolioFile) validate() error {
	imports, err := validateImports(f.Imports, importer.ImportTypePortfolio)
	if err != nil {
		return err
	}
	ptxs := make(map[string]struct{}, len(f.Transactions))
	for _, t := range f.Transactions {
		if _, dup := ptxs[t.Ref]; dup || t.Ref == "" {
			return invalid("portfolio transaction ref %q is empty or repeated", t.Ref)
		}
		ptxs[t.Ref] = struct{}{}
		if t.Checksum == "" {
			return invalid("portfolio transaction %s has no checksum", t.Ref)
		}
		switch t.Type {
		case portfolio.TxBuy, portfolio.TxSell, portfolio.TxDividend, portfolio.TxTax, portfolio.TxFee, portfolio.TxCash:
		default:
			return invalid("portfolio transaction %s has type %q", t.Ref, t.Type)
		}
		switch t.Origin {
		case portfolio.TransactionOriginImport:
			if t.ImportRef == nil {
				return invalid("imported portfolio transaction %s names no import", t.Ref)
			}
			if _, ok := imports[*t.ImportRef]; !ok {
				return invalid("portfolio transaction %s points at import %q, which is not in the file", t.Ref, *t.ImportRef)
			}
		case portfolio.TransactionOriginManual:
			if t.ImportRef != nil {
				return invalid("manual portfolio transaction %s names an import", t.Ref)
			}
		default:
			return invalid("portfolio transaction %s has origin %q", t.Ref, t.Origin)
		}
	}
	return nil
}

func (f *AssetsFile) validate() error {
	classes := make(map[string]struct{}, len(f.Classes))
	for _, c := range f.Classes {
		if strings.TrimSpace(c.Name) == "" {
			return invalid("an asset class has no name")
		}
		if _, dup := classes[c.Name]; dup {
			return invalid("asset class %q is repeated", c.Name)
		}
		classes[c.Name] = struct{}{}
		items := make(map[string]struct{}, len(c.Items))
		for _, i := range c.Items {
			if strings.TrimSpace(i.Name) == "" {
				return invalid("an item in asset class %q has no name", c.Name)
			}
			if _, dup := items[i.Name]; dup {
				return invalid("item %q in asset class %q is repeated", i.Name, c.Name)
			}
			items[i.Name] = struct{}{}
			for _, m := range i.Mutations {
				if m.ChangeType != assets.ChangeTypeSet && m.ChangeType != assets.ChangeTypeAdjust {
					return invalid("a mutation of item %q has change type %q", i.Name, m.ChangeType)
				}
			}
			for _, p := range i.Purchases {
				if p.Quantity <= 0 || p.UnitPrice < 0 {
					return invalid("a purchase of item %q has quantity %v at %d", i.Name, p.Quantity, p.UnitPrice)
				}
			}
		}
	}
	return nil
}

func (f *EODFile) validate() error {
	seen := make(map[ListingRef]struct{}, len(f.Listings))
	for _, l := range f.Listings {
		if strings.TrimSpace(l.Symbol) == "" || l.Source == "" {
			return invalid("a listing has no symbol or source")
		}
		if _, dup := seen[l.ListingRef]; dup {
			return invalid("listing %s (%s) is repeated", l.Symbol, l.Source)
		}
		seen[l.ListingRef] = struct{}{}
		if l.Instrument == nil {
			continue
		}
		if !l.Source.IsValid() {
			return invalid("listing %s has unknown source %q", l.Symbol, l.Source)
		}
		if strings.TrimSpace(l.Name) == "" {
			return invalid("listing %s (%s) has no name", l.Symbol, l.Source)
		}
		if c := l.Instrument.Currency; c != nil && !c.IsValid() {
			return invalid("listing %s (%s) has unknown currency %q", l.Symbol, l.Source, *c)
		}
	}
	return nil
}

func validDirection(d cashflow.CashFlowDirection) bool {
	return d == cashflow.CashIn || d == cashflow.CashOut
}
