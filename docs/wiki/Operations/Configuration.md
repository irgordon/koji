# Configuration

[Home](../Home.md) | Related: [Configuration Reference](../Reference/Configuration-Reference.md), [Service Allowlists](../Security/Service-Allowlists.md), [Sessions](../Security/Sessions.md), [Magic Tokens](../Security/Magic-Tokens.md)

## What Problem This Solves

Configuration makes runtime policy explicit: paths, sessions, magic-token lifetime, allowlists, process visibility, static assets, and agent mutation guardrails.

## How It Works

`kojid` loads `/etc/koji/koji.yaml`. `koji-agent` loads `/etc/koji/agent.yaml`. Startup fails when required values are invalid.

## What Protects It

Production requires absolute runtime paths and an explicit service allowlist. Agent mutation is disabled by default.

## What Can Fail

Invalid ports, relative paths, missing production allowlists, invalid service names, invalid magic-token TTL, or idle timeout greater than session TTL can block startup.

## How To Diagnose It

Read service logs and compare values with [Configuration Reference](../Reference/Configuration-Reference.md).

## How To Recover

Correct the invalid field, keep production paths absolute, and restart the affected service.
