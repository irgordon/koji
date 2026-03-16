# CLAUDE.md — KOJI Repository Instructions

This repository implements **KOJI**, a layered microkernel operating system.

Claude Code must operate under strict architectural and task discipline.

---

## Toolchain & Language Constraints

* **Kernel**

  * Language: **Odin only**
  * No other languages permitted

* **Userspace**

  * Language: **Go only**
  * No other languages permitted

* **Build Environment**

  * Platform: **macOS toolchains only**
  * No cross-platform build logic in this phase
  * No alternative compilers or environments

* **Python Tooling**

  * Python usage limited to tooling only
  * Must run inside a **virtual environment (venv)**
  * Dependency management via **PyYAML**
  * No global Python dependencies allowed

---

## Invariants

* Kernel and userspace languages must never mix
* Python must never be part of runtime or kernel code
* All Python tooling must be isolated to venv
* No implicit or undeclared toolchain dependencies

---

## Validation

* Verify no non-Odin files exist in `kernel/`
* Verify no non-Go files exist in `userspace/`
* Verify Python usage is limited to `tools/` and uses venv
* Verify PyYAML is declared and isolated

