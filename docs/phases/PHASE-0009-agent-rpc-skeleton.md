# PHASE-0009: Agent RPC Skeleton

## Goal

Create the real local process boundary between `kojid` and `koji-agent` while keeping service mutation disabled.

## Non-Goals

This phase does not enable `systemctl` mutation.

This phase does not add TCP transport.

This phase does not let `kojid` mutate host state directly.

## Invariants Preserved

- The web server is never privileged.
- Privileged work crosses the agent boundary.
- The agent socket is local-only.
- The agent API is allowlisted.
- Agent failure does not trigger a local fallback bypass.
- Service-control mutation remains disabled.

## Negative Patterns Avoided

- No `exec.Command` in `internal/http` or `internal/agent`.
- No direct `systemctl` path from `kojid`.
- No TCP listener.
- No removal of non-socket files at the agent socket path.
- No service-control success response from the skeleton agent.

## Design Summary

Phase 9 adds a JSON RPC skeleton over a Unix-domain socket. `kojid` builds an `internal/agent.Client` from configured `AgentSocketPath`. `koji-agent` starts an `internal/agent.Server` on that socket and responds to valid service-control requests with `not_implemented`.

Socket validation requires an absolute path, rejects unsafe world-writable parent directories unless sticky/root-owned checks pass, and only removes stale paths when they are actual socket files.

The HTTP service-control handler continues to audit denied, rejected, unavailable, and not-implemented outcomes.

## Files Changed

- `cmd/koji-agent/main.go`
- `cmd/kojid/main.go`
- `internal/agent`
- `internal/config`
- `internal/http`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
gofmt
GOCACHE=/tmp/koji-go-cache go test ./...
rg -n "systemctl|exec\.Command" internal/http internal/agent
```

## Changelog

Added the Unix socket agent RPC skeleton, socket path configuration, client error classification, and server-side `not_implemented` service-control response.

## Summary

`kojid` now talks to `koji-agent` through the real local transport shape. The agent owns the future mutation boundary, but Phase 9 still returns `not_implemented` for service-control RPC.

## Notes / Deviations

The managed workspace sandbox does not permit Unix socket bind, so socket integration tests skip only when the OS returns a bind permission error. The full test command still passes in this environment.
