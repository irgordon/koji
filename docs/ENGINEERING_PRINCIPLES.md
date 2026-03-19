# KOJI Engineering Principles

Version: 1.0  
Status: Project Standard

This document defines the engineering principles governing the KOJI operating system.

These principles are **normative** and binding. They are committed in-repo under `docs/ENGINEERING_PRINCIPLES.md`. They exist to preserve long-term maintainability, correctness, and architectural integrity.

Any change that makes the system harder to reason about is incorrect.

---

# 1. Core Architectural Model

KOJI is a strict layered microkernel system.

* Ring 0: Microkernel (mechanism only)
* Ring 3: Higher Substrate (system implementation)
* Ring 3: Linux Compatibility Layer (translation only)

Layers must remain strictly separated and independently evolvable.

---

# 2. Architectural Law

**Anything that can run in Ring 3 must not run in Ring 0.**

The kernel provides mechanisms only.  
All policy must exist in user mode.

---

# 3. System Trust Model

The system assumes:

* Kernel is fully trusted
* User-mode is untrusted and potentially adversarial
* IPC participants may be malicious
* All inputs crossing privilege boundaries are untrusted

All kernel behavior must be correct under adversarial input.

---

# 4. Non-Negotiable Constraints

The kernel must never:

* implement policy
* interpret high-level semantics (filesystems, paths, user intent)
* perform implicit retries or recovery logic
* allocate resources on behalf of user-mode policy (except explicit ownership transfer)
* expose internal pointers or memory layout
* depend on user-mode services for forward progress (outside defined IPC)
* contain compatibility logic (e.g., Linux behavior)

All violations must be rejected.

---

# 5. Boundary Enforcement

Layer separation must be mechanically enforced.

* Kernel and user-mode must compile separately
* No shared implementation across privilege boundaries
* All interaction occurs via defined syscalls and IPC only
* Shared structures must be ABI-defined, versioned, and validated
* No implicit cross-layer dependencies
* No shared mutable state across privilege boundaries
* Governing documents must be committed in-repo; architecture must not depend on external or unpublished documents

---

# 6. Microkernel Responsibilities

The kernel is minimal and deterministic.

Allowed:

* interrupt handling
* trap handling and bootstrapping
* scheduling primitives (mechanism only)
* IPC transport
* capability validation and resolution
* address space management
* thread lifecycle primitives
* hardware abstraction primitives (strictly defined below)

Not allowed:

* filesystems
* networking
* system orchestration
* resource policy
* compatibility layers
* service frameworks
* buffering, retry logic, or device policy

---

## 6.1 Hardware Abstraction Boundary

Hardware abstraction primitives are strictly limited to:

Allowed:

* register access
* interrupt routing and acknowledgement
* MMIO mapping
* DMA setup and teardown
* minimal device initialization required for system bring-up

Not allowed:

* device-specific policy
* buffering strategies
* retry or recovery logic
* protocol interpretation

All higher-level device logic must exist in user mode.

---

# 7. Capability Model

Capabilities are the sole authority mechanism.

## 7.1 Representation

* opaque index into a kernel-managed table (CNode)
* no pointer exposure to user space
* no encoding of kernel memory addresses

## 7.2 Validation (Required at Every Ingress)

Validation must include:

* existence check
* rights mask enforcement
* generation/version check (prevents reuse after revocation)

Validation must occur at every syscall and IPC boundary.  
Prior validation must never be trusted.

## 7.3 Delegation

* capabilities may only be transferred via IPC
* rights must be subset-restricted on delegation
* amplification of authority is forbidden

## 7.4 Lifecycle

Capabilities must support:

* creation via explicit kernel operation
* delegation via IPC
* revocation (defined below)
* reuse prevention via generation counters

## 7.5 Revocation Model

* revocation must be supported and explicit
* revocation invalidates all future use immediately
* in-flight operations must define behavior (fail or complete deterministically)
* stale capabilities must fail validation via generation mismatch

Lazy revocation is permitted only if externally indistinguishable from immediate revocation.

---

# 8. IPC Model

IPC is the primary communication mechanism.

## 8.1 Core Semantics

* default: synchronous message passing
* sender blocks until receiver replies or failure occurs
* message size must be bounded and ABI-defined
* all messages validated on send and receive

## 8.2 Failure Model

All IPC must define:

* timeout behavior (explicit)
* cancellation behavior (explicit)
* failure return codes (deterministic)
* deadlock visibility (must be observable, not hidden)

Kernel must not resolve deadlocks.

## 8.3 Scheduling Interaction

* blocking state must be visible to user-mode schedulers
* kernel must not implement priority inheritance unless explicitly defined as mechanism-only
* kernel may expose scheduling primitives (e.g., priority representation) without encoding policy
* no implicit scheduling policy

## 8.4 Data Transfer Semantics

* copy semantics are default
* shared memory requires:
  * explicit ownership definition
  * explicit synchronization model
  * explicit lifecycle management

---

# 9. Memory Model

The system defines explicit memory behavior.

## 9.1 Kernel Guarantees

