# Phase 22: Job Approval and Rejection Workflow

## Goal

Add an explicit human approval boundary before queued service-control jobs can move toward any future execution path.

## Scope

- Add `jobs.approve`.
- Persist approval and rejection decision metadata on jobs.
- Add protected approval and rejection endpoints.
- Update the Jobs page with queued-only decision controls.
- Audit approval, rejection, and approval denial.

## Boundaries

- No `systemctl` execution.
- No agent mutation.
- No automatic creator approval.
- Only `queued` jobs can transition to `approved` or `rejected`.
- Existing authentication, CSRF, capability checks, and audit requirements remain in force.

## Result

Queued service-control jobs now require a durable human decision before any later worker or agent execution phase can consume them.
