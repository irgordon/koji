# Security Review

[Home](../Home.md) | Related: [Threat Model](Threat-Model.md), [Audit](Audit.md), [Agent Mutation Controls](Agent-Mutation-Controls.md)

## Purpose

This review validates whether the implementation still matches Koji's intended threat model after authentication, sessions, CSRF, capabilities, audit, jobs, approvals, agent RPC, magic tokens, observability, backups, and upgrade safety were added.

## Review Summary

Koji has a strong defense-in-depth posture for a local host control panel. The most important residual risks are not missing route checks; they are operational authority risks: who receives high-risk capabilities, how magic tokens are transmitted, who can read backups, whether filesystem ownership is correct, and whether agent mutation is enabled with an overbroad allowlist.

## Findings

| Threat | Impact | Likelihood | Existing Control | Residual Risk | Recommended Action |
| --- | --- | --- | --- | --- | --- |
| Capability escalation by an identity administrator | High | Medium | `identity.users.manage`, CSRF, audit, self-lockout prevention. | A legitimate identity admin can intentionally grant powerful capabilities. | Add operational review and alerting for high-risk grants. |
| Magic token leakage | High | Medium | Hash-only storage, one-time consumption, TTL, audit. | Raw token delivery channel is outside Koji. | Document approved delivery channels and keep TTL short. |
| Approval bypass by DB tampering | High | Low to Medium | Store-level state transitions and DB migration checks. | Local write access to SQLite bypasses application logic. | Add runtime DB ownership/mode checks. |
| Agent socket replacement/access | High | Low to Medium | Absolute path validation, parent directory checks, stale socket safety. | Socket file mode after bind depends on runtime environment. | Set and verify socket file permissions explicitly. |
| Overbroad service mutation allowlist | High | Low by default, Medium when enabled | Mutation disabled by default and agent allowlist required. | Operator can still configure unsafe scope. | Add mutation enablement checklist and warnings. |
| Audit write failure | Medium | Medium | Audit counters and readiness visibility. | Some high-risk actions may continue if audit persistence fails. | Decide whether selected actions should fail closed on audit write failure. |
| Backup theft | High | Medium | Structured backup/restore verification. | Archive confidentiality is external. | Add encrypted backup guidance or tooling. |
| Dev mode in production | Critical | Low | Explicit config and dev bypass markers. | Misconfiguration can defeat auth gate. | Add production startup refusal or warning conditions. |

## Assumption Break Tests

| Scenario | Expected Control | Residual Concern |
| --- | --- | --- |
| Attacker is authenticated with no capabilities. | Protected handlers deny per-surface access. | Session theft still matters until expiration/revocation. |
| Attacker has `jobs.read`. | Can view jobs but not approve or create service-control intent. | Job metadata visibility may reveal operations. |
| Attacker has `identity.users.manage`. | Actions are audited and final admin lockout is blocked. | Capability assignment is intentionally powerful. |
| Agent socket is replaced by a file. | Agent refuses to remove non-socket path. | Socket permissions still need deployment verification. |
| Backup is stolen. | No runtime control applies after theft. | Encryption/access control must be operational. |
| Magic token leaks. | Token expires and can be consumed once. | Attacker can create a session before legitimate user. |
| OpenAPI is public. | Server-side auth/capability/CSRF still enforce access. | Attackers can enumerate protected routes. |

## Recommended Follow-Up Queue

1. Add runtime permission checks for `/etc/koji`, `/var/lib/koji`, `/run/koji`, and the agent socket.
2. Add explicit socket file mode/ownership enforcement after agent bind.
3. Add encrypted backup guidance or optional encrypted backup output.
4. Add audit alerting guidance for high-risk identity and job events.
5. Decide whether high-risk identity/job mutations should fail closed when audit writes fail.
6. Add production startup guardrails for `dev_mode=true`.

## Current Conclusion

The implementation matches the intended layered threat model for a local governed control panel. Koji is not free of residual risk, but the remaining high-value security work is now mostly operational hardening, permission verification, and alerting rather than missing core auth or authorization controls.
