package assets

import (
	"context"

	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	"github.com/lennardclaproth/ta11y/internal/logging"
)

// SnapshotsRebuildRequestedHandler rebuilds account-level assets snapshots in
// response to a rebuild request (published by the assets write side after a
// class or asset mutation) and announces completion so the realtime layer can
// notify clients.
type SnapshotsRebuildRequestedHandler struct {
	holdings *assets.HoldingsSyncer
	builder  *assets.Builder
	bus      eventbus.Bus
	logger   logging.Logger
}

// NewSnapshotsRebuildRequestedHandler constructs a SnapshotsRebuildRequestedHandler.
// The bus may be nil, in which case the SnapshotsRebuilt event is not published;
// the holdings syncer may be nil, in which case daily-priced items are not refreshed.
func NewSnapshotsRebuildRequestedHandler(
	holdings *assets.HoldingsSyncer,
	builder *assets.Builder,
	bus eventbus.Bus,
	logger logging.Logger,
) *SnapshotsRebuildRequestedHandler {
	return &SnapshotsRebuildRequestedHandler{holdings: holdings, builder: builder, bus: bus, logger: logger}
}

// Handle refreshes the account's daily-priced items, rebuilds its assets
// snapshots, then publishes assets.SnapshotsRebuilt on success.
//
// The holdings sync runs first because it writes the mutations the snapshot
// rebuild reads: running them the other way round would date every snapshot one
// request behind the prices it is supposed to reflect.
func (h *SnapshotsRebuildRequestedHandler) Handle(ctx context.Context, evt assets.SnapshotsRebuildRequested, _ eventbus.Metadata) error {
	if h.holdings != nil {
		if err := h.holdings.SyncAccount(ctx, evt.AccID); err != nil {
			if h.logger != nil {
				h.logger.Error(ctx, "assets holdings sync failed", err, "account_id", evt.AccID.String())
			}
			return err
		}
	}
	if err := h.builder.RebuildAll(ctx, evt.AccID); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "assets snapshots rebuild failed", err, "account_id", evt.AccID.String())
		}
		return err
	}
	publishSnapshotsRebuilt(ctx, h.bus, evt.AccID)
	return nil
}
