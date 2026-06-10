# PHASE-0035: Backup, Restore, and Disaster Recovery

## Goal

Create a documented, tested, repeatable recovery process for Koji.

## Non-Goals

- No replication.
- No clustering.
- No cloud backup integration.
- No API, UI, auth, audit, worker, job, agent, or service mutation behavior changes.

## Invariants Preserved

- `kojid` does not execute privileged service mutation.
- The agent boundary is unchanged.
- Authentication, capabilities, CSRF, audit, jobs, and approvals are unchanged.
- Recovery tooling does not require source repository knowledge at restore time.

## Negative Patterns Avoided

- No plain copy of an active SQLite database for backup creation.
- No distro-specific package manager assumptions.
- No developer-local paths.
- No source-control-dependent recovery process.

## Design Summary

Backup uses SQLite `.backup` to create a consistent database copy, stores daemon and agent configuration, writes metadata, and compresses the artifact. Restore validates the archive structure and database integrity before installing the database and configuration. Verification checks integrity, schema version, and metadata-backed record counts for governance data.

## Files Changed

- `packaging/scripts/backup.sh`
- `packaging/scripts/restore.sh`
- `packaging/scripts/verify_restore.sh`
- `packaging/packaging_test.go`
- `Makefile`
- `docs/wiki/Operations/Backup-and-Recovery.md`
- `docs/wiki/Operations/Disaster-Recovery.md`
- `docs/wiki/Operations/Release-Rollback.md`
- `docs/wiki/Home.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
gofmt -w packaging/packaging_test.go
GOCACHE=/tmp/koji-go-cache go test ./...
npm run build
packaging/scripts/verify_docs.sh
git diff --check
```

## Changelog

- Added backup, restore, and restore verification scripts.
- Added Makefile targets for backup, restore, and verification.
- Added automated recovery test coverage.
- Added disaster recovery and release rollback operations documentation.

## Summary

Koji now has a deterministic backup artifact, a restore path, and verification that restored governance data is present.

## Notes / Deviations

Restore verification embeds the current expected schema migration name. Future schema migrations must update the expected value.
