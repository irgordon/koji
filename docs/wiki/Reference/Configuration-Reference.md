# Configuration Reference

[Home](../Home.md) | Related: [Configuration](../Operations/Configuration.md), [Service Allowlists](../Security/Service-Allowlists.md)

## Daemon Fields

| Field | Default | Purpose |
| --- | --- | --- |
| `port` | `8443` | Daemon listen port |
| `dev_mode` | `false` | Explicit local development bypass behavior |
| `database_path` | `/var/lib/koji/koji.db` | SQLite DB path |
| `agent_socket_path` | `/run/koji/agent.sock` | Unix socket path |
| `static_asset_dir` | `/usr/share/koji/dist` | Production SPA assets |
| `dev_proxy_url` | `http://localhost:5173` | Dev frontend proxy |
| `session_ttl` | `12h` | Absolute session lifetime |
| `session_idle_ttl` | `30m` | Idle timeout |
| `service_allowlist` | none in production | Daemon service visibility and intent allowlist |
| `process_visibility_mode` | `summary` | `summary`, `owner`, or `all` |
| `include_command_line` | `false` | Whether process command lines are exposed |
| `max_processes` | `200` | Process response limit |

## Agent Fields

| Field | Default | Purpose |
| --- | --- | --- |
| `agent_socket_path` | `/run/koji/agent.sock` | Agent Unix socket |
| `agent_mutation_enabled` | `false` | Enables guarded mutation when true |
| `agent_service_allowlist` | required when mutation enabled | Agent mutation allowlist |
| `agent_command_timeout` | `3s` | Command timeout |
| `agent_command_output_limit` | `65536` | Output limit in bytes |

## Validation

Production paths must be absolute. Production daemon config requires an explicit service allowlist. Agent mutation requires an agent service allowlist.
