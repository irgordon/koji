# User Administration

[Home](../Home.md) | Related: [Capabilities](../Security/Capabilities.md), [Magic Tokens](../Security/Magic-Tokens.md), [Administration Page](../User-Guide/Administration-Page.md)

## What Problem This Solves

Koji needs a governed way to create managed users, assign capabilities, disable access, and issue passwordless sign-in tokens after bootstrap.

## How It Works

The first bootstrap account becomes the Super Admin and receives identity administration authority. Users with `identity.users.manage` can:

- list users;
- create passwordless managed users;
- enable or disable users;
- list and change user capabilities;
- issue one-time magic tokens.

Administrative changes are durable and audited. Mutating requests require a valid session, CSRF token, and `identity.users.manage`.

## Self-Lockout Protection

Koji prevents disabling the final active Super Admin and prevents removing the final active `identity.users.manage` administrator. These blocked changes are audited as `identity.self_lockout_prevented`.

## What Can Fail

Administration can fail because the operator lacks `identity.users.manage`, the CSRF token is stale, the target user does not exist, a capability name is invalid, or the requested change would lock out the final identity administrator.

## How To Diagnose It

Use the Administration page error text, Activity entries, and the [Capability Reference](../Reference/Capability-Reference.md).

## How To Recover

Sign in with an identity administrator account, issue a fresh magic token for the intended user, or restore from a verified backup if all administrative access is lost.
