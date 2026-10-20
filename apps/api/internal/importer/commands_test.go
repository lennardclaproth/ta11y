package importer

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

type recordingImportStore struct {
	imp *Import
}

func (s *recordingImportStore) Create(_ context.Context, _ *Import) error { return nil }

func (s *recordingImportStore) FetchByID(_ context.Context, _ uuid.UUID) (*Import, error) {
	return s.imp, nil
}

func (s *recordingImportStore) UpdateState(_ context.Context, _ *Import) error { return nil }

type recordingRemover struct {
	removed []string
}

func (r *recordingRemover) Remove(path string) error {
	r.removed = append(r.removed, path)
	return nil
}

type stubProcessor struct {
	result ProcessResult
	err    error
}

func (p stubProcessor) Process(_ context.Context, _ *Import) (ProcessResult, error) {
	return p.result, p.err
}

func processCommands(imp *Import, processor Processor, remover FileRemover) *Commands {
	return NewCommands(
		&recordingImportStore{imp: imp},
		nil,
		remover,
		vendor.Queries{},
		account.Queries{},
		marketdata.Queries{},
		nil,
		WithProcessors(processor, processor, processor),
	)
}

func pendingImport() *Import {
	accountID := uuid.New()
	return &Import{
		ID:        uuid.New(),
		Type:      ImportTypeCashflow,
		Status:    ImportStatusPending,
		Path:      "uploads/example.csv",
		AccountID: &accountID,
	}
}

func TestProcess_RemovesTheUploadOnceItCompletes(t *testing.T) {
	// The rows have been taken over by the target feature and nothing reads the file
	// again, so a copy of somebody's account statement must not stay on disk.
	imp := pendingImport()
	remover := &recordingRemover{}
	commands := processCommands(imp, stubProcessor{result: ProcessResult{TotalRows: 2, Imported: 2}}, remover)

	if err := commands.Process(context.Background(), imp.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(remover.removed) != 1 || remover.removed[0] != imp.Path {
		t.Fatalf("expected the upload to be removed, got %v", remover.removed)
	}
}

func TestProcess_RemovesTheUploadWhenItFails(t *testing.T) {
	// A refused file is exactly the one nobody will look at again.
	imp := pendingImport()
	remover := &recordingRemover{}
	commands := processCommands(imp, stubProcessor{err: fmt.Errorf("%w: missing required header: Date", ErrImportFileNotRecognised)}, remover)

	if err := commands.Process(context.Background(), imp.ID); err == nil {
		t.Fatalf("expected the processing error to be returned")
	}
	if len(remover.removed) != 1 || remover.removed[0] != imp.Path {
		t.Fatalf("expected the upload to be removed, got %v", remover.removed)
	}
	if reason := FailureReason(imp); reason != FailureReasonNotRecognised {
		t.Fatalf("expected the failure to be classified as %q, got %q", FailureReasonNotRecognised, reason)
	}
}

func TestProcess_KeepsTheCountersOfAFailedImport(t *testing.T) {
	// A cashflow import whose ignore-rule pass fails has already inserted its rows, and
	// the rules that ran before the failure have already taken rows out of the monthly
	// totals. The counters a failing processor reports therefore have to be written as
	// they are, or the import review would claim nothing was ignored while rows were.
	imp := pendingImport()
	result := ProcessResult{TotalRows: 40, Imported: 30, Duplicates: 10, AutoIgnored: 7}
	commands := processCommands(imp, stubProcessor{result: result, err: fmt.Errorf("apply ignore rules: boom")}, &recordingRemover{})

	if err := commands.Process(context.Background(), imp.ID); err == nil {
		t.Fatalf("expected the processing error to be returned")
	}
	if imp.Status != ImportStatusFailed {
		t.Fatalf("status = %q, want %q", imp.Status, ImportStatusFailed)
	}
	if imp.Imported != 30 || imp.Duplicates != 10 || imp.AutoIgnored != 7 {
		t.Fatalf("imported=%d duplicates=%d auto_ignored=%d, want the counters the processor reported", imp.Imported, imp.Duplicates, imp.AutoIgnored)
	}
}

func TestProcess_LeavesAnImportThatIsNotPendingAlone(t *testing.T) {
	// Redelivery must not re-run work or delete an upload a live run still needs.
	imp := pendingImport()
	imp.Status = ImportStatusProcessing
	remover := &recordingRemover{}
	commands := processCommands(imp, stubProcessor{}, remover)

	if err := commands.Process(context.Background(), imp.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(remover.removed) != 0 {
		t.Fatalf("expected nothing to be removed, got %v", remover.removed)
	}
}
