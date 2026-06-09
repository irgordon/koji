# PHASE-0021: Service Control Jobs and Approval Workflow

## Goal

Convert service-control intent into a tracked job lifecycle instead of a synchronous request path.

## Non-Goals

This phase does not execute `systemctl`.

This phase does not weaken authentication, capability checks, audit behavior, CSRF, or the agent boundary.

This phase does not implement approval endpoints or a background executor.

## Invariants Preserved

- Service mutation remains disabled.
- `host.services.control` is still required to create service-control jobs.
- `jobs.read` is required to read jobs.
- Service-control intent remains validated and allowlisted.
- Job creation and status changes are audited.

## Negative Patterns Avoided

- No synchronous privileged execution from HTTP.
- No arbitrary job query input.
- No in-memory-only operational intent.
- No fake approval or privilege model.

## Design Summary

Phase 21 adds a durable `jobs` table and `internal/jobs` store. Service-control requests now validate capability, service name, action, and service allowlist, then create a queued job and return the job ID with HTTP `202 Accepted`.

The HTTP request no longer waits on agent RPC. The initial queued lifecycle state is durable and audited with `job.created` and `job.status_changed`.

The phase also adds protected job read endpoints:

- `GET /api/jobs`
- `GET /api/jobs/{id}`

Both require `jobs.read` and emit `job.viewed` audit records.

## Files Changed

- `internal/db/migrations.go`
- `internal/caps/caps.go`
- `internal/audit/audit.go`
- `internal/jobs/jobs.go`
- `internal/http/handlers_services.go`
- `internal/http/handlers_jobs.go`
- `internal/http/handlers_audit.go`
- `internal/http/routes.go`
- `internal/http/jobs_test.go`
- `internal/http/service_control_test.go`
- `internal/http/phase8_test.go`
- `web/src/types.ts`
- `web/src/api.ts`
- `web/src/App.tsx`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
npm run build
gofmt
go test ./....
```

## Changelog

Added durable queued jobs, job read capability, protected job APIs, service-control job creation, job audit events, persistence tests, and the frontend Jobs page.

## Summary

Service-control intent now enters a durable governed job lifecycle before any privileged execution exists.

## Notes / Deviations

Approval and execution workers are intentionally deferred. New service-control jobs start in `queued` with `pending_approval`.
