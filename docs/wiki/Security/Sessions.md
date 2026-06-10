# Sessions

[Home](../Home.md) | Related: [Authentication](Authentication.md), [CSRF](CSRF.md), [Configuration Reference](../Reference/Configuration-Reference.md)

## What Problem This Solves

Sessions keep authenticated access bounded in time and revocable.

## How It Works

Sessions record `created_at`, `last_seen_at`, `expires_at`, and `revoked_at`. Valid requests refresh `last_seen_at`. Expired, idle-expired, or revoked sessions are rejected.

## What Protects It

Cookies are HttpOnly, SameSite Strict, path-scoped to `/`, and Secure in production.

## What Can Fail

Users can be signed out by absolute TTL, idle timeout, revocation, or cookie loss.

## How To Diagnose It

Use `/api/session`, safe error text, and Activity entries for logout or failed authenticated operations.

## How To Recover

Sign in again. If failures repeat, verify system time, cookie settings, and configured session TTL values.
