package portfolio

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lennardclaproth/ta11y/internal/files"
	"github.com/lennardclaproth/ta11y/internal/importer"
	portfoliodomain "github.com/lennardclaproth/ta11y/internal/portfolio"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

// Processor parses and persists accepted portfolio CSV imports.
type Processor struct {
	vendors  *vendor.Queries
	files    *files.Queries
	parsers  importer.PortfolioParserFactory
	commands *portfoliodomain.Commands
}

// NewProcessor constructs the portfolio CSV import processor.
func NewProcessor(
	vendors *vendor.Queries,
	files *files.Queries,
	parsers importer.PortfolioParserFactory,
	commands *portfoliodomain.Commands,
) *Processor {
	return &Processor{
		vendors:  vendors,
		files:    files,
		parsers:  parsers,
		commands: commands,
	}
}

// Process parses the portfolio CSV into a batch of transaction data and hands the whole
// batch to the portfolio commands, which persist it in a single bulk insert. It requires
// a brokerage vendor and stamps each row with its CSV row number plus the dedup sequence
// that lets a partly overlapping export be recognised as already imported.
func (p *Processor) Process(ctx context.Context, imp *importer.Import) (importer.ProcessResult, error) {
	accountID, err := imp.RequireAccountID()
	if err != nil {
		return importer.ProcessResult{}, err
	}

	v, err := p.vendors.GetById(ctx, imp.VendorID)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("fetch vendor: %w", err)
	}
	if v.Type != vendor.VendorTypeBrokerage {
		return importer.ProcessResult{}, importer.ErrVendorNotBrokerage
	}
	parser, err := p.parsers(v.Name)
	if err != nil {
		return importer.ProcessResult{}, err
	}
	rc, err := p.files.ReadCsv(ctx, imp.Path)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("open csv: %w", err)
	}
	// The file is only read; a close error tells us nothing actionable and must not
	// mask the processing result.
	defer func() { _ = rc.Close() }()

	rows, err := parser.ParseAll(rc)
	if err != nil {
		// The parser only refuses the whole file when its headers are not the ones the
		// vendor's export carries, so this is the wrong export rather than a bad row.
		return importer.ProcessResult{}, fmt.Errorf("%w: %v", importer.ErrImportFileNotRecognised, err)
	}

	sequencer := importer.NewDedupSequencer()
	data := make([]portfoliodomain.TransactionData, 0)
	for rowNumber, row := range rows {
		row.RowNumber = rowNumber
		row.DedupSeq = sequencer.Next(
			row.OccurredAt.Format("20060102"),
			derefString(row.ISIN),
			derefString(row.Symbol),
			row.Description,
			string(row.Type),
			strconv.FormatFloat(row.Quantity, 'f', 8, 64),
			strconv.FormatFloat(row.Price, 'f', 8, 64),
			strconv.FormatFloat(row.Amount, 'f', 8, 64),
			row.ExternalRef,
		)
		data = append(data, row)
	}
	if len(data) == 0 {
		return importer.ProcessResult{}, nil
	}

	result, err := p.commands.CreateMany(ctx, imp.ID, &accountID, data)
	if err != nil {
		return importer.ProcessResult{TotalRows: len(data), Failed: len(data)}, fmt.Errorf("import portfolio: %w", err)
	}
	return importer.ProcessResult{
		TotalRows:  len(data),
		Imported:   result.Imported,
		Duplicates: result.Duplicates,
	}, nil
}

// derefString reads an optional identifier into the dedup content key, where a missing
// value has to compare equal across exports rather than blow up.
func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
