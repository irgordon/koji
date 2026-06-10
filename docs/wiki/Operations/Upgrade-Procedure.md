# Upgrade Procedure

[Home](../Home.md) | Related: [Backup and Recovery](Backup-and-Recovery.md), [Release Rollback](Release-Rollback.md), [Release Operations](Release-Operations.md)

## Purpose

The upgrade procedure protects Koji state before a new release opens or migrates the SQLite database.

Koji migrations are forward-only. A prior release must not start against a database that has a newer schema.

## Pre-Upgrade

1. Record the currently installed Koji release.
2. Create a backup with `make backup`.
3. Verify the backup with `make verify-restore` or `packaging/scripts/verify_restore.sh`.
4. Review release notes for schema or operational changes.
5. Run the compatibility check:

```sh
make pre-upgrade-check
```

The check reports:

- Current schema.
- Target schema.
- Applied migration count.
- Status: `ok`, `migration_required`, or `future_schema_detected`.
- Backup requirement.

If the check reports `future_schema_detected`, stop. The database belongs to a newer Koji release.

## Upgrade

1. Stop `kojid`.
2. Stop `koji-agent`.
3. Install the new release artifact.
4. Start `koji-agent`.
5. Start `kojid`.
6. Let `kojid` run validated forward migrations during startup.

Startup refuses corrupt migration history and future schemas before migrations run.

## Post-Upgrade

Run:

```sh
make verify-upgrade
```

Then verify:

- `/healthz` returns live.
- `/readyz` reports database and migration checks as healthy.
- Jobs are readable.
- Activity is readable.
- Observability loads for an account with `observability.metrics.read`.
- Agent readiness is expected for the deployment.
- Pending approvals still show correct status.

## Failure Handling

If startup or verification fails:

1. Stop `kojid` and `koji-agent`.
2. Keep the failed database unchanged for investigation.
3. Restore the pre-upgrade backup.
4. Install the prior release artifact.
5. Run restore verification.
6. Start services and check readiness.

Do not assume rollback can undo migrations. Restore is the supported rollback path after schema changes.
