package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type Migration struct {
	Name string
	SQL  string
}

func InitialMigrations() []Migration {
	return []Migration{
		{
			Name: "0001_foundation",
			SQL: `
CREATE TABLE audit_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	actor TEXT NOT NULL,
	action TEXT NOT NULL,
	target TEXT NOT NULL,
	status TEXT NOT NULL,
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	disabled_at TEXT
);

CREATE TABLE bootstrap_state (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	completed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	csrf_token_hash TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	expires_at TEXT NOT NULL,
	revoked_at TEXT,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE capabilities (
	name TEXT PRIMARY KEY,
	description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE user_capabilities (
	user_id INTEGER NOT NULL,
	capability_name TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	PRIMARY KEY (user_id, capability_name),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (capability_name) REFERENCES capabilities(name) ON DELETE CASCADE
);
`,
		},
		{
			Name: "0002_capability_audit_fields",
			SQL: `
ALTER TABLE audit_events ADD COLUMN user_id INTEGER;
ALTER TABLE audit_events ADD COLUMN outcome TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_events ADD COLUMN reason_code TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_events ADD COLUMN request_id TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_events ADD COLUMN remote_addr TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_events ADD COLUMN dev_bypass INTEGER NOT NULL DEFAULT 0;

INSERT INTO capabilities (name, description) VALUES
	('host.metrics.read', 'Read host CPU, memory, and uptime metrics'),
	('host.disk.read', 'Read host disk metrics'),
	('host.services.read', 'Read host service status'),
	('host.processes.read', 'Read host process listings'),
	('host.services.control', 'Request host service start, stop, or restart');
`,
		},
		{
			Name: "0003_session_last_seen",
			SQL: `
ALTER TABLE sessions ADD COLUMN last_seen_at TEXT NOT NULL DEFAULT '1970-01-01T00:00:00Z';
`,
		},
		{
			Name: "0004_audit_events_read_capability",
			SQL: `
INSERT INTO capabilities (name, description) VALUES
	('audit.events.read', 'Read normalized audit activity events');
`,
		},
		{
			Name: "0005_jobs_foundation",
			SQL: `
CREATE TABLE jobs (
	id TEXT PRIMARY KEY,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	created_by INTEGER,
	action TEXT NOT NULL,
	target TEXT NOT NULL,
	status TEXT NOT NULL,
	status_reason TEXT NOT NULL DEFAULT '',
	request_id TEXT NOT NULL DEFAULT '',
	FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
);

INSERT INTO capabilities (name, description) VALUES
	('jobs.read', 'Read job lifecycle records');
`,
		},
		{
			Name: "0006_job_approval_fields",
			SQL: `
ALTER TABLE jobs ADD COLUMN approved_by INTEGER;
ALTER TABLE jobs ADD COLUMN approved_at TEXT;
ALTER TABLE jobs ADD COLUMN rejected_by INTEGER;
ALTER TABLE jobs ADD COLUMN rejected_at TEXT;
ALTER TABLE jobs ADD COLUMN decision_reason TEXT NOT NULL DEFAULT '';

INSERT INTO capabilities (name, description) VALUES
	('jobs.approve', 'Approve or reject queued jobs');
`,
		},
		{
			Name: "0007_job_worker_fields",
			SQL: `
ALTER TABLE jobs ADD COLUMN started_at TEXT;
`,
		},
		{
			Name: "0008_observability_metrics_read_capability",
			SQL: `
INSERT INTO capabilities (name, description) VALUES
	('observability.metrics.read', 'Read control-plane observability metrics');
`,
		},
		{
			Name: "0009_identity_magic_tokens",
			SQL: `
ALTER TABLE users ADD COLUMN is_super_admin INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN updated_at TEXT NOT NULL DEFAULT '1970-01-01T00:00:00Z';

UPDATE users
SET is_super_admin = 1,
	updated_at = created_at
WHERE id = (SELECT MIN(id) FROM users);

INSERT INTO capabilities (name, description) VALUES
	('identity.users.manage', 'Manage users, capabilities, sessions, and magic tokens');

INSERT OR IGNORE INTO user_capabilities (user_id, capability_name)
SELECT users.id, 'identity.users.manage'
FROM users
WHERE users.is_super_admin = 1;

CREATE TABLE magic_tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	token_hash TEXT NOT NULL UNIQUE,
	created_by INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
	expires_at TEXT NOT NULL,
	consumed_at TEXT,
	consumed_by_session_id TEXT,
	revoked_at TEXT,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
	FOREIGN KEY (consumed_by_session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

CREATE INDEX idx_magic_tokens_hash ON magic_tokens(token_hash);
CREATE INDEX idx_magic_tokens_user ON magic_tokens(user_id);
`,
		},
	}
}

func RunMigrations(ctx context.Context, conn *sql.DB, migrations []Migration) error {
	if err := ensureMigrationTable(ctx, conn); err != nil {
		return err
	}

	for _, migration := range migrations {
		if err := runMigration(ctx, conn, migration); err != nil {
			return err
		}
	}

	return nil
}

func CheckMigrationsCurrent(ctx context.Context, conn *sql.DB, migrations []Migration) error {
	if err := ensureMigrationTable(ctx, conn); err != nil {
		return err
	}
	for _, migration := range migrations {
		if err := checkMigrationCurrent(ctx, conn, migration); err != nil {
			return err
		}
	}
	return nil
}

func ensureMigrationTable(ctx context.Context, conn *sql.DB) error {
	_, err := conn.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	name TEXT PRIMARY KEY,
	checksum TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);`)
	if err != nil {
		return fmt.Errorf("ensure schema migrations table: %w", err)
	}
	return nil
}

func runMigration(ctx context.Context, conn *sql.DB, migration Migration) error {
	checksum := migrationChecksum(migration)

	applied, err := appliedMigration(ctx, conn, migration.Name)
	if err != nil {
		return err
	}
	if applied != "" {
		return verifyAppliedMigration(migration.Name, checksum, applied)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.Name, err)
	}
	defer rollbackMigration(tx)

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.Name, err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (name, checksum) VALUES (?, ?)", migration.Name, checksum); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Name, err)
	}

	return nil
}

func checkMigrationCurrent(ctx context.Context, conn *sql.DB, migration Migration) error {
	expected := migrationChecksum(migration)
	applied, err := appliedMigration(ctx, conn, migration.Name)
	if err != nil {
		return err
	}
	if applied == "" {
		return fmt.Errorf("migration %s is not applied", migration.Name)
	}
	return verifyAppliedMigration(migration.Name, expected, applied)
}

func appliedMigration(ctx context.Context, conn *sql.DB, name string) (string, error) {
	var checksum string
	err := conn.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE name = ?", name).Scan(&checksum)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read migration %s: %w", name, err)
	}
	return checksum, nil
}

func verifyAppliedMigration(name string, expected string, actual string) error {
	if expected != actual {
		return fmt.Errorf("migration %s checksum mismatch", name)
	}
	return nil
}

func rollbackMigration(tx *sql.Tx) {
	_ = tx.Rollback()
}

func migrationChecksum(migration Migration) string {
	sum := sha256.Sum256([]byte(migration.Name + "\n" + migration.SQL))
	return hex.EncodeToString(sum[:])
}
