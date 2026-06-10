# CSRF

[Home](../Home.md) | Related: [Sessions](Sessions.md), [Request Flow](../Architecture/Request-Flow.md), [Error Code Reference](../Reference/Error-Code-Reference.md)

## What Problem This Solves

CSRF protection prevents another site from causing authenticated state changes through a user's browser.

## How It Works

State-changing authenticated requests must send Koji's CSRF token. Login and bootstrap issue the token with the session response.

## What Protects It

CSRF validation is enforced before protected POST handlers can mutate state such as logout, service-control job creation, or job approval.

## What Can Fail

The token can be missing, stale, or mismatched after a session refresh.

## How To Diagnose It

Look for safe `CSRF token required` responses and frontend `csrf_missing_or_invalid` messages.

## How To Recover

Refresh the UI or sign in again to get a current session and CSRF token.
