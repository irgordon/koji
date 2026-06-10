# Architectural Inventory

[Home](../Home.md) | Related: [Backend Inventory](Backend-Inventory.md), [Frontend Inventory](Frontend-Inventory.md), [Phase History](Phase-History.md)

This inventory records the major Koji subsystems after Phase 36. It is architectural, not exhaustive source documentation.

| Subsystem | Purpose | Owner Package | Primary Interfaces | Failure Modes |
| --- | --- | --- | --- | --- |
| Auth | Bootstrap first user, authenticate credentials, create and revoke sessions. | `internal/auth`, `internal/http` | `auth.Store`, `/api/bootstrap`, `/api/login`, `/api/logout`, `/api/session` | Invalid credentials, disabled bootstrap, expired or revoked session, CSRF denial. |
| Sessions | Bound authenticated requests with TTL, idle timeout, CSRF token, and hardened cookies. | `internal/auth` | `SessionPolicy`, `ValidateSession`, `ValidateCSRF`, `RevokeSession` | Expired session, idle timeout, revoked session, invalid CSRF token. |
| Capabilities | Deny protected actions unless a user has the required capability. | `internal/caps`, `internal/http` | `caps.Store.Require`, protected route wrappers | Missing capability returns 403 and may audit denial. |
| Audit | Persist security and governance events, expose normalized activity. | `internal/audit` | `audit.Store.Record`, `ListRecent`, `/api/activity` | DB write failure, bounded activity listing, sensitive internals intentionally hidden. |
| Jobs | Persist service-control intent and decisions across restarts. | `internal/jobs`, `internal/http` | `jobs.Store`, `/api/jobs`, `/api/jobs/{id}` | Invalid transition, missing job, failed worker completion. |
| Worker | Poll approved jobs and advance them through the agent boundary. | `internal/jobs` | `jobs.Worker.Start`, `ClaimApproved`, `MarkCompleted`, `MarkFailed` | Agent unavailable, mutation disabled, timeout, command failure, context cancellation. |
| Agent | Own the local privileged boundary and guarded service-control RPC. | `internal/agent`, `cmd/koji-agent` | Unix socket server, `ServiceController`, `Executor` | Missing socket, refused connection, malformed response, mutation disabled, allowlist denial. |
| Observability | Report governed control-plane counters and job status aggregates. | `internal/observability`, `internal/http` | `Registry`, `/api/observability/metrics` | DB status-count query failure, missing capability, stale UI data. |
| Packaging | Stage Linux runtime layout with binaries, configs, systemd units, and static assets. | `Makefile`, `packaging/` | `make install`, `make release`, rootfs archive | Missing artifacts, invalid layout, forbidden local paths, checksum mismatch. |
| Release | Build, smoke-test, checksum, and publish tagged release artifacts. | `.github/workflows/release.yml`, `packaging/scripts` | `make release`, `make verify-release`, release smoke gate | CI failure, artifact permission loss, checksum failure, rootfs drift. |
| Backup | Create offline-restorable SQLite and configuration artifacts. | `packaging/scripts/backup.sh` | `make backup`, SQLite `.backup`, metadata counts | Missing DB/config, missing `sqlite3`, bad backup destination. |
| Recovery | Restore database/configuration and verify durable governance data. | `packaging/scripts/restore.sh`, `verify_restore.sh` | `make restore`, `make verify-restore` | Missing archive members, integrity failure, schema mismatch, count mismatch. |
| Upgrade Safety | Check schema compatibility before migrations and verify upgraded state. | `internal/db`, `packaging/scripts` | `CheckSchemaCompatibility`, `make pre-upgrade-check`, `make verify-upgrade` | Future schema, corrupt migration history, missing backup, unreadable core tables. |
| Frontend | Operator workspace for overview, services, processes, jobs, activity, and settings. | `web/src` | React views, typed API client, toast provider | Unauthorized requests, stale polling data, redacted process fields, network failure. |
| OpenAPI | Machine-readable API contract and generated reference docs. | `docs/api`, `packaging/scripts/generate_openapi_docs.mjs` | `make openapi`, `make verify-openapi` | Route drift, missing capability metadata, stale generated references. |

## Boundaries

- The browser is never authoritative.
- `kojid` may observe host state and coordinate governance.
- `koji-agent` owns privileged mutation.
- SQLite is the durable source for users, sessions, capabilities, audit, jobs, approvals, and migrations.
- OpenAPI documents HTTP contracts; it does not define runtime authorization by itself.
