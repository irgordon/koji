<div align="center">

# Koji

[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](https://react.dev/)
[![Tauri](https://img.shields.io/badge/Tauri-24C8DB?style=for-the-badge&logo=tauri&logoColor=white)](https://tauri.app/)
[![SQLite](https://img.shields.io/badge/SQLite-07405E?style=for-the-badge&logo=sqlite&logoColor=white)](https://sqlite.org/)
[![Bash](https://img.shields.io/badge/Bash-121011?style=for-the-badge&logo=gnubash&logoColor=white)](https://www.gnu.org/software/bash/)

</div>

![Koji browser UI mockup](docs/assets/koji-browser-mockup.svg)

Koji is a modern control panel built for operators who want clarity, confidence, and calm control over their servers. It turns noisy, high-risk host administration into a governed operating experience: see what matters, understand what changed, approve sensitive actions deliberately, recover from mistakes, and keep a durable record of the decisions that shape your environment. Koji is designed to feel trustworthy from the first screen: polished enough for daily operations, restrained enough for production systems, and structured around the idea that powerful infrastructure tools should be visible, auditable, recoverable, and easy to reason about.

## Architecture Overview

Koji is split into clear responsibility zones:

- `kojid` serves the authenticated web/API control plane.
- `koji-agent` owns the privileged local boundary.
- SQLite stores users, sessions, capabilities, magic tokens, audit events, and jobs.
- The React/TypeScript frontend provides the operator workspace.
- Identity administration is governed by capabilities, audit, self-lockout prevention, and one-time magic tokens.
- Service-control intent flows through capability checks, audit, durable jobs, approval, and the agent boundary.
- Direct host command execution is centralized behind bounded platform adapters.
- OpenAPI, backup/restore tooling, upgrade safety checks, packaging scripts, and release smoke gates support production operations.

Core invariant:

```text
The browser is never authoritative.
The web server is never privileged.
The agent is the only privileged execution surface.
```

## Installation

The repository includes a staging install layout for Linux packaging work and a smoke-gated GitHub release workflow.

For local development, build, test, and stage an install tree from the repository:

```sh
make test
make build
make install
```

The install target writes to `build/rootfs/` by default, including binaries, static assets, example configuration, and systemd units.

Operational helpers are available for release, backup, restore, and upgrade checks:

```sh
make release
make backup
make verify-restore
make pre-upgrade-check
make verify-upgrade
```

Before tagging a release candidate, use the [Release Candidate Checklist](docs/wiki/Operations/Release-Candidate-Checklist.md).

## Release

Tagged releases are built by GitHub Actions from a clean checkout. The workflow builds Linux amd64 binaries, packages a staged rootfs archive, verifies checksums, smoke-tests the downloaded artifact set, and only then publishes release assets:

```sh
git tag v0.1.0
git push origin v0.1.0
```

Verify downloaded release assets with:

```sh
shasum -a 256 -c SHA256SUMS.txt
```

The first validated release dry run completed with `v0.1.3`, proving the tag-triggered build, smoke-test, publish, and downloaded-asset verification path.

## License

Koji is released under the license included in this repository. See [LICENSE](LICENSE) for details.
