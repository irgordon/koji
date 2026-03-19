# KOJI Domain

Handles all interaction with the KOJI kernel via the native ABI.

Responsibilities:
- Capability operations
- IPC message construction
- KOJI syscall invocation
- Address space operations via KOJI ABI
- Thread control primitives
- Async notification consumption
- Handle lifecycle management

**Constraints:**
- Must not reference Linux structs
- Must not emit or interpret Linux errno values
- Must not encode Linux semantics
- Must not implement retry or restart behavior

See `docs/LINUX_COMPAT_ARCH.md` §3.2 for the full specification.
