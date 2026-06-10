# Changelog

All notable changes to this project should be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project follows semantic versioning.

This project follows a forward-only operational style. Database schema changes roll forward. Production down migrations are not supported.

## [v0.0.0]

### Added

- SQLite foundation with strict startup configuration loading and deterministic migrations.
- Deny-by-default session gate with bootstrap, login, logout, session status, and CSRF enforcement.
- Capability constants and SQLite-backed user capability checks for protected host APIs.
- Durable audit recording for auth lifecycle events, capability denial, service-control intent, and explicit dev-mode bypasses.
- Forward migration for Phase 8 capability seed data and audit event fields.
- Unix-domain socket RPC skeleton between `kojid` and `koji-agent`.
- Agent socket configuration with default `/run/koji/agent.sock`.
- Controlled agent client errors for missing socket, refused connection, timeout, malformed response, and `not_implemented`.
- Bounded read-only command runner under `internal/platform/command`.
- Service status observation now uses the bounded runner instead of direct `os/exec`.
- Buildable Vite/React/TypeScript frontend under `web/`.
- Typed frontend API client for metrics, disk, services, processes, and service-control requests.
- Production static asset directory configuration with default `/usr/share/koji/dist`.
- Production browser security headers and API no-store JSON responses.
- Configurable session TTL and idle timeout with durable `last_seen_at` tracking.
- Session cleanup for expired and revoked sessions.
- Request ID middleware with bounded inbound `X-Request-ID` preservation, generated IDs, response headers, structured request logs, and audit correlation.
- Configurable service allowlist for service status and service-control intent.
- Configurable process visibility policy with conservative summary defaults, command-line redaction, response limits, and audit events.
- Public non-telemetry `/healthz` and `/readyz` operational status endpoints.
- Frontend information architecture grouped into Overview, Services, Processes, Activity, and Settings.
- Reusable frontend UI primitives for gauges, metric cards, status badges, errors, tooltips, empty states, and loading states.
- Normalized typed frontend API errors with plain-language operator messages.
- Redaction-aware process table and safe process-state summary chart.
- `audit.events.read` capability and governed `/api/activity` audit read model.
- Frontend Activity page backed by normalized recent audit events.
- Durable service-control jobs with queued lifecycle state and `jobs.read` capability.
- Protected jobs read APIs and frontend Jobs page.
- `jobs.approve` capability with durable job approval and rejection decisions.
- Protected job approval/rejection APIs and Jobs page decision controls.
- Daemon-owned job worker skeleton for approved service-control jobs.
- Agent-side service mutation guardrails with disabled-by-default mutation, independent allowlist validation, bounded command runner integration, and normalized response codes.
- Controlled agent service mutation enablement through the guarded executor and platform command runner.
- Runtime packaging layout with systemd units, example configs, staging install target, and deployment filesystem paths.
- Release workflow with pinned toolchains, CI-built artifacts, rootfs archive assembly, checksum generation, and artifact validation.
- Release smoke-test stage with downloaded artifact validation, checksum verification, rootfs layout checks, systemd unit checks, forbidden path scanning, and workflow outputs.
- First tagged release dry run through GitHub Releases, validated with downloaded `v0.1.3` assets.
- Internal control-plane metrics registry for jobs, worker, agent RPC, authentication, audit writes, and readiness checks.
- `observability.metrics.read` capability and governed `/api/observability/metrics` endpoint.
- Frontend Overview cards for control-plane health, job flow, worker activity, agent RPC, audit persistence, authentication, and readiness counters.
- Frontend toast notifications, accessible feedback primitives, last-updated indicators, stale-data notices, and clearer policy/status explanations.
- Frontend Vitest/jsdom test harness with Testing Library coverage for UI feedback primitives, API error normalization, Jobs, Activity, and observability cards.
- Structured `docs/wiki` documentation portal for architecture, security, operations, user guide, developer guide, and references.
- Documentation validation script for wiki structure, Home links, stale terminology, and architecture diagrams.
- OpenAPI 3.1 contract under `docs/api` with generated YAML, endpoint docs, error docs, and wiki API reference.
- OpenAPI generation and validation scripts with route coverage and capability metadata checks.
- Backup, restore, and restore verification scripts for SQLite and Koji configuration recovery.
- Disaster recovery and release rollback operations documentation.
- Automated packaging recovery test for backup, database deletion, restore, and verification.

### Changed

- Governance terminology now consistently uses Koji, `kojid`, `koji-agent`, `/etc/koji/koji.yaml`, and `/var/lib/koji` paths.
- `koji-agent` now starts the local socket server skeleton while keeping service mutation disabled.
- Direct command execution is centralized in `internal/platform/command`.
- HTTP route registration, protected handlers, static serving, JSON helpers, and audit/capability helpers are split into focused files.
- Frontend API handling now surfaces typed failures in UI state instead of swallowing request errors.
- Production SPA serving no longer depends on the process working directory.
- Session validation now rejects expired, idle-expired, and revoked sessions while refreshing valid session activity.
- HTTP request logs now include stable request IDs for correlation with API responses and audit events.
- Service APIs now deny by default unless a systemd unit is explicitly allowlisted.
- Process API responses now apply configured visibility policy before returning host process metadata.
- Readiness now reports DB, migration, and agent reachability state without exposing protected host data.
- Activity now replaces the placeholder with a dedicated capability-protected audit view.
- Service-control intent now creates a durable queued job instead of synchronously waiting on agent RPC.
- Service-control agent errors now return safe specific messages for unavailable, disabled, validation, allowlist, timeout, and command failure states.
- Activity now replaces the placeholder with a dedicated capability-protected audit view.
- Queued jobs now require an explicit approval or rejection transition before any future execution path can consume them.
- Approved jobs can be claimed by the worker as `running` and then marked `completed` or `failed` from normalized agent outcomes.
- Agent service-control RPC now returns `mutation_disabled` by default and maps agent execution outcomes to safe reason codes.
- Build and install targets now separate frontend assets, Go binaries, configuration examples, and systemd unit files into package-oriented paths.
- Release targets now produce Linux amd64 binaries, a rootfs archive, and `SHA256SUMS.txt`.
- Release publishing now depends on a distinct CI smoke-test job before GitHub release assets are uploaded.
- Release smoke validation now restores executable bits after GitHub artifact download before running binary help checks.
- Audit, job, worker, auth, readiness, and agent RPC paths now update in-process operational counters without external telemetry dependencies.
- Frontend polling is less aggressive for higher-cost views, with page-local failure feedback instead of full-page blocking.
- Responsive layout, focus states, touch targets, tooltips, and table controls are improved for mobile, tablet, and desktop widths.
- Release workflow now runs frontend tests before the frontend production build.
- Documentation is now treated as a first-class subsystem with cross-linked operator, developer, and reviewer guidance.
- Release workflow now validates OpenAPI before building frontend assets and release artifacts.
- Makefile now includes `backup`, `restore`, and `verify-restore` operational targets.

### Removed

- Nothing yet.

## [0.5.0] - 2026-06-08

### Added

- Initial governance documents.
- Architecture constraints for daemon, agent, capabilities, configuration, database, sensors, jobs, and audit.
- Coding style rules for Go and TypeScript/React.
- Invariant model for privileged execution and control-plane safety.
- Phase reporting template.
