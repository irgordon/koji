# Local Development

[Home](../Home.md) | Related: [Repository Layout](Repository-Layout.md), [Testing Strategy](Testing-Strategy.md), [Frontend Architecture](Frontend-Architecture.md)

## What Problem This Solves

Local development lets maintainers change Koji while preserving security boundaries and release gates.

## How It Works

Run Go tests from the repo root. Run frontend tests and builds from `web/`. Development mode may use a local frontend proxy, but production static assets are served from an absolute configured path.

## What Protects It

Tests cover backend governance, agent boundaries, command ownership, frontend accessibility, and packaging layout.

## What Can Fail

Local config can differ from production, frontend dependencies can be missing, or generated `dist/` assets can become stale.

## How To Diagnose It

Run `npm run test`, `npm run build`, `GOCACHE=/tmp/koji-go-cache go test ./...`, and `git diff --check`.

## How To Recover

Install dependencies with `npm install`, rebuild assets, and rerun all validation before committing.
