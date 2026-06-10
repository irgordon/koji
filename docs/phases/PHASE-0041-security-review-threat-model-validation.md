# Phase 41: Security Review and Threat Model Validation

## Goal

Validate Koji's implementation against its intended threat model after identity administration, magic tokens, jobs, approvals, agent RPC, backup, upgrade safety, observability, and release workflows were added.

## Non-Goals

- No runtime behavior changes.
- No new endpoints.
- No new capabilities.
- No database schema changes.
- No authentication or authorization redesign.

## Invariants Preserved

- The browser is never authoritative.
- The web server is never privileged.
- Protected APIs require explicit capability checks.
- Privileged mutation crosses the agent boundary.
- Audit read APIs expose normalized fields only.

## Review Summary

The review attempted to break Koji's assumptions from authenticated, capability-bearing, local socket, backup theft, magic-token leakage, and OpenAPI enumeration perspectives.

The highest residual risks are operational:

- high-risk capability assignment;
- magic-token delivery outside Koji;
- agent socket permissions;
- backup confidentiality;
- audit write failure policy;
- accidental production dev mode.

## Files Changed

- `docs/wiki/Security/Threat-Model.md`
- `docs/wiki/Security/Security-Review.md`
- `docs/phases/PHASE-0040-security-review-threat-model-validation.md`
- `docs/wiki/Home.md`
- `docs/CHANGELOG.md`

## Commands Run

```text
packaging/scripts/verify_docs.sh
make verify-openapi
packaging/scripts/verify_code_quality.sh
git diff --check
```

## Changelog

- Added a repo-grounded threat model with assets, trust boundaries, attacker model, abuse cases, and manual review paths.
- Added a security review with prioritized residual risks and follow-up recommendations.

## Summary

Koji's implementation remains aligned with the intended defense-in-depth model. Follow-up work should focus on runtime permission verification, backup confidentiality, socket mode enforcement, audit-failure policy, and production dev-mode guardrails.
