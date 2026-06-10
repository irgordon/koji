# PHASE-0030: Operational Metrics and Control Plane Observability

## Objective

Make Koji observable without exposing sensitive operational data or adding an external telemetry backend.

## Scope

- Added `internal/observability` as the owner of in-process control-plane counters.
- Added fixed counters for job lifecycle, worker polling, agent RPC, authentication, audit writes, and readiness dependencies.
- Added `observability.metrics.read` as a dedicated capability.
- Added `GET /api/observability/metrics` as a governed read endpoint.
- Added Overview dashboard cards for control-plane health.

## Boundaries

- No Prometheus dependency.
- No OpenTelemetry dependency.
- No Grafana integration.
- No external collector or network telemetry sink.
- No raw audit internals, session data, user data, process data, service data, or SQL query input is exposed by the metrics endpoint.

## API

`GET /api/observability/metrics`

Requires:

- Valid session
- `observability.metrics.read`

Returns:

- `counters`: fixed in-process counter names and values
- `jobs_by_status`: bounded SQLite aggregate over durable job status values

## Operator Questions

The new surface is intended to answer:

- Is the worker polling and advancing jobs?
- Are jobs being created, approved, rejected, completed, or failed?
- Is the agent RPC path failing?
- Are authentication attempts succeeding or failing?
- Are audit writes succeeding?
- Are readiness checks detecting DB, migration, or agent dependency problems?

## Validation

- `npm run build`
- `gofmt`
- `GOCACHE=/tmp/koji-go-cache go test ./...`
