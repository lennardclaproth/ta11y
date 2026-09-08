package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/auth"
)

// SQLXAuthStore persists federated identities and server-side sessions. It satisfies
// the auth feature's storage contracts.
type SQLXAuthStore struct {
	db              *DB
	identitiesTable string
	sessionsTable   string
}

var (
	_ auth.IdentityStore = (*SQLXAuthStore)(nil)
	_ auth.SessionStore  = (*SQLXAuthStore)(nil)
)

// NewSQLXAuthStore creates an authentication store backed by SQLX.
func NewSQLXAuthStore(db *DB) *SQLXAuthStore {
	return &SQLXAuthStore{
		db:              db,
		identitiesTable: qualifyTable(db, SchemaAccount, TableIdentities),
		sessionsTable:   qualifyTable(db, SchemaAccount, TableSessions),
	}
}

// GetByIssuerSubject returns the identity a provider's (issuer, subject) pair maps to,
// or auth.ErrIdentityNotFound when this is a first login through that provider.
func (s *SQLXAuthStore) GetByIssuerSubject(ctx context.Context, issuer, subject string) (*auth.Identity, error) {
	var identity auth.Identity
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT id, account_id, issuer, subject, created_at, last_login_at
		FROM %s WHERE issuer = ? AND subject = ?
	`, s.identitiesTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &identity, query, issuer, subject); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrIdentityNotFound
		}
		return nil, err
	}
	return &identity, nil
}

// CreateIdentity inserts an identity, mapping a unique-constraint violation on
// (issuer, subject) to auth.ErrIdentityExists so a concurrent first login is not
// reported as a server fault.
func (s *SQLXAuthStore) CreateIdentity(ctx context.Context, identity *auth.Identity) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, issuer, subject, created_at, last_login_at)
		VALUES (:id, :account_id, :issuer, :subject, :created_at, :last_login_at)
	`, s.identitiesTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, identity); err != nil {
		if isUniqueViolation(err) {
			return auth.ErrIdentityExists
		}
		return err
	}
	return nil
}

// TouchLastLogin records that an identity was used to sign in.
func (s *SQLXAuthStore) TouchLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET last_login_at = ? WHERE id = ?`, s.identitiesTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, at, id)
	return err
}

// CreateSession inserts a session row. Only the token's hash is stored.
func (s *SQLXAuthStore) CreateSession(ctx context.Context, session *auth.Session) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, token_hash, issued_at, expires_at, last_seen_at, revoked_at)
		VALUES (:id, :account_id, :token_hash, :issued_at, :expires_at, :last_seen_at, :revoked_at)
	`, s.sessionsTable)
	_, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, session)
	return err
}

// GetByTokenHash returns the session a presented token hashes to, or
// auth.ErrSessionNotFound.
func (s *SQLXAuthStore) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.Session, error) {
	var session auth.Session
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT id, account_id, token_hash, issued_at, expires_at, last_seen_at, revoked_at
		FROM %s WHERE token_hash = ?
	`, s.sessionsTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &session, query, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

// Revoke ends one session. Revoking an unknown or already-revoked session is a no-op.
func (s *SQLXAuthStore) Revoke(ctx context.Context, tokenHash string, at time.Time) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`, s.sessionsTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, at, tokenHash)
	return err
}

// RevokeAllForAccount ends every live session an account holds.
func (s *SQLXAuthStore) RevokeAllForAccount(ctx context.Context, accountID uuid.UUID, at time.Time) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET revoked_at = ? WHERE account_id = ? AND revoked_at IS NULL`, s.sessionsTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, at, accountID)
	return err
}

// TouchLastSeen records recent activity on a session.
func (s *SQLXAuthStore) TouchLastSeen(ctx context.Context, id uuid.UUID, at time.Time) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET last_seen_at = ? WHERE id = ?`, s.sessionsTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, at, id)
	return err
}

// DeleteExpiredSessions removes sessions that lapsed before the cutoff, so the table
// does not grow without bound. Revoked sessions are kept until they also expire.
func (s *SQLXAuthStore) DeleteExpiredSessions(ctx context.Context, before time.Time) (int64, error) {
	query := s.db.Rebind(fmt.Sprintf(`DELETE FROM %s WHERE expires_at < ?`, s.sessionsTable))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
