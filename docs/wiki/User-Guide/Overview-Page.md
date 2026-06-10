# Overview Page

[Home](../Home.md) | Related: [Observability](../Operations/Observability.md), [Metrics Reference](../Reference/Metrics-Reference.md)

## What Problem This Solves

The Overview page answers whether the host and Koji control plane look healthy before an operator approves work.

## How It Works

The page shows CPU, memory, disk, uptime, health, readiness, job flow, worker, agent RPC, audit writes, auth outcomes, and readiness counters.

## What Protects It

Protected telemetry requires authentication and capabilities. Control-plane metrics require `observability.metrics.read`.

## What Can Fail

Cards may show permission errors, stale data, degraded readiness, or unavailable agent state.

## How To Diagnose It

Read tooltips, stale-data notices, and status badges. Use [Troubleshooting](../Operations/Troubleshooting.md) when degraded.

## How To Recover

Fix the dependency shown by the failed card before approving new jobs.
