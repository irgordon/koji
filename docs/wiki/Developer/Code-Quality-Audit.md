# Code Quality Audit

[Home](../Home.md) | Related: [Backend Inventory](Backend-Inventory.md), [Frontend Inventory](Frontend-Inventory.md), [Testing Strategy](Testing-Strategy.md)

## Purpose

This page records maintainability findings from Phase 39. It focuses on architecture conformance, single-level-of-abstraction boundaries, hidden side effects, validation drift, and file/function complexity.

## Audit Table

| Area | Finding | Severity | Fix | Status |
| ---- | ------- | -------- | --- | ------ |
| `web/src/App.tsx` | Administration UI added page-local state, API calls, token display, and capability controls inside the main app coordination file. | Medium | Extracted Administration into `web/src/AdminView.tsx` and kept `App.tsx` focused on shell routing, polling, and top-level data ownership. | Fixed |
| `internal/identity/store.go` | User lifecycle, capability assignment, lockout policy, token issue, row scanning, and token hashing lived in one file. | Medium | Split identity package into `store.go`, `capabilities.go`, `lockout.go`, `magic_tokens.go`, and `rows.go`. | Fixed |
| `internal/http/handlers_admin.go` | Admin handlers repeated user-target formatting while coordinating identity calls and audit. | Low | Added a shared `userTarget` helper and kept audit response mapping centralized. | Fixed |
| Frontend type safety | `: any` and `as any` must remain absent from TypeScript sources. | High | Added `packaging/scripts/verify_code_quality.sh` guard. | Guarded |
| Frontend swallowed errors | Empty catch blocks would hide operator-visible failures. | High | Added `packaging/scripts/verify_code_quality.sh` guard. | Guarded |
| Command ownership | HTTP, agent, system, and jobs packages must not own `systemctl` or direct command execution. | Critical | Added `packaging/scripts/verify_code_quality.sh` guard for forbidden packages and expected platform ownership. | Guarded |
| Documentation drift | Code quality expectations were not represented as a repeatable verifier. | Medium | Added code quality verification script and this wiki reference. | Fixed |

## Boundary Review

HTTP handlers coordinate request parsing, capability checks, store calls, audit, and JSON responses. Identity stores own persistence and identity policy. The frontend shell owns navigation and polling. The Administration view owns identity page state.

No API contracts, capabilities, database schema, service mutation behavior, or agent behavior changed in this phase.

## Validation Guard

Run:

```bash
packaging/scripts/verify_code_quality.sh
```

The script checks known complexity hotspots, frontend unsafe typing, swallowed frontend errors, command ownership boundaries, and required audit documentation.
