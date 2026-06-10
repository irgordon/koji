# Jobs Page

[Home](../Home.md) | Related: [Job Lifecycle](../Architecture/Job-Lifecycle.md), [Job State Reference](../Reference/Job-State-Reference.md), [Activity Page](Activity-Page.md)

## What Problem This Solves

The Jobs page shows service-control intent as a durable lifecycle instead of hiding it behind a synchronous button.

## How It Works

Queued jobs can be approved or rejected with a reason. Approved jobs can be claimed by the worker and advanced through the agent path.

## What Protects It

Viewing jobs requires `jobs.read`. Approving or rejecting jobs requires `jobs.approve` and CSRF.

## What Can Fail

Jobs can remain queued, remain approved, fail, or report not implemented.

## How To Diagnose It

Inspect status, reason, request ID, Activity rows, and observability counters.

## How To Recover

Approve or reject queued jobs, restart the worker daemon, fix the agent, or recreate failed jobs after resolving the cause.
