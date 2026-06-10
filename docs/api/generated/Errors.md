# Koji API Errors

Generated from `docs/api/openapi.json`.

The current server error body contains a safe `error` field and an `X-Request-ID` response header. Client integrations may normalize these into the documented error codes.

| Code | Meaning |
| --- | --- |
| `agent_not_implemented` | agent not implemented |
| `agent_unavailable` | agent unavailable |
| `csrf_missing_or_invalid` | csrf missing or invalid |
| `forbidden` | forbidden |
| `identity_user_not_found` | identity user not found |
| `job_conflict` | job conflict |
| `job_not_found` | job not found |
| `magic_token_expired` | magic token expired |
| `magic_token_invalid` | magic token invalid |
| `mutation_disabled` | mutation disabled |
| `network_error` | network error |
| `self_lockout_prevented` | self lockout prevented |
| `service_not_allowlisted` | service not allowlisted |
| `session_expired` | session expired |
| `unauthenticated` | unauthenticated |
| `unexpected_response` | unexpected response |
| `validation_error` | validation error |
