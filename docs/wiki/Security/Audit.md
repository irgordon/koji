# Audit

[Home](../Home.md) | Related: [Activity Page](../User-Guide/Activity-Page.md), [Audit Event Reference](../Reference/Audit-Event-Reference.md), [Observability](../Operations/Observability.md)

## What Problem This Solves

Audit records who attempted sensitive actions, what target was involved, the outcome, and the request ID needed for incident review.

## How It Works

Koji writes durable audit events for auth lifecycle, capability denials, service-control intent, job creation, job decisions, worker advancement, and process-list access.

## What Protects It

The Activity API exposes normalized fields only. Raw internals, remote address details, and sensitive errors are not exposed in the frontend read model.

## What Can Fail

Database write failures can prevent audit persistence.

## How To Diagnose It

Check `audit_writes_total`, `audit_write_failures_total`, `/readyz`, and the Activity page.

## How To Recover

Pause approvals, restore DB availability, fix filesystem ownership, and confirm audit counters resume.
