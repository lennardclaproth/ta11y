package portfolio

import (
	"context"
	"errors"

	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

// RestoredHandler rebuilds an account's portfolio after a portfolio backup brought
// transactions in: the file carries transactions only, so positions and snapshots are
// derived here. A
// successful rebuild publishes portfolio.Rebuilt, which drives the assets sync.
type RestoredHandler struct {
	builder *portfolio.Builder
	logger  logging.Logger
}

// NewRestoredHandler constructs a RestoredHandler.
func NewRestoredHandler(builder *portfolio.Builder, logger logging.Logger) *RestoredHandler {
	return &RestoredHandler{builder: builder, logger: logger}
}

// Handle rebuilds the restored account's portfolio.
func (h *RestoredHandler) Handle(ctx context.Context, evt backup.Restored, _ eventbus.Metadata) error {
	// An account with no buildable positions yields no snapshots; that is an expected
	// outcome, not a failure.
	if err := h.builder.Build(ctx, evt.AccID); err != nil && !errors.Is(err, portfolio.ErrPortfolioNoSnapshots) {
		if h.logger != nil {
			h.logger.Error(ctx, "portfolio rebuild after restore failed", err, "account_id", evt.AccID.String())
		}
		return err
	}
	return nil
}
