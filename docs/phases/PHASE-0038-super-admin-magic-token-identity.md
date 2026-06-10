# Phase 38: Super Admin Bootstrap and Magic Token Identity Management

## Objective

Add governed user administration after bootstrap while keeping password login limited to the first Super Admin account.

## What Changed

- Bootstrap now creates a Super Admin identity.
- Added `identity.users.manage`.
- Added passwordless managed users.
- Added one-time magic token login.
- Added identity administration APIs and UI.
- Added self-lockout protection for the final Super Admin and final identity manager.
- Added audit events for identity changes and magic-token lifecycle events.

## Security Notes

Non-Super Admin users cannot use password login. Magic tokens are stored as hashes, expire by `magic_token_ttl`, and are consumed once.

No service mutation, agent privilege, job approval, audit write, CSRF, or capability enforcement behavior was weakened.

## Validation

- `npm run build`
- `GOCACHE=/tmp/koji-go-cache go test ./...`
- `make verify-openapi`
- `packaging/scripts/verify_docs.sh`
