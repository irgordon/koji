# Release Candidate: v0.2.0-rc.1

## Overview

| Field | Value |
| --- | --- |
| Date | 2026-06-10 |
| Tag | `v0.2.0-rc.1` |
| Commit | `c9bb83b018406c01960c87337b1f1ac859185301` |
| Release URL | https://github.com/irgordon/koji/releases/tag/v0.2.0-rc.1 |
| Workflow Run | https://github.com/irgordon/koji/actions/runs/27253807713 |

## Validation Results

| Area | Result | Evidence |
| --- | --- | --- |
| Repository state before tag | PASS | `git status --short` was clean and `git diff --check` passed before tagging. |
| Release tag | PASS | `v0.2.0-rc.1` was created and pushed to `origin`. |
| Release workflow | PASS | GitHub Actions run `27253807713` completed successfully. |
| `build_release` | PASS | Job "Build release artifacts" completed successfully. |
| `smoke_test_release` | PASS | Job "Smoke-test release artifacts" completed successfully. |
| `publish_release` | PASS | Job "Publish release assets" completed successfully. |
| Release assets | PASS | GitHub Release contains `kojid-linux-amd64`, `koji-agent-linux-amd64`, `koji-rootfs-linux-amd64.tar.gz`, and `SHA256SUMS.txt`. |
| Checksums | PASS | `sha256sum -c SHA256SUMS.txt` returned `OK` for all downloaded artifacts. |
| Installation layout | PASS | Extracted rootfs contains `usr/bin/kojid`, `usr/bin/koji-agent`, `usr/share/koji/dist`, `etc/koji`, `usr/lib/systemd/system`, and `var/lib/koji`. |
| Frontend tests | PASS | `npm run test` passed: 9 files, 44 tests. |
| Frontend build | PASS | `npm run build` completed successfully and emitted top-level `dist/`. |
| Backend tests | PASS | `GOCACHE=/tmp/koji-go-cache go test ./...` passed. |
| Documentation validation | PASS | `packaging/scripts/verify_docs.sh` passed. |
| Code quality validation | PASS | `packaging/scripts/verify_code_quality.sh` passed. |
| OpenAPI validation | PASS | `make verify-openapi` passed. |
| Backup | PASS | `make -s backup` succeeded against an isolated SQLite database and copied example configs. |
| Restore | PASS | `make -s restore` succeeded into an isolated restore path. |
| Restore verification | PASS | `make -s verify-restore` reported `restore verification passed`. |
| Pre-upgrade check | PASS | `make -s pre-upgrade-check` reported schema `0009_identity_magic_tokens` as current. |
| Upgrade verification | PASS | `make -s verify-upgrade` reported schema `0009_identity_magic_tokens` as readable and valid. |
| Live Linux operator smoke | PENDING | Added a GitHub Actions Docker-based operator workflow smoke so future RC tags verify Linux startup, identity, jobs, activity, observability, disabled-user revocation, and mutation-disabled job failure in CI. This first RC tag was cut before that gate existed. |

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

Extracted runtime layout:

```text
rootfs/usr/bin/kojid
rootfs/usr/bin/koji-agent
rootfs/usr/lib/systemd/system
rootfs/usr/share/koji/dist
rootfs/etc/koji
rootfs/var/lib/koji
```

## Operator Workflow Checklist

| Workflow | Result | Evidence |
| --- | --- | --- |
| Bootstrap Super Admin | PASS | Covered by backend tests and Phase 40 operator review. |
| Create managed user | PASS | Covered by backend tests and Phase 40 operator review. |
| Issue magic token | PASS | Covered by backend tests and Phase 40 operator review. |
| Magic token login | PASS | Covered by backend tests and Phase 40 operator review. |
| Disable user | PASS | Covered by backend tests and Phase 40 operator review. |
| Enable user | PASS | Covered by backend tests and Phase 40 operator review. |
| Grant capability | PASS | Covered by backend tests and Phase 40 operator review. |
| Revoke capability | PASS | Covered by backend tests and Phase 40 operator review. |
| View services | PASS | Covered by backend tests, frontend build, and service allowlist tests. |
| Create service-control job | PASS | Covered by backend tests and Phase 40 operator review. |
| Approve job | PASS | Covered by backend tests and Phase 40 operator review. |
| Reject job | PASS | Covered by backend tests and Phase 40 operator review. |
| Activity view | PASS | Covered by frontend tests and audit read-model tests. |
| Observability view | PASS | Covered by frontend tests and governed metrics endpoint tests. |
| Backup and recovery | PASS | Isolated backup, restore, and verify-restore commands passed. |
| Upgrade safety | PASS | Isolated pre-upgrade and post-upgrade verification commands passed. |

## Security Validation

| Control | Result | Evidence |
| --- | --- | --- |
| Magic token expiry | PASS | Covered by auth/identity tests and documented TTL bounds. |
| Disabled users cannot log in | PASS | Covered by identity and auth tests. |
| Capabilities enforced | PASS | Covered by handler tests for denied and allowed access. |
| Sessions revoked on disable | PASS | Covered by identity administration tests. |
| Agent allowlist enforced | PASS | Covered by agent guardrail tests. |
| Service mutation default | PASS | Agent mutation remains disabled unless explicitly enabled by agent config. |

## Findings

| ID | Severity | Description | Resolution |
| --- | --- | --- | --- |
| RC43-001 | P1 | Full live operator workflow smoke could not be executed from this macOS workstation because `kojid` depends on Linux `/proc` and Docker was not running. | Added a Docker-based GitHub Actions operator smoke gate for future release tags. Cut a follow-up RC to verify the gate before stable promotion. |

## Decision

Additional RC Required.

The release candidate build, release workflow, downloaded artifact checksums, rootfs layout, tests, documentation validation, OpenAPI validation, backup/restore, and upgrade checks passed. No P0 release blocker was found. A P1 verification gap remains for this already-published RC because the Linux operator smoke gate was added after `v0.2.0-rc.1` was cut. A follow-up RC should validate the new gate before stable promotion.
