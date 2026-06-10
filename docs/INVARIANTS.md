# INVARIANTS.md

## 1. Purpose

Invariants are rules Koji must preserve as the code changes. They are stronger than preferences and should be testable, reviewable, or enforced by design.

## 2. Core Invariants

### The browser is never authoritative.

The browser may display state and request actions. It may not decide authorization, capability, trust, policy, approval, audit, or final state.

### The web server is never privileged.

`kojid` must not perform privileged operating system mutations directly. The web/API daemon must not directly execute privileged host mutations.

### The agent is the only privileged execution surface.

`koji-agent` is the only component allowed to own future privileged local mutation. It must expose narrow methods, not arbitrary shell execution.

## 3. Authentication and Sessions

- Production APIs are deny-by-default unless explicitly public.
- Bootstrap exists only until the first local user is created.
- Sessions have bounded absolute and idle lifetimes.
- Revoked, expired, and idle-expired sessions are rejected.
- Valid authenticated requests update session last-seen time.
- CSRF protection is required for state-changing authenticated browser requests.
- Production cookies are Secure, HttpOnly, SameSite=Strict, and scoped to `/`.

## 4. Capability Enforcement

- Authenticated users are not implicitly all-powerful.
- Protected APIs require explicit capability checks.
- Missing capability fails closed.
- Capability denial returns safe errors.
- Capability denial is audited where the surface requires audit.

## 5. Agent Boundary

- HTTP and agent packages do not own direct command execution.
- Direct command execution is centralized in `internal/platform/command`.
- Privileged service mutation crosses the Unix socket agent RPC boundary.
- Agent RPC does not expose TCP transport.
- Agent failure does not trigger local fallback mutation.
- Service mutation is disabled by default.
- Agent mutation requires agent-side service allowlist validation.
- Agent response codes are normalized.

## 6. Configuration

- Runtime config loads from `/etc/koji/koji.yaml` unless explicitly overridden.
- Agent config loads from `/etc/koji/agent.yaml` unless explicitly overridden.
- Unknown configuration fields are rejected.
- Duplicate configuration fields are rejected.
- Invalid configuration prevents startup.
- Database, agent socket, and production static asset paths are absolute.
- Production service APIs require explicit service allowlist.
- Process visibility defaults to summary metadata with command lines omitted.
- Process API result count is bounded.
- The web process must not write config files.

## 7. Database, Migrations, Backup, and Upgrade

- SQLite is authoritative for persisted governance state.
- Foreign keys are enabled.
- WAL mode is used.
- Migration checksums are immutable once applied.
- Future schemas prevent startup.
- Corrupt migration history prevents startup.
- Pending older schemas are validated before forward migration.
- Production down migrations are not supported.
- Backups must be offline-restorable from database, config, and metadata.
- Restore verification must not require manual SQLite table inspection.

## 8. Audit

- Audit history is append-only.
- Privileged intent is audited even when denied.
- Audit events include timestamp, action, target, outcome, reason code, request ID when available, and user ID when known.
- Dev-mode bypasses are explicitly marked in audit records.
- Audit read access requires `audit.events.read`.
- Audit read APIs expose normalized fields only.
- Audit read APIs do not expose raw actor metadata, user data, remote addresses, SQL errors, command output, or internal messages.

## 9. Jobs and Worker

- Service-control intent creates a durable job.
- Job read access requires `jobs.read`.
- Job approval and rejection require `jobs.approve`.
- Only queued jobs may be approved or rejected.
- Only approved jobs may be claimed by the worker.
- Worker state transitions are durable and audited.
- The worker delegates service mutation to the agent boundary.
- Terminal jobs are not retried in place.
- Worker shutdown uses context cancellation.

## 10. Service and Process Exposure

- Service APIs do not allow arbitrary systemd enumeration through user input.
- Service status and service-control intent require daemon service allowlist membership.
- Process API responses are shaped by configured visibility policy before serialization.
- Full command lines are hidden unless explicitly enabled.
- Redacted process fields cannot be recovered by the frontend.

## 11. External Commands

- No shell command is constructed from user input.
- External commands use explicit executable names and argument arrays.
- Command timeouts are required.
- Command output is bounded.
- Command executable names are allowlisted.
- Errors are normalized before returning to higher layers.

## 12. HTTP and UI

- Login surface is separate from authenticated SPA serving.
- Production SPA serving requires authentication.
- API JSON responses use no-store caching policy.
- API responses do not leak secrets.
- HTTP responses include `X-Request-ID`.
- Inbound `X-Request-ID` values are accepted only when valid and bounded.
- `/healthz` and `/readyz` do not expose protected host telemetry, users, sessions, or audit records.
- UI status is not authority; server responses remain authoritative.

## 13. Negative Patterns

The following patterns are invariant violations:

```text
HTTP handler -> privileged exec.Command
HTTP handler -> shell string
Browser state -> authorization decision
UI-only enforcement of privileged actions
Config POST body -> /etc/koji/koji.yaml
Root web process with broad privileges
Arbitrary command endpoint
Silent capability no-op
Mutable audit history
Unchecked migration checksum change
Future schema accepted by older binary
Unbounded command output
Unbounded process listing
Package-level mutable security policy
```
