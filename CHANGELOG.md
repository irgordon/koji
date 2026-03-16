# Changelog

All notable changes to KOJI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Versioning

KOJI uses a three-part version scheme: `vX.Y.Z`

- **vX.0.0** — Release (new major release milestone)
- **v0.X.0** — Major update (significant feature or architectural change)
- **v0.0.X** — Minor/security update (patches, fixes, security hardening)

## [v0.0.0] - 2026-03-16

### Added

- Canonical repository topology (phase 0.1)
- Directory structure: `kernel/`, `userspace/`, `abi/`, `tools/`, `docs/`, `tests/`
- `README.md` with project governance and separation statements
- Architecture placeholder (`docs/architecture/README.md`)
- Invariant template (`docs/invariants/TEMPLATE.md`)

### Removed

- Legacy bootstrap files (`Makefile`, `harness/`, `verify/`, `docs/roadmaps/`)
