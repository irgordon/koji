# Capabilities

[Home](../Home.md) | Related: [Capability Reference](../Reference/Capability-Reference.md), [Audit](Audit.md), [Request Flow](../Architecture/Request-Flow.md)

## What Problem This Solves

Capabilities separate identity from permission. Being logged in does not imply broad operational authority.

## How It Works

Protected handlers require a specific capability for their surface. Missing capabilities produce safe 403 responses and audit records.

## What Protects It

Koji denies by default. Capabilities are stored in SQLite and looked up per user.

## What Can Fail

A user may be authenticated but lack `identity.users.manage`, `jobs.read`, `jobs.approve`, `host.services.control`, or another needed capability.

## How To Diagnose It

Check the UI permission message and Activity entries for `capability.denied`.

## How To Recover

Grant the minimal needed capability through the Administration page or API. Avoid broad grants when a narrow capability is enough.
