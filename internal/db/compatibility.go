package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrFutureSchema            = errors.New("future_schema_detected")
	ErrCorruptMigrationHistory = errors.New("corrupt_migration_history")
)

type CompatibilityStatus string

const (
	CompatibilityOK                CompatibilityStatus = "ok"
	CompatibilityMigrationRequired CompatibilityStatus = "migration_required"
	CompatibilityFutureSchema      CompatibilityStatus = "future_schema_detected"
	CompatibilityCorruptHistory    CompatibilityStatus = "corrupt_migration_history"
)

type CompatibilityReport struct {
	CurrentSchema     string              `json:"currentSchema"`
	TargetSchema      string              `json:"targetSchema"`
	AppliedMigrations int                 `json:"appliedMigrations"`
	PendingMigrations int                 `json:"pendingMigrations"`
	Status            CompatibilityStatus `json:"status"`
	Reason            string              `json:"reason"`
}

type appliedRecord struct {
	name     string
	checksum string
}

func CheckSchemaCompatibility(ctx context.Context, conn *sql.DB, migrations []Migration) (CompatibilityReport, error) {
	report, err := buildCompatibilityReport(ctx, conn, migrations)
	if err != nil {
		return report, err
	}
	return report, compatibilityError(report)
}

func CurrentSchemaVersion(migrations []Migration) string {
	if len(migrations) == 0 {
		return ""
	}
	return migrations[len(migrations)-1].Name
}

func buildCompatibilityReport(ctx context.Context, conn *sql.DB, migrations []Migration) (CompatibilityReport, error) {
	applied, err := appliedMigrationRecords(ctx, conn)
	if err != nil {
		return CompatibilityReport{}, err
	}
	report := newCompatibilityReport(migrations, applied)
	classifyCompatibility(&report, migrations, applied)
	return report, nil
}

func newCompatibilityReport(migrations []Migration, applied []appliedRecord) CompatibilityReport {
	pending := len(migrations) - len(applied)
	if pending < 0 {
		pending = 0
	}
	return CompatibilityReport{
		CurrentSchema:     currentAppliedSchema(applied),
		TargetSchema:      CurrentSchemaVersion(migrations),
		AppliedMigrations: len(applied),
		PendingMigrations: pending,
		Status:            CompatibilityOK,
	}
}

func classifyCompatibility(report *CompatibilityReport, migrations []Migration, applied []appliedRecord) {
	known := knownMigrationChecksums(migrations)
	if reason := invalidAppliedMigration(known, applied); reason != "" {
		report.Status = corruptOrFutureStatus(report.CurrentSchema, report.TargetSchema)
		report.Reason = reason
		return
	}
	if reason := nonContiguousHistory(migrations, applied); reason != "" {
		report.Status = CompatibilityCorruptHistory
		report.Reason = reason
		return
	}
	if report.PendingMigrations > 0 {
		report.Status = CompatibilityMigrationRequired
		report.Reason = "pending_migrations"
		return
	}
	report.PendingMigrations = 0
	report.Reason = "schema_current"
}

func appliedMigrationRecords(ctx context.Context, conn *sql.DB) ([]appliedRecord, error) {
	exists, err := schemaMigrationTableExists(ctx, conn)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return queryAppliedMigrationRecords(ctx, conn)
}

func schemaMigrationTableExists(ctx context.Context, conn *sql.DB) (bool, error) {
	var count int
	err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("inspect schema migrations table: %w", err)
	}
	return count == 1, nil
}

func queryAppliedMigrationRecords(ctx context.Context, conn *sql.DB) ([]appliedRecord, error) {
	rows, err := conn.QueryContext(ctx, "SELECT name, checksum FROM schema_migrations ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("read migration history: %w", err)
	}
	defer rows.Close()
	return scanAppliedMigrationRecords(rows)
}

func scanAppliedMigrationRecords(rows *sql.Rows) ([]appliedRecord, error) {
	var applied []appliedRecord
	for rows.Next() {
		var record appliedRecord
		if err := rows.Scan(&record.name, &record.checksum); err != nil {
			return nil, fmt.Errorf("scan migration history: %w", err)
		}
		applied = append(applied, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate migration history: %w", err)
	}
	return applied, nil
}

func knownMigrationChecksums(migrations []Migration) map[string]string {
	known := make(map[string]string, len(migrations))
	for _, migration := range migrations {
		known[migration.Name] = migrationChecksum(migration)
	}
	return known
}

func invalidAppliedMigration(known map[string]string, applied []appliedRecord) string {
	for _, record := range applied {
		expected, ok := known[record.name]
		if !ok {
			return "unknown_migration_" + record.name
		}
		if expected != record.checksum {
			return "checksum_mismatch_" + record.name
		}
	}
	return ""
}

func nonContiguousHistory(migrations []Migration, applied []appliedRecord) string {
	appliedNames := make(map[string]bool, len(applied))
	for _, record := range applied {
		appliedNames[record.name] = true
	}
	seenPending := false
	for _, migration := range migrations {
		if !appliedNames[migration.Name] {
			seenPending = true
			continue
		}
		if seenPending {
			return "non_contiguous_migration_history"
		}
	}
	return ""
}

func corruptOrFutureStatus(current string, target string) CompatibilityStatus {
	if current > target {
		return CompatibilityFutureSchema
	}
	return CompatibilityCorruptHistory
}

func compatibilityError(report CompatibilityReport) error {
	switch report.Status {
	case CompatibilityFutureSchema:
		return fmt.Errorf("%w: current=%s target=%s", ErrFutureSchema, report.CurrentSchema, report.TargetSchema)
	case CompatibilityCorruptHistory:
		return fmt.Errorf("%w: %s", ErrCorruptMigrationHistory, report.Reason)
	default:
		return nil
	}
}

func currentAppliedSchema(applied []appliedRecord) string {
	if len(applied) == 0 {
		return ""
	}
	return applied[len(applied)-1].name
}
