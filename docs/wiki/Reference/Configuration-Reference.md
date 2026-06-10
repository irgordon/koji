# Configuration Reference

[Home](../Home.md) | Related: [Configuration](../Operations/Configuration.md), [Service Allowlists](../Security/Service-Allowlists.md), [Upgrade Procedure](../Operations/Upgrade-Procedure.md)

This reference matches `internal/config.Config` and the strict parser in `internal/config/config.go`.

## Server

| Field | Default | Required | Validation Rules | Security Impact |
| --- | --- | --- | --- | --- |
| `port` | `8443` | Yes | Must be 1 through 65535. | Controls daemon listen port. |
| `dev_mode` | `false` | No | Boolean. | Enables explicit dev behavior, including dev service allowlist defaults and relaxed dev static flow. |
| `static_asset_dir` | `/usr/share/koji/dist` | Yes in production | Must be absolute unless dev mode is enabled. | Production SPA is served from this path after auth. |
| `dev_proxy_url` | `http://localhost:5173` | No | Must be absolute URL when set. | Dev-only frontend proxy; should not be used as production asset serving. |

## Database

| Field | Default | Required | Validation Rules | Security Impact |
| --- | --- | --- | --- | --- |
| `database_path` | `/var/lib/koji/koji.db` | Yes | Must be absolute. | Holds users, sessions, capabilities, audit, jobs, approvals, and migrations. |

## Sessions

| Field | Default | Required | Validation Rules | Security Impact |
| --- | --- | --- | --- | --- |
| `session_ttl` | `12h` | Yes | Positive duration. | Absolute session lifetime. |
| `session_idle_ttl` | `30m` | Yes | Positive duration, must not exceed `session_ttl`. | Limits stale authenticated sessions. |
| `session_idle_timeout` | alias for `session_idle_ttl` | No | Same as `session_idle_ttl`. | Backward-compatible spelling for idle timeout. |
| `magic_token_ttl` | `15m` | Yes | Positive duration no greater than `1h`. | Bounds passwordless login token lifetime. |

## Services

| Field | Default | Required | Validation Rules | Security Impact |
| --- | --- | --- | --- | --- |
| `service_allowlist` | none in production; `ssh.service`, `kojid.service` in dev mode when omitted | Required in production | Each value must pass service-name validation. Supports comma list or YAML-style list. | Prevents arbitrary systemd status/control intent over user input. |

## Processes

| Field | Default | Required | Validation Rules | Security Impact |
| --- | --- | --- | --- | --- |
| `process_visibility_mode` | `summary` | Yes | One of `summary`, `owner`, or `all`. | Controls how much process metadata is exposed. |
| `include_command_line` | `false` | No | Boolean. | Full command lines remain hidden unless explicitly enabled. |
| `max_processes` | `200` | Yes | 1 through 1000. | Bounds process response size. |

## Agent

| Field | Default | Required | Validation Rules | Security Impact |
| --- | --- | --- | --- | --- |
| `agent_socket_path` | `/run/koji/agent.sock` | Yes | Must be absolute. | Defines daemon-to-agent Unix socket boundary. |
| `agent_mutation_enabled` | `false` | No | Boolean. | Service mutation remains disabled unless explicitly enabled in agent config. |
| `agent_service_allowlist` | empty | Required only when mutation is enabled | Each value must pass service-name validation. | Agent independently enforces mutation allowlist. |
| `agent_command_timeout` | `3s` | Yes for agent | Positive duration. | Bounds privileged command execution time. |
| `agent_command_output_limit` | `65536` | Yes for agent | 1 through 1048576 bytes. | Bounds privileged command output capture. |

## Observability

There are no runtime observability configuration fields yet. Control-plane metrics are capability-protected by `observability.metrics.read`.

## Packaging

Packaging paths are controlled by `Makefile` variables at build/install time, not daemon runtime config:

| Variable | Default | Purpose |
| --- | --- | --- |
| `PREFIX` | `/usr` | Binary and shared asset prefix. |
| `SYSCONFDIR` | `/etc` | Configuration root. |
| `LOCALSTATEDIR` | `/var/lib` | Database/state root. |
| `SYSTEMDUNITDIR` | `$(PREFIX)/lib/systemd/system` | Unit installation directory. |
| `DESTDIR` | `build/rootfs` | Staging root for package assembly. |

## Parser Behavior

- Unknown fields are rejected.
- Duplicate fields are rejected.
- List fields support comma-separated values or YAML-style `- value` entries.
- Production daemon config requires an explicit service allowlist.
- Agent mutation requires an explicit agent service allowlist.
