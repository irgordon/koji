# Agent RPC

[Home](../Home.md) | Related: [Agent Architecture](../Architecture/Agent-Architecture.md), [Agent Mutation Controls](../Security/Agent-Mutation-Controls.md)

## What Problem This Solves

Agent RPC gives Koji a local process boundary for privileged operations without using TCP or direct HTTP command execution.

## How It Works

`kojid` connects to a Unix socket and sends service-control requests. The agent validates the request and returns normalized result codes.

## What Protects It

Socket path validation, stale socket checks, request validation, allowlists, mutation-disabled default, timeouts, and bounded output protect the path.

## What Can Fail

Missing socket, refused connection, timeout, malformed response, non-allowlisted service, unsupported action, or mutation disabled.

## How To Diagnose It

Use job status reason, agent RPC metrics, and `/readyz`.

## How To Recover

Fix socket path and permissions, start `koji-agent`, and keep mutation disabled until intentionally enabled.
