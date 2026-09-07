-- +goose Up
-- +goose StatementBegin

-- Accounts carry an explicit admin flag so admin-only screens (listings, dailies,
-- provider credentials) can be hidden from accounts that should not see them. The
-- bootstrapped account is promoted to admin; every other account defaults to false.
--
-- Note: the API is unauthenticated today, so this flag describes intent rather than
-- enforcing it. It must be paired with real authentication before the API is exposed
-- beyond localhost.
ALTER TABLE account.accounts ADD COLUMN admin BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE account.accounts DROP COLUMN IF EXISTS admin;

-- +goose StatementEnd
