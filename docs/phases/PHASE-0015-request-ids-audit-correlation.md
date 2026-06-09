# PHASE-0015: Request IDs and Audit Correlation

## Goal

Give every HTTP request a stable request ID and carry it through responses, request logs, and audit events.

## Non-Goals

This phase does not enable service mutation.

This phase does not change authentication, authorization, capabilities, or agent privilege behavior.

This phase does not change API response bodies.

## Invariants Preserved

- Auth remains deny-by-default.
- Capability checks remain deny-by-default.
- The web/API daemon still does not directly execute privileged host mutations.
- Audit fields are preserved and now include request correlation when events originate from HTTP.

## Negative Patterns Avoided

- No trusted unbounded caller-supplied request ID.
- No API response body contract changes.
- No service mutation.
- No agent privilege change.

## Design Summary

Phase 15 adds request ID middleware around the HTTP stack. A valid bounded inbound `X-Request-ID` is preserved; missing or invalid values are replaced with a generated ID. The selected ID is stored in request context, returned in the response header, included in structured request logs, and copied into audit events through existing audit helpers.

## Files Changed

- `internal/http/middleware.go`
- `internal/http/request_context.go`
- `internal/http/request_meta.go`
- `internal/http/request_id_test.go`
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

Added bounded request ID handling, response header propagation, structured request log correlation, and audit request ID tests.

## Summary

Each HTTP request now has one stable request ID that can be used to correlate the browser response, daemon logs, and audit events.

## Notes / Deviations

Inbound request IDs are accepted only when they are ASCII token-like values between 8 and 128 characters.
