# Services Page

[Home](../Home.md) | Related: [Service Allowlists](../Security/Service-Allowlists.md), [Jobs Page](Jobs-Page.md), [Agent Mutation Controls](../Security/Agent-Mutation-Controls.md)

## What Problem This Solves

The Services page limits operators to explicitly allowlisted systemd units.

## How It Works

Koji lists allowlisted services and lets an authorized user create a start, stop, or restart job. Buttons create durable intent; they do not execute directly.

## What Protects It

Service reads require `host.services.read`. Service-control intent requires `host.services.control`, CSRF, audit, allowlist checks, and job creation.

## What Can Fail

A service can be hidden by allowlist policy, control can be denied, or the resulting job can fail at the agent boundary.

## How To Diagnose It

Use the page notice, Jobs page, and Activity page.

## How To Recover

Add the service to the correct allowlist or request the required capability. Approve jobs only after understanding impact.
