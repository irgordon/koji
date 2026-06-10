# Koji Documentation Portal

Koji is a governed host control panel for operators who need safe visibility, durable intent, human approval, auditability, and a strict daemon-to-agent privilege boundary.

## Operator Quick Start

1. Install Koji from a verified release artifact.
2. Configure `/etc/koji/koji.yaml` and `/etc/koji/agent.yaml`.
3. Start `kojid` and `koji-agent` with the packaged systemd units.
4. Verify `/healthz`, `/readyz`, and the web UI.
5. Operate through allowlisted services, approval-backed jobs, Activity, and observability cards.

Start with [Installation](Operations/Installation.md), [Configuration](Operations/Configuration.md), [Health and Readiness](Operations/Health-and-Readiness.md), and [Troubleshooting](Operations/Troubleshooting.md).

## Developer Quick Start

1. Clone the repository.
2. Run backend tests with `GOCACHE=/tmp/koji-go-cache go test ./...`.
3. Run frontend tests with `npm run test` from `web/`.
4. Build the frontend with `npm run build` from `web/`.
5. Use `make release` and `make verify-release` for release artifacts.

Read [Local Development](Developer/Local-Development.md), [Repository Layout](Developer/Repository-Layout.md), [Testing Strategy](Developer/Testing-Strategy.md), and [Release Workflow](Developer/Release-Workflow.md).

## Architecture Overview

Koji is split into:

- `kojid`: web/API daemon, authentication, capabilities, audit, jobs, worker, and read-only host observation.
- `koji-agent`: local privileged boundary for future service mutation.
- Jobs: durable service-control intent that survives daemon restart.
- Approvals: human authorization before a queued job can advance.
- Audit: durable record of sensitive intent and governance events.
- Observability: governed control-plane counters that answer whether Koji itself is healthy.

Primary architecture pages:

- [Overview](Architecture/Overview.md)
- [Trust Boundaries](Architecture/Trust-Boundaries.md)
- [Request Flow](Architecture/Request-Flow.md)
- [Data Flow](Architecture/Data-Flow.md)
- [Job Lifecycle](Architecture/Job-Lifecycle.md)
- [Agent Architecture](Architecture/Agent-Architecture.md)
- [Packaging and Deployment](Architecture/Packaging-and-Deployment.md)
- [Release Architecture](Architecture/Release-Architecture.md)

## Common Links

- [Installation](Operations/Installation.md)
- [Troubleshooting](Operations/Troubleshooting.md)
- [Configuration](Operations/Configuration.md)
- [Jobs Page](User-Guide/Jobs-Page.md)
- [Release Operations](Operations/Release-Operations.md)

## Security

- [Authentication](Security/Authentication.md)
- [Sessions](Security/Sessions.md)
- [CSRF](Security/CSRF.md)
- [Capabilities](Security/Capabilities.md)
- [Audit](Security/Audit.md)
- [Service Allowlists](Security/Service-Allowlists.md)
- [Agent Mutation Controls](Security/Agent-Mutation-Controls.md)
- [Threat Model](Security/Threat-Model.md)

## Operations

- [Installation](Operations/Installation.md)
- [Configuration](Operations/Configuration.md)
- [Health and Readiness](Operations/Health-and-Readiness.md)
- [Observability](Operations/Observability.md)
- [Backup and Recovery](Operations/Backup-and-Recovery.md)
- [Disaster Recovery](Operations/Disaster-Recovery.md)
- [Release Rollback](Operations/Release-Rollback.md)
- [Release Operations](Operations/Release-Operations.md)
- [Troubleshooting](Operations/Troubleshooting.md)

## User Guide

- [Overview Page](User-Guide/Overview-Page.md)
- [Services Page](User-Guide/Services-Page.md)
- [Processes Page](User-Guide/Processes-Page.md)
- [Jobs Page](User-Guide/Jobs-Page.md)
- [Activity Page](User-Guide/Activity-Page.md)
- [Settings Page](User-Guide/Settings-Page.md)

## Developer

- [Local Development](Developer/Local-Development.md)
- [Repository Layout](Developer/Repository-Layout.md)
- [Frontend Architecture](Developer/Frontend-Architecture.md)
- [Backend Architecture](Developer/Backend-Architecture.md)
- [Agent RPC](Developer/Agent-RPC.md)
- [Job System](Developer/Job-System.md)
- [Database Schema](Developer/Database-Schema.md)
- [Testing Strategy](Developer/Testing-Strategy.md)
- [Release Workflow](Developer/Release-Workflow.md)

## Reference

- [Configuration Reference](Reference/Configuration-Reference.md)
- [API Reference](Reference/API-Reference.md)
- [Capability Reference](Reference/Capability-Reference.md)
- [Audit Event Reference](Reference/Audit-Event-Reference.md)
- [Job State Reference](Reference/Job-State-Reference.md)
- [Error Code Reference](Reference/Error-Code-Reference.md)
- [Metrics Reference](Reference/Metrics-Reference.md)
