# PHASE-0010: Bounded System Command Execution

## Goal

Create one safe command-execution helper for read-only system inspection and use it for service status calls.

## Non-Goals

This phase does not enable service mutation.

This phase does not add `systemctl start`, `stop`, or `restart`.

This phase does not weaken auth, capability, audit, or agent boundaries.

## Invariants Preserved

- The web server is never privileged.
- The agent remains the future mutation boundary.
- HTTP and agent packages do not call `os/exec` directly.
- Command output is size-bounded.
- Commands have context timeouts.
- Errors are normalized before crossing package boundaries.

## Negative Patterns Avoided

- No shell command construction.
- No direct command execution in `internal/http`.
- No direct command execution in `internal/agent`.
- No service mutation.
- No unbounded stdout or stderr collection.

## Design Summary

Phase 10 adds `internal/platform/command`, a bounded read-only command runner. The runner uses `exec.CommandContext` only inside the platform package, applies context timeouts, limits stdout and stderr bytes, rejects executables outside an allowlist, and returns normalized error values.

Read-only service status now calls the platform runner for `systemctl show` and preserves existing service-name validation.

## Files Changed

- `internal/platform/command`
- `internal/system/services.go`
- `internal/system/services_test.go`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/SECURITY.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
gofmt
go test ./....
rg -n "exec\.Command|CommandContext" internal/http internal/agent internal/system
```

## Changelog

Added a bounded read-only command runner and refactored service status observation to use it.

## Summary

Only `internal/platform/command` owns direct command execution. Service status observation remains read-only, bounded, and validated.

## Notes / Deviations

The broader scan still finds `exec.CommandContext` in `internal/platform/command`, which is the expected ownership point for direct command execution.
