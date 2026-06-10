# Service Allowlists

[Home](../Home.md) | Related: [Services Page](../User-Guide/Services-Page.md), [Configuration](../Operations/Configuration.md), [Agent Mutation Controls](Agent-Mutation-Controls.md)

## What Problem This Solves

Allowlists prevent Koji from becoming an arbitrary systemd inspection or mutation surface.

## How It Works

`kojid` uses `service_allowlist` for service status and service-control intent. `koji-agent` independently validates `agent_service_allowlist` before mutation can run.

## What Protects It

Service names are validated, production requires an explicit daemon allowlist, and agent mutation requires its own allowlist when enabled.

## What Can Fail

A service can be hidden or rejected because it is missing from the daemon or agent allowlist.

## How To Diagnose It

Check service API errors, Activity denied events, and configuration files.

## How To Recover

Add only the intended systemd unit to the correct allowlist and restart the affected service.
