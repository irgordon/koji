# AGENTS.md

## 1. Required Reading

Before changing code, read:

1. `ARCHITECTURE.md`
2. `INVARIANTS.md`
3. `CODING_STYLE.md`
4. `SECURITY.md`
5. `PHASEMAP.md`

These files are part of the project contract.

## 2. Operating Rules

- One task, one surface.
- Preserve all invariants.
- No bypasses.
- No hidden behavior.
- No dead code.
- No TODO placeholders without an associated issue or phase note.
- No speculative abstractions.
- No framework magic.
- No arbitrary shell execution.
- No UI-only security enforcement.

## 3. Architecture Rules

- The browser is never authoritative.
- The web server is never privileged.
- The agent is the only privileged execution surface.
- Every privileged action requires authentication, authorization, capability check, validation, execution through the allowed boundary, and audit.

## 4. Coding Rules

Follow `CODING_STYLE.md`.

Pay particular attention to:

- Top-down code organization.
- Early returns.
- Shallow nesting.
- Human-readable names.
- Explicit types.
- No boolean flag arguments.
- No global mutable state.
- Clear error context.
- Negative-path tests.

## 5. Testing Rules

Changes must include tests when they affect:

- Authentication.
- Authorization.
- Capabilities.
- Agent boundary behavior.
- Configuration validation.
- Database migrations.
- Audit events.
- Job state transitions.
- Sensor degradation.
- External command execution.

Invariant-preserving tests are mandatory for invariant-sensitive changes.

## 6. Documentation Rules

If a change modifies architecture, trust boundaries, capabilities, configuration, migrations, privileged operations, or invariants, update the relevant governance document in the same change.

If a change violates or weakens an invariant, stop. Do not implement it unless the invariant is intentionally revised with clear justification.

## 7. Required Phase Report

Every implementation phase must produce a phase note using `PHASEMAP.md`.

The phase note must include:

- Goal.
- Non-goals.
- Invariants preserved.
- Negative patterns avoided.
- Commands run.
- Changelog.
- Summary.
- Notes and deviations.

## 8. Prohibited Shortcuts

Do not use:

```text
quick hack
temporary bypass
unsafe shell path
root web server fallback
TODO security later
silent no-op
unbounded buffer
manual DB edit
config write from web request
```
