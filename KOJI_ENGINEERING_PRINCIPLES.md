# KOJI Engineering Principles

Version: 1.0  
Status: Project Standard

This document defines the engineering principles governing the KOJI operating system.

The goal of these principles is long-term maintainability, correctness, and architectural integrity.  
An operating system that becomes difficult to reason about will eventually become unsafe to modify.

---

# 1. Core Architectural Model

KOJI is built as a strict layered microkernel system.

- Ring 0: Microkernel
- Ring 3: Higher Substrate
- Ring 3: Linux Compatibility Layer

The layers must remain strictly separated.

---

# 2. Architectural Law

**Anything that can run in Ring 3 must not run in Ring 0.**

The microkernel exists to provide mechanisms only.  
Policy belongs outside the kernel.

---

# 3. Microkernel Responsibilities

Allowed responsibilities:

- hardware abstraction primitives
- interrupt handling
- scheduling primitives
- inter-process communication
- capability validation
- address space management
- thread primitives
- minimal boot and trap handling
- deterministic timing primitives required by scheduling

The kernel must not contain:

- filesystems
- network stacks
- complex drivers unless required for early boot
- system orchestration
- resource policy
- Linux compatibility logic
- service frameworks
- convenience abstractions

---

# 4. Layer Responsibilities

## Microkernel
Purpose: Provide minimal, deterministic primitives.

## Higher Substrate
Purpose: Implement drivers, filesystems, networking, and system services.

## Linux Compatibility Layer
Purpose: Translate Linux semantics into KOJI primitives.

This layer must adapt Linux behavior to KOJI, not the reverse.

---

# 5. Engineering Priorities

Priority order:

1. Correctness of invariants
2. Simplicity of control flow
3. Explicit ownership and boundaries
4. Observability and diagnosability
5. Performance
6. Stylistic elegance

---

# 6. System Design Rules

## Explicit Boundaries
Every subsystem must clearly define:

- ownership
- input contracts
- output guarantees
- failure modes
- concurrency model
- security assumptions

## Mechanism vs Policy
Kernel components provide mechanism.  
User-mode services provide policy.

## Explicit Control Flow
Avoid implicit side effects, hidden control flow, and clever abstractions that obscure behavior.

## Invariant Enforcement
Critical invariants must be validated at boundaries.

## Failure Transparency
Failure must always be explicit.

---

# 7. Concurrency Doctrine

Concurrency must be explicit and documented.

Guidelines:

- prefer message passing across subsystems
- prefer single-owner models
- avoid lock-free algorithms unless necessary
- document lock ordering rules
- avoid hidden cross-thread mutation

---

# 8. Code Structure

## Naming
Names must communicate purpose, subsystem, and mutation behavior.

Examples:

- `cap_validate_handle`
- `ipc_copy_in_message`
- `sched_enqueue_runnable`
- `linux_open_translate_flags`

## Functions
Functions should:

- perform one conceptual operation
- maintain one invariant story
- keep control flow obvious
- validate inputs before mutation

## Data Structures
Kernel data structures must be small, explicit, and ownership-safe.

## Source Layout
Organize code by subsystem.  
Architecture-specific code must not leak into generic subsystems.

---

# 9. Comments

Comments are required for:

- invariants
- security assumptions
- concurrency behavior
- hardware constraints
- non-obvious design choices

A comment should explain why, not what.

---

# 10. Testing Principles

Tests must verify behavior and invariants.

Test categories:

- unit tests
- invariant tests
- IPC contract tests
- compatibility tests
- fault injection tests

A test should verify one property or invariant.

---

# 11. Code Smells

- rigidity
- fragility
- immobility
- needless complexity
- needless repetition
- opacity

---

# 12. System-Specific Smells

- policy logic inside kernel
- architecture assumptions in generic code
- hidden allocations in critical paths
- implicit ownership transfer
- compatibility hacks bypassing kernel primitives
- ambiguous error codes
- helper layers hiding privilege transitions
- temporary kernel shortcuts for user-mode features

---

# 13. Code Review Standard

Code is acceptable only if a future developer can:

1. understand the subsystem without external explanation
2. trace control flow deterministically
3. identify ownership boundaries
4. modify behavior without violating invariants

---

# 14. Final Principle

KOJI must remain understandable.

A system that cannot be understood cannot be safely changed.  
A system that cannot be safely changed will eventually fail.