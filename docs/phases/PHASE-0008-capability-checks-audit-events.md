# PHASE-0008: Capability Checks and Audit Events

## Goal

Add authorization and durable audit recording for protected API access and service-control intent.

## Non-Goals

This phase does not enable actual service mutation.

This phase does not weaken the daemon/agent boundary.

This phase does not make authenticated users implicitly all-powerful.

## Invariants Preserved

- The browser is never authoritative.
- The web server is never privileged.
- Privileged work crosses the agent boundary.
- Capability denial fails closed.
- Privileged intent is audited even when denied.
- Audit records do not expose sensitive internal errors.

## Negative Patterns Avoided

- No direct privileged execution from HTTP handlers.
- No shell command construction in HTTP.
- No UI-only enforcement.
- No local fallback bypass when the agent is unavailable.
- No unaudited service-control intent.

## Design Summary

Phase 8 adds `internal/caps` for SQLite-backed user capability checks and `internal/audit` for durable audit event writes. Auth middleware now attaches the authenticated principal to request context. Protected handlers require one explicit capability before reading host data or accepting service-control intent. Dev mode remains explicit and writes audit events with a bypass marker.

Service control remains agent-bound. The default agent client still reports unavailable instead of mutating host services.

## Files Changed

- `internal/caps`
- `internal/audit`
- `internal/db/migrations.go`
- `internal/http`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
gofmt -w internal/caps internal/audit internal/http internal/db
GOCACHE=/tmp/koji-go-cache go test ./...
rg -n "systemctl|exec\.Command" internal/http
```

## Changelog

Added SQLite-backed capability checks and durable audit events for auth, capability denial, service-control intent, and dev-mode bypass markers.

## Summary

Protected host APIs now deny by default unless the authenticated user has the required capability. Service-control intent is audited for denial, validation rejection, agent unavailability, and accepted agent-bound intent.

## Notes / Deviations

Read-only service status still lives in `internal/system` temporarily. Service mutation remains behind `internal/agent` and is not enabled by this phase.
