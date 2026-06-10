# Frontend Inventory

[Home](../Home.md) | Related: [Frontend Architecture](Frontend-Architecture.md), [API Reference](../Reference/API-Reference.md), [Architectural Inventory](Architectural-Inventory.md)

The frontend is a React and TypeScript operator workspace under `web/src`. It uses plain CSS and the typed API helpers in `api.ts`.

## Pages And Views

| View | Component | Purpose | Data Source | Failure Surface |
| --- | --- | --- | --- | --- |
| Overview | `Overview`, `ControlPlaneObservability` | Host summary, health/readiness, and control-plane counters. | `/api/v1/metrics`, `/api/v1/disk`, `/healthz`, `/readyz`, `/api/observability/metrics` | Permission denial, network failure, degraded readiness, stale data. |
| Services | `ServicesView` | Show allowlisted services and create service-control jobs. | `/api/v1/services`, `/api/services/{name}/{action}` | Service not allowlisted, missing capability, CSRF failure, agent disabled later in job lifecycle. |
| Processes | `ProcessesView`, `ProcessSummaryChart` | Show redaction-aware process rows and safe process-state summary. | `/api/v1/processes` | Hidden fields, missing capability, policy limits. |
| Jobs | `JobsView`, `JobDecisionControls` | Show durable jobs and approve/reject queued jobs. | `/api/jobs`, `/api/jobs/{id}/approve`, `/api/jobs/{id}/reject` | Missing `jobs.read` or `jobs.approve`, CSRF failure, invalid transition. |
| Activity | `ActivityView` | Show normalized audit activity. | `/api/activity` | Missing `audit.events.read`, empty audit stream. |
| Administration | `AdminView` | Create managed users, enable/disable accounts, grant/revoke capabilities, and issue one-time magic tokens. | `/api/admin/users`, `/api/admin/users/{id}/capabilities`, `/api/admin/users/{id}/magic-token` | Missing `identity.users.manage`, CSRF failure, self-lockout prevention, invalid capability. |
| Settings | `SettingsView`, `PolicyCard` | Explain read-only policy surfaces currently enforced by backend config. | Static UI text. | None; does not expose mutable settings. |

## Reusable Components

| Component | Purpose | Accessibility Behavior |
| --- | --- | --- |
| `ToastProvider` | Own toast state and expose `useToast`. | Uses `ToastRegionContent` with live-region behavior. |
| `ToastRegionContent` | Render success, error, warning, and info notices. | Error toasts use alert semantics; all toasts have dismiss buttons. |
| `StatusBadge` | Show non-color-only status labels. | Includes text prefixes such as `OK`, `WARN`, `FAIL`, `RUN`, and `DONE`. |
| `Tooltip` | Explain technical fields in plain language. | Focusable help affordance with `aria-describedby`. |
| `ErrorBanner` | Show normalized safe errors. | Uses `role="alert"`. |
| `MetricCard` | Frame a labeled metric and optional tooltip. | Labels and details remain visible without hover. |
| `Gauge` | Show bounded percentage telemetry. | Uses visible label and percentage text, not color alone. |
| `EmptyState` | Explain an empty list or inaccessible data surface. | Plain text title and detail. |
| `LoadingState` | Show loading state for async panels. | Uses `role="status"` and `aria-live="polite"`. |
| `InlineError` | Show panel-scoped failures. | Keeps one failed panel from blocking the page. |
| `LastUpdated` / `StaleDataNotice` | Show recency and stale-data warning. | Plain text time and warning. |

## Hooks And State Ownership

| Surface | Owner | Notes |
| --- | --- | --- |
| Hash navigation | `useHashNavigation` | Keeps navigation client-local and bookmarkable. |
| Polling | `usePolling` | Owns interval polling, abort handling, stale errors, and polling toasts. |
| Toasts | `ToastProvider`, `useToast` | Centralized feedback for login/session/API/job actions. |
| API normalization | `api.ts` | Converts backend and network errors to safe plain-language messages. |
| Job decision form state | `JobsView` | Reason text is local per visible job row. |
| Identity administration state | `AdminView` | User creation, capability selection, per-user capability lists, and one-time token display are local to the page. |

## Polling Behavior

- Overview telemetry and health/readiness use moderate polling.
- Jobs use moderate polling because job state can change by worker action.
- Activity uses slower polling.
- Administration uses moderate polling and refreshes after mutating identity actions.
- Processes are fetched through the same polling helper but remain redaction-aware.
- One failed panel stores its own error instead of blocking the full app.

## API Client Surfaces

`api.ts` exports typed helpers for metrics, disk, services, processes, health, readiness, activity, jobs, observability, job decisions, service-control job creation, identity administration, capability assignment, and magic-token login.

Runtime guards validate response shapes and return `unexpected_response` when a payload does not match the expected contract.
