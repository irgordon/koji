# Troubleshooting

[Home](../Home.md) | Related: [Health and Readiness](Health-and-Readiness.md), [Observability](Observability.md), [Error Code Reference](../Reference/Error-Code-Reference.md)

## Agent Unavailable

Symptoms: `/readyz` is degraded, jobs fail with agent unavailable, agent RPC failures increase.

Likely causes: `koji-agent` is stopped, socket path is wrong, or `/run/koji` ownership is wrong.

Verification steps: check `/readyz`, `systemctl status koji-agent`, and `agent_rpc_failures_total`.

Recovery steps: start `koji-agent`, correct socket path and runtime directory ownership, then retry by creating a new job.

## Readiness Degraded

Symptoms: `/readyz` returns degraded.

Likely causes: agent socket unreachable while DB and migrations are healthy.

Verification steps: inspect `/readyz` checks and agent service status.

Recovery steps: start or repair `koji-agent`.

## Jobs Stuck Queued

Symptoms: jobs remain `queued`.

Likely causes: no user with `jobs.approve` has approved or rejected them.

Verification steps: inspect Jobs page and Activity for missing `job.approved` or `job.rejected`.

Recovery steps: approve or reject the job with a reason.

## Jobs Stuck Approved

Symptoms: jobs remain `approved`.

Likely causes: worker is not polling or `kojid` is stopped.

Verification steps: check `worker_polls_total`, `systemctl status kojid`, and service logs.

Recovery steps: restart `kojid` and verify the worker claims approved jobs.

## Job Failed

Symptoms: job status is `failed`.

Likely causes: agent unavailable, mutation disabled, validation failure, command failure, or timeout.

Verification steps: inspect job `status_reason`, Activity events, and agent metrics.

Recovery steps: fix the underlying reason and create a new job.

## Login Failures

Symptoms: login returns safe failure text.

Likely causes: wrong credentials or bootstrap already completed.

Verification steps: check Activity for `auth.login` failure and auth metrics.

Recovery steps: use valid credentials or restore the admin data path.

## Session Expiration

Symptoms: UI asks the user to sign in again.

Likely causes: idle timeout, absolute TTL, revoked session, or cookie loss.

Verification steps: call `/api/session` and inspect frontend message.

Recovery steps: sign in again.

## CSRF Failures

Symptoms: state-changing request returns CSRF error.

Likely causes: stale token after session change or missing header.

Verification steps: inspect frontend error and request headers.

Recovery steps: refresh and sign in again if needed.

## Missing Capability

Symptoms: view or action returns permission denied.

Likely causes: authenticated user lacks the required capability.

Verification steps: inspect Activity for `capability.denied` and compare [Capability Reference](../Reference/Capability-Reference.md).

Recovery steps: grant the minimal required capability.

## Mutation Disabled

Symptoms: job fails with `mutation_disabled`.

Likely causes: `agent_mutation_enabled` remains false.

Verification steps: inspect agent config and job status reason.

Recovery steps: keep disabled unless intentionally deploying mutation; if enabling, set agent allowlist and review guardrails.

## Release Workflow Failure

Symptoms: tag workflow fails before publishing.

Likely causes: tests, build, artifact assembly, or smoke checks failed.

Verification steps: inspect the failed GitHub Actions job.

Recovery steps: fix the failing gate and create a new tag.

## Artifact Smoke Test Failure

Symptoms: smoke job fails after artifacts are built.

Likely causes: checksum mismatch, missing executable bits, invalid rootfs layout, systemd path drift, or forbidden local paths.

Verification steps: run `packaging/scripts/ci_verify_release_outputs.sh build/release`.

Recovery steps: fix packaging scripts or layout and rerun release validation.
