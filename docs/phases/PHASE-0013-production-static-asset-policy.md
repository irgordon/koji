# PHASE-0013: Production Static Asset Policy

## Goal

Make production SPA serving explicit, safe, and independent of the process working directory.

## Non-Goals

This phase does not change API response contracts.

This phase does not weaken auth, capability, audit, or agent behavior.

This phase does not enable service mutation.

## Invariants Preserved

- Production SPA remains authenticated.
- Dev proxy behavior remains separate from production static serving.
- API response shapes are unchanged.
- Production static serving does not depend on the process working directory.
- API JSON responses use no-store caching policy.

## Negative Patterns Avoided

- No relative production static asset directory.
- No path traversal fallback to the SPA shell.
- No exposure of SPA assets to unauthenticated production requests.
- No service mutation.
- No backend privilege changes.

## Design Summary

Phase 13 adds `StaticAssetDir` and `DevProxyURL` runtime configuration. Production static assets default to `/usr/share/koji/dist`, while the frontend build still emits to top-level `dist/` for packaging.

Production responses now include browser security headers. JSON response helpers set `Cache-Control: no-store`. Static serving rejects unsafe paths before SPA fallback.

## Files Changed

- `internal/config/config.go`
- `cmd/kojid/main.go`
- `internal/http/handlers_static.go`
- `internal/http/middleware.go`
- `internal/http/json.go`
- `internal/http/static_security_test.go`
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

Added explicit production static asset directory configuration, production security headers, no-store JSON responses, and path traversal safeguards for SPA serving.

## Summary

Production static serving now reads from configured absolute paths and no longer relies on starting `kojid` from the repository root.

## Notes / Deviations

Tests exercise static serving through the real auth and security middleware without constructing Linux system probes, which keeps them portable on non-Linux development hosts.
