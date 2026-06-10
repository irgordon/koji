# Phase 43: Release Candidate Cut, Verification, and Production Readiness Gate

## Goal

Create the first formal Koji release candidate and validate production readiness with release, artifact, documentation, backup, restore, upgrade, and test evidence.

## Non-Goals

- No feature development.
- No new APIs.
- No new capabilities.
- No security redesign.
- No architecture refactor.
- No privileged behavior changes.

## Release Candidate

| Field | Value |
| --- | --- |
| Version | `v0.2.0-rc.1` |
| Commit | `c9bb83b018406c01960c87337b1f1ac859185301` |
| Release URL | https://github.com/irgordon/koji/releases/tag/v0.2.0-rc.1 |
| Workflow Run | https://github.com/irgordon/koji/actions/runs/27253807713 |

Follow-up RC after adding the operator smoke gate:

| Field | Value |
| --- | --- |
| Version | `v0.2.0-rc.2` |
| Commit | `79b9f3eeafdf51023cfc810cf2eb6c974a728618` |
| Release URL | https://github.com/irgordon/koji/releases/tag/v0.2.0-rc.2 |
| Workflow Run | https://github.com/irgordon/koji/actions/runs/27254987692 |

## Invariants Preserved

- `kojid` still does not execute `systemctl`.
- `koji-agent` remains the only future host mutation owner.
- Agent mutation remains disabled unless explicitly configured.
- Authentication, sessions, CSRF, capabilities, audit, jobs, approval, and service allowlists were not weakened.
- OpenAPI, docs, and release artifacts were validated without changing runtime behavior.

## Verification Summary

| Check | Result |
| --- | --- |
| Clean repository before tag | PASS |
| `git diff --check` before tag | PASS |
| GitHub release workflow | PASS |
| Release asset download from GitHub | PASS |
| `sha256sum -c SHA256SUMS.txt` | PASS |
| Rootfs install layout | PASS |
| `npm run test` | PASS |
| `npm run build` | PASS |
| `GOCACHE=/tmp/koji-go-cache go test ./...` | PASS |
| `packaging/scripts/verify_docs.sh` | PASS |
| `packaging/scripts/verify_code_quality.sh` | PASS |
| `make verify-openapi` | PASS |
| Isolated backup/restore verification | PASS |
| Isolated upgrade verification | PASS |
| Live Linux operator smoke from this workstation | NOT APPLICABLE |
| GitHub Actions Docker operator smoke gate | ADDED |
| RC.2 GitHub Actions Docker operator smoke gate | PASS |

## Findings

| ID | Severity | Finding | Outcome |
| --- | --- | --- | --- |
| RC43-001 | P1 | macOS cannot run the Linux host telemetry path because `/proc/stat` is unavailable, and Docker was not running for a Linux container smoke. | Added `operator_smoke_release`, a Docker-based GitHub Actions gate that runs the downloaded release artifacts in Ubuntu and exercises operator workflows before publish. |

## Decision

RC.1: Additional RC Required.

No P0 release blocker was found. The remaining P1 item was resolved for future tags by adding a Linux Docker operator smoke gate to the release workflow. Because `v0.2.0-rc.1` was cut before that gate existed, stable promotion required a follow-up RC.

RC.2: Ready for Production.

The follow-up RC proved the build, artifact smoke, operator smoke, and publish sequencing gates in GitHub Actions, then published downloadable assets that passed checksum and rootfs layout verification.

## Files Changed

- `docs/releases/RC-0.2.0-rc.1.md`
- `docs/releases/RC-0.2.0-rc.2.md`
- `docs/phases/PHASE-0043-release-candidate-verification.md`
- `docs/CHANGELOG.md`
- `.github/workflows/release.yml`
- `packaging/ci/operator-smoke.Dockerfile`
- `packaging/scripts/ci_operator_smoke.sh`

## Commands Run

```text
git status --short
git diff --check
git tag v0.2.0-rc.1
git push origin v0.2.0-rc.1
curl GitHub Actions API for release workflow status
curl GitHub Release API for release assets
sha256sum -c SHA256SUMS.txt
tar -xzf koji-rootfs-linux-amd64.tar.gz
npm run test
npm run build
GOCACHE=/tmp/koji-go-cache go test ./...
packaging/scripts/verify_docs.sh
packaging/scripts/verify_code_quality.sh
make verify-openapi
make -s backup
make -s restore
make -s verify-restore
make -s pre-upgrade-check
make -s verify-upgrade
docker context ls
```

## Summary

Phase 43 cut `v0.2.0-rc.1`, found the Linux operator smoke gap, added a Docker-based GitHub Actions operator gate, then cut `v0.2.0-rc.2`. RC.2 passed build, artifact smoke, operator smoke, publish, asset, checksum, and install-layout verification.
