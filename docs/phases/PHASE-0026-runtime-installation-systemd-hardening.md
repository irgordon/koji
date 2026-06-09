# Phase 26: Runtime Installation, systemd Hardening, and Production Deployment Layout

## Goal

Prepare Koji for Linux deployment by defining installation paths, service accounts, systemd units, filesystem ownership expectations, runtime directories, and package-oriented build targets.

## Scope

- Add `packaging/systemd/kojid.service`.
- Add `packaging/systemd/koji-agent.service`.
- Add daemon and agent configuration examples.
- Add a staging install tree under `build/rootfs/`.
- Add package-oriented Makefile targets.
- Validate packaging files and example configs in tests.
- Tighten agent socket parent ownership validation.

## Runtime Layout

```text
/usr/bin/kojid
/usr/bin/koji-agent
/usr/share/koji/dist
/etc/koji/koji.yaml
/etc/koji/agent.yaml
/var/lib/koji
/run/koji
/usr/lib/systemd/system
```

## Boundaries

- No custom daemonization.
- No PID files.
- No custom log files.
- systemd owns supervision and restart behavior.
- Runtime paths are absolute.
- Services run as `koji:koji`.

## Result

Koji now has a deterministic deployment layout suitable for later RPM, DEB, or other package-specific work.
