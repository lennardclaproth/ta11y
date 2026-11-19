package importer

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

// QueryStore reads import records.
type QueryStore interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*Import, error)
}

// unlinkedProductReader names the products an import brought in that no listing
// matches. It is the portfolio feature's answer; the importer only reports it.
type unlinkedProductReader interface {
	UnlinkedProducts(ctx context.Context, accountID, importID uuid.UUID) ([]portfolio.ImportedProduct, error)
}

// Queries exposes CSV import read-side use cases.
type Queries struct {
	qs        QueryStore
	portfolio unlinkedProductReader
}

// NewQueries creates CSV import read-side use cases.
func NewQueries(qs QueryStore, pf unlinkedProductReader) *Queries {
	return &Queries{qs: qs, portfolio: pf}
}

// ImportResult is one import's outcome: its lifecycle state, the counters the
// processor produced, and — for a portfolio import — the products it brought in that
// no listing matches and that therefore do not count towards performance yet.
type ImportResult struct {
	Import           *Import
	UnlinkedProducts []portfolio.ImportedProduct
}

// FailureReasonNotRecognised marks an import the vendor's parser refused because the
// file is not the export that vendor produces, which is worth a different answer than
// a generic failure: a different file, not a retry.
const FailureReasonNotRecognised = "file_not_recognised"

// Result returns one import's outcome, scoped to the account that uploaded it.
// Imports that belong to another account, or that have no account at all (EOD uploads
// are listing-scoped), are reported as not found rather than leaked.
func (q *Queries) Result(ctx context.Context, importID, accountID uuid.UUID) (*ImportResult, error) {
	imp, err := q.qs.FetchByID(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("import result: %w", err)
	}
	if imp == nil || imp.AccountID == nil || *imp.AccountID != accountID {
		return nil, ErrImportNotFound
	}

	result := &ImportResult{Import: imp, UnlinkedProducts: []portfolio.ImportedProduct{}}
	// Only a finished portfolio import has rows to look at; asking earlier would
	// report every product as unlinked simply because nothing is stored yet.
	if imp.Type != ImportTypePortfolio || imp.Status != ImportStatusCompleted || q.portfolio == nil {
		return result, nil
	}
	products, err := q.portfolio.UnlinkedProducts(ctx, accountID, imp.ID)
	if err != nil {
		return nil, fmt.Errorf("import result: %w", err)
	}
	result.UnlinkedProducts = products
	return result, nil
}

// FailureReason classifies a failed import so a caller can answer it, rather than
// repeating a wrapped Go error at the user. An empty string means the failure has no
// classification beyond its message.
func FailureReason(imp *Import) string {
	if imp == nil || imp.Status != ImportStatusFailed {
		return ""
	}
	if strings.Contains(imp.StatusMsg, ErrImportFileNotRecognised.Error()) {
		return FailureReasonNotRecognised
	}
	return ""
}
