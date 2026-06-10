package identity

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

	"koji/internal/caps"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserDisabled      = errors.New("user disabled")
	ErrSelfLockout       = errors.New("self lockout prevented")
	ErrInvalidCapability = errors.New("invalid capability")
)

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
	Disabled     bool   `json:"disabled"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type MagicToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, username, is_super_admin, disabled_at IS NOT NULL, created_at, updated_at
FROM users
ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

func (s *Store) CreateManagedUser(ctx context.Context, username string) (User, error) {
	if username == "" {
		return User{}, fmt.Errorf("username is required")
	}
	result, err := s.db.ExecContext(ctx, `
INSERT INTO users (username, password_hash, is_super_admin, updated_at)
VALUES (?, '', 0, ?)`, username, formatTime(time.Now().UTC()))
	if err != nil {
		return User{}, fmt.Errorf("create managed user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("read managed user id: %w", err)
	}
	return s.GetUser(ctx, id)
}

func (s *Store) GetUser(ctx context.Context, id int64) (User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
SELECT id, username, is_super_admin, disabled_at IS NOT NULL, created_at, updated_at
FROM users
WHERE id = ?`, id).Scan(&user.ID, &user.Username, &user.IsSuperAdmin, &user.Disabled, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *Store) DisableUser(ctx context.Context, id int64) (User, error) {
	if err := s.preventDisableLockout(ctx, id); err != nil {
		return User{}, err
	}
	now := formatTime(time.Now().UTC())
	result, err := s.db.ExecContext(ctx, `
UPDATE users
SET disabled_at = ?,
	updated_at = ?
WHERE id = ? AND disabled_at IS NULL`, now, now, id)
	if err != nil {
		return User{}, fmt.Errorf("disable user: %w", err)
	}
	if err := requireAffected(result); err != nil {
		return User{}, err
	}
	if _, err := s.db.ExecContext(ctx, "UPDATE sessions SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL", now, id); err != nil {
		return User{}, fmt.Errorf("revoke disabled user sessions: %w", err)
	}
	return s.GetUser(ctx, id)
}

func (s *Store) EnableUser(ctx context.Context, id int64) (User, error) {
	result, err := s.db.ExecContext(ctx, `
UPDATE users
SET disabled_at = NULL,
	updated_at = ?
WHERE id = ?`, formatTime(time.Now().UTC()), id)
	if err != nil {
		return User{}, fmt.Errorf("enable user: %w", err)
	}
	if err := requireAffected(result); err != nil {
		return User{}, err
	}
	return s.GetUser(ctx, id)
}

func (s *Store) ListAvailableCapabilities(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT name FROM capabilities ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list capabilities: %w", err)
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (s *Store) ListUserCapabilities(ctx context.Context, id int64) ([]string, error) {
	if _, err := s.GetUser(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT capability_name
FROM user_capabilities
WHERE user_id = ?
ORDER BY capability_name`, id)
	if err != nil {
		return nil, fmt.Errorf("list user capabilities: %w", err)
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (s *Store) GrantCapability(ctx context.Context, id int64, capability string) ([]string, error) {
	if err := validateCapability(capability); err != nil {
		return nil, err
	}
	if _, err := s.GetUser(ctx, id); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO user_capabilities (user_id, capability_name)
VALUES (?, ?)`, id, capability); err != nil {
		return nil, fmt.Errorf("grant capability: %w", err)
	}
	return s.ListUserCapabilities(ctx, id)
}

func (s *Store) RevokeCapability(ctx context.Context, id int64, capability string) ([]string, error) {
	if err := validateCapability(capability); err != nil {
		return nil, err
	}
	if capability == string(caps.IdentityUsersManage) {
		if err := s.preventManageCapabilityLockout(ctx, id); err != nil {
			return nil, err
		}
	}
	result, err := s.db.ExecContext(ctx, `
DELETE FROM user_capabilities
WHERE user_id = ? AND capability_name = ?`, id, capability)
	if err != nil {
		return nil, fmt.Errorf("revoke capability: %w", err)
	}
	if err := requireAffected(result); err != nil {
		return nil, err
	}
	return s.ListUserCapabilities(ctx, id)
}

func (s *Store) IssueMagicToken(ctx context.Context, userID int64, createdBy int64, ttl time.Duration) (MagicToken, error) {
	if ttl <= 0 {
		return MagicToken{}, fmt.Errorf("magic token ttl must be positive")
	}
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return MagicToken{}, err
	}
	if user.Disabled {
		return MagicToken{}, ErrUserDisabled
	}
	token, err := randomToken()
	if err != nil {
		return MagicToken{}, err
	}
	expiresAt := time.Now().UTC().Add(ttl)
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO magic_tokens (user_id, token_hash, created_by, expires_at)
VALUES (?, ?, ?, ?)`, userID, hashToken(token), createdBy, formatTime(expiresAt)); err != nil {
		return MagicToken{}, fmt.Errorf("issue magic token: %w", err)
	}
	return MagicToken{
		Token:     token,
		ExpiresAt: formatTime(expiresAt),
	}, nil
}

func (s *Store) preventDisableLockout(ctx context.Context, id int64) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	if user.IsSuperAdmin {
		count, err := s.activeSuperAdminCount(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrSelfLockout
		}
	}
	hasManage, err := s.userHasCapability(ctx, id, string(caps.IdentityUsersManage))
	if err != nil {
		return err
	}
	if hasManage {
		count, err := s.activeIdentityManagerCount(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrSelfLockout
		}
	}
	return nil
}

func (s *Store) preventManageCapabilityLockout(ctx context.Context, id int64) error {
	hasManage, err := s.userHasCapability(ctx, id, string(caps.IdentityUsersManage))
	if err != nil {
		return err
	}
	if !hasManage {
		return nil
	}
	count, err := s.activeIdentityManagerCount(ctx)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrSelfLockout
	}
	return nil
}

func (s *Store) activeSuperAdminCount(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_super_admin = 1 AND disabled_at IS NULL").Scan(&count); err != nil {
		return 0, fmt.Errorf("count super admins: %w", err)
	}
	return count, nil
}

func (s *Store) activeIdentityManagerCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM users
JOIN user_capabilities ON user_capabilities.user_id = users.id
WHERE users.disabled_at IS NULL
	AND user_capabilities.capability_name = ?`, string(caps.IdentityUsersManage)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count identity managers: %w", err)
	}
	return count, nil
}

func (s *Store) userHasCapability(ctx context.Context, id int64, capability string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM user_capabilities
WHERE user_id = ? AND capability_name = ?`, id, capability).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("lookup capability: %w", err)
	}
	return count > 0, nil
}

func scanUsers(rows *sql.Rows) ([]User, error) {
	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.IsSuperAdmin, &user.Disabled, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func scanStrings(rows *sql.Rows) ([]string, error) {
	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("scan string: %w", err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate strings: %w", err)
	}
	return values, nil
}

func validateCapability(capability string) error {
	for _, known := range caps.All() {
		if string(known) == capability {
			return nil
		}
	}
	return ErrInvalidCapability
}

func requireAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
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
