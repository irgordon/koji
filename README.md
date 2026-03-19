# KOJI

KOJI is a layered microkernel operating system writtin in Odin and Go.

## Governance

The [KOJI Engineering Principles](docs/ENGINEERING_PRINCIPLES.md) are binding on all contributions to this repository.

## Architecture

Kernel and userspace are strictly separated. The `kernel/` and `userspace/` directories represent distinct privilege domains and must never share implementation code. The `abi/` directory is the sole legal location for ABI definitions shared between kernel and userspace.
