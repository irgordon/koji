# Translation Boundary

The only layer allowed to depend on both the Linux and KOJI domains.

Responsibilities:
- Linux struct → KOJI message translation
- KOJI result → Linux struct / errno translation
- Handle ↔ file descriptor mapping (see: FD → Capability Mapping Table, §4)
- Timeout and cancellation mapping
- Signal-related state transitions (in cooperation with signal subsystem)

**Constraints:**
- Stateless by default; state allowed only where explicitly defined and owned
- Must not accumulate policy
- Must not leak types across domains
- Must not introduce implicit behavior
- Must never be behaviorally stateful (no history-dependent semantics)

See `docs/LINUX_COMPAT_ARCH.md` §3.3 for the full specification.
