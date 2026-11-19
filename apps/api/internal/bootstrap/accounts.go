package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/auth"
	"github.com/lennardclaproth/ta11y/internal/logging"
)

var (
	defaultAccountID   = uuid.MustParse("64a3d50f-c71a-4015-9ee1-45572147ce56")
	defaultAccountName = "Lennard Claproth"
)

// Accounts bootstraps the default account through the account feature boundary.
//
// adminEmail, when set, is stamped on the account so the first sign-in with that
// address adopts it by email instead of provisioning a fresh one -- without which every
// import, portfolio and asset already attached to the seeded account would be orphaned
// the moment authentication is switched on.
func Accounts(ctx context.Context, commands *account.Commands, queries *account.Queries, adminEmail string, logger logging.Logger) {
	if commands == nil || queries == nil {
		panic(fmt.Errorf("bootstrap accounts: account commands/queries are required"))
	}

	if existing, err := queries.GetByID(ctx, defaultAccountID); err == nil {
		claimBootstrapEmail(ctx, commands, existing, adminEmail, logger)
		logger.Info(ctx, "account already exists, skipping bootstrap", "account_id", defaultAccountID.String())
		return
	} else if !errors.Is(err, account.ErrAccountNotFound) {
		panic(fmt.Errorf("bootstrap accounts: fetch by id %s: %w", defaultAccountID, err))
	}

	options := []account.AccountOption{account.AsAdmin()}
	if adminEmail != "" {
		options = append(options, account.WithEmail(adminEmail))
	}

	if _, err := commands.Create(ctx, &defaultAccountID, nil, defaultAccountName, options...); err != nil {
		if errors.Is(err, account.ErrAccountAlreadyExists) {
			logger.Info(ctx, "account already exists by unique constraint, skipping bootstrap", "account_name", defaultAccountName)
			return
		}
		panic(fmt.Errorf("bootstrap accounts: create account %s: %w", defaultAccountName, err))
	}

	logger.Info(ctx, "bootstrapped account", "account_id", defaultAccountID.String(), "name", defaultAccountName)
}

// claimBootstrapEmail attaches the configured admin email to an already-seeded account
// that predates authentication. It never overwrites an email the account already has.
func claimBootstrapEmail(
	ctx context.Context,
	commands *account.Commands,
	existing *account.Account,
	adminEmail string,
	logger logging.Logger,
) {
	if adminEmail == "" || existing.Email != nil {
		return
	}
	if err := commands.SetEmail(ctx, existing.ID, adminEmail); err != nil {
		if errors.Is(err, account.ErrAccountAlreadyExists) {
			logger.Warn(ctx, "bootstrap admin email is already claimed by another account",
				"account_id", existing.ID.String())
			return
		}
		panic(fmt.Errorf("bootstrap accounts: claim admin email: %w", err))
	}
	logger.Info(ctx, "claimed bootstrap admin email", "account_id", existing.ID.String())
}

// FallbackPrincipal returns the bootstrapped account as an authenticated caller, for
// running with sign-in switched off. It exists so the unauthenticated mode is one
// explicit, logged decision at start-up rather than an implicit hole in the middleware.
func FallbackPrincipal(ctx context.Context, queries *account.Queries) (*auth.Principal, error) {
	acc, err := queries.GetByID(ctx, defaultAccountID)
	if err != nil {
		return nil, fmt.Errorf("bootstrap fallback: load account %s: %w", defaultAccountID, err)
	}
	email := ""
	if acc.Email != nil {
		email = *acc.Email
	}
	return &auth.Principal{AccountID: acc.ID, Email: email, Admin: acc.Admin}, nil
}
