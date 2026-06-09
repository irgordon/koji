package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "koji_session"
	CSRFCookieName    = "koji_csrf"
	CSRFHeaderName    = "X-CSRF-Token"
)

var (
	ErrBootstrapDisabled = errors.New("bootstrap disabled")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrInvalidSession    = errors.New("invalid session")
	ErrInvalidCSRF       = errors.New("invalid csrf token")
)

type Store struct {
	db     *sql.DB
	policy SessionPolicy
}

type SessionPolicy struct {
	TTL         time.Duration
	IdleTimeout time.Duration
}

type Session struct {
	ID         string
	CSRFToken  string
	UserID     int64
	Username   string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
}

type Principal struct {
	UserID   int64
	Username string
}

func NewStore(db *sql.DB) *Store {
	return NewStoreWithPolicy(db, DefaultSessionPolicy())
}

func NewStoreWithPolicy(db *sql.DB, policy SessionPolicy) *Store {
	return &Store{
		db:     db,
		policy: normalizedPolicy(policy),
	}
}

func DefaultSessionPolicy() SessionPolicy {
	return SessionPolicy{
		TTL:         12 * time.Hour,
		IdleTimeout: 30 * time.Minute,
	}
}

func (s *Store) HasUsers(ctx context.Context) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE disabled_at IS NULL").Scan(&count); err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return count > 0, nil
}

func (s *Store) Bootstrap(ctx context.Context, username string, password string) (Session, error) {
	if err := validateCredentialInput(username, password); err != nil {
		return Session{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, fmt.Errorf("begin bootstrap: %w", err)
	}
	defer rollback(tx)

	hasUsers, err := txHasUsers(ctx, tx)
	if err != nil {
		return Session{}, err
	}
	if hasUsers {
		return Session{}, ErrBootstrapDisabled
	}
	if err := txClaimBootstrap(ctx, tx); err != nil {
		return Session{}, err
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return Session{}, err
	}

	result, err := tx.ExecContext(ctx, "INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		return Session{}, fmt.Errorf("create bootstrap user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return Session{}, fmt.Errorf("read bootstrap user id: %w", err)
	}

	session, err := txCreateSession(ctx, tx, userID, username, s.policy)
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("commit bootstrap: %w", err)
	}

	return session, nil
}

func (s *Store) Login(ctx context.Context, username string, password string) (Session, error) {
	if err := validateCredentialInput(username, password); err != nil {
		return Session{}, ErrInvalidCredential
	}

	userID, passwordHash, err := s.lookupUser(ctx, username)
	if err != nil {
		return Session{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return Session{}, ErrInvalidCredential
	}

	return s.createSession(ctx, userID, username)
}

func (s *Store) ValidateSession(ctx context.Context, sessionID string) (Principal, error) {
	if sessionID == "" {
		return Principal{}, ErrInvalidSession
	}

	var principal Principal
	var expiresAtRaw string
	var lastSeenAtRaw string
	err := s.db.QueryRowContext(ctx, `
SELECT users.id, users.username, sessions.expires_at, sessions.last_seen_at
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.id = ?
	AND sessions.revoked_at IS NULL
	AND users.disabled_at IS NULL`, sessionID).Scan(&principal.UserID, &principal.Username, &expiresAtRaw, &lastSeenAtRaw)
	if err == sql.ErrNoRows {
		return Principal{}, ErrInvalidSession
	}
	if err != nil {
		return Principal{}, fmt.Errorf("validate session: %w", err)
	}
	if err := s.validateSessionBounds(expiresAtRaw, lastSeenAtRaw); err != nil {
		return Principal{}, err
	}
	if err := s.touchSession(ctx, sessionID); err != nil {
		return Principal{}, err
	}
	return principal, nil
}

func (s *Store) ValidateCSRF(ctx context.Context, sessionID string, token string) error {
	if token == "" {
		return ErrInvalidCSRF
	}

	var expectedHash string
	err := s.db.QueryRowContext(ctx, `
SELECT csrf_token_hash
FROM sessions
WHERE id = ?
	AND revoked_at IS NULL
	AND expires_at > ?`, sessionID, formatTime(time.Now().UTC())).Scan(&expectedHash)
	if err == sql.ErrNoRows {
		return ErrInvalidCSRF
	}
	if err != nil {
		return fmt.Errorf("read csrf token: %w", err)
	}
	if expectedHash != hashToken(token) {
		return ErrInvalidCSRF
	}
	return nil
}

func (s *Store) RevokeSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrInvalidSession
	}

	result, err := s.db.ExecContext(ctx, "UPDATE sessions SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL", formatTime(time.Now().UTC()), sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read revoked session count: %w", err)
	}
	if affected == 0 {
		return ErrInvalidSession
	}
	return nil
}

