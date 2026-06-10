# Observability

[Home](../Home.md) | Related: [Metrics Reference](../Reference/Metrics-Reference.md), [Overview Page](../User-Guide/Overview-Page.md), [Health and Readiness](Health-and-Readiness.md)

## What Problem This Solves

Observability answers whether Koji's control plane is healthy without requiring log inspection or external telemetry.

## How It Works

Koji maintains in-process counters for jobs, worker polling, agent RPC, auth outcomes, audit writes, and readiness checks. The UI reads them through a governed API requiring `observability.metrics.read`.

## What Protects It

Metrics expose fixed control-plane counters and job status aggregates only. They do not expose user data, sessions, raw audit internals, or host process details.

## What Can Fail

Counters can show agent RPC failures, audit write failures, worker errors, or readiness dependency failures.

## How To Diagnose It

Use the Overview control-plane cards and [Metrics Reference](../Reference/Metrics-Reference.md).

## How To Recover

Fix the failing dependency, then confirm counters stop increasing for failure paths.
