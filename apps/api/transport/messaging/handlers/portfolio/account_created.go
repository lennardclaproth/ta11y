package portfolio

import (
	"context"

	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
)

// AccountCreatedHandler creates the portfolio projection for a new account. The
// projection holds portfolio-specific per-account state (the rebuild lock), so
// it must exist before a portfolio can be built for the account.
type AccountCreatedHandler struct {
	commands *portfolio.Commands
	logger   logging.Logger
}

// NewAccountCreatedHandler constructs an AccountCreatedHandler.
func NewAccountCreatedHandler(commands *portfolio.Commands, logger logging.Logger) *AccountCreatedHandler {
	return &AccountCreatedHandler{commands: commands, logger: logger}
}

// Handle creates the portfolio projection for the newly created account.
func (h *AccountCreatedHandler) Handle(ctx context.Context, evt account.Created, _ eventbus.Metadata) error {
	if _, err := h.commands.CreateAccount(ctx, evt.AccID); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "portfolio account projection creation failed", err, "account_id", evt.AccID.String())
		}
		return err
	}
	return nil
}
