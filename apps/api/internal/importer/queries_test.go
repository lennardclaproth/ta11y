package importer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

type stubImportStore struct {
	imp *Import
	err error
}

func (s *stubImportStore) FetchByID(_ context.Context, _ uuid.UUID) (*Import, error) {
	return s.imp, s.err
}

type stubProductReader struct {
	products []portfolio.ImportedProduct
	calls    int
	err      error
}

func (s *stubProductReader) UnlinkedProducts(_ context.Context, _, _ uuid.UUID) ([]portfolio.ImportedProduct, error) {
	s.calls++
	return s.products, s.err
}

func completedPortfolioImport(accountID uuid.UUID) *Import {
	account := accountID
	return &Import{
		ID:        uuid.New(),
		Type:      ImportTypePortfolio,
		Status:    ImportStatusCompleted,
		AccountID: &account,
	}
}

func TestResult_ReportsUnlinkedProductsForACompletedPortfolioImport(t *testing.T) {
	accountID := uuid.New()
	imp := completedPortfolioImport(accountID)
	products := &stubProductReader{products: []portfolio.ImportedProduct{{Name: "Example Fund", Transactions: 3}}}
	queries := NewQueries(&stubImportStore{imp: imp}, products)

	result, err := queries.Result(context.Background(), imp.ID, accountID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.UnlinkedProducts) != 1 {
		t.Fatalf("expected the import's unlinked products, got %d", len(result.UnlinkedProducts))
	}
}

func TestResult_AnotherAccountsImportIsNotFound(t *testing.T) {
	// Imports are account data: answering with somebody else's counts would leak them.
	imp := completedPortfolioImport(uuid.New())
	queries := NewQueries(&stubImportStore{imp: imp}, &stubProductReader{})

	_, err := queries.Result(context.Background(), imp.ID, uuid.New())
	if !errors.Is(err, ErrImportNotFound) {
		t.Fatalf("expected ErrImportNotFound, got %v", err)
	}
}

func TestResult_AnImportWithoutAnAccountIsNotFound(t *testing.T) {
	// EOD uploads are listing-scoped and belong to no account, so this route never
	// serves them rather than serving them to whoever asks first.
	imp := &Import{ID: uuid.New(), Type: ImportTypeEOD, Status: ImportStatusCompleted}
	queries := NewQueries(&stubImportStore{imp: imp}, &stubProductReader{})

	_, err := queries.Result(context.Background(), imp.ID, uuid.New())
	if !errors.Is(err, ErrImportNotFound) {
		t.Fatalf("expected ErrImportNotFound, got %v", err)
	}
}

func TestResult_MissingImportIsNotFound(t *testing.T) {
	queries := NewQueries(&stubImportStore{}, &stubProductReader{})

	_, err := queries.Result(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrImportNotFound) {
		t.Fatalf("expected ErrImportNotFound, got %v", err)
	}
}

func TestResult_DoesNotAskForProductsBeforeThePortfolioImportFinished(t *testing.T) {
	// Nothing is stored yet while the import is running, so asking now would report
	// every product as unlinked.
	accountID := uuid.New()
	imp := completedPortfolioImport(accountID)
	imp.Status = ImportStatusProcessing
	products := &stubProductReader{}
	queries := NewQueries(&stubImportStore{imp: imp}, products)

	result, err := queries.Result(context.Background(), imp.ID, accountID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if products.calls != 0 {
		t.Fatalf("expected no product lookup for an unfinished import, got %d", products.calls)
	}
	if result.UnlinkedProducts == nil {
		t.Fatalf("expected an empty list rather than nil")
	}
}

func TestResult_DoesNotAskForProductsForACashflowImport(t *testing.T) {
	accountID := uuid.New()
	imp := completedPortfolioImport(accountID)
	imp.Type = ImportTypeCashflow
	products := &stubProductReader{}
	queries := NewQueries(&stubImportStore{imp: imp}, products)

	if _, err := queries.Result(context.Background(), imp.ID, accountID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if products.calls != 0 {
		t.Fatalf("expected no product lookup for a cashflow import, got %d", products.calls)
	}
}

func TestFailureReason_ClassifiesAnUnreadableFile(t *testing.T) {
	// A file the parser refuses needs a different answer than a generic failure: a
	// different file, not a retry.
	imp := &Import{Status: ImportStatusFailed}
	imp.StatusMsg = fmt.Errorf("%w: missing required header: Date", ErrImportFileNotRecognised).Error()

	if got := FailureReason(imp); got != FailureReasonNotRecognised {
		t.Fatalf("expected %q, got %q", FailureReasonNotRecognised, got)
	}
}

func TestFailureReason_LeavesOtherOutcomesUnclassified(t *testing.T) {
	failedForAnotherReason := &Import{Status: ImportStatusFailed, StatusMsg: "import portfolio: create many: boom"}
	if got := FailureReason(failedForAnotherReason); got != "" {
		t.Fatalf("expected no classification, got %q", got)
	}

	completed := &Import{Status: ImportStatusCompleted}
	if got := FailureReason(completed); got != "" {
		t.Fatalf("expected no classification for a completed import, got %q", got)
	}
}
