# PHASE-0016: Configurable Service Allowlist

## Goal

Ensure service status and service-control intent only apply to explicitly allowed systemd units.

## Non-Goals

This phase does not enable service mutation.

This phase does not change authentication, authorization, audit semantics, capabilities, or agent privileges.

This phase does not add arbitrary service discovery.

## Invariants Preserved

- Auth remains deny-by-default.
- Capability checks remain deny-by-default.
- The web/API daemon still does not directly execute privileged host mutations.
- Privileged service-control intent remains agent-bound.
- Non-allowlisted service-control intent is audited as denied.

## Negative Patterns Avoided

- No arbitrary systemd inspection surface through API input.
- No fallback service mutation path.
- No implicit all-services access for authenticated users.
- No capability weakening.

## Design Summary

Phase 16 adds `service_allowlist` to runtime config. Production config must explicitly name eligible systemd units, and each entry is validated with the existing service-name validator. Development mode may use a narrow default when no allowlist is provided.

The services API lists only allowlisted units. Service-control intent validates the request, checks capability, enforces the allowlist before the agent call, and audits non-allowlisted units as denied.

## Files Changed

- `internal/config/config.go`
- `cmd/kojid/main.go`
- `internal/http/mux.go`
- `internal/http/handlers_services.go`
- `internal/http/service_allowlist_test.go`
- `internal/http/service_control_test.go`
- `internal/http/phase8_test.go`
- `internal/config/config_test.go`
- `docs/ARCHITECTURE.md`
- `docs/INVARIANTS.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
npm run build
gofmt
GOCACHE=/tmp/koji-go-cache go test ./...
rg -n "systemctl|exec\.Command|CommandContext" internal/http internal/agent
```

## Changelog

Added production service allowlist validation, dev-mode narrow allowlist defaults, service list filtering, service-control allowlist enforcement, and non-allowlisted audit coverage.

## Summary

Koji no longer exposes service status or service-control intent for arbitrary systemd unit names. Service APIs now fail closed unless the unit is allowlisted.

## Notes / Deviations

The config parser accepts `service_allowlist` as a comma-separated value, with optional square brackets and quotes for simple YAML-like input.
