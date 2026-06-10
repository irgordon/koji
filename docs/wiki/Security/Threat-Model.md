# Threat Model

[Home](../Home.md) | Related: [Trust Boundaries](../Architecture/Trust-Boundaries.md), [Capabilities](Capabilities.md), [Agent Mutation Controls](Agent-Mutation-Controls.md)

## What Problem This Solves

The threat model explains what Koji protects against and where risk remains.

## Assets

- Sessions and CSRF tokens
- Capability assignments
- Audit records
- Service-control jobs
- Agent socket
- Host service state
- Configuration and SQLite database

## Attacker Capabilities

Assume attackers may reach the web UI, attempt credential abuse, reuse stale browser state, craft API requests, or attempt to trigger privileged host actions through `kojid`.

## Protections

Authentication, CSRF, capabilities, audit, allowlists, jobs, approval, bounded command execution, and the Unix socket agent boundary reduce blast radius.

## Failure Modes

Misconfigured allowlists, overbroad capabilities, weak host filesystem permissions, unavailable audit storage, or intentionally enabled mutation without review increase risk.

## Diagnosis and Recovery

Use Activity, request IDs, readiness, metrics, and release provenance. Revoke sessions, remove broad capabilities, restore DB backups, and disable agent mutation when risk is unclear.
