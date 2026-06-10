# SECURITY.md

## 1. Security Posture

Koji is control-plane software. It can observe host state and request privileged service-control intent through a governed workflow.

The security posture is conservative:

- Minimize privilege.
- Isolate privileged mutation in the local agent.
- Require authentication, CSRF, capabilities, allowlists, jobs, approvals, and audit.
- Fail closed.
- Keep runtime behavior explicit and testable.

## 2. Trust Boundaries

```text
Browser -> kojid -> koji-agent -> Operating system
```

The browser is untrusted. `kojid` is unprivileged. `koji-agent` is the privileged boundary and must stay narrow.

## 3. Authentication and Sessions

Browser sessions use `koji_session` and CSRF uses `koji_csrf` plus `X-CSRF-Token`.

Production session cookies are HttpOnly, SameSite=Strict, Secure, and scoped to `/`. Development mode keeps local HTTP testing explicit.

Sessions have absolute and idle timeouts. Revoked, expired, and idle-expired sessions are invalid.

Bootstrap is available only while no users exist.

## 4. Authorization and Capabilities

Every protected action requires a server-side capability check. The UI may hide controls, but it is not an enforcement point.

High-risk capabilities include:

- `host.services.control`
- `jobs.approve`
- `audit.events.read`
- `host.processes.read`
- `observability.metrics.read`

Grant the minimum capability set required for the task.

## 5. Agent Security

The daemon communicates with the agent over a Unix-domain socket.

The agent:

- Validates socket path ownership and stale socket safety.
- Validates service name and action.
- Enforces an agent-side allowlist.
- Keeps mutation disabled by default.
- Uses bounded command execution through `internal/platform/command` when mutation is enabled.
- Returns normalized reason codes.

Forbidden:

```text
run arbitrary command
execute shell string
write arbitrary file
read arbitrary file without policy
accept unsanitized user input
fallback to daemon-side mutation
```

## 6. External Commands

External command execution is centralized in `internal/platform/command`.

Requirements:

- No shell.
- Explicit executable names and argument arrays.
- Executable allowlists.
- Validated inputs.
- Context timeouts.
- Bounded stdout and stderr.
- Normalized errors.

## 7. Web Security

The unauthenticated login/bootstrap surface is minimal.

The production SPA is served only after authentication.

API JSON responses use no-store caching. Production responses include browser security headers including CSP, content-type protection, referrer policy, frame restrictions, and permissions policy.

API responses must not expose password hashes, session IDs, CSRF secrets, SQL errors, command output, or internal stack traces.

## 8. Audit Requirements

Audit events are append-only and record security-sensitive actions, privileged intent, capability denial, job lifecycle changes, auth events, process-list access, and dev-mode bypasses.

Audit records include action, target, outcome, reason code, timestamp, request ID when available, and user ID when known.

The Activity API exposes only normalized fields.

## 9. Database, Backup, and Upgrade Security

SQLite lives at `/var/lib/koji/koji.db` by default and stores users, sessions, capabilities, audit, jobs, approvals, bootstrap state, and migrations.

Migrations are embedded and checksummed. Checksum mismatch, future schema, or corrupt migration history prevents startup.

Backups use SQLite `.backup` and include configuration plus metadata. Restores must be verified before normal operation.

Production down migrations are not supported. Rollback after schema changes requires restoring a pre-upgrade backup and installing the prior release artifact.

## 10. Secrets and Sensitive Data

Secrets must not be logged.

Support bundles and config dumps must redact secrets.

API responses must not expose password hashes, session secrets, CSRF secrets, or bootstrap internals.

## 11. Reporting Security Issues

Until a public process exists, report security issues through the private project maintainer channel.

Include:

- Affected version.
- Reproduction steps.
- Expected result.
- Actual result.
- Logs or support bundle if safe to share.

Do not include secrets in reports.
