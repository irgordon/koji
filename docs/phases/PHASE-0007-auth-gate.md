# PHASE-0007: Auth Gate and Bootstrap Surface

## Goal

Add a deny-by-default session gate before host telemetry, process data, service data, service-control intent routes, and production SPA serving.

## Non-Goals

This phase does not implement role or capability authorization.

This phase does not enable service mutation.

## Invariants Preserved

- The browser is never authoritative.
- The web server is never privileged.
- The agent is the only privileged execution surface.
- CSRF protection is required for authenticated browser mutations.

## Negative Patterns Avoided

- No UI-only enforcement.
- No public host telemetry APIs.
- No permanent bootstrap surface.
- No direct privileged execution from HTTP handlers.

## Design Summary

Sessions are stored in SQLite. Bootstrap is available only before the first user exists and is claimed by durable bootstrap state. Login creates a session with a CSRF token. Logout revokes the session. Production middleware denies non-auth routes unless a valid session exists, and state-changing authenticated requests require a CSRF header.

## Files Changed

- `internal/auth`
- `internal/http`
- `internal/db/migrations.go`
- `cmd/kojid/main.go`
- `docs/CHANGELOG.md`

## Commands Run

```text
gofmt
GOCACHE=/tmp/koji-go-cache go test ./...
rg -n "systemctl|exec\.Command" internal/http
```

## Changelog

Added deny-by-default auth gate, bootstrap, login, logout, session status, and CSRF enforcement.

## Summary

Unauthenticated production requests no longer reach protected API or SPA handlers.

## Notes / Deviations

Development mode intentionally bypasses the auth gate and marks responses with `X-Koji-Auth-Bypass: dev`.
