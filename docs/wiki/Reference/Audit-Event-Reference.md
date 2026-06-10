# Audit Event Reference

[Home](../Home.md) | Related: [Audit](../Security/Audit.md), [Activity Page](../User-Guide/Activity-Page.md), [Architectural Inventory](../Developer/Architectural-Inventory.md)

Audit actions are defined in `internal/audit`. Audit records include timestamp, action, target, outcome, reason code, request ID, user ID when known, remote address when available, and a dev-bypass marker when relevant.

| Action | When Emitted | Required Fields | Purpose |
| --- | --- | --- | --- |
| `auth.bootstrap` | Bootstrap success or failure. | action, target `auth`, outcome, reason code, request ID. | Prove first-user creation attempts. |
| `auth.login` | Login success or failure. | action, target `auth`, outcome, reason code, request ID. | Track authentication attempts without exposing credentials. |
| `auth.logout` | Logout success or failure. | action, target `auth`, outcome, reason code, request ID. | Track session revocation attempts. |
| `capability.denied` | Protected request lacks capability. | user ID when known, action, target, outcome `denied`, reason code. | Prove deny-by-default authorization. |
| `capability.bypass` | Dev-mode bypass is used. | action, target, outcome `accepted`, dev bypass marker. | Keep dev bypass visible. |
| `service.control` | Service-control intent accepted, denied, or rejected before/while creating a job. | user ID, service target, outcome, reason code, request ID. | Track privileged intent even when denied. |
| `process.list` | Process listing succeeds. | user ID, target `host.processes`, outcome `success`. | Record access to process metadata. |
| `job.created` | Durable service-control job is created. | user ID, `jobs:<id>`, outcome `accepted`, job status reason. | Link service-control intent to a durable job. |
| `job.viewed` | Job list or detail is viewed. | user ID, target, outcome `success`. | Track access to operational intent records. |
| `job.status_changed` | Job status changes through approval or worker action. | user ID when available, `jobs:<id>`, outcome `accepted`, current status. | Correlate lifecycle changes. |
| `job.approved` | Queued job is approved. | approving user ID, `jobs:<id>`, outcome `accepted`, status `approved`. | Record human authorization. |
| `job.rejected` | Queued job is rejected. | rejecting user ID, `jobs:<id>`, outcome `rejected`, status `rejected`. | Record human rejection. |
| `job.approval_denied` | Approval/rejection attempt is denied or conflicts. | user ID when known, job target, outcome `denied`, reason code. | Track failed authority attempts. |
| `job.started` | Worker claims a job. | job target, outcome `accepted`, status reason. | Prove worker advancement. |
| `job.not_implemented` | Reserved action for not-implemented agent outcome. | job target, outcome, reason. | Historical action name; current worker records failed status path. |
| `job.completed` | Worker marks a running job completed. | job target, outcome `success`, status reason. | Prove successful completion. |
| `job.failed` | Worker marks a running job failed. | job target, outcome `failure`, status reason. | Record safe failure. |
| `job.command_failed` | Agent command path returns command failure. | job target, outcome `failure`, reason code. | Distinguish command failure from transport failure. |
| `job.command_timeout` | Agent command path times out. | job target, outcome `failure`, reason code. | Track bounded execution timeout. |
| `job.mutation_disabled` | Agent rejects mutation because it is disabled. | job target, outcome `failure`, reason code. | Prove default-disabled mutation guardrail. |

## Outcomes

`success`, `failure`, `denied`, `accepted`, and `rejected`.

## Exposed Activity Fields

The Activity API returns only timestamp, action, target, outcome, reason code, and request ID.

## Protected Internals

Raw internal errors, detailed remote metadata, SQL errors, command output, password data, and session token material are not exposed through Activity.
