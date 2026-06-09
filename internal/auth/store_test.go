package auth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"koji/internal/db"
)

func TestExpiredSessionRejected(t *testing.T) {
	store, cleanup := newLifecycleStore(t)
	defer cleanup()
	session := bootstrapLifecycleSession(t, store)

	setSessionTime(t, store, session.ID, "expires_at", time.Now().UTC().Add(-time.Minute))

	_, err := store.ValidateSession(context.Background(), session.ID)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected invalid session, got %v", err)
	}
}

func TestIdleExpiredSessionRejected(t *testing.T) {
	store, cleanup := newLifecycleStore(t)
	defer cleanup()
	session := bootstrapLifecycleSession(t, store)

	setSessionTime(t, store, session.ID, "last_seen_at", time.Now().UTC().Add(-31*time.Minute))

	_, err := store.ValidateSession(context.Background(), session.ID)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected idle-expired session rejection, got %v", err)
	}
}

func TestValidSessionUpdatesLastSeenAt(t *testing.T) {
	store, cleanup := newLifecycleStore(t)
	defer cleanup()
	session := bootstrapLifecycleSession(t, store)
	original := time.Now().UTC().Add(-5 * time.Minute)
	setSessionTime(t, store, session.ID, "last_seen_at", original)

	if _, err := store.ValidateSession(context.Background(), session.ID); err != nil {
		t.Fatalf("validate session: %v", err)
	}

	updated := readSessionTime(t, store, session.ID, "last_seen_at")
	if !updated.After(original) {
		t.Fatalf("expected last_seen_at to advance past %s, got %s", original, updated)
	}
}

func TestRevokedSessionRejected(t *testing.T) {
	store, cleanup := newLifecycleStore(t)
	defer cleanup()
	session := bootstrapLifecycleSession(t, store)

	if err := store.RevokeSession(context.Background(), session.ID); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	_, err := store.ValidateSession(context.Background(), session.ID)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected revoked session rejection, got %v", err)
	}
}

func TestCleanupExpiredAndRevokedSessions(t *testing.T) {
	store, cleanup := newLifecycleStore(t)
	defer cleanup()
	expired := bootstrapLifecycleSession(t, store)
	revoked := loginLifecycleSession(t, store)
	active := loginLifecycleSession(t, store)
	setSessionTime(t, store, expired.ID, "expires_at", time.Now().UTC().Add(-time.Minute))
	if err := store.RevokeSession(context.Background(), revoked.ID); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	removed, err := store.CleanupExpiredAndRevokedSessions(context.Background())
	if err != nil {
		t.Fatalf("cleanup sessions: %v", err)
	}
	if removed != 2 {
		t.Fatalf("expected 2 removed sessions, got %d", removed)
	}
	if _, err := store.ValidateSession(context.Background(), active.ID); err != nil {
		t.Fatalf("expected active session to remain valid: %v", err)
	}
}

func newLifecycleStore(t *testing.T) (*Store, func()) {
	t.Helper()

	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "koji.db"))
	if err != nil {
		t.Fatalf("open lifecycle db: %v", err)
	}
	store := NewStoreWithPolicy(database, SessionPolicy{
		TTL:         time.Hour,
		IdleTimeout: 30 * time.Minute,
	})
	return store, func() {
		_ = database.Close()
	}
}

func bootstrapLifecycleSession(t *testing.T, store *Store) Session {
	t.Helper()

	session, err := store.Bootstrap(context.Background(), "admin", "secret-password")
	if err != nil {
		t.Fatalf("bootstrap session: %v", err)
	}
	return session
}

func loginLifecycleSession(t *testing.T, store *Store) Session {
	t.Helper()

	session, err := store.Login(context.Background(), "admin", "secret-password")
	if err != nil {
		t.Fatalf("login session: %v", err)
	}
	return session
}

func setSessionTime(t *testing.T, store *Store, sessionID string, column string, value time.Time) {
	t.Helper()

	query := "UPDATE sessions SET " + column + " = ? WHERE id = ?"
	if _, err := store.db.ExecContext(context.Background(), query, formatTime(value), sessionID); err != nil {
		t.Fatalf("set %s: %v", column, err)
	}
}

func readSessionTime(t *testing.T, store *Store, sessionID string, column string) time.Time {
	t.Helper()

	var raw string
	query := "SELECT " + column + " FROM sessions WHERE id = ?"
	if err := store.db.QueryRowContext(context.Background(), query, sessionID).Scan(&raw); err != nil {
		t.Fatalf("read %s: %v", column, err)
	}
	value, err := parseStoredTime(raw)
	if err != nil {
		t.Fatalf("parse %s: %v", column, err)
	}
	return value
}
