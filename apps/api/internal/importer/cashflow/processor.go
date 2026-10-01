package cashflow

import (
	"context"
	"fmt"
	"strconv"

	cashflowdomain "github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/files"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

// Processor parses and persists accepted cashflow CSV imports.
type Processor struct {
	vendors  *vendor.Queries
	files    *files.Queries
	parsers  importer.CashflowParserFactory
	commands *cashflowdomain.Commands
}

// NewProcessor constructs the cashflow CSV import processor.
func NewProcessor(
	vendors *vendor.Queries,
	files *files.Queries,
	parsers importer.CashflowParserFactory,
	commands *cashflowdomain.Commands,
) *Processor {
	return &Processor{
		vendors:  vendors,
		files:    files,
		parsers:  parsers,
		commands: commands,
	}
}

// Process parses the cashflow CSV into a batch of transaction data and hands the whole
// batch to the cashflow commands, which persist it in a single bulk insert. Rows are
// stamped with the vendor as their source, their CSV row number, and the dedup
// sequence that lets a partly overlapping export be recognised as already imported.
func (p *Processor) Process(ctx context.Context, imp *importer.Import) (importer.ProcessResult, error) {
	accountID, err := imp.RequireAccountID()
	if err != nil {
		return importer.ProcessResult{}, err
	}

	v, err := p.vendors.GetById(ctx, imp.VendorID)
	if err != nil {
		return importer.ProcessResult{}, fmt.Errorf("fetch vendor: %w", err)
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

	source := string(v.Name)
	sequencer := importer.NewDedupSequencer()
	data := make([]cashflowdomain.TransactionData, 0)
	for rowNumber, row := range rows {
		row.Source = source
		row.RowNumber = rowNumber
		// The note carries the vendor's own reference (DEGIRO's order id), so the
		// content key already distinguishes two same-day transfers of equal size.
		row.DedupSeq = sequencer.Next(
			row.Date.Format("20060102"),
			row.Description,
			row.Note,
			string(row.Direction),
			strconv.FormatInt(int64(row.Amount), 10),
		)
		data = append(data, row)
	}
	if len(data) == 0 {
		return importer.ProcessResult{}, nil
	}

	result, err := p.commands.CreateMany(ctx, accountID, &imp.ID, data)
	if err != nil {
		return importer.ProcessResult{TotalRows: len(data), Failed: len(data)}, fmt.Errorf("import cashflow: %w", err)
	}
	return importer.ProcessResult{
		TotalRows:  len(data),
		Imported:   result.Imported,
		Duplicates: result.Duplicates,
	}, nil
}
