# Release Operations

[Home](../Home.md) | Related: [Release Architecture](../Architecture/Release-Architecture.md), [Release Workflow](../Developer/Release-Workflow.md), [Release Candidate Checklist](Release-Candidate-Checklist.md), [Troubleshooting](Troubleshooting.md)

## What Problem This Solves

Release operations ensure users receive artifacts that were built, tested, checksummed, smoke-tested, and published by CI.

## How It Works

Tags matching `v*` trigger the release workflow. CI builds binaries and rootfs, verifies checksums and layout, then publishes assets after smoke tests pass.

Before tagging, run the [Release Candidate Checklist](Release-Candidate-Checklist.md) against the candidate build and staging state.

## What Protects It

Frontend tests, frontend build, Go tests, artifact checksums, rootfs validation, systemd unit validation, and forbidden path scans block release.

## What Can Fail

Tests can fail, artifacts can lose executable bits, checksums can mismatch, or smoke tests can detect layout drift.

## How To Diagnose It

Inspect the failing CI job and compare output with [Release Architecture](../Architecture/Release-Architecture.md).

## How To Recover

Fix the gate, push a commit, create a new tag, and verify downloaded assets before distribution.
