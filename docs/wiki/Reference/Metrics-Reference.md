# Metrics Reference

[Home](../Home.md) | Related: [Observability](../Operations/Observability.md), [Overview Page](../User-Guide/Overview-Page.md), [Architectural Inventory](../Developer/Architectural-Inventory.md)

Metrics are internal control-plane counters from `internal/observability`. Koji does not expose Prometheus, OpenTelemetry, Grafana, or external collectors in this phase.

## Counters

| Metric | Meaning | Source |
| --- | --- | --- |
| `agent_rpc_failures_total` | Agent RPC calls that failed or returned an error condition. | Agent client and worker paths. |
| `agent_rpc_requests_total` | Agent RPC requests attempted. | Agent client and worker paths. |
| `audit_write_failures_total` | Audit writes that failed. | `audit.Store.Record`. |
| `audit_writes_total` | Audit writes that succeeded. | `audit.Store.Record`. |
| `auth_login_failure_total` | Failed login attempts. | Auth handlers. |
| `auth_login_success_total` | Successful login attempts. | Auth handlers. |
| `jobs_approved_total` | Jobs approved by a user with `jobs.approve`. | `jobs.Store.Approve`. |
| `jobs_claimed_total` | Approved jobs claimed by the worker. | `jobs.Store.ClaimApproved`. |
| `jobs_completed_total` | Jobs completed successfully. | `jobs.Store.MarkCompleted`. |
| `jobs_created_total` | Durable jobs created. | `jobs.Store.Create`. |
| `jobs_failed_total` | Jobs marked failed. | `jobs.Store.MarkFailed`. |
| `jobs_rejected_total` | Jobs rejected by a user with `jobs.approve`. | `jobs.Store.Reject`. |
| `readiness_agent_degraded_total` | Readiness checks where agent reachability is degraded. | `/readyz`. |
| `readiness_checks_total` | Readiness checks performed. | `/readyz`. |
| `readiness_db_failures_total` | Readiness checks where DB access failed. | `/readyz`. |
| `readiness_migration_failures_total` | Readiness checks where migration currentness failed. | `/readyz`. |
| `worker_errors_total` | Worker loop errors. | Job worker. |
| `worker_polls_total` | Worker polling attempts. | Job worker. |

## Job Status Aggregate

`jobs_by_status` returns durable job counts grouped by the current job status in SQLite.

## Protection

Metrics require `observability.metrics.read` and do not expose sessions, users, raw audit details, process rows, command output, or service-control payload internals.
