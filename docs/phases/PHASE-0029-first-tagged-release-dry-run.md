# Phase 29: First Tagged Release Dry Run

## Status

Implemented and verified.

## Summary

Phase 29 performed the first end-to-end GitHub Releases dry run for Koji. The release process was validated using real tags, GitHub Actions, published GitHub Release assets, checksum verification, rootfs inspection, systemd unit checks, and forbidden-path scans.

## Tags

The first attempted tag was `v0.1.0`, as planned. It exposed CI portability issues in tests and did not publish a release.

Corrective dry-run tags:

- `v0.1.1`: verified the first test portability fix but exposed a second socket lifecycle assumption.
- `v0.1.2`: passed build and artifact upload, then exposed GitHub artifact permission normalization in the smoke gate.
- `v0.1.3`: completed successfully.

## Successful Release

- Tag: `v0.1.3`
- Workflow run: <https://github.com/irgordon/koji/actions/runs/27245563509>
- Release URL: <https://github.com/irgordon/koji/releases/tag/v0.1.3>
- Commit: `31bbaa563b2a1f8beb2de35510551cccaafa1b11`

## Workflow Results

The successful `v0.1.3` release workflow completed in the required order:

```text
build_release -> smoke_test_release -> publish_release
```

Results:

- `build_release`: success
- `smoke_test_release`: success
- `publish_release`: success

The smoke-test job downloaded the build artifact produced by `build_release`, validated checksums, validated the rootfs layout, validated systemd units, scanned for forbidden developer-local paths, and set the configured workflow outputs.

## Assets

GitHub Release `v0.1.3` contains:

- `kojid-linux-amd64`
- `koji-agent-linux-amd64`
- `koji-rootfs-linux-amd64.tar.gz`
- `SHA256SUMS.txt`

All assets were present and non-empty on the GitHub Release.

## Downloaded Asset Verification

Assets were downloaded from the GitHub Release page into `/tmp/koji-release-v0.1.3`.

Checksum verification:

```text
kojid-linux-amd64: OK
koji-agent-linux-amd64: OK
koji-rootfs-linux-amd64.tar.gz: OK
```

Rootfs layout verification confirmed:

```text
rootfs/usr/bin/kojid
rootfs/usr/bin/koji-agent
rootfs/usr/share/koji/dist
rootfs/etc/koji
rootfs/usr/lib/systemd/system
rootfs/var/lib/koji
```

Systemd unit verification confirmed:

```text
ExecStart=/usr/bin/kojid
ExecStart=/usr/bin/koji-agent -config /etc/koji/agent.yaml
```

Forbidden-path scans found no matches for:

```text
/Users
Documents/Projects
godzilla
zuki
/home/runner/work
```

## Issues Encountered

### v0.1.0

`go test ./...` failed in CI because agent socket tests used `/private/tmp`, which exists on macOS but not on Ubuntu runners. The service status failure test also depended on host `systemctl` behavior.

Corrective action:

- Use `/tmp` for short Unix socket test paths.
- Use a fake `systemctl` in `PATH` for deterministic command-failure mapping.

### v0.1.1

`go test ./...` still failed because a closed Unix socket listener produced different dial behavior on Ubuntu than on macOS.

Corrective action:

- Replace the platform-dependent closed-listener integration expectation with deterministic dial error classification coverage.

### v0.1.2

The smoke-test job failed because GitHub artifact upload/download does not preserve executable bits on binaries.

Corrective action:

- Restore executable bits in the smoke verifier before running safe `--help` checks.

## Notes

The release binaries are Linux amd64 ELF files. Local verification on macOS validated file type, checksums, rootfs contents, systemd paths, and path hygiene. Binary `--help` execution was validated in the Ubuntu smoke-test job.

## Non-Goals

No runtime application behavior changed during this phase. The only code changes were release/test portability corrections required to complete the release-process validation.
