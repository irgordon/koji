# CLAUDE.md — KOJI Repository Instructions

This repository implements **KOJI**, a layered microkernel operating system.

Claude Code must operate under strict architectural and task discipline.

---

## Source of Authority

The following files are authoritative and must be read before any work:

- `AGENTS.md` (repository root)
- `kernel/AGENTS.md`
- `tests/AGENTS.md`
- `tools/AGENTS.md`
- `harness/AGENTS.md`
- `harness/PHASES.yaml`

If any instruction conflicts, **AGENTS.md files take precedence**.

---

## Operating Rules

Claude Code must:

- Follow the active phase defined in `harness/PHASES.yaml`
- Execute only explicitly provided tasks
- Modify only files listed in the task’s allowed paths
- Stop immediately once acceptance checks pass
- Refuse speculative or exploratory changes

Claude Code must not:

- Invent tasks or expand scope
- Modify forbidden paths
- Edit generated files directly
- Introduce kernel policy or Linux semantics

---

## Architecture Reminder

- Ring 0: Microkernel (mechanism only)
- Ring 3: Higher Substrate
- Ring 3: Linux Compatibility Layer

Anything that can run in Ring 3 must not run in Ring 0.

---

## Definition of Done

Work is complete only when:

- Required outputs exist
- Acceptance checks pass
- Scope boundaries were respected
- No unrelated repository drift occurred