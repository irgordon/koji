# Testing Strategy

[Home](../Home.md) | Related: [Frontend Architecture](Frontend-Architecture.md), [Backend Architecture](Backend-Architecture.md), [Release Workflow](Release-Workflow.md)

## What Problem This Solves

Tests prevent security, workflow, packaging, and UI regressions from reaching release artifacts.

## How It Works

Go tests cover backend packages. Vitest covers frontend behavior and accessibility primitives. Packaging scripts validate release layout, checksums, documentation, OpenAPI, and code-quality guardrails.

## What Protects It

CI runs frontend tests before build and release packaging. Release smoke tests validate downloaded artifacts before publish. Maintainers should use the [Release Candidate Checklist](../Operations/Release-Candidate-Checklist.md) before tagging or promoting a release.

## What Can Fail

Tests can miss untested behavior, generated assets can drift, or new code can bypass package ownership.

## How To Diagnose It

Run all validation commands from the phase report and release workflow.

## How To Recover

Add a focused regression test for the behavior that escaped.
