# Authentication

[Home](../Home.md) | Related: [Sessions](Sessions.md), [Capabilities](Capabilities.md), [Audit](Audit.md)

## What Problem This Solves

Authentication identifies the caller before Koji exposes protected host data or accepts operational intent.

## How It Works

Koji supports bootstrap only until the first user exists. Login creates a server-backed session and CSRF token. Public API access is limited to bootstrap, login, logout, session status, health, and readiness.

## What Protects It

Password handling is server-side, sessions are stored durably, and protected routes deny by default.

## What Can Fail

Bootstrap may be disabled after the first user, credentials may be wrong, or session state may expire.

## How To Diagnose It

Check login error text, Activity entries for `auth.login` or `auth.bootstrap`, and `auth_login_failure_total` metrics.

## How To Recover

Use the existing admin path for credentials, restore the DB from backup if the first user state is lost, or sign in again after session expiration.
