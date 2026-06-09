# INVARIANTS.md

## 1. Purpose

Invariants are rules the system must always preserve.

They are stronger than preferences. They define what must remain true as the code changes.

Every invariant should be testable, reviewable, or enforceable by design.

## 2. Core Invariants

### The browser is never authoritative.

The browser may display state and request actions.

The browser may not decide authorization, capability, trust, policy, or final state.

### The web server is never privileged.

`kojid` must not perform privileged operating system actions directly.

The web/API daemon must not directly execute privileged host mutations.

Privileged work must cross the agent boundary.

### The agent is the only privileged execution surface.

`koji-agent` is the only component allowed to perform privileged local operations.

The agent must expose narrow methods, not arbitrary shell execution.

## 3. Authentication and Authorization

- Every privileged action requires authentication.
- Every privileged action requires authorization.
- Authorization is enforced server-side.
- UI-hidden buttons are not authorization controls.
- Session state is validated before use.
- CSRF protection is required for browser session mutations.
- Bootstrap login is disabled by default.
- One-time bootstrap tokens must be stored with restrictive permissions.
- Sessions have bounded absolute and idle lifetimes.
- Revoked, expired, and idle-expired sessions are rejected.
- Valid authenticated requests update session last-seen time.

## 4. Capability Enforcement

Allowed action is always:

```text
compile-time capability ∩ runtime policy ∩ user authorization = allowed action
```

Required invariants:

- Build-disabled capabilities cannot be enabled by config.
- Runtime-disabled capabilities cannot be used by handlers.
- Unauthorized users cannot use enabled capabilities.
- Handlers must call the capability layer.
- Handlers must not read feature flags directly.
- Capability denial must fail closed.
- Authenticated users are not implicitly all-powerful.

## 5. Agent Boundary

- The web process never calls `exec.Command` for privileged actions.
- The web process never uses shell strings for system mutation.
- Privileged actions traverse the agent RPC boundary.
- The agent socket is local-only.
- Agent RPC uses Unix-domain socket transport on Unix-like systems.
- Agent RPC does not expose TCP transport.
- The agent API is allowlisted.
- The agent does not expose arbitrary command execution.
- Agent failure does not trigger a local fallback bypass.
- Service-control RPC returns controlled `mutation_disabled` unless agent mutation is explicitly enabled.
- Agent service mutation requires an agent-side allowlist.
- Service-control intent is denied when the requested unit is not allowlisted.

## 6. Configuration

- Runtime config is loaded from `/etc/koji/koji.yaml` unless explicitly overridden by command-line flag.
- Agent socket defaults to `/run/koji/agent.sock`.
- Production static assets default to `/usr/share/koji/dist`.
- Production service APIs require an explicit service allowlist.
- Process visibility defaults to summary metadata.
- Process command lines are omitted unless explicitly enabled.
- Process API result count is bounded.
- Unknown configuration fields are rejected.
- Invalid configuration prevents startup.
- Capability mismatches prevent startup.
- Runtime config is policy, not authority.
- The web process must not write config files.
- Config reload, if implemented, re-reads from disk and validates before applying.

## 7. Database and Migrations

- SQLite is authoritative for persisted state.
- Migrations are embedded in the binary.
- Migration checksums are immutable once applied.
- A changed checksum for an applied migration prevents startup or migration.
- Production down migrations are not supported.
- Normal startup refuses pending migrations unless first-boot auto-init is explicitly allowed.
- Database writes use explicit stores or repositories.
- Foreign keys are enabled.
- WAL mode is used unless the platform cannot support it.

## 8. Audit

- Audit history is append-only.
- Prior audit events are not mutated.
- Privileged mutations produce audit events.
- Privileged intent is audited even when denied.
- Security-sensitive reads produce audit events when required by policy.
- Audit events include timestamp, action, target, outcome, reason code, and user ID when known.
- Audit events include the stable request ID when an event originates from HTTP.
- Process listing access produces an audit event.
- Dev-mode bypasses are explicitly marked in audit records.
- Audit write failure prevents privileged mutation unless the action is explicitly marked audit-best-effort.
- Audit read access requires `audit.events.read`.
- Audit read APIs expose normalized fields only.
- Audit read APIs do not expose raw actor metadata, user data, remote addresses, or internal messages.

## 9. Jobs

- Long-running operations execute as jobs.
- Mutating operations that may block execute as jobs.
- Service-control intent creates a durable job before execution.
- Job read access requires `jobs.read`.
- Job approval and rejection require `jobs.approve`.
- Only queued jobs may be approved or rejected.
- The job worker may claim only approved jobs.
- The job worker delegates service mutation to the agent boundary.
- Agent service mutation is disabled by default and must be explicitly enabled.
- Jobs emit events.
- Job events preserve progress and failure state.
- Job lifecycle status changes are audited.
- Job output is bounded.
- Job artifacts are stored outside the primary database.
- Job cancellation uses context cancellation or equivalent controlled shutdown.

## 10. Sensors

- Sensor collection failure does not crash the daemon.
- Missing sensor sources degrade gracefully.
- Unsupported sensor sources do not produce fake zero values.
- Sensor IDs are stable across reboots where possible.
- `hwmon` indexes are not durable IDs.
- High-cardinality time-series data is not persisted by default.
- Live streams use bounded ring buffers.

## 11. External Commands

- No shell command is constructed from user input.
- Direct command execution is owned by `internal/platform/command`.
- External commands use explicit executable and argument arrays.
- Inputs are validated before reaching an adapter.
- Command timeouts are required.
- Command output is bounded.
- Command executable names are allowlisted.
- Slow commands do not block hot polling loops.

## 12. HTTP and UI

- Login surface is separate from authenticated SPA serving.
- Unauthenticated routes do not expose SPA chunks.
- CSP is strict on unauthenticated routes.
- Production static serving does not depend on the process working directory.
- API JSON responses use no-store caching policy.
- Authenticated APIs re-check authentication and authorization.
- API responses do not leak secrets.
- Error messages are useful but do not reveal sensitive internals.
- Production session cookies are Secure, HttpOnly, SameSite=Strict, and scoped to `/`.
- HTTP responses include `X-Request-ID`.
- Inbound `X-Request-ID` values are accepted only when valid and bounded.
- Service APIs do not accept arbitrary unit enumeration through user input.
- Process API responses are shaped by configured visibility policy before serialization.
- `/healthz` and `/readyz` are non-telemetry operational endpoints.
- Health endpoints do not expose metrics, service data, process data, users, sessions, or audit records.

## 13. Security-Sensitive Files

- `/etc/koji/koji.yaml` is root-owned and not writable by the web process.
- `/var/lib/koji/bootstrap.token` is one-time use and permission-restricted.
- Unix sockets use restrictive permissions.
- Stale socket cleanup removes only actual socket files.
- Job artifacts are not web-served without authorization.
- Support bundles must redact secrets.

## 14. Negative Patterns

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
Sensor missing -> fake zero value
Unbounded job output
Unbounded live stream buffer
Package-level mutable registry
```
