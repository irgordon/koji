# Agent Mutation Controls

[Home](../Home.md) | Related: [Agent Architecture](../Architecture/Agent-Architecture.md), [Service Allowlists](Service-Allowlists.md), [Troubleshooting](../Operations/Troubleshooting.md)

## What Problem This Solves

Mutation controls ensure privileged service actions are explicit, bounded, and agent-owned.

## How It Works

The agent rejects mutation unless `agent_mutation_enabled` is true. When enabled, the agent validates action, service name, agent allowlist, command timeout, and output limit before using the platform command runner.

## What Protects It

`kojid` never executes `systemctl`. The agent uses normalized response codes such as `mutation_disabled`, `service_not_allowlisted`, `unsupported_action`, `command_failed`, and `command_timeout`.

## What Can Fail

Mutation can be disabled, non-allowlisted, timed out, or rejected by the command runner.

## How To Diagnose It

Use job status reason, Activity events, agent RPC metrics, and systemd status for `koji-agent`.

## How To Recover

Keep mutation disabled unless intentionally deploying it. When enabled, verify allowlists, command timeout, and agent permissions.
