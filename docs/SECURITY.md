# SECURITY.md

## 1. Security Posture

Koji is control-plane software. It runs on servers and can request sensitive system actions.

The security posture is conservative:

- Minimize privilege.
- Isolate privileged work.
- Require explicit authorization.
- Record audit events.
- Fail closed.
- Keep the codebase small enough to review.

## 2. Trust Boundaries

```text
Browser -> kojid -> koji-agent -> Operating system
```

The browser is untrusted.

`kojid` is unprivileged.

`koji-agent` is privileged and narrow.

The operating system is the final execution layer.

## 3. Authentication

Browser sessions must use secure cookies.

CSRF protection is required for browser-originated mutations.

Session cookies use `SameSite=Strict` because Koji is an administrative control plane and does not need cross-site browser workflows. Development mode explicitly omits the `Secure` cookie flag for local HTTP testing; production keeps it enabled.

Sessions must have both an absolute TTL and an idle timeout. Revoked, expired, and idle-expired sessions are invalid.

Bootstrap login is disabled by default.

A first-boot bootstrap token, if generated, must be:

- One-time use.
- Stored outside the web root.
- Permission-restricted.
- Removed or invalidated after use.

## 4. Authorization

Every privileged action requires server-side authorization.

The UI may hide controls but must not be treated as the enforcement point.

Authorization must be checked before job creation or agent execution.

## 5. Capability Enforcement

Every privileged action requires a capability check.

Allowed action is:

```text
compile-time capability ∩ runtime policy ∩ user authorization
```

If any layer denies the action, the action is denied.

No handler may bypass the capability layer.

## 6. Agent Security

The agent must communicate over a local Unix domain socket by default.

Socket permissions must restrict access to the daemon and intended service accounts.

The agent must expose explicit methods only.

Forbidden:

```text
run arbitrary command
execute shell string
write arbitrary file
read arbitrary file without policy
accept unsanitized user input
```

## 7. External Commands

External command use must be rare and isolated to system adapter packages.

Requirements:

- Use `exec.CommandContext` or equivalent.
- Use argument arrays.
- Use the centralized platform command runner.
- Enforce executable allowlists.
- Validate all user-controlled inputs.
- Use timeouts.
- Bound output size.
- Return structured errors.

Forbidden:

```go
exec.Command("sh", "-c", userInput)
```

## 8. Web Security

The unauthenticated login surface should be minimal.

The authenticated SPA should only be served after authentication.

Unauthenticated routes should not expose:

- JavaScript chunks.
- Route names.
- API schemas.
- Initial application state.
- Internal model names.

Use strict CSP on unauthenticated routes.

Use appropriate cache headers:

- Login HTML: no-store.
- Authenticated SPA shell: no-store.
- Static JS/CSS assets: cacheable with ETags or content hashes.
- API JSON responses: no-store.

Production responses must include browser security headers:

- Content-Security-Policy.
- X-Content-Type-Options.
- Referrer-Policy.
- X-Frame-Options or frame-ancestors.
- Permissions-Policy.

## 9. Audit Requirements

Audit events are append-only.

Record privileged mutations and security-sensitive actions.

Audit records must include:

- Actor.
- Action.
- Target.
- Status.
- Timestamp.
- Error summary when applicable.

Do not store secrets in audit payloads.

## 10. Secrets and Sensitive Data

Secrets must not be logged.

Support bundles must redact secrets.

Config dumps must redact secrets.

API responses must not expose password hashes, session secrets, CSRF secrets, MFA secrets, or bootstrap tokens.

## 11. Database Security

SQLite files must live under `/var/lib/koji/` by default.

File permissions must prevent non-service users from reading security state.

Migrations are embedded and checksummed.

Checksum mismatch is a security-relevant failure.

## 12. Reporting Security Issues

Until a public process exists, report security issues through the private project maintainer channel.

Include:

- Affected version.
- Reproduction steps.
- Expected result.
- Actual result.
- Logs or support bundle if safe to share.

Do not include secrets in reports.
