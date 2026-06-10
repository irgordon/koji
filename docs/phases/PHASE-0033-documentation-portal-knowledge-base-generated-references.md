# PHASE-0033: Documentation Portal, Knowledge Base, and Generated References

## Objective

Create a structured documentation portal that lets operators, administrators, developers, maintainers, and security reviewers understand Koji without reading source code.

## Scope

- Added `docs/wiki`.
- Added Architecture, Security, Operations, User Guide, Developer, and Reference sections.
- Added Mermaid diagrams for request flow, job lifecycle, agent boundary, release flow, data flow, and trust boundaries.
- Added semi-generated reference pages for capabilities, audit events, job states, metrics, API routes, configuration, and frontend error codes.
- Added required troubleshooting topics for agent, readiness, jobs, auth, session, CSRF, capabilities, mutation, release, and artifact smoke failures.
- Added `packaging/scripts/verify_docs.sh`.

## Protections

The portal explains auth, sessions, CSRF, capabilities, audit, jobs, allowlists, agent mutation controls, release validation, and observability without weakening runtime behavior.

## Validation

- `git diff --check`
- `npm run test`
- `npm run build`
- `GOCACHE=/tmp/koji-go-cache go test ./...`
- `packaging/scripts/verify_docs.sh`

## Known Limitations

- Reference pages are semi-generated from current source constants and routes, not produced by an automated generator.
- The portal is Markdown-only and is suitable for GitHub/wiki rendering; no static site generator is introduced in this phase.
