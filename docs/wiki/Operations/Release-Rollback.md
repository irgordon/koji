# Release Rollback

[Home](../Home.md) | Related: [Release Operations](Release-Operations.md), [Upgrade Procedure](Upgrade-Procedure.md), [Backup and Recovery](Backup-and-Recovery.md), [Disaster Recovery](Disaster-Recovery.md)

## Purpose

Rollback returns Koji to a prior release artifact and a matching backup when an upgrade fails.

Koji database migrations are forward-only. Production down migrations are not supported.

## Before Upgrade

Before installing a new release:

1. Create a backup.
2. Verify the backup.
3. Record the current Koji version.
4. Keep the current release artifact available.
5. Confirm `/readyz` is healthy before changing binaries.

## Rollback Workflow

If an upgrade fails:

1. Stop `kojid` and `koji-agent`.
2. Restore the backup taken before the upgrade.
3. Install the prior release artifact.
4. Run restore verification and upgrade verification for the restored release.
5. Start `koji-agent`.
6. Start `kojid`.
7. Check `/healthz`.
8. Check `/readyz`.
9. Confirm Jobs and Activity show expected pre-upgrade state.

## Verification

After rollback, verify:

- Binaries come from the prior release artifact.
- `/etc/koji/koji.yaml` and `/etc/koji/agent.yaml` match the restored backup.
- `/var/lib/koji/koji.db` passes restore verification.
- Jobs are visible.
- Audit records are visible.
- Control-plane observability counters load.

## Limits

Rollback cannot preserve data created after the backup point. If an upgrade partially ran and accepted new operator actions, restoring the pre-upgrade backup intentionally discards those later changes.

Rollback requires restore when a release applied forward migrations. Koji does not support production down migrations.
