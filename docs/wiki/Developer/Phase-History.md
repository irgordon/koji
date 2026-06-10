# Phase History

[Home](../Home.md) | Related: [Architectural Inventory](Architectural-Inventory.md), [Repository Layout](Repository-Layout.md), [Testing Strategy](Testing-Strategy.md)

This is a contributor-oriented evolution map. It summarizes architectural milestones through Phase 42 without requiring every phase report to be read first.

| Phase | Purpose | Major Deliverables | Breaking Changes | Documentation Impact |
| --- | --- | --- | --- | --- |
| 1 | Establish project direction. | Initial Koji structure and governance intent. | None recorded. | Seeded architecture direction. |
| 2 | Align naming and layout. | Koji product naming, daemon/agent names, target directories. | Replaced pre-Koji terminology. | Governance docs renamed paths and binaries. |
| 3 | Preserve daemon/agent boundary. | Early service-control boundary analysis. | None. | Invariants clarified. |
| 4 | Prepare production shape. | Runtime directories and docs alignment. | None. | Architecture docs tightened. |
| 5 | Move service mutation behind agent intent. | Agent client boundary for service control; HTTP stopped owning mutation. | Direct HTTP mutation path removed. | Daemon/agent boundary documented. |
| 6 | Add durable storage foundation. | SQLite open, PRAGMAs, migrations, core tables, config validation. | Startup fails on invalid config or DB init. | Database and config docs introduced. |
| 7 | Deny public access. | Session auth, bootstrap, login/logout, CSRF, production SPA auth gate. | Protected APIs require valid session. | Auth and session docs updated. |
| 8 | Add authorization and audit. | Capability checks, audit events, initial capability seeds. | Authenticated users are not implicitly privileged. | Capability and audit references added. |
| 9 | Add real agent transport shape. | Unix socket RPC client/server skeleton. | TCP transport excluded. | Agent architecture documented. |
| 10 | Bound command execution. | Central platform command runner, timeouts, output limits. | Direct command ownership centralized. | Command-execution invariants documented. |
| 11 | Decompose HTTP package. | Focused route and handler files. | None. | Backend architecture easier to navigate. |
| 12 | Restore frontend build. | Vite/React/TypeScript build and typed API helpers. | None. | Frontend build docs updated. |
| 13 | Harden static serving. | Absolute static asset config and security headers. | Production runtime no longer depends on repo root. | Packaging and static asset docs updated. |
| 14 | Harden sessions. | TTL, idle timeout, revoked sessions, secure cookies. | Expired/idle sessions rejected. | Session lifecycle documented. |
| 15 | Add request correlation. | Request IDs in responses, logs, and audit. | None. | Incident review docs updated. |
| 16 | Add service allowlist. | Explicit service allowlist for status and control intent. | Arbitrary service input denied. | Service allowlist docs updated. |
| 17 | Add process visibility policy. | Summary/owner/all modes, command-line redaction, process limits. | Process metadata redacted by default. | Process and settings docs updated. |
| 18 | Add health/readiness. | `/healthz`, `/readyz`, DB/migration/agent checks. | None. | Operations health docs added. |
| 19 | Improve UI information architecture. | Overview, Services, Processes, Activity, Settings, visual telemetry. | None. | User guide updated. |
| 20 | Add audit read model. | `audit.events.read`, `/api/activity`, Activity page. | Audit reads require dedicated capability. | Activity and audit docs updated. |
| 21 | Add durable jobs. | Jobs table, job APIs, Jobs page, service-control creates jobs. | Service-control no longer waits synchronously on agent. | Job lifecycle docs added. |
| 22 | Add approval workflow. | `jobs.approve`, approve/reject APIs, decisions and audit. | Queued jobs require explicit decision. | Approval docs updated. |
| 23 | Add worker skeleton. | Approved job claim, running/final states, worker audit. | Execution lifecycle leaves request path. | Worker lifecycle documented. |
| 24 | Add agent mutation guardrails. | Disabled-by-default mutation, agent allowlist, bounded runner path. | Mutation remains disabled unless explicitly configured. | Agent mutation controls documented. |
| 25 | Enable guarded mutation path. | Agent can run guarded service mutation when config enables it. | Actual mutation requires explicit agent config. | Mutation enablement docs updated. |
| 26 | Add runtime installation layout. | systemd units, install paths, example configs, rootfs layout. | Runtime paths standardized. | Packaging docs updated. |
| 27 | Add release workflow. | Tagged release CI, checksums, rootfs archive. | None. | Release workflow documented. |
| 28 | Add artifact smoke gate. | Downloaded artifact validation before publish. | Failed smoke blocks release. | Release operations updated. |
| 29 | Run first tagged release dry run. | Validated release assets through `v0.1.3`. | None. | Release lessons captured. |
| 30 | Add control-plane observability. | Metrics registry, governed metrics endpoint, UI cards. | Metrics require capability. | Metrics reference added. |
| 31 | Improve accessible feedback. | Toasts, tooltips, responsive UI, normalized status language. | None. | Accessibility notes added. |
| 32 | Add frontend regression tests. | Vitest/jsdom and component/API error tests. | None. | Testing docs updated. |
| 33 | Add documentation portal. | Structured wiki, references, validation script. | Docs validation becomes a gate. | Portal becomes primary documentation surface. |
| 34 | Add OpenAPI contract. | OpenAPI JSON/YAML, generated API reference, route coverage validation. | None. | API docs generated from contract. |
| 35 | Add backup and recovery. | Backup, restore, verify scripts and recovery docs. | None. | Operations recovery docs added. |
| 36 | Add upgrade safety. | Schema compatibility checks, pre-upgrade and verify-upgrade scripts. | Future/corrupt schemas refuse startup. | Upgrade and rollback docs updated. |
| 37 | Synchronize architecture documentation. | Implementation inventories, references, wiki validation, and docs reconciliation. | None. | Documentation drift reduced across wiki, OpenAPI, and governance docs. |
| 38 | Add governed identity administration. | Super Admin bootstrap, `identity.users.manage`, managed users, magic tokens, admin APIs, Administration page, self-lockout prevention. | Non-Super Admin password login is denied. | Identity, capability, audit, API, and user-guide docs updated. |
| 39 | Reduce complexity after identity work. | Administration view extraction, focused identity files, code-quality verifier. | None. | Code quality audit docs added. |
| 40 | Validate operator workflows. | Production readiness review and clearer operator-facing UI language. | None. | Phase 40 readiness report added. |
| 41 | Validate security model. | Threat model and security review covering residual risks and recommended actions. | None. | Security review and threat model updated. |
| 42 | Refresh documentation for release-candidate readiness. | Documentation reconciliation, release candidate checklist, stronger docs validation. | None. | Wiki, operations, references, and validation script aligned. |

## Reading Order

Start with [Architectural Inventory](Architectural-Inventory.md), then read the subsystem-specific reference page for the area you are changing.
