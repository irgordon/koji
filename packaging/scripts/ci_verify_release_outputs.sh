#!/usr/bin/env sh
set -eu

artifact_dir="${1:-build/release}"
extract_dir="${RUNNER_TEMP:-/tmp}/koji-release-rootfs"

checksums_status="fail"
rootfs_status="fail"
systemd_status="fail"
forbidden_status="fail"
forbidden_found="true"

record_outputs() {
	if [ -n "${GITHUB_OUTPUT:-}" ]; then
		{
			printf 'checksums_valid=%s\n' "$(status_bool "$checksums_status")"
			printf 'rootfs_layout_valid=%s\n' "$(status_bool "$rootfs_status")"
			printf 'systemd_units_valid=%s\n' "$(status_bool "$systemd_status")"
			printf 'forbidden_paths_found=%s\n' "$forbidden_found"
		} >> "$GITHUB_OUTPUT"
	fi
}

write_summary() {
	if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
		{
			printf '## Artifact Smoke Test Summary\n'
			printf -- '- Checksums: %s\n' "$checksums_status"
			printf -- '- Rootfs layout: %s\n' "$rootfs_status"
			printf -- '- systemd units: %s\n' "$systemd_status"
			printf -- '- Forbidden path scan: %s\n' "$forbidden_status"
			printf -- '- Artifact names verified\n'
		} >> "$GITHUB_STEP_SUMMARY"
	fi
}

status_bool() {
	if [ "$1" = "pass" ]; then
		printf true
	else
		printf false
	fi
}

finish() {
	record_outputs
	write_summary
}
trap finish EXIT

require_artifacts() {
	for artifact in kojid-linux-amd64 koji-agent-linux-amd64 koji-rootfs-linux-amd64.tar.gz SHA256SUMS.txt; do
		if [ ! -s "$artifact_dir/$artifact" ]; then
			echo "missing or empty release artifact: $artifact" >&2
			exit 1
		fi
	done
}

verify_executable() {
	for binary in kojid-linux-amd64 koji-agent-linux-amd64; do
		if [ ! -x "$artifact_dir/$binary" ]; then
			echo "release binary is not executable: $binary" >&2
			exit 1
		fi
		"$artifact_dir/$binary" --help >/dev/null 2>&1
	done
}

extract_rootfs() {
	rm -rf "$extract_dir"
	mkdir -p "$extract_dir"
	tar -xzf "$artifact_dir/koji-rootfs-linux-amd64.tar.gz" -C "$extract_dir"
}

require_artifacts
verify_executable
packaging/scripts/ci_verify_checksums.sh "$artifact_dir"
checksums_status="pass"
extract_rootfs
packaging/scripts/ci_verify_rootfs_layout.sh "$extract_dir/rootfs"
rootfs_status="pass"
packaging/scripts/ci_verify_systemd_units.sh "$extract_dir/rootfs"
systemd_status="pass"
packaging/scripts/verify_release.sh "$artifact_dir" "$extract_dir/rootfs" >/dev/null
forbidden_status="pass"
forbidden_found="false"
