# Error Code Reference

[Home](../Home.md) | Related: [Troubleshooting](../Operations/Troubleshooting.md), [API Reference](API-Reference.md)

## Frontend Normalized Errors

| Code | Operator Meaning |
| --- | --- |
| `unauthenticated` | Sign in before using the view |
| `forbidden` | The account lacks permission |
| `csrf_missing_or_invalid` | Refresh or sign in again |
| `agent_unavailable` | The local agent cannot be reached |
| `agent_not_implemented` | Service control is not enabled in this build path |
| `mutation_disabled` | Agent mutation is disabled by configuration |
| `service_not_allowlisted` | The service is not in Koji's allowlist |
| `job_conflict` | The job is no longer queued |
| `validation_error` | One or more request fields were invalid |
| `network_error` | Browser cannot reach Koji |
| `unexpected_response` | Response shape was not expected |
| `session_expired` | Sign in again before continuing |
| `self_lockout_prevented` | Koji blocked removing the final identity administrator |

## Protected Internals

SQL errors, Go stack traces, command stderr, and platform command details must not be shown to operators.
