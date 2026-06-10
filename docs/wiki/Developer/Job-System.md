# Job System

[Home](../Home.md) | Related: [Job Lifecycle](../Architecture/Job-Lifecycle.md), [Job State Reference](../Reference/Job-State-Reference.md)

## What Problem This Solves

The job system converts sensitive service-control intent into durable, reviewable, auditable workflow.

## How It Works

Service-control API creates queued jobs. Approval moves queued jobs to approved. The worker claims approved jobs and calls the agent. Completion or failure updates the durable row.

## What Protects It

State transitions are constrained, persisted, and audited. The creator does not bypass approval.

## What Can Fail

Invalid transitions, worker stoppage, agent failure, or DB errors can block progression.

## How To Diagnose It

Use Jobs page, tests under `internal/jobs`, and audit events.

## How To Recover

Fix the failed dependency, restart the worker daemon, or reject stale queued jobs.
