# General Availability Decision: v0.2.0

## Summary

| Field | Value |
| --- | --- |
| Date | 2026-06-10 |
| Decision | Approved for `v0.2.0` |
| Decision Basis | RC.2 resolved the RC.1 operator-smoke gap and passed all release gates. |
| Planned Tag | `v0.2.0` |
| Planned Commit | `4f7e81f971e2710327b37dbcf66c3011217321c7` |
| Planned Release URL | https://github.com/irgordon/koji/releases/tag/v0.2.0 |
| Workflow URL | Pending tag-triggered release workflow |

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

Koji has passed the required validation gates, the RC.1 P1 release gap was resolved by RC.2, and no P0 or unresolved P1 issue remains. The next step is to tag `v0.2.0`, verify the GA release workflow, verify downloaded release assets, and update this decision record with the final workflow result.
