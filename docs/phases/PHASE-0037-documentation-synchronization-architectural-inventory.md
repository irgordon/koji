# PHASE-0037: Documentation Synchronization and Architectural Inventory

## Goal

Synchronize project documentation with the actual Koji implementation after Phases 1 through 36.

## Non-Goals

- No runtime behavior changes.
- No API behavior changes.
- No database schema changes.
- No new backend or frontend features.

## Invariants Preserved

- The browser is never authoritative.
- The web server is never privileged.
- The agent remains the privileged boundary.
- OpenAPI remains route-coverage validated.
- Documentation validation remains automated.

## Negative Patterns Avoided

- No phantom routes.
- No undocumented implemented subsystems.
- No stale pre-Koji terminology.
- No references to reserved features as implemented behavior.
- No manual source archaeology requirement for major subsystems.

## Design Summary

This phase adds architectural, backend, frontend, and phase-history inventories. It refreshes governance and reference docs to match implemented authentication, sessions, CSRF, capabilities, audit, jobs, approvals, worker lifecycle, agent RPC, mutation controls, observability, packaging, release automation, OpenAPI, backup/recovery, and upgrade safety.

The docs validator now requires the inventory pages, required Mermaid diagrams, Home links, stale terminology checks, and local wiki link checks.

## Files Changed

- `README.md`
- `docs/assets/koji-browser-mockup.svg`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/SECURITY.md`
- `docs/CHANGELOG.md`
- `docs/README.md`
- `docs/wiki/Home.md`
- `docs/wiki/Developer/Architectural-Inventory.md`
- `docs/wiki/Developer/Backend-Inventory.md`
- `docs/wiki/Developer/Frontend-Inventory.md`
- `docs/wiki/Developer/Phase-History.md`
- `docs/wiki/Reference/Configuration-Reference.md`
- `docs/wiki/Reference/Capability-Reference.md`
- `docs/wiki/Reference/Job-State-Reference.md`
- `docs/wiki/Reference/Audit-Event-Reference.md`
- `docs/wiki/Reference/Metrics-Reference.md`
- `docs/wiki/Operations/Troubleshooting.md`
- `packaging/scripts/verify_docs.sh`

## Commands Run

```text
packaging/scripts/verify_docs.sh
make verify-openapi
npm run test
npm run build
GOCACHE=/tmp/koji-go-cache go test ./...
git diff --check
```

## Changelog

- Added architectural, backend, frontend, and phase-history inventories.
- Reconciled reference docs with current config, capability, job, audit, and metric constants.
- Refreshed top-level architecture, invariants, and security docs.
- Updated README and SVG browser mockup for the current governed control panel.
- Strengthened documentation validation for required inventory pages and local link drift.

## Summary

Future contributors now have synchronized governance pillars across code, tests, documentation, and OpenAPI.

## Notes / Deviations

`docs/api` did not require contract changes because `make verify-openapi` still reports route coverage and generated references as current.
