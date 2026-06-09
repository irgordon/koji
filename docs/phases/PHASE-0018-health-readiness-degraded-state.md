# PHASE-0018: Health, Readiness, and Degraded-State Reporting

## Goal

Add explicit operational status endpoints without exposing protected host telemetry.

## Non-Goals

This phase does not expose metrics, process data, service data, user data, sessions, or audit records through health endpoints.

This phase does not change authentication, authorization, capabilities, audit semantics, or agent privileges.

This phase does not enable service mutation.

## Invariants Preserved

- Protected host telemetry remains behind auth and capability checks.
- Health endpoints return only compact named check outcomes.
- Service mutation remains disabled.
- Agent unavailability does not trigger a fallback path.

## Negative Patterns Avoided

- No unauthenticated telemetry.
- No health endpoint audit spam.
- No sensitive failure details in readiness responses.
- No service-control or agent mutation probe.

## Design Summary

Phase 18 adds two unauthenticated operational endpoints for local supervisors and packaging:

- `GET /healthz`: minimal liveness response.
- `GET /readyz`: checks database reachability, migration currentness, and agent socket reachability.

Readiness returns `fail` when the database or migrations are unavailable. Agent unavailability reports `degraded` because the daemon may still serve read-only authenticated surfaces while privileged agent-backed operations are unavailable.

## Files Changed

- `internal/http/handlers_health.go`
- `internal/http/health_test.go`
- `internal/http/routes.go`
- `internal/http/middleware.go`
- `internal/http/mux.go`
- `internal/agent/client.go`
- `internal/db/migrations.go`
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

Added `/healthz`, `/readyz`, DB readiness checks, migration currentness checks, agent reachability degradation, and non-telemetry health endpoint tests.

## Summary

Koji now exposes safe operational health and readiness endpoints suitable for systemd, packaging, and local operator checks.

## Notes / Deviations

Readiness is unauthenticated by design. The endpoint is safe only because it reports named check statuses without protected telemetry or sensitive error details.
