# Magic Tokens

[Home](../Home.md) | Related: [Authentication](Authentication.md), [Sessions](Sessions.md), [User Administration](../Operations/User-Administration.md)

## What Problem This Solves

Magic tokens let managed users sign in without storing or rotating user passwords after bootstrap.

## How It Works

The first bootstrap user is the Super Admin and may use password login. Other users are passwordless managed users. An identity administrator issues a one-time magic token for a target user, Koji stores only the token hash, and the raw token is shown once.

`POST /api/login/magic-token` consumes an unexpired token and creates a normal session with a CSRF token. A consumed, expired, revoked, or disabled-user token cannot be reused.

## What Protects It

- Tokens are random and stored as hashes.
- Tokens have a bounded `magic_token_ttl`, defaulting to `15m`.
- Tokens are single-use.
- Disabled users cannot sign in with tokens.
- Token issue and consumption are audited.
- Password login is denied for non-Super Admin users.

## What Can Fail

The token may be expired, already consumed, revoked, typed incorrectly, or attached to a disabled user.

## How To Diagnose It

Check the UI error text and Activity entries for `auth.magic_token_failure`, `auth.magic_token_success`, `identity.magic_token_issued`, `identity.magic_token_consumed`, or `identity.magic_token_expired`.

## How To Recover

An identity administrator can issue a new token if the user should still have access. Do not reuse or store old token values.
