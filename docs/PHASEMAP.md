# PHASEMAP.md

## 1. Purpose

`PHASEMAP.md` defines how project phases are planned, implemented, validated, and recorded.

Each phase should be small enough to review.

Each phase should preserve the architecture and invariants unless the phase explicitly changes them.

## 2. Product Roadmap After v0.2.0

`v0.2.0` is the Production Baseline.

Koji now has the core production foundation:

- Authentication.
- Magic tokens.
- Capabilities.
- Audit.
- Jobs.
- Approvals.
- Agent boundary.
- Service control.
- Observability.
- Backup and restore.
- Upgrade safety.
- OpenAPI.
- Documentation portal.
- Release automation.
- Artifact smoke testing.
- Operator smoke testing.
- Accessibility.
- Frontend test suite.
- Code quality audit.

After `v0.2.0`, Koji should be treated as a product rather than a continuous architecture-construction project.

Major architecture work is frozen for the `v0.3.0` cycle unless a release-blocking production issue requires it. New work should improve operator visibility, daily workflows, search, retention, usability, and release readiness.

### v0.2.x: Maintenance

Scope:

- Bug fixes.
- Documentation corrections.
- Packaging corrections.
- Release metadata corrections.
- Security fixes.
- Test fixes.
- Operational fixes discovered from production use.

Non-goals:

- No new subsystems.
- No broad refactors.
- No schema redesign.
- No authentication model redesign.
- No agent architecture redesign.

### v0.3.0: Operational Excellence

Goal:

Improve daily operator experience on top of the `v0.2.0` production foundation.

The `v0.3.0` roadmap is:

| Phase | Name | Purpose | Primary Deliverables |
| --- | --- | --- | --- |
| 45 | Session Management and Active Session Visibility | Show operators who is logged in and allow governed session revocation. | Session administration page, session list API, session revoke API, audit events. |
| 46 | Audit Search and Filtering | Make the Activity view usable for incident review and operational investigation. | Search, filtering, time ranges, actor filters, action filters. |
| 47 | Job Search, Filtering, and Retention | Make job history manageable as operational volume grows. | Status filters, date filters, actor filters, retention policy and retention management. |
| 48 | Capability Templates | Reduce assignment mistakes while preserving explicit capability enforcement. | Read-only Operator, Service Operator, Identity Administrator, and Auditor templates. |
| 49 | UI/UX Performance and Mobile Polish | Improve production usability under realistic data volume and small screens. | Polling review, large-table behavior, mobile navigation, touch interactions, loading states, empty states. |
| 50 | v0.3.0 Release Readiness | Repeat the release candidate discipline used for `v0.2.0`. | RC decision report, release workflow verification, artifact smoke, operator smoke, production readiness decision. |

### v0.3.0 Scope Guardrails

Allowed:

- Operator workflow improvements.
- Read-model and filtering improvements.
- UI clarity and performance improvements.
- Governance-preserving administration features.
- Documentation updates that describe the implemented product.

Not allowed without a separate release decision:

- SSO.
- MFA.
- OIDC.
- LDAP.
- Containers as a runtime deployment model.
- Kubernetes.
- HA or clustering.
- Major database redesign.
- Agent trust-boundary redesign.
- Large architecture refactors.

The goal of `v0.3.0` is operational excellence, not infrastructure expansion.

## 3. Phase Discipline

Every phase follows one rule:

```text
one task, one surface
```

A phase should not combine unrelated work.

Examples of valid phases:

- Add strict config loader.
- Add embedded migration runner.
- Add capability model.
- Add login route and session table.
- Add procfs memory collector.
- Add agent ping RPC.

Examples of invalid phases:

- Add auth, sensors, migrations, and UI dashboard.
- Refactor everything before adding service restart.
- Add a temporary shell endpoint for testing.

## 4. Required Phase Report Template

Create a phase note under `docs/phases/` using this naming pattern:

```text
docs/phases/PHASE-0001-short-name.md
```

Use this template:

