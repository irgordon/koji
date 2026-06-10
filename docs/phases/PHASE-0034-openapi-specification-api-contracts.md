# PHASE-0034: OpenAPI Specification, API Contracts, and Generated Reference Documentation

## Objective

Create a machine-readable API contract for Koji and make it the source for generated API documentation.

## Scope

- Added `docs/api/openapi.json`.
- Added generated `docs/api/openapi.yaml`.
- Added generated API reference pages under `docs/api/generated`.
- Updated `docs/wiki/Reference/API-Reference.md` from the OpenAPI contract.
- Added `packaging/scripts/generate_openapi_docs.mjs`.
- Added `packaging/scripts/verify_openapi.sh`.
- Added `make openapi` and `make verify-openapi`.
- Added OpenAPI validation to the release workflow before frontend build.

## Contract Coverage

The OpenAPI contract covers the actual registered API routes:

- Health and readiness
- Bootstrap, login, logout, and session status
- Host metrics and disk metrics
- Allowlisted service status and service-control job creation
- Process listing
- Activity
- Observability metrics
- Jobs list, detail, approval, and rejection

## Protections

Each operation documents:

- Authentication requirement
- CSRF requirement where applicable
- Capability requirement via `x-koji-capability`
- Request body
- Response body
- Safe error codes

## Validation

- `make openapi`
- `make verify-openapi`
- `npm run test`
- `npm run build`
- `GOCACHE=/tmp/koji-go-cache go test ./...`
- `packaging/scripts/verify_docs.sh`
- `git diff --check`

## Known Limitations

- Koji does not currently register settings or config API endpoints, so none are documented.
- No SDK, CLI, Swagger UI, or hosted API browser is introduced in this phase.
- The YAML file is generated from the JSON contract and should not be edited by hand.
