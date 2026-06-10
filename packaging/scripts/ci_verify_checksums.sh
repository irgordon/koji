#!/usr/bin/env sh
set -eu

artifact_dir="${1:-build/release}"
checksum_file="$artifact_dir/SHA256SUMS.txt"

if [ ! -s "$checksum_file" ]; then
	echo "missing or empty checksum file: $checksum_file" >&2
	exit 1
fi

for artifact in kojid-linux-amd64 koji-agent-linux-amd64 koji-rootfs-linux-amd64.tar.gz; do
	if ! grep -E "  $artifact$" "$checksum_file" >/dev/null; then
		echo "missing checksum entry: $artifact" >&2
		exit 1
	fi
done

cd "$artifact_dir"
if command -v sha256sum >/dev/null 2>&1; then
	sha256sum -c SHA256SUMS.txt
else
	shasum -a 256 -c SHA256SUMS.txt
fi
