# Capability Reference

[Home](../Home.md) | Related: [Capabilities](../Security/Capabilities.md), [API Reference](API-Reference.md)

Semi-generated from `internal/caps`.

| Capability | Protects |
| --- | --- |
| `host.metrics.read` | Host metrics API |
| `host.disk.read` | Disk metrics API |
| `host.services.read` | Allowlisted service status |
| `host.processes.read` | Process listing |
| `host.services.control` | Service-control job creation |
| `audit.events.read` | Activity audit read model |
| `jobs.read` | Jobs list and detail APIs |
| `jobs.approve` | Job approval and rejection |
| `observability.metrics.read` | Control-plane metrics API |

## Failure Modes

Missing capability returns a safe permission error and records a denied audit event where the surface is audited.

## Recovery

Grant only the minimal required capability for the task.
