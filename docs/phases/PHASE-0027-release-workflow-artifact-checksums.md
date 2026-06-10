# Phase 27: Release Workflow, Build Provenance, and Artifact Checksums

## Goal

Create a deterministic release workflow that builds Koji artifacts in CI, validates package structure, generates SHA256 checksums, and publishes release assets from version tags.

## Trigger

Push a tag matching:

```text
v*
```

Example:

```sh
git tag v0.1.0
git push origin v0.1.0
```

## Release Artifacts

```text
kojid-linux-amd64
koji-agent-linux-amd64
koji-rootfs-linux-amd64.tar.gz
SHA256SUMS.txt
```

## Local Verification

```sh
make release
make verify-release
```

Checksum verification:

```sh
cd build/release
shasum -a 256 -c SHA256SUMS.txt
```

## Guardrails

- CI starts from a clean checkout.
- Go and Node versions are pinned.
- Frontend assets are built during release.
- Rootfs layout is validated before upload.
- Checksums are generated for every release artifact.
- Artifacts are scanned for developer-local path contamination.
- Releases are published automatically by GitHub Actions.

## Non-Goals

- RPM or DEB packaging.
- Container image publishing.
- Signing or provenance attestations.
- Multi-architecture releases.
- SBOM enforcement.
