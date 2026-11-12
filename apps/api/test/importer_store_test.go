//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/storage"
	"github.com/lennardclaproth/ta11y/internal/vendor"
)

// TestImporterListUnfinished checks that start-up recovery sees only pending or
// processing imports that have been idle since before the cutoff.
func TestImporterListUnfinished(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		now := time.Now().UTC()
		v := &vendor.Vendor{ID: uuid.New(), Name: vendor.VendorING, Active: true, Type: vendor.VendorTypeBank, CreatedAt: now, UpdatedAt: now}
		if err := storage.NewSQLXVendorStore(db).Create(ctx, v); err != nil {
			t.Fatalf("create vendor: %v", err)
		}
		store := storage.NewSQLXImporterStore(db)

		create := func(status importer.ImportStatus, updatedAt time.Time) uuid.UUID {
			t.Helper()
			imp := &importer.Import{
				ID: uuid.New(), Type: importer.ImportTypeCashflow, VendorID: v.ID, Path: "upload.csv",
				Status: status, CreatedAt: updatedAt, UpdatedAt: updatedAt,
			}
			if err := store.Create(ctx, imp); err != nil {
				t.Fatalf("create import: %v", err)
			}
			return imp.ID
		}
		stalePending := create(importer.ImportStatusPending, now.Add(-time.Hour))
		staleProcessing := create(importer.ImportStatusProcessing, now.Add(-time.Hour))
		create(importer.ImportStatusPending, now)
		create(importer.ImportStatusCompleted, now.Add(-time.Hour))
		create(importer.ImportStatusFailed, now.Add(-time.Hour))

		unfinished, err := store.ListUnfinished(ctx, now.Add(-time.Minute))
		if err != nil {
			t.Fatalf("list unfinished: %v", err)
		}
		got := map[uuid.UUID]bool{}
		for _, imp := range unfinished {
			got[imp.ID] = true
		}
		if len(unfinished) != 2 || !got[stalePending] || !got[staleProcessing] {
			t.Fatalf("expected only the two stale unfinished imports, got %d rows", len(unfinished))
		}
	})
}
