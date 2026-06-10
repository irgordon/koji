package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

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
	return s.createMagicToken(ctx, userID, createdBy, ttl)
}

func (s *Store) createMagicToken(ctx context.Context, userID int64, createdBy int64, ttl time.Duration) (MagicToken, error) {
	token, err := randomToken()
	if err != nil {
		return MagicToken{}, err
	}
	expiresAt := time.Now().UTC().Add(ttl)
	if err := s.insertMagicToken(ctx, userID, createdBy, token, expiresAt); err != nil {
		return MagicToken{}, err
	}
	return MagicToken{
		Token:     token,
		ExpiresAt: formatTime(expiresAt),
	}, nil
}

func (s *Store) insertMagicToken(ctx context.Context, userID int64, createdBy int64, token string, expiresAt time.Time) error {
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO magic_tokens (user_id, token_hash, created_by, expires_at)
VALUES (?, ?, ?, ?)`, userID, hashToken(token), createdBy, formatTime(expiresAt)); err != nil {
		return fmt.Errorf("issue magic token: %w", err)
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
