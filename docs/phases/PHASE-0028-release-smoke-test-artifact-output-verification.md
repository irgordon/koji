# Phase 28: Release Smoke Test and Artifact Output Verification

## Status

Implemented.

## Summary

Phase 28 adds a distinct CI smoke-test stage between release artifact creation and GitHub release publication. The release workflow now builds artifacts, uploads them as GitHub Actions artifacts, downloads those artifacts in a separate smoke-test job, validates the downloaded files, and only publishes release assets after smoke validation succeeds.

## Release Job Flow

```text
build_release -> smoke_test_release -> publish_release
```

The `publish_release` job depends on `smoke_test_release`, so release assets are not uploaded unless CI has verified the generated artifact set.

## Smoke Validation

The CI smoke scripts validate:

- required artifact names and non-empty files
- executable `kojid-linux-amd64` and `koji-agent-linux-amd64`
- safe `--help` execution for both binaries
- `SHA256SUMS.txt` coverage and checksum validity
- extracted rootfs layout
- installed systemd unit paths
- forbidden developer-local path contamination

## Workflow Outputs

The smoke-test job exposes:

- `checksums_valid`
- `rootfs_layout_valid`
- `systemd_units_valid`
- `forbidden_paths_found`

Successful smoke tests report:

```text
checksums_valid=true
rootfs_layout_valid=true
systemd_units_valid=true
forbidden_paths_found=false
```

## GitHub Step Summary

The smoke-test job writes an `Artifact Smoke Test Summary` to `$GITHUB_STEP_SUMMARY` with checksum, rootfs, systemd, forbidden-path, and artifact-name results.

## Non-Goals

Phase 28 does not add package formats, signing, SBOM enforcement, container images, multi-architecture builds, or systemd service startup in CI.
