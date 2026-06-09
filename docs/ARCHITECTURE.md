# ARCHITECTURE.md

## 1. Purpose

Koji is a small server control panel for bare-metal and VPS systems. It provides a secure web interface for system visibility and controlled operations without requiring a large runtime, container stack, or framework-heavy deployment model.

The project is built as a surgical tool, not a cathedral. The system must stay small, auditable, resource-efficient, and explicit.

## 2. Design Goals

- Ship as a small static Go binary where practical.
- Keep the web/API process unprivileged.
- Isolate privileged work behind a narrow local agent boundary.
- Prefer Linux-native data sources such as `/proc` and `/sys` before external tools.
- Keep the React UI dynamic but non-authoritative.
- Use SQLite for durable state.
- Use bounded memory structures for live streams.
- Fail closed when configuration, policy, authorization, or capability checks do not pass.
- Make operational failures visible and diagnosable.

## 3. Non-Goals

- Do not become a general-purpose remote shell.
- Do not expose arbitrary command execution through HTTP.
- Do not replace full observability platforms such as Netdata, Prometheus, Grafana, or Cockpit.
- Do not require Docker, Node.js, Python, or a package-manager runtime on target servers.
- Do not assume every host has hardware sensors.
- Do not silently ignore missing capabilities.
- Do not allow the browser to become authoritative for state, policy, or security decisions.

## 4. Repository Layout

```text
/cmd/kojid             # Web/API daemon, unprivileged
/cmd/koji-agent        # Privileged helper
/internal/auth         # Sessions, CSRF, MFA
/internal/http         # Routes, middleware, SPA serving
/internal/caps         # CapabilitySet enforcement
/internal/config       # Strict runtime configuration loading and validation
/internal/db           # SQLite, embedded migrations
/internal/agent        # Client for agent RPC
/internal/system       # procfs, sysfs, smart, ipmi backends
/internal/sensors      # Collection, normalization, degradation logic
/internal/stream       # WebSocket/SSE, ring buffers, live updates
/internal/jobs         # Long-running operations and job events
/internal/audit        # Append-only audit events
/internal/platform     # OS-specific helpers and build-tagged files
/web                   # TypeScript and React UI
/docs                  # Additional design notes and phase reports
```

## 5. Component Model

### `kojid`

`kojid` is the web/API daemon. It serves the login page, authenticated SPA shell, JSON APIs, live streams, and operational commands.

It must run without broad privileges.

It owns:

- Authentication.
- Session management.
- CSRF enforcement.
- Role-based authorization.
- Capability checks.
- Job creation.
- Audit event recording.
- SQLite state access.
- UI asset serving.

It must not directly perform privileged OS actions.

### `koji-agent`

`koji-agent` is the privileged helper. It exposes a narrow local RPC surface over a Unix domain socket.

It owns:

- Protected log reads.
- Service control.
- SMART reads when elevated access is required.
- IPMI calls when enabled.
- Other explicit privileged operations.

It must not expose a general shell interface.

Service-control mutation is disabled by default. Valid service-control RPC requests return controlled reason codes such as `mutation_disabled`, `service_not_allowlisted`, `unsupported_action`, `command_failed`, or `command_timeout`.

### React UI

The UI is a TypeScript and pure React interface. It displays state and requests intent. It does not enforce security policy.

The UI must assume that every server response can deny an action, even if the UI showed the action as available.

## 6. Trust Boundaries

```text
Browser
  -> HTTPS / reverse proxy
  -> kojid, unprivileged
  -> Unix socket RPC
  -> koji-agent, privileged
  -> Operating system
```

The important trust boundary is between `kojid` and `koji-agent`.

The browser is untrusted.

The web process is not privileged.

The agent is the only privileged execution surface.

## 7. Authentication and Authorization

Authentication proves who is making the request.

Authorization decides what that identity may do.

Capability enforcement decides whether the build and deployment permit the requested action.

Audit records what was attempted.

Every HTTP request has a bounded request ID. The request ID is returned in the `X-Request-ID` response header, included in request logs, and copied into audit events.

