# Koji API Errors

Generated from `docs/api/openapi.json`.

The current server error body contains a safe `error` field and an `X-Request-ID` response header. Client integrations may normalize these into the documented error codes.

| Code | Meaning |
| --- | --- |
| `unauthenticated` | unauthenticated |
| `forbidden` | forbidden |
| `csrf_missing_or_invalid` | csrf missing or invalid |
| `session_expired` | session expired |
| `agent_unavailable` | agent unavailable |
| `agent_not_implemented` | agent not implemented |
| `mutation_disabled` | mutation disabled |
| `service_not_allowlisted` | service not allowlisted |
| `validation_error` | validation error |
| `job_not_found` | job not found |
| `job_conflict` | job conflict |
| `network_error` | network error |
| `unexpected_response` | unexpected response |
