# KOJI ABI Boundary Model

Version: 1.0
Status: Binding Constraint

---

## 1. Dual ABI Surfaces

KOJI defines two completely independent ABI surfaces.

### 1.1 KOJI Native ABI

Defined by: `abi/KOJI_ABI_V1.h`

Characteristics:
- minimal
- capability-based
- object-oriented (handles, rights, endpoints)
- explicitly versioned
- deterministic and stable
- designed for correctness and auditability

This is the only ABI the kernel understands.

### 1.2 Linux Compatibility ABI

Defined by: Linux syscall numbers, calling conventions, struct layouts, and semantics.

Characteristics:
- large
- legacy-constrained
- behavior-rich and implicit
- not versioned by KOJI
- defined externally (Linux kernel ABI)

This is not implemented in the kernel.

---

## 2. Strict Separation Rule

The kernel must only implement the KOJI ABI.

The kernel must:
- not implement Linux syscalls
- not interpret Linux structs
- not emulate Linux behavior
- not branch on Linux semantics
- not contain compatibility code

Any violation is a kernel policy leak.

---

## 3. Compatibility Layer Role

The Linux compatibility layer is a Ring 3 component.

It is the only component allowed to speak both ABIs.

**Responsibilities:**
- translate Linux syscalls to KOJI syscalls
- map Linux handles to KOJI capabilities
- emulate Linux semantics in user space
- manage Linux process expectations (signals, errno, etc.)

**Constraints:**
- must not extend kernel behavior
- must not bypass capability checks
- must not introduce implicit kernel policy

---

## 4. Translation Boundary

All translation occurs at a single, explicit boundary:

```
Linux Process
    | (Linux ABI)
Compatibility Layer (Ring 3)
    | (KOJI ABI)
Kernel (Ring 0)
```

There is no alternate path.

---

## 5. Non-Negotiable Invariants

### 5.1 Kernel Purity
- kernel only understands KOJI ABI
- kernel structures must never encode Linux semantics
- kernel syscall table must contain only KOJI syscalls

### 5.2 One-Way Knowledge
- kernel knows nothing about Linux ABI
- compatibility layer knows both
- Linux processes know nothing about KOJI ABI

### 5.3 No Shared ABI Structures
- Linux structs must never appear in `abi/`
- KOJI structs must never be exposed to Linux processes
- no dual-purpose structures

### 5.4 No Implicit Translation
- no "helpful" kernel behavior for Linux compatibility
- no fallback paths
- no hidden emulation

All translation must be explicit and visible in the compat layer.

---

## 6. Failure Model

If translation fails:
- compat layer returns Linux-compatible error
- kernel returns KOJI error codes only

There must never be mixed error domains.

---

## 7. Build and Repository Enforcement

This separation must be enforced mechanically.

**Directory isolation:**
- `kernel/` — KOJI ABI only
- `userspace/compat_linux/` — Linux ABI handling
- `abi/` — KOJI ABI only

**CI enforcement:**
- reject any reference to Linux ABI headers in `kernel/`
- reject any reference to KOJI ABI in Linux-facing public interfaces
- enforce that only compat layer imports both domains

**Toolchain enforcement:**
- kernel build must not include Linux headers
- compat layer may include both Linux headers and KOJI generated bindings

---

## 8. Design Consequences

This model forces:
- a clean microkernel boundary
- explicit translation cost
- no accidental ABI coupling
- easier auditing and verification
- ability to evolve KOJI ABI independently of Linux

---

## 9. Anti-Patterns (Must Reject)

The following are violations:
- kernel syscall numbers aligned to Linux syscalls
- kernel structs shaped like Linux structs
- "fast path" Linux syscalls in kernel
- shared headers between kernel and Linux compatibility layer
- conditional compilation inside kernel for Linux behavior

---

## 10. Deletion Test (Invariant)

At any point:

> Remove `userspace/compat_linux/`

The system must:
- still compile
- still boot
- still operate using KOJI-native ABI

If not, Linux semantics have leaked into the kernel or substrate.

---

## 11. Future Extensibility

This model allows:
- multiple compatibility layers (Linux, POSIX-lite, custom runtime)
- removal or replacement of Linux compat without kernel change
- strict ABI versioning for KOJI independent of Linux

---

## Final Principle

The kernel is not a Linux kernel.

It is a capability microkernel that exposes a minimal, explicit ABI.

Linux compatibility is a translation problem, not a kernel feature.
