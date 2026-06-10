# Backup and Recovery

[Home](../Home.md) | Related: [Database Schema](../Developer/Database-Schema.md), [Audit](../Security/Audit.md), [Jobs Page](../User-Guide/Jobs-Page.md)

## What Problem This Solves

The SQLite database holds users, sessions, capabilities, audit events, and jobs. Losing it loses governance state.

## How It Works

Back up `/var/lib/koji/koji.db` and related SQLite WAL files using a process that preserves consistency.

## What Protects It

Durable schema migrations and foreign keys preserve consistency when the DB is healthy.

## What Can Fail

Filesystem corruption, ownership changes, disk exhaustion, or bad restores can break startup or audit writes.

## How To Diagnose It

Check `/readyz`, service logs, and audit write failure counters.

## How To Recover

Stop services, restore the DB files from a known-good backup, correct ownership, start services, and verify readiness.
