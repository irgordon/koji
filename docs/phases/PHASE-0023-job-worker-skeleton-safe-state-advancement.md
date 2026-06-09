# Phase 23: Job Worker Skeleton and Safe State Advancement

## Goal

Add a daemon-owned worker loop that advances approved jobs toward execution without enabling privileged service mutation.

## Scope

- Claim one approved job atomically and mark it `running`.
- Persist `started_at`.
- Mark running jobs as `not_implemented` when the agent returns that controlled result.
- Mark running jobs as `failed` when the agent is unavailable or returns another controlled failure.
- Audit `job.started`, `job.not_implemented`, `job.failed`, and lifecycle status changes.
- Start and stop the worker with `kojid`.

## Boundaries

- No `systemctl` execution.
- No agent mutation.
- No bypass of authentication, capability checks, CSRF, audit, or approval.
- Queued jobs are never claimed directly by the worker.

## Result

Koji now has a durable execution lifecycle skeleton. Approved jobs can be claimed and safely advanced through non-mutating worker outcomes while preserving the daemon/agent boundary.
