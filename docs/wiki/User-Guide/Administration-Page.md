# Administration Page

[Home](../Home.md) | Related: [User Administration](../Operations/User-Administration.md), [Capabilities](../Security/Capabilities.md), [Magic Tokens](../Security/Magic-Tokens.md)

## What Problem This Solves

The Administration page gives identity administrators a UI for controlled user and capability management.

## How It Works

The page lists users, shows whether each account is enabled, marks Super Admin accounts, and exposes controls for user creation, enable/disable, capability grant/revoke, and magic token issue.

Magic tokens are displayed once after issuance. Copy them immediately through the approved operational channel.

## What Protects It

The page is only a client surface. The server still requires authentication, CSRF protection, and `identity.users.manage` before changing users, capabilities, or tokens.

## What Can Fail

An action may be denied because the session expired, the account lacks permission, the CSRF token is invalid, the target user is missing, or Koji blocked a self-lockout risk.

## How To Diagnose It

Read the inline error or toast, then check Activity for matching `identity.*`, `auth.magic_token_*`, or `capability.denied` events.

## How To Recover

Refresh the page after a successful change. For expired sessions, sign in again. For permission denial, ask an existing identity administrator to grant the minimal required capability.
