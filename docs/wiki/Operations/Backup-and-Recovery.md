# Backup and Recovery

[Home](../Home.md) | Related: [Database Schema](../Developer/Database-Schema.md), [Upgrade Procedure](Upgrade-Procedure.md), [Release Candidate Checklist](Release-Candidate-Checklist.md), [Disaster Recovery](Disaster-Recovery.md), [Release Rollback](Release-Rollback.md)

## What Is Backed Up

Koji backups include:

- SQLite database: users, sessions, capabilities, jobs, approvals, audit records, migration state, and stored governance data.
- Daemon configuration: `/etc/koji/koji.yaml`.
- Agent configuration: `/etc/koji/agent.yaml`.
- Metadata: timestamp, Koji version, schema version, backup format version, and record counts used for restore verification.

Backups do not include operating system packages, systemd runtime state, journal logs, release binaries, or external snapshots.

## Backup Artifact

`packaging/scripts/backup.sh` writes an offline-restorable archive:

```text
koji-backup-YYYYMMDD-HHMMSS.tar.gz
```

The archive contains:

```text
koji-backup-YYYYMMDD-HHMMSS/
├── database/
│   └── koji.db
├── config/
│   ├── koji.yaml
│   └── agent.yaml
└── metadata.json
```

The database copy is created with SQLite `.backup`; do not use a plain `cp` of an active database as the backup mechanism.

## Create A Backup

Run from an installed host:

```sh
make backup
```

Equivalent direct command:

```sh
packaging/scripts/backup.sh /secure/backup/location
```

The script defaults to:

- Database: `/var/lib/koji/koji.db`
- Config directory: `/etc/koji`
- Output directory: `build/backups`

For staging or tests, override paths:

```sh
KOJI_DB_PATH=/path/to/koji.db \
KOJI_CONFIG_DIR=/path/to/etc/koji \
packaging/scripts/backup.sh /path/to/backups
```

## Restore A Backup

Restore expects either the backup directory or the `.tar.gz` archive:

```sh
BACKUP=/secure/backup/location/koji-backup-YYYYMMDD-HHMMSS.tar.gz make restore
```

Equivalent direct command:

```sh
packaging/scripts/restore.sh /secure/backup/location/koji-backup-YYYYMMDD-HHMMSS.tar.gz
```

The restore script validates the backup structure, validates the SQLite file, restores the database and configuration, and runs restore verification.

## Verify A Restore

Run:

```sh
make verify-restore
```

Equivalent direct command:

```sh
packaging/scripts/verify_restore.sh /var/lib/koji/koji.db
```

Verification checks:

- SQLite opens successfully.
- `PRAGMA integrity_check` returns `ok`.
- The restored migration version matches the expected Koji schema.
- Users, jobs, audit events, and capability assignments are present when backup metadata says they existed.

## Recovery Order

Use this order during production recovery:

1. Stop `kojid` and `koji-agent`.
2. Restore the backup archive.
3. Confirm ownership and permissions for `/var/lib/koji` and `/etc/koji`.
4. Start `koji-agent`.
5. Start `kojid`.
6. Check `/healthz` and `/readyz`.
7. Confirm Jobs, Activity, and Observability pages show expected state.

## Operator Notes

Keep at least one backup outside the host being protected. A backup stored only on the same failed disk is not a disaster recovery artifact.

Before every upgrade, create and verify a backup. If migration compatibility checks fail or a future schema is detected, do not start the new release against the database.
