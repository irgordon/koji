# Metrics Reference

[Home](../Home.md) | Related: [Observability](../Operations/Observability.md), [Overview Page](../User-Guide/Overview-Page.md)

Semi-generated from `internal/observability`.

## Counters

| Metric | Meaning |
| --- | --- |
| `jobs_created_total` | Jobs created |
| `jobs_approved_total` | Jobs approved |
| `jobs_rejected_total` | Jobs rejected |
| `jobs_completed_total` | Jobs completed |
| `jobs_failed_total` | Jobs failed |
| `jobs_claimed_total` | Approved jobs claimed by the worker |
| `worker_polls_total` | Worker polling attempts |
| `worker_errors_total` | Worker loop errors |
| `agent_rpc_requests_total` | Agent RPC requests |
| `agent_rpc_failures_total` | Agent RPC failures |
| `auth_login_success_total` | Successful login attempts |
| `auth_login_failure_total` | Failed login attempts |
| `audit_writes_total` | Audit writes |
| `audit_write_failures_total` | Audit write failures |
| `readiness_checks_total` | Readiness checks |
| `readiness_db_failures_total` | DB readiness failures |
| `readiness_agent_degraded_total` | Agent readiness degradations |
| `readiness_migration_failures_total` | Migration readiness failures |

## Job Status Aggregate

`jobs_by_status` returns durable job counts grouped by job status.

## Protection

Metrics require `observability.metrics.read` and do not expose sessions, users, raw audit details, or host process rows.
