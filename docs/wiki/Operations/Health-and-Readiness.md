# Health and Readiness

[Home](../Home.md) | Related: [Observability](Observability.md), [Troubleshooting](Troubleshooting.md), [Request Flow](../Architecture/Request-Flow.md)

## What Problem This Solves

Health endpoints let supervisors and operators distinguish process liveness from dependency readiness without exposing protected telemetry.

## How It Works

`GET /healthz` reports minimal liveness. `GET /readyz` checks DB reachability, migrations, and agent socket reachability.

## What Protects It

The endpoints return compact status checks only. They do not expose metrics, processes, services, users, sessions, or audit records.

## What Can Fail

Readiness can fail because the DB is unreachable or migrations are not current. It can degrade when the agent is unavailable.

## How To Diagnose It

Call `/readyz` and inspect `db`, `migrations`, and `agent` check statuses.

## How To Recover

Restore DB access, apply migrations through normal startup, or start/fix `koji-agent`.
