# Audit Event Reference

[Home](../Home.md) | Related: [Audit](../Security/Audit.md), [Activity Page](../User-Guide/Activity-Page.md), [Architectural Inventory](../Developer/Architectural-Inventory.md)

Audit actions are defined in `internal/audit`. Audit records include timestamp, action, target, outcome, reason code, request ID, user ID when known, remote address when available, and a dev-bypass marker when relevant.

| Action | When Emitted | Required Fields | Purpose |
| --- | --- | --- | --- |
| `auth.bootstrap` | Bootstrap success or failure. | action, target `auth`, outcome, reason code, request ID. | Prove first-user creation attempts. |
| `auth.login` | Login success or failure. | action, target `auth`, outcome, reason code, request ID. | Track authentication attempts without exposing credentials. |
| `auth.magic_token_success` | Magic token login succeeds. | user ID, target `auth`, outcome `success`, reason code, request ID. | Track passwordless sign-in. |
| `auth.magic_token_failure` | Magic token login fails. | action, target `auth`, outcome `failure`, reason code, request ID. | Track passwordless login failure without exposing token values. |
| `auth.password_denied_non_super_admin` | Non-Super Admin password login is attempted. | action, target `auth`, outcome `failure`, reason code, request ID. | Prove password login remains limited to Super Admin accounts. |
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
| `identity.user_created` | Identity administrator creates a managed user. | user ID, target user, outcome `success`, reason code, request ID. | Track user creation. |
| `identity.user_disabled` | Identity administrator disables a user. | user ID, target user, outcome `success`, reason code, request ID. | Track access removal. |
| `identity.user_enabled` | Identity administrator enables a user. | user ID, target user, outcome `success`, reason code, request ID. | Track access restoration. |
| `identity.capability_granted` | Identity administrator grants a user capability. | user ID, target user/capability, outcome `success`, reason code, request ID. | Track authorization expansion. |
| `identity.capability_revoked` | Identity administrator revokes a user capability. | user ID, target user/capability, outcome `success`, reason code, request ID. | Track authorization reduction. |
| `identity.magic_token_issued` | Identity administrator issues a one-time magic token. | user ID, target user, outcome `success`, reason code, request ID. | Track passwordless token issuance. |
| `identity.magic_token_consumed` | A magic token creates a session and is consumed. | user ID, target `auth`, outcome `success`, reason code, request ID. | Track one-time token use. |
| `identity.magic_token_expired` | Expired magic token login is attempted. | action, target `auth`, outcome `failure`, reason code, request ID. | Track expired token use attempts. |
| `identity.magic_token_revoked` | A magic token is revoked. | user ID when known, target token/user, outcome, reason code. | Reserved for token revocation workflows. |
| `identity.self_lockout_prevented` | Koji blocks disabling or de-authorizing the final identity administrator. | user ID, target user/capability, outcome `denied`, reason code, request ID. | Prove self-lockout protection. |

## Outcomes

`success`, `failure`, `denied`, `accepted`, and `rejected`.

## Exposed Activity Fields

The Activity API returns only timestamp, action, target, outcome, reason code, and request ID.

## Protected Internals

Raw internal errors, detailed remote metadata, SQL errors, command output, password data, and session token material are not exposed through Activity.
