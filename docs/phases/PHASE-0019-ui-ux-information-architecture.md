# PHASE-0019: UI/UX Information Architecture

## Goal

Make the frontend easier to operate by grouping pages logically, normalizing errors into plain language, and adding safe visual telemetry.

## Non-Goals

This phase does not weaken backend auth, capability checks, or audit behavior.

This phase does not enable service mutation.

This phase does not expose backend-redacted process data.

This phase does not add Tailwind or a large UI framework.

## Invariants Preserved

- Frontend displays only data returned by existing authorized APIs.
- Redacted process fields remain redacted.
- Service-control intent still uses the existing CSRF-protected API path.
- Backend service mutation remains disabled.

## Negative Patterns Avoided

- No client-side reconstruction of redacted fields.
- No swallowed frontend errors.
- No broad UI framework dependency.
- No backend privilege expansion.

## Design Summary

The frontend now uses grouped navigation:

- Overview
- Services
- Processes
- Activity
- Settings

Overview presents CPU, memory, disk, uptime, health, and readiness. Services show allowlisted service status and plain-language control errors. Processes show a redaction-aware table and a state summary chart that uses only fields available in summary mode. Activity and Settings remain explicit placeholders/read-only policy surfaces.

`api.ts` now normalizes backend and network failures into typed error codes and plain-language messages. The UI adds reusable components for gauges, metric cards, status badges, error banners, tooltips, empty states, and loading states.

## Files Changed

- `web/src/App.tsx`
- `web/src/App.css`
- `web/src/api.ts`
- `web/src/types.ts`
- `internal/http/handlers_services.go`
- `docs/CHANGELOG.md`

## Commands Run

```text
npm run build
gofmt
go test ./....
```

## Changelog

Added grouped frontend IA, normalized errors, gauges, operational status panels, safe process charting, redaction tooltips, and explicit service-control unavailable messaging.

## Summary

The Koji UI is now easier to scan and safer to operate without changing backend authorization or privileged behavior.

## Notes / Deviations

This phase was corrected from Phase 18 to Phase 19 to avoid tracking drift with the operational health Phase 18.
