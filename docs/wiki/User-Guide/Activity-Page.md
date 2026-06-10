# Activity Page

[Home](../Home.md) | Related: [Audit](../Security/Audit.md), [Audit Event Reference](../Reference/Audit-Event-Reference.md)

## What Problem This Solves

The Activity page provides a governed audit read model for operators and reviewers.

## How It Works

It lists recent normalized audit events with timestamp, action, target, outcome, reason code, and request ID.

## What Protects It

Activity requires `audit.events.read` and excludes sensitive raw internals.

## What Can Fail

The view can be empty, denied by capability, or stale if audit writes fail.

## How To Diagnose It

Check audit counters and request IDs. Compare actions with [Audit Event Reference](../Reference/Audit-Event-Reference.md).

## How To Recover

Restore DB availability for audit writes or grant the minimal read capability.
