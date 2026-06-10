# Release Candidate: v0.2.0-rc.2

## Overview

| Field | Value |
| --- | --- |
| Date | 2026-06-10 |
| Tag | `v0.2.0-rc.2` |
| Commit | `79b9f3eeafdf51023cfc810cf2eb6c974a728618` |
| Release URL | https://github.com/irgordon/koji/releases/tag/v0.2.0-rc.2 |
| Workflow Run | https://github.com/irgordon/koji/actions/runs/27254987692 |

## Validation Results

| Area | Result | Evidence |
| --- | --- | --- |
| Release workflow | PASS | GitHub Actions run `27254987692` completed successfully. |
| `build_release` | PASS | Job "Build release artifacts" completed successfully. |
| `smoke_test_release` | PASS | Job "Smoke-test release artifacts" completed successfully. |
| `operator_smoke_release` | PASS | Job "Smoke-test operator workflows" completed successfully. |
| `publish_release` sequencing | PASS | Publish started after both smoke gates completed successfully. |
| `publish_release` | PASS | Job "Publish release assets" completed successfully. |
| Release assets | PASS | GitHub Release contains `kojid-linux-amd64`, `koji-agent-linux-amd64`, `koji-rootfs-linux-amd64.tar.gz`, and `SHA256SUMS.txt`. |
| Checksums | PASS | `sha256sum -c SHA256SUMS.txt` returned `OK` for all downloaded artifacts. |
| Installation layout | PASS | Extracted rootfs contains executable `usr/bin/kojid`, executable `usr/bin/koji-agent`, frontend assets, config directory, systemd unit directory, and data directory. |

## Workflow Job Order

| Job | Started | Completed | Result |
| --- | --- | --- | --- |
| Build release artifacts | 2026-06-10T05:18:22Z | 2026-06-10T05:20:35Z | PASS |
| Smoke-test release artifacts | 2026-06-10T05:20:37Z | 2026-06-10T05:20:43Z | PASS |
| Smoke-test operator workflows | 2026-06-10T05:20:45Z | 2026-06-10T05:21:01Z | PASS |
| Publish release assets | 2026-06-10T05:21:03Z | 2026-06-10T05:21:13Z | PASS |

## Operator Smoke Coverage

The RC.2 release workflow ran the downloaded Linux release artifacts inside an Ubuntu Docker smoke image and verified:

- `koji-agent` starts and exposes its Unix socket.
- `kojid` starts on Linux and reports healthy readiness.
- Bootstrap Super Admin succeeds.
- Managed user creation succeeds.
- Capability grants and revocation succeed.
- Disabled users cannot receive new magic tokens.
- Magic token login succeeds.
- Service-control intent creates durable jobs.
- Approved service-control jobs reach safe `failed` status with `mutation_disabled` while agent mutation remains disabled.
- Rejected jobs reach `rejected`.
- Activity records are readable through the governed endpoint.
- Observability metrics show job and agent RPC counters.
- Disabled-user sessions are rejected on protected endpoints.

## Release Asset Verification

Downloaded from the GitHub Release, not local build output:

```text
koji-agent-linux-amd64
koji-rootfs-linux-amd64.tar.gz
kojid-linux-amd64
SHA256SUMS.txt
```

Checksum output:

```text
kojid-linux-amd64: OK
koji-agent-linux-amd64: OK
koji-rootfs-linux-amd64.tar.gz: OK
```

## Findings

| ID | Severity | Description | Resolution |
| --- | --- | --- | --- |
| RC43-001 | P1 | RC.1 lacked a live Linux operator workflow gate. | Resolved in RC.2 by adding and passing `operator_smoke_release`. |

## Decision

Ready for Production.

The release workflow now validates build, artifact smoke, Linux operator workflow smoke, publish sequencing, published assets, checksums, and install layout. No P0 or unresolved P1 release blocker remains from the RC.2 evidence.