```markdown
# PHASE-0000: Short Name

## Goal

State the concrete outcome of this phase.

## Non-Goals

State what this phase intentionally does not do.

## Invariants Preserved

List the invariants this phase touches or protects.

## Negative Patterns Avoided

List the forbidden shortcuts this phase avoids.

## Design Summary

Explain the design in plain language.

## Files Changed

List the major files or directories changed.

## Commands Run

Record validation commands exactly.

Example:

```text
go fmt ./...
go test ./...
npm test
npm run build
```

## Changelog

List user-visible, operator-visible, or architecture-visible changes.

## Summary

State what now works.

## Notes / Deviations

Record deviations from the original plan, follow-up work, risks, or known limitations.
```

## 5. Goal

The goal must be concrete.

Bad:

```text
Improve security.
```

Good:

```text
Add capability checks before service restart job creation.
```

## 6. Non-Goals

Non-goals protect scope.

Example:

```text
This phase does not implement service restart execution. It only adds the capability model and tests.
```

## 7. Invariants

Each phase must list relevant invariants from `INVARIANTS.md`.

Example:

```text
- The browser is never authoritative.
- The web server is never privileged.
- Capability denial fails closed.
```

If the phase changes an invariant, the phase must update `INVARIANTS.md` and explain why.

## 8. Negative Patterns

Each phase must explicitly state avoided negative patterns.

Example:

```text
- No arbitrary command endpoint.
- No direct privileged execution from HTTP handlers.
- No UI-only enforcement.
```

This forces reviewers to check for shortcuts.

## 9. Commands Run

Every phase must record commands actually run.

Do not invent validation.

If a command was not run, say so.

Example:

```text
go test ./...        # run, passed
npm run build        # not run, UI not touched
```

## 10. Changelog

Use `CHANGELOG.md` for release-level history.

Use phase notes for implementation-level history.

A phase changelog should say what changed, not why the whole project exists.

## 11. Summary

The summary states the result.

Example:

```text
The daemon now refuses startup when runtime config enables a capability that is not present in the build.
```

## 12. Notes / Deviations

Use this section for:

- Known limitations.
- Follow-up work.
- Risk accepted.
- Differences from plan.
- Test gaps.
- Platform gaps.

Deviations are allowed. Hidden deviations are not.

## 13. Historical Initial Phase Suggestions

This section is retained for historical context. It described early project bootstrapping before the `v0.2.0` Production Baseline.

### PHASE-0001: Repository Skeleton

Goal:

Create the initial Go and web repository layout with governance documents.

Non-goals:

No runtime behavior.

### PHASE-0002: Config Loader

Goal:

Load `/etc/koji/koji.yaml`, reject unknown fields, validate safe defaults, and implement `-configtest`.

Non-goals:

No web UI and no database.

### PHASE-0003: Capability Model

Goal:

Implement compile-time capability, runtime policy, and authorization intersection.

Non-goals:

No privileged execution.

### PHASE-0004: Database and Migrations

Goal:

Add SQLite open path, PRAGMAs, embedded migrations, checksum validation, and `-migrate`.

Non-goals:

No auth tables beyond initial schema unless included by migration plan.

### PHASE-0005: Auth Bootstrap

Goal:

Add users, sessions, CSRF, login route, and first-user bootstrap flow.

Non-goals:

No MFA unless scoped into this phase.

### PHASE-0006: Agent Ping Boundary

Goal:

Add local Unix socket agent, ping RPC, socket permissions, and daemon client.

Non-goals:

No privileged operations.

### PHASE-0007: Audit Events

Goal:

Add append-only audit event storage and API/event write path.

Non-goals:

No support bundle yet.

### PHASE-0008: Job Event Model

Goal:

Add jobs, job events, bounded output, cancellation, and status API.

Non-goals:

No service control yet.

### PHASE-0009: Procfs Sensors

Goal:

Add CPU, memory, load, disk, and network read-only collectors.

Non-goals:

No hwmon, SMART, or IPMI.

### PHASE-0010: Authenticated SPA Shell

Goal:

Serve login pre-auth and SPA post-auth with correct cache headers and CSP.

Non-goals:

No full dashboard.

## 13. Release Discipline

A release should include:

- Version.
- Commit hash.
- Migration version.
- Capability inventory.
- Operator-visible changes.
- Known deviations.

Do not release with undocumented invariant violations.
