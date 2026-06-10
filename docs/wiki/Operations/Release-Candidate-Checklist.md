# Release Candidate Checklist

[Home](../Home.md) | Related: [Release Operations](Release-Operations.md), [Testing Strategy](../Developer/Testing-Strategy.md), [Upgrade Procedure](Upgrade-Procedure.md), [Backup and Recovery](Backup-and-Recovery.md)

## What Problem This Solves

The release candidate checklist gives operators and maintainers one repeatable gate before tagging or distributing a Koji release.

## Required Checks

| Check | Command Or Evidence | Pass Criteria |
| --- | --- | --- |
| Backend tests | `GOCACHE=/tmp/koji-go-cache go test ./...` | All Go packages pass. |
| Frontend tests | `npm run test` from `web/` | Vitest suite passes. |
| Frontend build | `npm run build` from `web/` | Top-level `dist/` is regenerated successfully. |
| OpenAPI validation | `make verify-openapi` | Routes, OpenAPI, and generated references match. |
| Documentation validation | `packaging/scripts/verify_docs.sh` | Required wiki pages exist, links resolve, stale terms are absent. |
| Code quality guard | `packaging/scripts/verify_code_quality.sh` | No unsafe frontend typing, empty catches, command-boundary drift, or required-doc gaps. |
| Backup verification | `make backup` and `make verify-restore` or equivalent staging commands. | Backup archive is created and restore verification passes. |
| Upgrade verification | `make pre-upgrade-check` and `make verify-upgrade` against a staging DB. | Compatibility and post-upgrade checks pass. |
| Release workflow | Tagged GitHub Actions release workflow. | Build, checksum, artifact, smoke, and publish jobs pass. |
| Downloaded artifacts | `shasum -a 256 -c SHA256SUMS.txt` on downloaded release assets. | Checksums match downloaded files. |

## Operator Workflow Checks

Before release, verify the UI supports these workflows with clear messages:

- Bootstrap Super Admin.
- Create a managed user.
- Issue a magic token and sign in with it.
- Grant and revoke a capability.
- Disable and re-enable a user.
- Create a service-control job.
- Approve and reject queued jobs.
- Understand agent unavailable, mutation disabled, failed, and not-implemented job states.
- Confirm Overview readiness, worker, agent, audit, auth, and job counters are understandable.

## Release Decision

Do not promote a release candidate when any required check fails. Fix the failure, rerun the checklist, then tag or publish the next candidate.
