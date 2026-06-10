package db

import (
	"context"
	"database/sql"
	"errors"
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

func TestCheckSchemaCompatibilityReportsCurrentSchema(t *testing.T) {
	ctx := context.Background()
	conn := openTestDB(t)
	defer conn.Close()

	if err := RunMigrations(ctx, conn, InitialMigrations()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	report, err := CheckSchemaCompatibility(ctx, conn, InitialMigrations())
	if err != nil {
		t.Fatalf("check compatibility: %v", err)
	}
	if report.Status != CompatibilityOK {
		t.Fatalf("expected current schema, got %s", report.Status)
	}
	if report.CurrentSchema != CurrentSchemaVersion(InitialMigrations()) {
		t.Fatalf("unexpected current schema %s", report.CurrentSchema)
	}
	if report.PendingMigrations != 0 {
		t.Fatalf("expected no pending migrations, got %d", report.PendingMigrations)
	}
}

func TestCheckSchemaCompatibilityReportsUpgradeNeeded(t *testing.T) {
	ctx := context.Background()
	conn := openTestDB(t)
	defer conn.Close()

	oldMigrations := InitialMigrations()[:2]
	if err := RunMigrations(ctx, conn, oldMigrations); err != nil {
		t.Fatalf("run old migrations: %v", err)
	}

	report, err := CheckSchemaCompatibility(ctx, conn, InitialMigrations())
	if err != nil {
		t.Fatalf("check compatibility: %v", err)
	}
	if report.Status != CompatibilityMigrationRequired {
		t.Fatalf("expected migration required, got %s", report.Status)
	}
	if report.PendingMigrations != len(InitialMigrations())-len(oldMigrations) {
		t.Fatalf("unexpected pending migration count %d", report.PendingMigrations)
	}
}

func TestOpenUpgradesOlderSchema(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "koji.db")
	conn := openRawInitializedDB(t, path)
	if err := RunMigrations(ctx, conn, InitialMigrations()[:2]); err != nil {
		t.Fatalf("run old migrations: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close old database: %v", err)
	}

	upgraded, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open upgraded database: %v", err)
	}
	defer upgraded.Close()

	report, err := CheckSchemaCompatibility(ctx, upgraded, InitialMigrations())
	if err != nil {
		t.Fatalf("check upgraded database: %v", err)
	}
	if report.Status != CompatibilityOK {
		t.Fatalf("expected upgraded schema current, got %s", report.Status)
	}
}

func TestOpenRejectsFutureSchema(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "koji.db")
	conn := openRawInitializedDB(t, path)
	if err := RunMigrations(ctx, conn, InitialMigrations()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if _, err := conn.Exec("INSERT INTO schema_migrations (name, checksum) VALUES ('9999_future', 'future')"); err != nil {
		t.Fatalf("insert future migration: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close future database: %v", err)
	}

	if _, err := Open(ctx, path); !errors.Is(err, ErrFutureSchema) {
		t.Fatalf("expected future schema error, got %v", err)
	}
}

func TestOpenRejectsCorruptMigrationHistory(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "koji.db")
	conn := openRawInitializedDB(t, path)
	if err := RunMigrations(ctx, conn, InitialMigrations()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if _, err := conn.Exec("UPDATE schema_migrations SET checksum = 'changed' WHERE name = '0001_foundation'"); err != nil {
		t.Fatalf("corrupt migration checksum: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close corrupt database: %v", err)
	}

	if _, err := Open(ctx, path); !errors.Is(err, ErrCorruptMigrationHistory) {
		t.Fatalf("expected corrupt history error, got %v", err)
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

func openRawInitializedDB(t *testing.T, path string) *sql.DB {
	t.Helper()

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
