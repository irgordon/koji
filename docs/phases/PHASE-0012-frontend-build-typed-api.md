# PHASE-0012: Frontend Build and Typed API

## Goal

Restore a complete buildable frontend under `web/` and replace loose API handling with typed request and response helpers.

## Non-Goals

This phase does not change backend behavior.

This phase does not weaken auth, capability, audit, or agent boundaries.

This phase does not enable service mutation.

This phase does not introduce Tailwind or a large UI framework.

## Invariants Preserved

- Backend API contracts are preserved.
- Production frontend output still emits to top-level `dist/`.
- Service-control requests still go through existing backend authorization and agent paths.
- Privileged backend behavior is unchanged.

## Negative Patterns Avoided

- No `any` in `App.tsx`.
- No empty catch blocks.
- No raw API handling in `App.tsx`.
- No new UI framework dependency.
- No backend privilege or service mutation changes.

## Design Summary

Phase 12 adds the missing Vite, TypeScript, HTML, CSS, and package metadata required to build the frontend from the repository. `web/src/api.ts` centralizes typed fetch helpers and runtime response validation. `App.tsx` now consumes typed helpers, keeps hash-based navigation, preserves dashboard/services/processes behavior, and displays API failures through UI state.

The Vite build writes to `../dist`, matching the backend static serving contract.

## Files Changed

- `web/package.json`
- `web/package-lock.json`
- `web/tsconfig.json`
- `web/vite.config.ts`
- `web/index.html`
- `web/src/main.tsx`
- `web/src/App.tsx`
- `web/src/App.css`
- `web/src/api.ts`
- `web/src/types.ts`
- `Makefile`
- `docs/CHANGELOG.md`

## Commands Run

```text
npm install
npm run build
gofmt
go test ./....
```

## Changelog

Added a buildable frontend skeleton and typed API client. Refactored the app to remove loose API handling and surface API failures in UI state.

## Summary

`web/` now builds from repo-local metadata, and production assets emit to top-level `dist/` for the existing backend contract.

## Notes / Deviations

`npm install` was used because no lockfile existed at the start of the phase.