All privileged actions must pass:

```text
authentication -> authorization -> capability check -> request validation -> job/agent execution -> audit event
```

The order matters. Do not execute work before all checks pass.

## 8. Capability Model

Allowed action is the intersection of three controls:

```text
compile-time capability ∩ runtime policy ∩ user authorization = allowed action
```

Compile-time capability defines what the binary can physically do.

Runtime policy defines what the deployment permits.

User authorization defines what the current user may request.

Handlers must call the capability layer. Handlers must not read feature flags directly.

Initial host capabilities:

```text
host.metrics.read
host.disk.read
host.services.read
host.processes.read
host.services.control
audit.events.read
jobs.read
jobs.approve
```

Authenticated users have no implicit administrative capability. Missing capability denies by default.

## 9. Configuration Model

Runtime configuration lives at:

```text
/etc/koji/koji.yaml
```

Default agent socket:

```text
/run/koji/agent.sock
```

Default production static asset directory:

```text
/usr/share/koji/dist
```

Service APIs are constrained by an explicit `service_allowlist` in production. Development mode may use a narrow local default for convenience, but production must name eligible systemd units explicitly.

Process APIs are constrained by `process_visibility_mode`, `include_command_line`, and `max_processes`. Production defaults are conservative: summary visibility, command lines omitted, and bounded result count.

The config loader must:

- Reject unknown fields.
- Validate all paths, durations, booleans, listeners, and feature settings.
- Fail closed on invalid combinations.
- Fail if runtime policy enables a capability not present in the build.
- Require absolute database and agent socket paths.
- Require an absolute static asset directory in production.
- Require an explicit service allowlist in production.
- Validate process visibility mode, command-line exposure, and process result limits.

The web process must never write `/etc/koji/koji.yaml`.

If reload is supported, it must re-read from disk after a signal or explicit admin action. It must not accept POST-body-to-disk configuration writes.

## 10. Database Architecture

SQLite is the durable state store.

Default path:

```text
/var/lib/koji/koji.db
```

SQLite stores:

- Users.
- Roles.
- Sessions.
- API tokens.
- Audit events.
- Jobs.
- Job events.
- Config revisions.
- Agent inventory.
- Sensor last-seen state.

SQLite should not store high-cardinality live metrics by default. Live streams use bounded in-memory ring buffers. Optional sensor history must be explicitly enabled.

SQLite must use:

```text
WAL journal mode
foreign_keys=ON
busy_timeout
explicit migrations
```

## 11. Migration Architecture

Migrations are embedded in the binary.

Migrations are immutable after release.

Applied migration checksums must never change.

No production down migrations are supported. Roll forward with a new migration.

Normal daemon startup should refuse to run if pending migrations exist, unless the deployment explicitly allows first-boot initialization.

Required commands:

```text
kojid -configtest /etc/koji/koji.yaml
kojid -migrate /etc/koji/koji.yaml
kojid -doctor /etc/koji/koji.yaml
kojid -print-capabilities
kojid -version
```

## 12. Sensor Architecture

Sensors are normalized into one contract:

```go
type SensorSample struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Kind      string    `json:"kind"`
    Value     float64   `json:"value"`
    Unit      string    `json:"unit"`
    Source    string    `json:"source"`
    Timestamp time.Time `json:"timestamp"`
}
```

Sensor IDs must be stable. Do not use unstable `hwmon0` indexes as durable IDs. Prefer source, device name, and label.

Sensor collection must degrade gracefully:

```text
hwmon available      -> show temperature, fan, voltage, power sensors
procfs available     -> show CPU, memory, load, disk, network counters
smartctl available   -> show disk health and wear
ipmitool available   -> show BMC and out-of-band readings
virtualized host     -> suppress unsupported hardware sensor cards
```

A missing sensor source is not a daemon failure.

## 13. Job and Event Model

Long-running or mutating operations must execute as jobs.

Service-control requests create queued durable jobs before any privileged execution is attempted.

Queued service-control jobs require an explicit human approval or rejection decision before any future worker or agent execution path may consume them.

