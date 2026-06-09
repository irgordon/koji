# Phase 25: Controlled Agent Mutation Enablement

## Goal

Enable real service mutation exclusively inside `koji-agent`, only when explicitly configured, and only through the existing guarded execution path.

## Scope

- Keep mutation disabled by default.
- Require agent-side mutation allowlists when enabled.
- Validate service name, action, and allowlist inside the agent.
- Run service mutation through `internal/platform/command`.
- Advance successful jobs to `completed`.
- Advance disabled, validation, allowlist, timeout, and command failures to `failed` with normalized reason codes.
- Audit completion and normalized command failure outcomes.

## Boundaries

- `kojid` does not execute `systemctl`.
- `internal/agent` does not call `exec.Command` or `exec.CommandContext`.
- No shell execution.
- No raw command output is returned through RPC or HTTP.
- Supported service actions remain `start`, `stop`, and `restart`.

## Result

Koji now has a real privileged service mutation path, but it is reachable only through the full governed chain: authenticated request, capability check, audit, durable job, human approval, worker, agent RPC, agent guardrails, and bounded platform command execution.
