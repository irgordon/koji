# PHASE-0020: Activity and Audit Read Model

## Goal

Replace the Activity placeholder with a governed, read-only audit activity view.

## Non-Goals

This phase does not expose raw sensitive audit internals.

This phase does not weaken authentication, capability checks, CSRF, or audit write behavior.

This phase does not enable service mutation.

This phase does not add arbitrary SQL or query input.

## Invariants Preserved

- Audit history remains append-only.
- Audit write behavior is unchanged.
- Activity read access requires `audit.events.read`.
- Activity responses include normalized fields only.
- Service mutation remains disabled.

## Negative Patterns Avoided

- No arbitrary audit query API.
- No raw actor, user, remote address, or internal message exposure.
- No unauthenticated audit access.
- No implicit access for authenticated users.

## Design Summary

Phase 20 adds `audit.events.read` as a dedicated capability and seeds it through a forward migration. `internal/audit` now exposes a bounded recent-events read model ordered newest-first. The HTTP layer adds `GET /api/activity`, protected by the new capability.

The frontend Activity page fetches this endpoint only when the Activity view is active. It renders a normalized table with timestamp, action, target, outcome, reason code, and request ID.

## Files Changed

- `internal/caps/caps.go`
- `internal/db/migrations.go`
- `internal/audit/audit.go`
- `internal/http/handlers_activity.go`
- `internal/http/routes.go`
- `internal/http/activity_test.go`
- `web/src/types.ts`
- `web/src/api.ts`
- `web/src/App.tsx`
- `web/src/App.css`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
npm run build
gofmt
go test ./....
```

## Changelog

Added `audit.events.read`, a bounded audit read model, protected `/api/activity`, backend tests for governed access and sensitive-field exclusion, and a frontend Activity table.

## Summary

Koji now has a governed operator activity view instead of a placeholder.

## Notes / Deviations

The prior UI/UX phase was corrected to Phase 19. This activity phase is recorded as Phase 20 to prevent duplicate Phase 19 tracking.
