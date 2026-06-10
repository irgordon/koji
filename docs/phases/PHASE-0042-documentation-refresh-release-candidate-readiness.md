# Phase 42: Documentation Refresh, Operational Alignment, and Release Candidate Readiness

## Goal

Reconcile documentation after identity management, refactoring, production-readiness review, and security-review phases so code, OpenAPI, wiki, operator workflows, and security guidance describe the same system.

## Non-Goals

- No new features.
- No new APIs.
- No new capabilities.
- No authentication or authorization changes.
- No agent, database, packaging, or release redesign.

## Invariants Preserved

- Documentation must describe server-side enforcement, not UI-only behavior.
- Managed users use magic-token login only.
- Super Admin password login remains the documented password path.
- Job state documentation must match the API contract.
- Metrics documentation must match `internal/observability`.

## Design Summary

Phase 42 updates release-readiness guidance, current operator workflow language, reference pages, and phase history. It also strengthens `packaging/scripts/verify_docs.sh` so required security, operations, user-guide, developer, and reference pages remain present and linked.

## Files Changed

- `docs/wiki/Home.md`
- `docs/wiki/Operations/Release-Candidate-Checklist.md`
- `docs/wiki/Operations/*`
- `docs/wiki/User-Guide/*`
- `docs/wiki/Developer/Phase-History.md`
- `docs/wiki/Reference/*`
- `packaging/scripts/verify_docs.sh`
- `docs/phases/PHASE-0042-documentation-refresh-release-candidate-readiness.md`

## Commands Run

```text
packaging/scripts/verify_docs.sh
make verify-openapi
npm run test
npm run build
GOCACHE=/tmp/koji-go-cache go test ./...
git diff --check
rg -n "<pre-Koji product and binary names>" docs README.md
rg -n "password login|password" docs/wiki
```

## Changelog

- Added release candidate checklist.
- Updated user-guide and reference language for current operator workflows.
- Extended phase history through Phase 42.
- Strengthened documentation validation requirements.

## Summary

Koji documentation now reflects the current authentication model, user administration flow, job lifecycle, observability counters, security review, and release-candidate readiness workflow.
