# General Availability Decision: v0.2.0

## Summary

| Field | Value |
| --- | --- |
| Date | 2026-06-10 |
| Decision | Approved for `v0.2.0` |
| Decision Basis | RC.2 resolved the RC.1 operator-smoke gap and passed all release gates. |
| Tag | `v0.2.0` |
| Commit | `ee0dde054e31e7eee1680f24e1ccfec4a3142d08` |
| Release URL | https://github.com/irgordon/koji/releases/tag/v0.2.0 |
| Workflow URL | https://github.com/irgordon/koji/actions/runs/27272287368 |

## Release Candidates Reviewed

| Candidate | Decision | Evidence |
| --- | --- | --- |
| `v0.2.0-rc.1` | Additional RC Required | Build, artifact smoke, publish, checksums, install layout, tests, docs, OpenAPI, backup, restore, and upgrade checks passed. Live Linux operator workflow validation was missing. |
| `v0.2.0-rc.2` | Ready for Production | Build, artifact smoke, Docker-based Linux operator smoke, publish sequencing, published assets, checksums, and install layout passed. |

## Validation Matrix

| Area | Result | Evidence |
| --- | --- | --- |
| Backend Tests | PASS | `GOCACHE=/tmp/koji-go-cache go test ./...` |
| Frontend Tests | PASS | `npm run test` from `web/` |
| Frontend Build | PASS | `npm run build` from `web/` |
| OpenAPI | PASS | `make verify-openapi` |
| Documentation | PASS | `packaging/scripts/verify_docs.sh` |
| Code Quality | PASS | `packaging/scripts/verify_code_quality.sh` |
| Backup/Restore | PASS | Isolated `make -s backup`, `make -s restore`, and `make -s verify-restore` |
| Upgrade Validation | PASS | Isolated `make -s pre-upgrade-check` and `make -s verify-upgrade` |
| RC.2 Release Workflow | PASS | https://github.com/irgordon/koji/actions/runs/27254987692 |
| RC.2 Artifact Smoke | PASS | `smoke_test_release` passed. |
| RC.2 Operator Smoke | PASS | `operator_smoke_release` passed. |
| RC.2 Publish Gate | PASS | `publish_release` started after artifact and operator smoke gates passed. |
| GA Release Workflow | PASS | https://github.com/irgordon/koji/actions/runs/27272287368 |
| GA Artifact Smoke | PASS | `smoke_test_release` passed. |
| GA Operator Smoke | PASS | `operator_smoke_release` passed. |
| GA Publish Gate | PASS | `publish_release` started after artifact and operator smoke gates passed. |
| GA Release Assets | PASS | Downloaded `v0.2.0` assets include both binaries, rootfs archive, and `SHA256SUMS.txt`. |
| GA Checksums | PASS | `sha256sum -c SHA256SUMS.txt` returned `OK` for all downloaded GA artifacts. |
| GA Install Layout | PASS | Extracted rootfs contains binaries, frontend assets, config directory, systemd unit directory, and data directory. |

## GA Workflow Job Order

| Job | Started | Completed | Result |
| --- | --- | --- | --- |
| Build release artifacts | 2026-06-10T11:13:39Z | 2026-06-10T11:15:51Z | PASS |
| Smoke-test release artifacts | 2026-06-10T11:15:54Z | 2026-06-10T11:15:59Z | PASS |
| Smoke-test operator workflows | 2026-06-10T11:16:02Z | 2026-06-10T11:16:25Z | PASS |
| Publish release assets | 2026-06-10T11:16:27Z | 2026-06-10T11:16:35Z | PASS |

## GA Asset Verification

Downloaded from the GitHub Release:

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

## Operator Smoke Evidence

RC.2 verified these workflows through downloaded Linux release artifacts running in an Ubuntu Docker smoke environment:

- Bootstrap Super Admin.
- Create managed user.
- Issue magic token.
- Magic token login.
- Capability grant and revoke.
- Disable and enable user.
- Create service-control job.
- Approve job.
- Reject job.
- Activity read model.
- Observability metrics.
- Disabled-user protected endpoint rejection.
- Mutation-disabled service-control outcome through the agent boundary.

## Findings

### P0

None.

### P1

None remaining.

Resolved:

| ID | Finding | Resolution |
| --- | --- | --- |
| RC43-001 | RC.1 lacked a live Linux operator workflow gate. | RC.2 added and passed `operator_smoke_release`. |

### P2

Deferred roadmap items from the production-readiness review remain future enhancements:

- Backup/restore UI status.
- Guided first-run setup.
- Per-job timeline.
- Capability presets.
- Richer agent recovery hints.

These are not production blockers for `v0.2.0`.

## Final Decision

Approved for `v0.2.0`.

Koji passed the required local validation gates, the RC.1 P1 release gap was resolved by RC.2, the GA release workflow passed, and downloaded GA artifacts verified successfully. No P0 or unresolved P1 issue remains.
