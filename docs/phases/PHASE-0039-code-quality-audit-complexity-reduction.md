# Phase 39: Code Quality Audit and Complexity Reduction

## Goal

Reduce complexity introduced by recent identity and frontend administration work without changing runtime behavior or public API contracts.

## Non-Goals

- No new endpoints.
- No new capabilities.
- No schema changes.
- No service mutation changes.
- No authentication, authorization, audit, jobs, or agent redesign.

## Invariants Preserved

- The browser is never authoritative.
- The web server is never privileged.
- Protected APIs require explicit capability checks.
- User and capability management requires `identity.users.manage`.
- Command execution remains centralized outside HTTP, agent, system, and jobs packages.

## Negative Patterns Avoided

- No rewrite of working subsystems.
- No public contract change.
- No speculative abstraction.
- No UI-only enforcement.
- No direct privileged execution.

## Design Summary

The audit identified two concrete complexity hotspots: the Administration view inside `App.tsx` and the overloaded identity store file. The fix split Administration into its own frontend view and split identity responsibilities into user lifecycle, capabilities, lockout policy, magic tokens, and row helpers.

The phase also adds `packaging/scripts/verify_code_quality.sh` so objective quality checks can run repeatedly.

## Files Changed

- `web/src/App.tsx`
- `web/src/AdminView.tsx`
- `web/src/App.css`
- `internal/identity/*.go`
- `internal/http/handlers_admin.go`
- `packaging/scripts/verify_code_quality.sh`
- `docs/wiki/Developer/Code-Quality-Audit.md`

## Commands Run

```text
gofmt -w cmd internal
npm run test
npm run build
GOCACHE=/tmp/koji-go-cache go test ./...
make verify-openapi
packaging/scripts/verify_docs.sh
packaging/scripts/verify_code_quality.sh
git diff --check
```

## Changelog

- Extracted Administration UI into a focused frontend view.
- Split identity store responsibilities across focused package files.
- Added repeatable code-quality validation.
- Added developer code-quality audit documentation.

## Summary

Koji keeps the Phase 38 identity behavior while reducing file-level concentration and adding a maintainability gate for future phases.

## Notes / Deviations

`App.tsx` remains the frontend shell and still owns several established views. This phase reduces the newest source of complexity without doing a broad UI module rewrite.
