# Capability Reference

[Home](../Home.md) | Related: [Capabilities](../Security/Capabilities.md), [API Reference](API-Reference.md), [Architectural Inventory](../Developer/Architectural-Inventory.md)

Capabilities are defined in `internal/caps`. Authentication proves identity; capabilities authorize individual surfaces. Deny is the default.

| Capability | Purpose | Protected Endpoints | Risk Level |
| --- | --- | --- | --- |
| `host.metrics.read` | Read CPU, memory, and uptime telemetry. | `GET /api/v1/metrics` | Medium: host telemetry visibility. |
| `host.disk.read` | Read filesystem usage telemetry. | `GET /api/v1/disk` | Medium: storage visibility. |
| `host.services.read` | Read allowlisted service status. | `GET /api/v1/services` | Medium: service topology visibility. |
| `host.processes.read` | Read process listing subject to visibility policy. | `GET /api/v1/processes` | High: process metadata may reveal workload details. |
| `host.services.control` | Create durable service-control jobs. | `POST /api/services/{name}/{action}` | High: privileged intent; still requires allowlist, audit, job, approval, worker, and agent guardrails. |
| `jobs.read` | Read job list and details. | `GET /api/jobs`, `GET /api/jobs/{id}` | Medium: operational intent visibility. |
| `jobs.approve` | Approve or reject queued jobs. | `POST /api/jobs/{id}/approve`, `POST /api/jobs/{id}/reject` | High: human authorization boundary for future mutation. |
| `audit.events.read` | Read normalized audit activity. | `GET /api/activity` | High: governance history visibility. |
| `observability.metrics.read` | Read control-plane counters and job status aggregates. | `GET /api/observability/metrics` | Medium: operational health visibility. |

## Failure Modes

- Missing capability returns a safe permission error.
- Capability denial records audit where the surface is audited.
- Dev-mode bypass remains explicit and visibly marked in audit events.

## Assignment Guidance

Grant the smallest capability set required for the operator task. In particular, keep `jobs.approve`, `host.services.control`, `audit.events.read`, and `host.processes.read` tightly scoped.
