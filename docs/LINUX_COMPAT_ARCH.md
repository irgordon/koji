# KOJI Linux Compatibility Layer — Internal Architecture

Version: 1.0
Status: Binding Constraint

---

## 1. Purpose

This document defines the internal structure of the Linux compatibility layer
located under `userspace/compat_linux/`.

The goal is to prevent semantic drift inside the compat layer itself by
enforcing a strict internal boundary between:
- Linux-facing semantics
- KOJI-native interactions

The compat layer must remain a translation system, not a hybrid semantic environment.

---

## 2. Core Model

The compat layer is divided into three distinct domains:

```
Linux Domain  →  Translation Boundary  →  KOJI Domain
```

### 2.1 Linux Domain

Handles all Linux-visible behavior.

### 2.2 Translation Boundary

Explicit, narrow, and audited conversion layer.

### 2.3 KOJI Domain

Handles all interaction with the KOJI kernel via the native ABI.

---

## 3. Domain Responsibilities

### 3.1 Linux Domain (Linux-Facing)

Responsible for:
- syscall entry and dispatch (Linux syscall numbers)
- Linux struct parsing and validation
- errno semantics and mapping
- signal semantics (delivery, masking, restart behavior)
- ptrace-visible behavior (v2+; see §13)
- file descriptor table semantics
- process/thread lifecycle semantics (fork, clone, exec, wait)
- Linux-specific blocking and retry semantics (EINTR, restartable syscalls)

Constraints:
- must not construct KOJI IPC messages directly
- must not manipulate capabilities directly
- must not depend on KOJI object models
- must not assume KOJI scheduling or memory behavior

### 3.2 KOJI Domain (Kernel-Facing)

Responsible for:
- capability operations
- IPC message construction
- KOJI syscall invocation
- address space operations via KOJI ABI
- thread control primitives
- async notification consumption
- handle lifecycle management

Constraints:
- must not reference Linux structs
- must not emit or interpret Linux errno values
- must not encode Linux semantics
- must not implement retry or restart behavior

### 3.3 Translation Boundary

The only layer allowed to depend on both domains.

Responsible for:
- Linux struct → KOJI message translation
- KOJI result → Linux struct / errno translation
- handle ↔ file descriptor mapping
- timeout and cancellation mapping
- signal-related state transitions (in cooperation with signal subsystem)

Constraints:
- stateless by default; state allowed only where explicitly defined and owned
- must not accumulate policy
- must not leak types across domains
- must not introduce implicit behavior
- must never be behaviorally stateful (no history-dependent semantics)

---

## 4. Defined Stateful Exception: FD → Capability Mapping Table

This is the only required stateful structure at the boundary in v1.

**Properties:**
- Owned by: Linux domain (not KOJI domain)
- Accessed via: translation layer
- Represents: Linux-visible resource identity
- Backed by: KOJI capabilities

**Invariants:**
- no KOJI capability may exist in Linux domain without an fd mapping
- no fd may reference more than one live capability
- mapping must be explicitly mutated (no implicit duplication)
- mapping must enforce rights via capability restrictions
- mapping must be per-process isolated (one compat instance per Linux process)

**Lifecycle:**
- created on open / create / receive
- duplicated via `dup*` semantics (explicit copy or shared ref)
- transferred via fork/clone semantics (subject to CCR-005 process instantiation support)
- destroyed on close / process exit

**Prohibited boundary state:**
- cached Linux structs
- partial syscall state
- retry bookkeeping
- signal queues (belong to signal subsystem, not translation)
- KOJI object ownership state

Any additional state must be explicitly justified and documented at the same
level as this table.

---

## 5. Internal Boundary Rules

### 5.1 No Cross-Domain Leakage
- Linux types must not appear in KOJI domain modules
- KOJI types must not appear in Linux domain modules
- only the translation layer may reference both

### 5.2 Single Translation Pass per Operation

Each syscall or operation must:
1. enter via Linux domain
2. translate once
3. execute in KOJI domain
4. translate once back

Multiple translation passes are not allowed.

### 5.3 Error Domain Isolation
- KOJI domain returns KOJI error codes only
- Linux domain exposes Linux errno only
- translation layer performs the mapping

Mixed error domains are forbidden.

### 5.4 No Implicit Semantics
- no hidden retries
- no silent conversions
- no fallback paths

All behavior must be explicit and traceable.

---

## 6. Signal Model Requirements

Linux signals are asynchronous control-flow interventions, not simple messages.

The compat layer must synthesize:
- asynchronous delivery
- interruption of blocking syscalls (EINTR)
- handler redirection
- alternate stack (`sigaltstack`)
- interaction with thread state

### 6.1 KOJI Upstream Dependencies (CCR)

