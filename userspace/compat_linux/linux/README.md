# Linux Domain

Handles all Linux-visible behavior:

- Syscall entry and dispatch (Linux syscall numbers)
- Linux struct parsing and validation
- errno semantics and mapping
- Signal semantics (delivery, masking, restart behavior)
- File descriptor table semantics
- Process/thread lifecycle semantics (fork, clone, exec, wait)
- Linux-specific blocking and retry semantics (EINTR, restartable syscalls)

**Constraints:**
- Must not construct KOJI IPC messages directly
- Must not manipulate capabilities directly
- Must not depend on KOJI object models
- Must not assume KOJI scheduling or memory behavior

See `docs/LINUX_COMPAT_ARCH.md` §3.1 for the full specification.
