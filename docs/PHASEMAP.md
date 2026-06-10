# PHASEMAP.md

## 1. Purpose

`PHASEMAP.md` defines how project phases are planned, implemented, validated, and recorded.

Each phase should be small enough to review.

Each phase should preserve the architecture and invariants unless the phase explicitly changes them.

## 2. Phase Discipline

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

## 3. Required Phase Report Template

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

## 4. Goal

The goal must be concrete.

Bad:

```text
Improve security.
```

Good:

```text
Add capability checks before service restart job creation.
```

## 5. Non-Goals

Non-goals protect scope.

Example:

```text
This phase does not implement service restart execution. It only adds the capability model and tests.
```

## 6. Invariants

Each phase must list relevant invariants from `INVARIANTS.md`.

Example:

```text
- The browser is never authoritative.
- The web server is never privileged.
- Capability denial fails closed.
```

If the phase changes an invariant, the phase must update `INVARIANTS.md` and explain why.

## 7. Negative Patterns

Each phase must explicitly state avoided negative patterns.

Example:

```text
- No arbitrary command endpoint.
- No direct privileged execution from HTTP handlers.
- No UI-only enforcement.
```

This forces reviewers to check for shortcuts.

## 8. Commands Run

Every phase must record commands actually run.

Do not invent validation.

If a command was not run, say so.

Example:

```text
go test ./...        # run, passed
npm run build        # not run, UI not touched
```

## 9. Changelog

Use `CHANGELOG.md` for release-level history.

Use phase notes for implementation-level history.

A phase changelog should say what changed, not why the whole project exists.

## 10. Summary

The summary states the result.

Example:

```text
The daemon now refuses startup when runtime config enables a capability that is not present in the build.
```

## 11. Notes / Deviations

Use this section for:

- Known limitations.
- Follow-up work.
- Risk accepted.
- Differences from plan.
- Test gaps.
- Platform gaps.

Deviations are allowed. Hidden deviations are not.

## 12. Suggested Initial Phases

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
