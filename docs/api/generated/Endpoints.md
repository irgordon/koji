# Koji API Endpoints

Generated from `docs/api/openapi.json`.

## GET /healthz

Returns minimal process liveness without protected host telemetry.

- Authentication: Public
- Capability: `public`
- CSRF required: no
- Request: None
- Response: `HealthOK`
- Errors: `unexpected_response`

## GET /readyz

Reports DB, migration, and agent dependency state without protected host telemetry.

- Authentication: Public
- Capability: `public`
- CSRF required: no
- Request: None
- Response: `ReadinessOK`
- Errors: `unexpected_response`

## POST /api/bootstrap

Creates the first local user only while no users exist.

- Authentication: Public
- Capability: `public-until-first-user`
- CSRF required: no
- Request: `Credentials`
- Response: `AuthSession`
- Errors: `validation_error`, `job_conflict`, `unexpected_response`

## POST /api/login

Authenticates a user and creates a session plus CSRF token.

- Authentication: Public
- Capability: `public`
- CSRF required: no
- Request: `Credentials`
- Response: `AuthSession`
- Errors: `unauthenticated`, `validation_error`, `unexpected_response`

## POST /api/logout

Revokes the current session.

- Authentication: Required
- Capability: `authenticated`
- CSRF required: yes
- Request: None
- Response: `Status`
- Errors: `unauthenticated`, `csrf_missing_or_invalid`, `session_expired`, `unexpected_response`

## GET /api/session

Reports whether the browser has a valid session and whether bootstrap is required.

- Authentication: Public
- Capability: `public`
- CSRF required: no
- Request: None
- Response: `SessionStatus`
- Errors: `unexpected_response`

## GET /api/v1/metrics

Returns authorized CPU, memory, and uptime telemetry.

- Authentication: Required
- Capability: `host.metrics.read`
- CSRF required: no
- Request: None
- Response: `SystemMetrics`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## GET /api/v1/disk

Returns authorized disk usage telemetry.

- Authentication: Required
- Capability: `host.disk.read`
- CSRF required: no
- Request: None
- Response: `DiskMetrics`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## GET /api/v1/services

Returns status for services in the daemon allowlist.

- Authentication: Required
- Capability: `host.services.read`
- CSRF required: no
- Request: None
- Response: `ServiceList`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## POST /api/services/{name}/{action}

Creates a durable service-control job. The request does not execute systemctl directly.

- Authentication: Required
- Capability: `host.services.control`
- CSRF required: yes
- Request: None
- Response: `ServiceControlJob`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `service_not_allowlisted`, `validation_error`, `unexpected_response`

## GET /api/v1/processes

Returns a process list after applying backend visibility policy.

- Authentication: Required
- Capability: `host.processes.read`
- CSRF required: no
- Request: None
- Response: `ProcessList`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## GET /api/activity

Returns normalized recent audit activity.

- Authentication: Required
- Capability: `audit.events.read`
- CSRF required: no
- Request: None
- Response: `Activity`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## GET /api/observability/metrics

Returns governed Koji control-plane counters and job status aggregates.

- Authentication: Required
- Capability: `observability.metrics.read`
- CSRF required: no
- Request: None
- Response: `ObservabilityMetrics`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## GET /api/jobs

Returns recent durable service-control jobs.

- Authentication: Required
- Capability: `jobs.read`
- CSRF required: no
- Request: None
- Response: `JobList`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## GET /api/jobs/{id}

Returns a single durable job by ID.

- Authentication: Required
- Capability: `jobs.read`
- CSRF required: no
- Request: None
- Response: `Job`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `job_not_found`, `unexpected_response`

## POST /api/jobs/{id}/approve

Moves a queued job to approved.

- Authentication: Required
- Capability: `jobs.approve`
- CSRF required: yes
- Request: `JobDecision`
- Response: `Job`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `job_not_found`, `job_conflict`, `validation_error`, `unexpected_response`

## POST /api/jobs/{id}/reject

Moves a queued job to rejected.

- Authentication: Required
- Capability: `jobs.approve`
- CSRF required: yes
- Request: `JobDecision`
- Response: `Job`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `job_not_found`, `job_conflict`, `validation_error`, `unexpected_response`

## POST /api/login/magic-token

Consumes a one-time magic token and creates a session.

- Authentication: Public
- Capability: `public`
- CSRF required: no
- Request: `MagicTokenLoginRequest`
- Response: `AuthSessionResponse`
- Errors: `magic_token_invalid`, `magic_token_expired`, `validation_error`, `unexpected_response`

## GET /api/admin/users

Lists Super Admin and managed users.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: no
- Request: None
- Response: `AdminUserListResponse`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `unexpected_response`

## POST /api/admin/users

Creates a passwordless managed user.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: yes
- Request: `AdminUserCreateRequest`
- Response: `AdminUser`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `validation_error`, `unexpected_response`

## POST /api/admin/users/{id}/disable

Disables a user and revokes active sessions. Final Super Admin and final identity manager are protected.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: yes
- Request: None
- Response: `AdminUser`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `identity_user_not_found`, `self_lockout_prevented`, `unexpected_response`

## POST /api/admin/users/{id}/enable

Enables a disabled managed user.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: yes
- Request: None
- Response: `AdminUser`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `identity_user_not_found`, `unexpected_response`

## GET /api/admin/users/{id}/capabilities

Lists assigned and available capabilities for a user.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: no
- Request: None
- Response: `CapabilityListResponse`
- Errors: `unauthenticated`, `forbidden`, `session_expired`, `identity_user_not_found`, `unexpected_response`

## POST /api/admin/users/{id}/capabilities

Grants a capability to a user.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: yes
- Request: `CapabilityGrantRequest`
- Response: `CapabilitiesResponse`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `identity_user_not_found`, `validation_error`, `unexpected_response`

## DELETE /api/admin/users/{id}/capabilities/{capability}

Revokes a capability from a user. Revoking the final identity manager is blocked.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: yes
- Request: None
- Response: `CapabilitiesResponse`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `identity_user_not_found`, `self_lockout_prevented`, `validation_error`, `unexpected_response`

## POST /api/admin/users/{id}/magic-token

Issues a one-time passwordless login token. The raw token is returned only once and is never stored.

- Authentication: Required
- Capability: `identity.users.manage`
- CSRF required: yes
- Request: None
- Response: `MagicTokenResponse`
- Errors: `unauthenticated`, `forbidden`, `csrf_missing_or_invalid`, `session_expired`, `identity_user_not_found`, `unexpected_response`

