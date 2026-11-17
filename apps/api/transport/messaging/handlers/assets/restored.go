package assets

import (
	"context"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
)

// RestoredHandler rebuilds an account's assets snapshots, and the mutations of its
// daily-priced items, after an assets backup brought items in. It runs the same rebuild
// a rebuild request does.
type RestoredHandler struct {
	rebuild *SnapshotsRebuildRequestedHandler
}

// NewRestoredHandler constructs a RestoredHandler around the snapshot rebuild.
func NewRestoredHandler(rebuild *SnapshotsRebuildRequestedHandler) *RestoredHandler {
	return &RestoredHandler{rebuild: rebuild}
}

// Handle rebuilds the restored account's assets snapshots.
func (h *RestoredHandler) Handle(ctx context.Context, evt backup.Restored, meta eventbus.Metadata) error {
	return h.rebuild.Handle(ctx, assets.SnapshotsRebuildRequested{AccID: evt.AccID}, meta)
}
