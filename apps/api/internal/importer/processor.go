package importer

import (
	"context"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
	eodparsers "github.com/lennardclaproth/ta11y/internal/importer/eod/parsers"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

// CashflowParserFactory resolves a CSV parser for cashflow imports by vendor.
type CashflowParserFactory func(id vendor.VendorID) (cashflow.CsvParser, error)

// PortfolioParserFactory resolves a CSV parser for portfolio imports by vendor.
type PortfolioParserFactory func(id vendor.VendorID) (portfolio.CsvParser, error)

// EODParserFactory resolves a CSV parser for EOD imports by listing source.
type EODParserFactory func(source marketdata.Source) (eodparsers.EODParser, error)

// ProcessResult captures counters produced by one import processor.
type ProcessResult struct {
	TotalRows  int
	Imported   int
	Duplicates int
	Failed     int
}

// Processor handles the type-specific parsing and persistence for an import.
type Processor interface {
	Process(ctx context.Context, imp *Import) (ProcessResult, error)
}
