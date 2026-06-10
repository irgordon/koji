# API Reference

[Home](../Home.md) | Related: [Request Flow](../Architecture/Request-Flow.md), [Capability Reference](Capability-Reference.md), [Error Code Reference](Error-Code-Reference.md)

Semi-generated from `internal/http/routes.go`.

| Method | Path | Purpose | Capability Required | Response Shape |
| --- | --- | --- | --- | --- |
| GET | `/healthz` | Liveness | Public | `{status, checks}` |
| GET | `/readyz` | Dependency readiness | Public | `{status, checks}` |
| GET | `/api/v1/metrics` | Host CPU, memory, uptime | `host.metrics.read` | `SystemMetrics` |
| GET | `/api/v1/disk` | Disk usage | `host.disk.read` | `DiskMetrics` |
| GET | `/api/v1/services` | Allowlisted service status | `host.services.read` | `{services}` |
| POST | `/api/services/{name}/{action}` | Create service-control job | `host.services.control` plus CSRF | `{jobId, status}` |
| GET | `/api/v1/processes` | Process list | `host.processes.read` | `ProcessInfo[]` |
| GET | `/api/activity` | Recent audit activity | `audit.events.read` | `{events}` |
| GET | `/api/observability/metrics` | Control-plane metrics | `observability.metrics.read` | `{counters, jobs_by_status}` |
| GET | `/api/jobs` | Recent jobs | `jobs.read` | `{jobs}` |
| GET | `/api/jobs/{id}` | Job detail | `jobs.read` | `Job` |
| POST | `/api/jobs/{id}/approve` | Approve queued job | `jobs.approve` plus CSRF | `Job` |
| POST | `/api/jobs/{id}/reject` | Reject queued job | `jobs.approve` plus CSRF | `Job` |
| POST | `/api/bootstrap` | First user bootstrap | Public until first user | Auth session response |
| POST | `/api/login` | Login | Public | Auth session response |
| POST | `/api/logout` | Logout | Valid session plus CSRF | `{status}` |
| GET | `/api/session` | Session status | Public | `{authenticated, bootstrapRequired, username?}` |

## Protection

Protected API routes are wrapped by auth middleware and capability checks. State-changing authenticated routes require CSRF.

## Failure Modes

Routes return safe JSON errors for authentication, permission, CSRF, validation, allowlist, agent, and unexpected response failures.
