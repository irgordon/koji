# Database Schema

[Home](../Home.md) | Related: [Data Flow](../Architecture/Data-Flow.md), [Backup and Recovery](../Operations/Backup-and-Recovery.md)

## What Problem This Solves

SQLite stores the durable state needed for sessions, capabilities, audit, jobs, and migrations.

## How It Works

Migrations are ordered, checksummed, and tracked in `schema_migrations`. Core tables include `users`, `sessions`, `capabilities`, `user_capabilities`, `audit_events`, and `jobs`.

## What Protects It

WAL, foreign keys, busy timeout, deterministic migrations, and startup failure on DB initialization errors protect consistency.

## What Can Fail

Checksum mismatch, missing migrations, DB path ownership problems, or disk exhaustion.

## How To Diagnose It

Use `/readyz` migration checks and DB migration tests.

## How To Recover

Restore a consistent DB backup or repair the migration source before startup.
