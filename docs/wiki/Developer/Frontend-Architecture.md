# Frontend Architecture

[Home](../Home.md) | Related: [Testing Strategy](Testing-Strategy.md), [Overview Page](../User-Guide/Overview-Page.md)

## What Problem This Solves

The frontend provides an operator UI that explains policy, redaction, jobs, approval, audit, and control-plane health in plain language.

## How It Works

The React app uses typed API helpers, normalized error mapping, hash-based navigation, reusable feedback primitives, and plain CSS. Tests use Vitest, Testing Library, and jsdom.

## What Protects It

The UI does not bypass backend auth, capabilities, CSRF, allowlists, or agent boundaries. It displays only backend-approved data.

## What Can Fail

Errors can be unclear, tests can miss a new accessibility regression, or generated assets can be stale.

## How To Diagnose It

Run `npm run test`, `npm run build`, and inspect UI messages for raw backend internals.

## How To Recover

Add component tests for new UI primitives and keep error messages normalized.
