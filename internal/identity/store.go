package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
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
