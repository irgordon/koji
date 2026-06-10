# Release Workflow

[Home](../Home.md) | Related: [Release Architecture](../Architecture/Release-Architecture.md), [Release Operations](../Operations/Release-Operations.md)

## What Problem This Solves

The release workflow ensures artifacts are built and validated consistently from tags.

## How It Works

Tags matching `v*` run Go setup, Node setup, dependency install, backend tests, frontend tests, frontend build, release assembly, verification, artifact upload, smoke tests, and GitHub Release publication.

## What Protects It

Smoke tests block publishing when checksums, rootfs layout, systemd units, executable bits, or forbidden paths are wrong.

## What Can Fail

Any validation step can fail and stop publication.

## How To Diagnose It

Read the failing workflow step and compare it with [Release Operations](../Operations/Release-Operations.md).

## How To Recover

Fix the root cause and create a new tag for the next dry run or release.