The KOJI native ABI must provide primitives sufficient for:
- interrupting a blocked thread (CCR-001)
- scheduling a user-mode execution redirection (CCR-002)
- retrieving thread context for user-mode reconstruction (CCR-003)
- distinguishing timeout vs cancellation vs interruption (CCR-004 — satisfied)

The kernel must not implement signals, but must provide sufficient primitives
to build them.

### 6.2 Compat Layer Responsibilities

The compat layer must:
- maintain signal state per thread
- decide delivery timing
- synthesize handler frames
- manage restart vs interrupt semantics
- map KOJI interruption primitives to Linux-visible behavior

---

## 7. Runtime Architecture

### 7.1 Process Model (Binding Decision)

**One compat instance per Linux process.**

```
Linux Process (PID X)  ↔  compat instance X  ↔  KOJI kernel
```

A shared global compat server is rejected because it would require:
- holding capabilities for multiple isolated processes
- mediating access across isolation boundaries
- acting as a capability broker

This violates KOJI principles: breaks ownership clarity, introduces ambient
authority, complicates revocation, and creates cross-process coupling.

Each compat instance:
- owns its own capability set
- maintains its own fd table
- operates within one KOJI address space
- maps directly to KOJI process/thread primitives

### 7.2 Internal Decomposition (Mandatory)

Even within a single process, the following subsystems must be isolated:
- syscall dispatch
- memory management translation
- file descriptor / handle mapping
- signal subsystem
- process/thread lifecycle
- KOJI IPC adapter

No shared mutable state without explicit ownership.

### 7.3 Future Decomposition (Allowed)

Subsystems may later be split into separate Ring 3 services:
- filesystem emulation
- networking
- `/proc` and metadata services

Constraints:
- must not change kernel behavior
- must preserve ABI boundaries
- must not introduce implicit coupling

---

## 8. Fork / Clone Translation

Under the per-process model, fork/clone requires the parent compat instance
to create a new child compat instance.

This requires:
- creation of a new KOJI address space
- creation of at least one runnable thread
- explicit capability provisioning for the child
- explicit fd table duplication per fork semantics

The exact mechanism is governed by CCR-005 (compat instance creation primitive),
which is an open requirement against the native ABI. The fd table lifecycle
section is intentionally incomplete until CCR-005 is resolved.

---

## 9. Blocking and Restart Semantics

Linux allows:
- syscall interruption (EINTR)
- restartable syscalls
- cancellation

KOJI provides:
- explicit timeout
- explicit cancellation
- deterministic return paths (distinct: TIMEOUT, CANCELLED, INTERRUPTED)

The compat layer must:
- translate KOJI return conditions into Linux semantics
- implement restart logic entirely in user space
- never require kernel retry behavior

---

## 10. Directory Structure and Enforcement

```
userspace/compat_linux/
  linux/       Linux domain only
  translate/   boundary only
  koji/        KOJI domain only
```

**CI rules:**
- reject Linux headers or types in `koji/`
- reject KOJI ABI usage in `linux/`
- allow both only in `translate/`

**Build rules:**
- modules must compile independently where possible
- no implicit cross-imports

---

## 11. Anti-Patterns (Must Reject)

- Linux structs passed into KOJI domain modules
- KOJI handles exposed directly to Linux-facing code
- retry logic implemented in KOJI-facing modules
- signal logic embedded in KOJI interaction layer
- direct syscall translation without explicit boundary layer
- shared "common" structs mixing Linux and KOJI semantics
- behaviorally stateful translation layer (history-dependent semantics)

---

## 12. Deletion Test (Invariant)

At any point:

> Remove `userspace/compat_linux/`

The system must:
- still compile
- still boot
- still operate using KOJI-native ABI

If not, Linux semantics have leaked into the kernel or substrate.

---

## 13. ptrace Scope

ptrace is explicitly out of scope for v1.

v1 provides:
- enough primitives for Go runtime
- standard Linux application support
- signal handling
- basic process control

Full ptrace (v2+) requires:
- cross-thread register inspection
- cross-thread memory inspection
- syscall interception
- execution control
- capability model extensions
- security model decisions

> ptrace is not part of KOJI v1 compatibility guarantees.

---

## 14. Open Requirements (CCR)

| ID      | Description                           | Blocks                                    |
|---------|---------------------------------------|-------------------------------------------|
| CCR-001 | Thread interruption primitive         | Signal delivery, EINTR                    |
| CCR-002 | Execution redirection primitive       | Signal handler dispatch                   |
| CCR-003 | Execution context access              | Handler frame construction                |
| CCR-004 | Timeout / cancellation / interruption | **Satisfied** — distinct error codes      |
| CCR-005 | Compat instance creation primitive    | fork/clone translation, fd table transfer |

---

## Final Principle

The compatibility layer must remain a pure translation system with explicit,
bounded state.

If it becomes a shared authority, a policy engine, or a semantic hybrid,
the architecture has failed.
