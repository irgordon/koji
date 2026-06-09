# PHASE-0011: HTTP Package Decomposition

## Goal

Split `internal/http/mux.go` into focused files without changing behavior.

## Non-Goals

This phase does not add features.

This phase does not change auth, capability, audit, or agent behavior.

This phase does not enable service mutation.

## Invariants Preserved

- Route paths are preserved.
- Response shapes are preserved.
- Auth, capability, and audit checks remain in place.
- Service mutation remains disabled.
- HTTP and agent packages do not own direct command execution.

## Negative Patterns Avoided

- No route behavior changes.
- No broad refactor outside `internal/http`.
- No command execution in HTTP.
- No service-control mutation.

## Design Summary

`mux.go` now coordinates dependency construction only. Route registration moved to `routes.go`. Protected handler surfaces moved into focused files for metrics, services, processes, static assets, JSON helpers, audit helpers, and capability helpers.

Shared dependencies remain explicit through `routeDependencies` and `protectedHandlers`.

## Files Changed

- `internal/http/mux.go`
- `internal/http/routes.go`
- `internal/http/handlers_metrics.go`
- `internal/http/handlers_services.go`
- `internal/http/handlers_processes.go`
- `internal/http/handlers_static.go`
- `internal/http/handlers_authz.go`
- `internal/http/handlers_audit.go`
- `internal/http/json.go`
- `internal/http/request_meta.go`
- `docs/CHANGELOG.md`

## Commands Run

```text
gofmt
go test ./....
rg -n "exec\.Command|CommandContext" internal/http internal/agent internal/system
rg -n "systemctl" internal/http internal/agent
```

## Changelog

Split HTTP routing and handler code into focused files.

## Summary

`internal/http/mux.go` is now a small coordination file. Handler behavior remains covered by the existing tests.

## Notes / Deviations

No new tests were added because the split did not reveal missing behavior coverage.
