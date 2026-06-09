# Phase 24: Agent-Side Service Mutation Guardrails

## Goal

Prepare the privileged agent execution path with strict guardrails while keeping service mutation disabled by default.

## Scope

- Add agent-specific config validation.
- Require explicit `agent_mutation_enabled` before service mutation can proceed.
- Require an agent-side service allowlist when mutation is enabled.
- Validate service names and service actions inside the agent.
- Route service-control RPC through a guarded executor.
- Use `internal/platform/command` for the future mutation command runner.
- Return normalized agent reason codes only.

## Boundaries

- `kojid` still never executes `systemctl`.
- `koji-agent` owns the future mutation point.
- Disabled mutation returns `mutation_disabled`.
- Non-allowlisted services return `service_not_allowlisted`.
- Unsupported actions return `unsupported_action`.
- Command failures return controlled command reason codes.

## Result

Koji now has the final agent-side safety envelope before any real `start`, `stop`, or `restart` behavior is enabled.
