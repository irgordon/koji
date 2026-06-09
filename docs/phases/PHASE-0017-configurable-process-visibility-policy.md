# PHASE-0017: Configurable Process Visibility Policy

## Goal

Prevent the process API from exposing arbitrary process metadata by default.

## Non-Goals

This phase does not enable service mutation.

This phase does not change authentication, authorization, capabilities, or agent privileges.

This phase does not add process mutation or process control.

## Invariants Preserved

- Auth remains deny-by-default.
- `host.processes.read` is still required before process listing.
- The web/API daemon still does not directly execute privileged host mutations.
- Command lines are omitted unless explicitly configured.
- Process listing access is audited.

## Negative Patterns Avoided

- No full command-line exposure by default.
- No unbounded process API response.
- No capability weakening.
- No agent boundary change.

## Design Summary

Phase 17 adds process visibility configuration:

- `process_visibility_mode`: `summary`, `owner`, or `all`
- `include_command_line`: default `false`
- `max_processes`: default `200`

The collector may read process metadata, but the HTTP layer applies a response policy before serialization. Summary mode returns only PID, name, and state. Owner mode adds UID. All mode adds detailed parent, CPU, RSS, and memory fields. Command lines remain omitted unless explicitly enabled.

## Files Changed

- `internal/config/config.go`
- `cmd/kojid/main.go`
- `internal/system/processes.go`
- `internal/http/mux.go`
- `internal/http/handlers_processes.go`
- `internal/http/process_visibility_test.go`
- `internal/config/config_test.go`
- `internal/audit/audit.go`
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
GOCACHE=/tmp/koji-go-cache go test ./...
```

## Changelog

Added configurable process visibility, command-line omission by default, max process response limits, process response redaction, and process-list audit events.

## Summary

The process API now returns bounded, policy-shaped process metadata and no longer exposes detailed process fields by default.

## Notes / Deviations

Tests inject a process lister for deterministic policy and audit coverage instead of depending on host `/proc` availability.
