-- +goose Up
-- +goose StatementBegin

-- Accounts carry an explicit admin flag so admin-only screens (listings, dailies,
-- provider credentials) can be hidden from accounts that should not see them. The
-- bootstrapped account is promoted to admin; every other account defaults to false.
--
-- Note: written before authentication existed, when the flag only described intent.
-- Since [030] the API enforces it -- admin-only routes are registered through the
-- adminOnly tier and answer 403 -- so hiding the screens is a convenience, not the
-- control. Comment corrected in place; the DDL below is unchanged.
ALTER TABLE account.accounts ADD COLUMN admin BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE account.accounts DROP COLUMN IF EXISTS admin;

-- +goose StatementEnd