The daemon worker may claim only approved jobs. Service-control jobs can advance to `running`, then to `completed` or `failed` based on normalized agent RPC results.

The agent independently validates service name, action, and its own service allowlist before mutation can run. When mutation is disabled, requests stop at `mutation_disabled`.

Examples:

- Restart service.
- Run package upgrade.
- Collect support bundle.
- Read slow disk health source.
- Change firewall policy.

Jobs should emit events such as:

```text
job.created
job.approved
job.rejected
job.approval_denied
job.started
job.not_implemented
job.stdout
job.stderr
job.completed
job.failed
job.cancelled
```

Current job state may be derived from job events or stored as a cached summary. The event history is the source for replay and diagnosis.

## 14. Audit Model

Audit history is append-only.

Do not mutate prior audit events. Add a new event.

Audit events must record:

- User ID when known.
- Action.
- Target.
- Timestamp.
- Outcome.
- Reason code.
- Request ID when present.
- Remote address when available.
- Dev-mode bypass marker when applicable.
- Source address when trustworthy.
- User agent when available.
- Error summary when applicable.

Audit logging must happen for privileged, destructive, security-sensitive, and policy-relevant actions.

Audit read access is governed separately from audit writes. The activity API exposes only normalized read-model fields, not raw actor metadata, user records, remote addresses, or internal messages.

## 15. Deployment Model

Default paths:

```text
/etc/koji/koji.yaml
/var/lib/koji/koji.db
/var/lib/koji/bootstrap.token
/var/lib/koji/artifacts/
/run/koji/koji.sock
/run/koji/agent.sock
```

Use systemd-managed directories where possible:

```ini
RuntimeDirectory=koji
RuntimeDirectoryMode=0750
StateDirectory=koji
StateDirectoryMode=0750
ConfigurationDirectory=koji
ConfigurationDirectoryMode=0750
```

The deployment should work on bare metal and VPS hosts. Hardware-specific features must degrade on virtualized systems.

## 16. Security Model

The system is designed to reduce blast radius.

Required security properties:

- The web process is unprivileged.
- Privileged actions traverse the agent boundary.
- The agent exposes narrow operations, not shell access.
- Agent RPC uses a local Unix-domain socket, not TCP.
- All mutations require authorization and audit.
- CSRF is enforced for browser sessions.
- Login surface is served separately from the authenticated SPA.
- The unauthenticated surface does not expose SPA chunks, route names, or internal state.
- `/healthz` and `/readyz` are unauthenticated for local supervisor and packaging integration.
- Health endpoints return compact non-telemetry check outcomes only.
- CSP is strict on unauthenticated routes.
- Production responses include browser security headers.
- API JSON responses are not cached.
- External process execution uses argument arrays, not shell strings.
- External command execution is centralized in `internal/platform/command`.
- Command execution has timeouts, stdout/stderr byte limits, and executable allowlists.
- User input is allowlisted before it reaches OS integration layers.

Session lifecycle rules:

- Sessions have an absolute TTL.
- Sessions have an idle timeout.
- Valid authenticated requests update `last_seen_at`.
- Expired, idle-expired, and revoked sessions are rejected.
- Production session cookies use `Secure`, `HttpOnly`, `SameSite=Strict`, and `Path=/`.
- Development mode explicitly omits `Secure` so local HTTP testing can work.

## 17. Failure Modes

The system must fail loudly and safely.

Examples:

- Invalid config prevents startup.
- Pending migrations prevent normal startup.
- Capability mismatch prevents startup.
- Agent unavailable causes privileged actions to fail, not bypass.
- Sensor failure emits degraded status, not fake zero values.
- Audit write failure prevents privileged mutation unless the action is explicitly marked audit-best-effort.

## 18. Forbidden Architecture

The following patterns are forbidden:

```text
HTTP handler -> exec.Command
HTTP handler -> shell string
Browser -> privileged action
UI state -> authorization decision
Config POST body -> /etc/koji/koji.yaml
Root web server with broad privileges
Arbitrary command endpoint
Silent capability no-op
Mutable audit history
Unchecked migration checksum changes
```
