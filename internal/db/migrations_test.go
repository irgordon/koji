package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestRunMigrationsAppliesOnce(t *testing.T) {
	ctx := context.Background()
	conn := openTestDB(t)
	defer conn.Close()

	migrations := []Migration{{
		Name: "0001_test",
		SQL:  "CREATE TABLE demo (id INTEGER PRIMARY KEY);",
	}}

	if err := RunMigrations(ctx, conn, migrations); err != nil {
		t.Fatalf("first migration run failed: %v", err)
	}
	if err := RunMigrations(ctx, conn, migrations); err != nil {
		t.Fatalf("second migration run failed: %v", err)
	}

	count := countAppliedMigrations(t, conn, "0001_test")
	if count != 1 {
		t.Fatalf("expected one migration record, got %d", count)
	}
}

func TestRunMigrationsRejectsChecksumMismatch(t *testing.T) {
	ctx := context.Background()
	conn := openTestDB(t)
	defer conn.Close()

	original := []Migration{{
		Name: "0001_test",
		SQL:  "CREATE TABLE demo (id INTEGER PRIMARY KEY);",
	}}
	changed := []Migration{{
		Name: "0001_test",
		SQL:  "CREATE TABLE demo (id INTEGER PRIMARY KEY, name TEXT);",
	}}

	if err := RunMigrations(ctx, conn, original); err != nil {
		t.Fatalf("initial migration failed: %v", err)
	}
	if err := RunMigrations(ctx, conn, changed); err == nil {
		t.Fatal("expected checksum mismatch")
	}
}

func TestOpenEnablesForeignKeys(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "koji.db")

	conn, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	if !foreignKeysEnabled(t, conn) {
		t.Fatal("expected foreign_keys pragma to be enabled")
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "koji.db")
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	conn.SetMaxOpenConns(1)
	if err := initializeConnection(context.Background(), conn); err != nil {
		t.Fatalf("initialize sqlite db: %v", err)
	}
	return conn
}

func countAppliedMigrations(t *testing.T, conn *sql.DB, name string) int {
	t.Helper()

	var count int
	err := conn.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ?", name).Scan(&count)
	if err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	return count
}

func foreignKeysEnabled(t *testing.T, conn *sql.DB) bool {
	t.Helper()

	var enabled int
	if err := conn.QueryRow("PRAGMA foreign_keys").Scan(&enabled); err != nil {
		t.Fatalf("read foreign_keys pragma: %v", err)
	}
	return enabled == 1
}