func (s *Store) lookupUser(ctx context.Context, username string) (int64, string, error) {
	var userID int64
	var passwordHash string
	err := s.db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE username = ? AND disabled_at IS NULL", username).Scan(&userID, &passwordHash)
	if err == sql.ErrNoRows {
		return 0, "", ErrInvalidCredential
	}
	if err != nil {
		return 0, "", fmt.Errorf("lookup user: %w", err)
	}
	return userID, passwordHash, nil
}

func (s *Store) createSession(ctx context.Context, userID int64, username string) (Session, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, fmt.Errorf("begin session: %w", err)
	}
	defer rollback(tx)

	session, err := txCreateSession(ctx, tx, userID, username, s.policy)
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("commit session: %w", err)
	}

	return session, nil
}

func txHasUsers(ctx context.Context, tx *sql.Tx) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE disabled_at IS NULL").Scan(&count); err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return count > 0, nil
}

func txClaimBootstrap(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, "INSERT INTO bootstrap_state (id) VALUES (1)"); err != nil {
		return ErrBootstrapDisabled
	}
	return nil
}

func txCreateSession(ctx context.Context, tx *sql.Tx, userID int64, username string, policy SessionPolicy) (Session, error) {
	sessionID, err := randomToken()
	if err != nil {
		return Session{}, err
	}
	csrfToken, err := randomToken()
	if err != nil {
		return Session{}, err
	}

	now := time.Now().UTC()
	session := Session{
		ID:         sessionID,
		CSRFToken:  csrfToken,
		UserID:     userID,
		Username:   username,
		CreatedAt:  now,
		LastSeenAt: now,
		ExpiresAt:  now.Add(policy.TTL),
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO sessions (id, user_id, csrf_token_hash, created_at, last_seen_at, expires_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID,
		session.UserID,
		hashToken(session.CSRFToken),
		formatTime(session.CreatedAt),
		formatTime(session.LastSeenAt),
		formatTime(session.ExpiresAt),
	)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

func (s *Store) validateSessionBounds(expiresAtRaw string, lastSeenAtRaw string) error {
	now := time.Now().UTC()
	expiresAt, err := parseStoredTime(expiresAtRaw)
	if err != nil {
		return fmt.Errorf("parse session expiry: %w", err)
	}
	lastSeenAt, err := parseStoredTime(lastSeenAtRaw)
	if err != nil {
		return fmt.Errorf("parse session last seen: %w", err)
	}
	if !expiresAt.After(now) {
		return ErrInvalidSession
	}
	if !lastSeenAt.Add(s.policy.IdleTimeout).After(now) {
		return ErrInvalidSession
	}
	return nil
}

func (s *Store) touchSession(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE sessions
SET last_seen_at = ?
WHERE id = ? AND revoked_at IS NULL`, formatTime(time.Now().UTC()), sessionID)
	if err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

func (s *Store) CleanupExpiredAndRevokedSessions(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
DELETE FROM sessions
WHERE revoked_at IS NOT NULL OR expires_at <= ?`, formatTime(time.Now().UTC()))
	if err != nil {
		return 0, fmt.Errorf("cleanup sessions: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read cleanup session count: %w", err)
	}
	return affected, nil
}

func validateCredentialInput(username string, password string) error {
	if username == "" || password == "" {
		return ErrInvalidCredential
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseStoredTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}

func normalizedPolicy(policy SessionPolicy) SessionPolicy {
	defaults := DefaultSessionPolicy()
	if policy.TTL <= 0 {
		policy.TTL = defaults.TTL
	}
	if policy.IdleTimeout <= 0 {
		policy.IdleTimeout = defaults.IdleTimeout
	}
	return policy
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}
