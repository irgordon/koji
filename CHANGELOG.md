# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `tools/abi/gen_abi.py`: ABI generator that produces Odin, Go, and JSON manifest from
  the canonical C header. Supports `--check` mode for CI drift detection.
- `abi/generated/go/abi_generated.go`: Go ABI bindings (auto-generated).
- `abi/generated/meta/abi_manifest.json`: ABI manifest with struct sizes, field offsets,
  and constant values (auto-generated).
- `userspace/compat_linux/{linux,translate,koji}/`: Stub directory structure for the
  Linux compatibility layer, per `docs/LINUX_COMPAT_ARCH.md`.

### Changed

- `abi/KOJI_ABI_V1.h` (v1.0.0 → v1.1.0): Reconciled canonical C header with kernel
  implementation. Breaking changes relative to the v1.0.0 draft:
  - Syscall numbers reorganized: handles first (0–2), lifecycle (3–9), IPC (10–13),
    ports (14–16), memory (17–22), IRQ (23–24), misc (25). Removed `NOOP` placeholder.
    Renamed `SYS_HANDLE_DUP` → `SYS_HANDLE_DUPLICATE`. Added `SYS_PROCESS_INFO`,
    `SYS_THREAD_SUSPEND`, `SYS_THREAD_RESUME`, `SYS_VMO_READ`, `SYS_VMO_WRITE`,
    `SYS_VMAR_PROTECT`, `SYS_ABI_INFO`. Count: 26.
  - Error code ordering corrected: `ERR_INVALID_ARGS=3`, `ERR_ACCESS_DENIED=4`,
    `ERR_NO_MEMORY=5`. Replaced `ERR_CANCELLED`/`ERR_PEER_CLOSED` with
    `ERR_CHANNEL_CLOSED`/`ERR_WOULD_BLOCK`/`ERR_INTERNAL`.
  - Syscall frame simplified: callee-saved registers (rbx, rbp, r12–r15) removed.
    Field order reordered to match NASM push sequence (syscall_num first). 80 bytes.
  - IPC header fields renamed and reordered: `data_size`, `handle_count`, `ordinal`,
    `flags`. `IPC_MAX_MSG_SIZE` renamed to `IPC_MAX_DATA_BYTES`.
  - Added typed casts on all constants (e.g. `((koji_u32)0xFF000000)`).
- `abi/generated/odin/abi_generated.odin`: Regenerated from updated C header.

## [v0.0.0] - 2026-03-16

### Added

- Canonical repository topology (phase 0.1)
- Directory structure: `kernel/`, `userspace/`, `abi/`, `tools/`, `docs/`, `tests/`
- `README.md` with project governance and separation statements
- Architecture placeholder (`docs/architecture/README.md`)
- Invariant template (`docs/invariants/TEMPLATE.md`)

### Removed

- Legacy bootstrap files (`Makefile`, `harness/`, `verify/`, `docs/roadmaps/`)

[0.0.0]: https://github.com/irgordon/koji/releases/tag/v0.0.0
