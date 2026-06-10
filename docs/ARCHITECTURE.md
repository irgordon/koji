# ARCHITECTURE.md

## 1. Purpose

Koji is a governed Linux control panel for host visibility, durable operational intent, human approval, auditability, and a narrow daemon-to-agent privilege boundary.

The system is intentionally small: a Go daemon, a local Go agent, SQLite, a React/TypeScript frontend, packaging scripts, and generated API documentation.

## 2. Design Goals

- Keep `kojid` unprivileged.
- Keep privileged mutation behind `koji-agent`.
- Deny by default for authentication, capabilities, service allowlists, and process visibility.
- Store governance state durably in SQLite.
- Make jobs, approvals, audit, request IDs, observability, backup, recovery, and upgrade safety operationally visible.
- Keep release artifacts and runtime paths independent of developer workstations.

## 3. Non-Goals

- General-purpose remote shell.
- Arbitrary command execution through HTTP.
- Prometheus, OpenTelemetry, Grafana, or external telemetry backends.
- High availability, replication, clustering, or distributed storage.
- Browser-side authorization.
- Production down migrations.

## 4. Repository Layout

```text
/cmd/kojid             # Web/API daemon, unprivileged
/cmd/koji-agent        # Local agent process
/internal/auth         # Bootstrap, login, sessions, CSRF
/internal/caps         # Capability constants and SQLite-backed checks
/internal/config       # Strict runtime configuration loading and validation
/internal/db           # SQLite, migrations, schema compatibility
/internal/http         # Routes, middleware, handlers, SPA serving
/internal/agent        # Unix socket RPC client/server and guarded executor
/internal/system       # Read-only host observation and policy shaping
/internal/jobs         # Durable jobs, approvals, worker lifecycle
/internal/audit        # Durable audit writes and activity read model
/internal/observability # Control-plane counters and snapshots
/internal/platform     # Bounded command execution
/web                   # React, TypeScript, plain CSS frontend
/packaging             # Install, release, backup, restore, upgrade scripts
/docs                  # Governance docs, wiki, phase reports, OpenAPI
```

## 5. Component Model

### `kojid`

`kojid` owns the web/API control plane:

- Config loading and startup validation.
- SQLite initialization, schema compatibility checks, and migrations.
- Authentication, sessions, CSRF, and cookies.
- Capability enforcement.
- Protected host observation APIs.
- Durable job creation and approval/rejection APIs.
- Job worker lifecycle.
- Audit writes and normalized Activity reads.
- Control-plane observability.
- Authenticated production SPA serving.

`kojid` must not execute privileged host mutations.

### `koji-agent`

`koji-agent` owns the local privileged boundary:

- Unix-domain socket RPC.
- Independent validation of service name, action, and allowlist.
- Disabled-by-default mutation.
- Bounded command-runner integration when mutation is explicitly enabled.
- Normalized response codes.

The agent does not expose arbitrary shell execution or TCP transport.

### Frontend

The frontend is a React/TypeScript operator workspace. It renders Overview, Services, Processes, Jobs, Activity, and Settings.

It may request actions and display state. It does not decide authentication, authorization, policy, approval, audit, or final state.

## 6. Trust Boundaries

```text
Browser
  -> kojid, unprivileged
  -> Unix socket RPC
  -> koji-agent, privileged boundary
  -> Operating system
```

The browser is untrusted. `kojid` is not privileged. `koji-agent` is the only privileged execution surface.

## 7. Authentication, Authorization, and Audit

Authentication proves identity.

Capabilities authorize surfaces.

Audit records security-sensitive actions and privileged intent.

State-changing authenticated browser requests require CSRF validation. Every request receives a bounded request ID that is returned in `X-Request-ID` and copied into audit events where applicable.

## 8. Capability Model

Protected surfaces require explicit user capabilities:

```text
host.metrics.read
host.disk.read
host.services.read
host.processes.read
host.services.control
jobs.read
jobs.approve
audit.events.read
observability.metrics.read
```

Authenticated users are not implicitly administrators.

## 9. Configuration Model

Default daemon config:

```text
/etc/koji/koji.yaml
```

Default database:

```text
/var/lib/koji/koji.db
```

Default agent socket:

```text
/run/koji/agent.sock
```

Default production static assets:

```text
/usr/share/koji/dist
```

Production requires an explicit `service_allowlist`. Process visibility defaults to summary mode, hides command lines, and bounds row count. Agent mutation is disabled unless explicitly enabled in agent config with an agent-side allowlist.

## 10. Database and Migrations

SQLite stores users, sessions, capabilities, audit events, jobs, approvals, bootstrap state, and migration state.

SQLite is opened with:

```text
WAL
foreign_keys=ON
busy_timeout
```

Startup performs:

```text
open database
  -> initialize connection
  -> check schema compatibility
  -> reject future/corrupt schema
  -> run pending forward migrations
```

Migrations are immutable after release. Applied migration checksum changes fail. Future schemas fail startup to prevent accidental downgrades.

## 11. Jobs and Worker

Service-control intent becomes a durable job:

```text
queued -> approved -> running -> completed
       -> rejected
                  -> failed
                  -> not_implemented
```

Only queued jobs can be approved or rejected. Only approved jobs can be claimed by the worker. The worker delegates service mutation to the agent boundary.

## 12. Service Control

Service status observation remains read-only and bounded.

Service-control intent requires:

```text
session
CSRF
host.services.control
daemon service allowlist
job creation
audit
approval
worker
agent RPC
agent mutation enabled
agent allowlist
```

No HTTP handler executes `systemctl`.

## 13. Observability

Koji exposes governed control-plane metrics through `/api/observability/metrics` with `observability.metrics.read`.

Metrics answer whether jobs are flowing, the worker is polling, agent RPC is failing, logins are succeeding or failing, audit writes are succeeding, and readiness dependencies are healthy.

## 14. Operations

Release artifacts are built and smoke-gated by CI.

Backups use SQLite `.backup`, include config, and write metadata.

Restore validates archive structure, database integrity, schema version, and expected governance row counts.

Upgrade checks report current versus target schema, require backup awareness, reject future schemas, and verify core tables after upgrade.

## 15. API Contract

`docs/api/openapi.json` is the authoritative machine-readable HTTP contract. Generated YAML and markdown references are validated against registered routes with `make verify-openapi`.
