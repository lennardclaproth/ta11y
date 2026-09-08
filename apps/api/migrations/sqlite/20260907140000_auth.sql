-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20260907140000_auth.sql. SQLite is schemaless, so
-- account.identities/account.sessions flatten to bare identities/sessions. See the
-- Postgres mirror for what is stored and, more importantly, what is not.
ALTER TABLE accounts ADD COLUMN email TEXT;
CREATE UNIQUE INDEX uq_accounts_email_lower ON accounts (LOWER(email));

CREATE TABLE identities (
    id            TEXT PRIMARY KEY,
    account_id    TEXT     NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    issuer        TEXT     NOT NULL,
    subject       TEXT     NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at DATETIME,
    CONSTRAINT uq_identities_issuer_subject UNIQUE (issuer, subject)
);

CREATE INDEX idx_identities_account ON identities (account_id);

CREATE TABLE sessions (
    id           TEXT PRIMARY KEY,
    account_id   TEXT     NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash   TEXT     NOT NULL UNIQUE,
    issued_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at   DATETIME NOT NULL,
    last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at   DATETIME
);

CREATE INDEX idx_sessions_account ON sessions (account_id);
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS identities;
DROP INDEX IF EXISTS uq_accounts_email_lower;
ALTER TABLE accounts DROP COLUMN email;

-- +goose StatementEnd
