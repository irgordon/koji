# Disaster Recovery

[Home](../Home.md) | Related: [Backup and Recovery](Backup-and-Recovery.md), [Health and Readiness](Health-and-Readiness.md), [Troubleshooting](Troubleshooting.md)

## Purpose

Disaster recovery restores Koji from a backup artifact and a release artifact without relying on the source repository, Git history, or a developer workstation.

## Supported Failure Cases

This process covers:

- Host loss.
- Disk loss.
- SQLite corruption.
- Accidental deletion.
- Failed upgrade.
- Configuration mistake.

It does not provide clustering, replication, high availability, object storage integration, or cross-region failover.

## Required Inputs

Recovery requires:

- A verified Koji backup archive.
- A release artifact for the Koji version being restored.
- The target host runtime layout: `/usr/bin`, `/etc/koji`, `/var/lib/koji`, `/run/koji`, and `/usr/share/koji`.
- Service user and group ownership matching the packaged runtime policy.

## Recovery Workflow

1. Prepare a host with the runtime directories used by the package.
2. Install the selected release artifact.
3. Restore the backup archive with `packaging/scripts/restore.sh`.
4. Verify the restored database with `packaging/scripts/verify_restore.sh`.
5. Start `koji-agent`.
6. Start `kojid`.
7. Check `/healthz`.
8. Check `/readyz`.
9. Confirm the Jobs page shows durable jobs.
10. Confirm the Activity page shows audit records.

## Validation Signals

A successful recovery should prove:

- Database integrity passes.
- Current migrations are present.
- Users exist if users existed at backup time.
- Capability assignments exist if they existed at backup time.
- Jobs exist if jobs existed at backup time.
- Audit records exist if audit records existed at backup time.
- Readiness reports database access and migration state.

## If Recovery Fails

Use controlled rollback steps:

- Do not start service mutation while the restored state is uncertain.
- Keep the failed backup artifact unchanged for investigation.
- Check filesystem ownership and permissions before changing configuration.
- Re-run restore verification after each correction.
- Use a prior known-good backup if integrity checks fail.
