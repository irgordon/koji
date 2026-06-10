# Authentication

[Home](../Home.md) | Related: [Sessions](Sessions.md), [Magic Tokens](Magic-Tokens.md), [Capabilities](Capabilities.md), [Audit](Audit.md)

## What Problem This Solves

Authentication identifies the caller before Koji exposes protected host data or accepts operational intent.

## How It Works

Koji supports bootstrap only until the first user exists. The bootstrap user becomes the Super Admin and may use password login. Managed users sign in with one-time magic tokens. Successful login creates a server-backed session and CSRF token.

Public API access is limited to bootstrap, password login, magic-token login, logout, session status, health, and readiness.

## What Protects It

Password handling is server-side for the Super Admin account, magic tokens are stored as hashes, sessions are stored durably, and protected routes deny by default.

## What Can Fail

Bootstrap may be disabled after the first user, credentials may be wrong, a magic token may be invalid or expired, non-Super Admin password login may be denied, or session state may expire.

## How To Diagnose It

Check login error text, Activity entries for `auth.login`, `auth.bootstrap`, `auth.magic_token_failure`, or `auth.password_denied_non_super_admin`, and `auth_login_failure_total` metrics.

## How To Recover

Use an identity administrator to issue a fresh magic token, restore the DB from backup if all identity administration is lost, or sign in again after session expiration.
