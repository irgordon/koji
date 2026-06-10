# Repository Layout

[Home](../Home.md) | Related: [Backend Architecture](Backend-Architecture.md), [Frontend Architecture](Frontend-Architecture.md), [Packaging and Deployment](../Architecture/Packaging-and-Deployment.md)

## What Problem This Solves

The layout separates command entrypoints, internal packages, frontend assets, packaging, docs, and generated distribution output.

## How It Works

- `cmd/kojid`: daemon entrypoint
- `cmd/koji-agent`: agent entrypoint
- `internal/`: backend packages
- `web/`: React/TypeScript frontend
- `dist/`: production frontend build output
- `packaging/`: examples, systemd units, verification scripts
- `docs/`: governance, phases, and wiki portal

## What Protects It

Internal packages keep privileged logic out of HTTP handlers and centralize command execution in the platform command runner.

## What Can Fail

Feature work can drift across package boundaries or create duplicated behavior.

## How To Diagnose It

Use package ownership docs and command scans from the phase validation prompts.

## How To Recover

Move behavior back to the owning package and add regression tests at the boundary.
