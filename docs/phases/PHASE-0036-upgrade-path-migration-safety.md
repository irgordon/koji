# PHASE-0036: Upgrade Path, Migration Safety, and Version Compatibility

## Goal

Add an upgrade safety gate that reports schema state, requires backup awareness, refuses future schemas, and verifies upgraded governance data.

## Non-Goals

- No live migrations.
- No zero-downtime upgrades.
- No multi-node upgrade orchestration.
- No database replication.
- No auth, capability, audit, job, worker, agent, API, or UI behavior changes.

## Invariants Preserved

- Database migrations are forward-only.
- Applied migration checksums cannot change silently.
- Future schemas prevent startup.
- Recovery requires backup and restore rather than production down migrations.

## Negative Patterns Avoided

- No blind startup migration without compatibility validation.
- No downgrade against a newer database.
- No manual SQLite inspection requirement for operators.
- No distro-specific package-manager assumptions.

## Design Summary

`internal/db` now checks schema compatibility before applying migrations. Current schemas pass, older schemas report `migration_required` and can migrate forward, future schemas fail with `future_schema_detected`, and checksum or history corruption fails with `corrupt_migration_history`.

Operator scripts expose the same lifecycle: `pre_upgrade_check.sh` reports current versus target schema and backup requirement, while `verify_upgrade.sh` checks that schema and core governance tables are readable after upgrade.

## Files Changed

- `internal/db/compatibility.go`
- `internal/db/db.go`
- `internal/db/migrations_test.go`
- `packaging/scripts/pre_upgrade_check.sh`
- `packaging/scripts/verify_upgrade.sh`
- `packaging/packaging_test.go`
- `Makefile`
- `docs/wiki/Operations/Upgrade-Procedure.md`
- `docs/wiki/Operations/Release-Rollback.md`
- `docs/wiki/Operations/Backup-and-Recovery.md`
- `docs/wiki/Home.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
gofmt -w internal/db packaging/packaging_test.go
GOCACHE=/tmp/koji-go-cache go test ./...
npm run test
npm run build
packaging/scripts/verify_docs.sh
make verify-openapi
KOJI_DB_PATH=<temp>/koji.db KOJI_CONFIG_DIR=<temp>/etc/koji make pre-upgrade-check
KOJI_DB_PATH=<temp>/koji.db make verify-upgrade
git diff --check
```

## Changelog

- Added startup schema compatibility validation before migrations.
- Added future schema and corrupt migration history rejection.
- Added pre-upgrade and verify-upgrade scripts.
- Added upgrade procedure documentation and rollback guidance.

## Summary

Koji now has a deterministic upgrade gate before migration, a post-upgrade verification path, and documented rollback boundaries.

## Notes / Deviations

`verify_upgrade.sh` can optionally check an observability URL through `KOJI_OBSERVABILITY_URL`; the governed metrics API still requires the existing authenticated access path.