* kernel synchronization semantics default to sequential consistency unless explicitly documented otherwise
* all kernel synchronization primitives must enforce defined ordering

## 9.2 User Memory

* no implicit ordering guarantees for shared memory
* all ordering must be explicitly defined by user-mode

## 9.3 Atomics

* restricted to well-defined kernel primitives
* relaxed ordering requires explicit justification and documentation

---

# 10. Scheduler Boundary

Scheduling is split between mechanism and policy.

## 10.1 Kernel Responsibilities

* run queue management
* context switching
* thread state transitions (ready, blocked, running)

## 10.2 User-Mode Responsibilities

* priority decisions
* fairness policies
* CPU allocation strategies

## 10.3 Constraints

* kernel must not encode scheduling policy
* kernel provides no fairness guarantees
* contention resolution is not handled in kernel

---

# 11. Concurrency Doctrine

Concurrency must be explicit and constrained.

* prefer message passing over shared memory
* no shared mutable state without explicit ownership
* lock-free structures require formal justification
* lock ordering must be documented and enforced
* avoid hidden cross-thread mutation
* priority inversion must be explicitly handled or prevented (not ignored)

---

# 12. Resource Ownership Model

All resources must have explicit ownership.

* kernel allocates only when explicitly requested
* ownership must be transferable via defined mechanisms
* resource lifetime must be deterministic
* leaked resources must be reclaimable via defined policy in user mode

Process and address-space creation authority must be explicitly modeled.  
Delegation, scope, and revocation of creation authority must be defined before such primitives are stabilized.

Kernel must not implement reclamation policy.

---

# 13. Determinism Definition

Determinism means:

* identical inputs and scheduling decisions produce identical outcomes
* all failure paths are explicit and reproducible

Determinism does not imply identical timing across hardware.

---

# 14. System Design Rules

## 14.1 Explicit Boundaries

Every subsystem must define:

* ownership
* input contracts
* output guarantees
* failure modes
* concurrency model
* security assumptions

---

## 14.2 Explicit Control Flow

* no hidden state
* no implicit side effects
* no opaque abstractions
* behavior must be traceable from entry to exit

---

## 14.3 Invariant Enforcement

All critical invariants must be validated at:

* syscall ingress
* IPC boundaries
* capability resolution
* scheduling transitions

No mutation before validation.

---

## 14.4 Failure Transparency

* all failures must be explicit
* no silent failure
* no partial mutation before validation
* deterministic error paths required

Undefined behavior is unacceptable.

---

# 15. Kernel Object Model

All kernel objects must be:

* opaque
* capability-addressable only
* inaccessible without valid capability
* lifecycle-managed (explicit creation and destruction)

No direct references are exposed to user space.

---

# 16. Bootstrapping and Trust Root

System initialization must define:

* initial capability set (root capabilities)
* first user-mode thread creation
* initial scheduler bootstrap
* kernel → user transition guarantees

Boot sequence must be deterministic and auditable.

---

# 17. Error Model

Errors must be:

* globally defined and enumerable
* deterministic and reproducible
* classified as:
  * terminal (must not retry)
  * retryable (user-mode decision only)

Kernel must not perform retries.

---

# 18. Code Structure

## Naming

Names must encode:

* subsystem
* behavior
* ownership

## Functions

* one conceptual operation
* explicit control flow
* validate before mutation

## Data Structures

* small and explicit
* clear ownership and lifecycle
* no global mutable structures

## Source Layout

* organized by subsystem
* architecture-specific code isolated
* no cross-layer contamination

---

# 19. Comments

Comments must explain:

* invariants
* security assumptions
* concurrency behavior
* non-obvious design decisions

Comments must not restate code.

---

# 20. Testing Requirements

Testing is mandatory.

Required:

* unit tests
* invariant validation tests
* syscall fuzzing
* IPC contract testing
* capability misuse testing
* fault injection at all boundaries

Required before advanced concurrency or SMP:

* deterministic replay for concurrency issues

Tests must be deterministic, isolated, and verify a single property.

---

# 21. Code Review Standard

Code is acceptable only if a reviewer can:

1. understand the subsystem without external explanation
2. trace all control flow deterministically
3. identify ownership boundaries
4. identify all invariants
5. determine how invariants could be violated
6. modify behavior without systemic risk

If not, the code must be rejected.

---

# 22. Kernel Change Justification

Any kernel addition must include:

* justification for why it cannot exist in Ring 3
* security impact analysis
* invariant impact analysis

Weak justification results in rejection.

---

# 23. Deletion Bias

* removing code is preferred to adding abstraction
* duplication is acceptable until invariants stabilize
* abstraction must follow understanding, not precede it

---

# 24. Architectural Drift Prevention

Indicators of failure:

* policy logic entering kernel
* expansion of hardware abstraction scope
* increasing cross-layer coupling
* hidden ownership or unclear boundaries
* convenience abstractions in critical paths

When detected:

* refactor immediately
* do not defer correction
* do not justify temporary violations

---

# Final Principle

A system that cannot be understood cannot be safely modified.

Maintainability is a security property.

Every line of code must preserve the ability for future engineers to reason about the system.