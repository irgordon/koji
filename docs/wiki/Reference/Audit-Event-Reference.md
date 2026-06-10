# Audit Event Reference

[Home](../Home.md) | Related: [Audit](../Security/Audit.md), [Activity Page](../User-Guide/Activity-Page.md)

Semi-generated from `internal/audit`.

## Actions

| Action | Meaning |
| --- | --- |
| `auth.login` | Login attempt |
| `auth.logout` | Logout attempt |
| `auth.bootstrap` | Initial user bootstrap attempt |
| `capability.denied` | Capability denied |
| `capability.bypass` | Explicit dev-mode bypass |
| `service.control` | Service-control intent accepted, denied, or rejected |
| `process.list` | Process list read |
| `job.created` | Durable job created |
| `job.viewed` | Job list or detail viewed |
| `job.status_changed` | Job status changed |
| `job.approved` | Job approved |
| `job.rejected` | Job rejected |
| `job.approval_denied` | Approval or rejection denied |
| `job.started` | Worker started a job |
| `job.not_implemented` | Agent reported not implemented |
| `job.completed` | Job completed |
| `job.failed` | Job failed |
| `job.command_failed` | Agent command failed |
| `job.command_timeout` | Agent command timed out |
| `job.mutation_disabled` | Agent mutation disabled |

## Outcomes

`success`, `failure`, `denied`, `accepted`, and `rejected`.

## Exposed Activity Fields

The Activity API returns timestamp, action, target, outcome, reason code, and request ID.

## Protected Internals

Raw internal errors, detailed remote metadata, SQL errors, and command output are not exposed through Activity.
