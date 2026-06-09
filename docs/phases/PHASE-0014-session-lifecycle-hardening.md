# PHASE-0014: Session Lifecycle Hardening

## Goal

Make session behavior explicit, bounded, and safer under production use.

## Non-Goals

This phase does not enable service mutation.

This phase does not change agent privilege behavior.

This phase does not weaken capabilities or authorization.

## Invariants Preserved

- Auth remains deny-by-default.
- Revoked sessions are rejected.
- Expired sessions are rejected.
- Idle-expired sessions are rejected.
- Production cookies are hardened.
- Development cookie behavior remains explicit for local HTTP.

## Negative Patterns Avoided

- No unbounded session lifetime.
- No implicit all-powerful authenticated user behavior.
- No service mutation.
- No agent privilege change.
- No deletion-based logout semantics.

## Design Summary

Phase 14 adds configurable absolute session TTL and idle timeout settings. Sessions now persist `created_at`, `last_seen_at`, `expires_at`, and `revoked_at`. Valid session validation updates `last_seen_at`; expired, idle-expired, and revoked sessions are rejected.

Logout continues to set `revoked_at`. A cleanup method removes expired and revoked sessions. Production session cookies use `Secure`, `HttpOnly`, `SameSite=Strict`, and `Path=/`; development mode explicitly omits `Secure` for local HTTP.

## Files Changed

- `internal/config/config.go`
- `cmd/kojid/main.go`
- `internal/db/migrations.go`
- `internal/auth/store.go`
- `internal/auth/store_test.go`
- `internal/http/auth_test.go`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/SECURITY.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
npm run build
gofmt
go test ./....
```

## Changelog

Added configurable session TTL, idle timeout tracking, durable `last_seen_at`, hardened cookie tests, and cleanup for expired/revoked sessions.

## Summary

Session validation is now bounded by both absolute lifetime and idle activity, and valid authenticated requests refresh session activity.

## Notes / Deviations

`SameSite=Strict` is retained because Koji is an administrative control plane with no cross-site browser workflow requirement.
