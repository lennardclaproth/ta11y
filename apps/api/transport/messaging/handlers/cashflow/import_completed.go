// Package cashflow holds event handlers that react to imports on behalf of cashflow.
package cashflow

import (
	"context"

	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/logging"
)

// ImportCompletedHandler attaches the rows of a finished cashflow import to the
// recurring items the account already confirmed, so a recurring payment has to be
// pointed at once rather than after every statement. It never creates an item:
// only items confirmed by the user take part, and an ended item takes nothing.
type ImportCompletedHandler struct {
	commands *cashflow.RecurringCommands
	logger   logging.Logger
}

// NewImportCompletedHandler constructs an ImportCompletedHandler.
func NewImportCompletedHandler(commands *cashflow.RecurringCommands, logger logging.Logger) *ImportCompletedHandler {
	return &ImportCompletedHandler{commands: commands, logger: logger}
}

// Handle links the transactions of a completed cashflow import. Other import
// types carry no cashflow rows and are ignored.
func (h *ImportCompletedHandler) Handle(ctx context.Context, evt importer.Completed, _ eventbus.Metadata) error {
	if evt.Type != importer.ImportTypeCashflow || evt.AccountID == nil {
		return nil
	}

	linked, err := h.commands.LinkImported(ctx, *evt.AccountID, evt.ImportID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "linking imported transactions to recurring items failed", err,
				"account_id", evt.AccountID.String(), "import_id", evt.ImportID.String())
		}
		return err
	}
	if linked > 0 && h.logger != nil {
		h.logger.Info(ctx, "linked imported transactions to recurring items",
			"account_id", evt.AccountID.String(), "import_id", evt.ImportID.String(), "linked", linked)
	}
	return nil
}
