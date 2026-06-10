#!/usr/bin/env sh
set -eu

artifact_dir="${1:-build/release}"
output_file="$artifact_dir/SHA256SUMS.txt"

cd "$artifact_dir"
rm -f SHA256SUMS.txt

for artifact in kojid-linux-amd64 koji-agent-linux-amd64 koji-rootfs-linux-amd64.tar.gz; do
	if [ ! -f "$artifact" ]; then
		echo "missing release artifact: $artifact" >&2
		exit 1
	fi
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$artifact" >> SHA256SUMS.txt
	else
		shasum -a 256 "$artifact" >> SHA256SUMS.txt
	fi
done

test -s "$(basename "$output_file")"
