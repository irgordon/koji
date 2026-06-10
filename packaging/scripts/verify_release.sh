#!/usr/bin/env sh
set -eu

artifact_dir="${1:-build/release}"
rootfs_dir="${2:-build/rootfs}"

require_file() {
	if [ ! -f "$1" ]; then
		echo "missing file: $1" >&2
		exit 1
	fi
}

require_dir() {
	if [ ! -d "$1" ]; then
		echo "missing directory: $1" >&2
		exit 1
	fi
}

reject_forbidden_paths() {
	target="$1"
	if strings "$target" | grep -E '/Users/|/home/|C:\\Users\\|Documents/Projects|godzilla|irgordon' >/dev/null; then
		echo "forbidden local path found in $target" >&2
		exit 1
	fi
}

require_file "$artifact_dir/kojid-linux-amd64"
require_file "$artifact_dir/koji-agent-linux-amd64"
require_file "$artifact_dir/koji-rootfs-linux-amd64.tar.gz"
require_file "$artifact_dir/SHA256SUMS.txt"

require_dir "$rootfs_dir/usr/bin"
require_file "$rootfs_dir/usr/bin/kojid"
require_file "$rootfs_dir/usr/bin/koji-agent"
require_dir "$rootfs_dir/usr/share/koji/dist"
require_dir "$rootfs_dir/etc/koji"
require_dir "$rootfs_dir/usr/lib/systemd/system"
require_dir "$rootfs_dir/var/lib/koji"

tar -tzf "$artifact_dir/koji-rootfs-linux-amd64.tar.gz" | grep -E '(^|/)usr/share/koji/dist/?$' >/dev/null
tar -tzf "$artifact_dir/koji-rootfs-linux-amd64.tar.gz" | grep -E '(^|/)usr/lib/systemd/system/?$' >/dev/null

reject_forbidden_paths "$artifact_dir/kojid-linux-amd64"
reject_forbidden_paths "$artifact_dir/koji-agent-linux-amd64"
reject_forbidden_paths "$artifact_dir/koji-rootfs-linux-amd64.tar.gz"

cd "$artifact_dir"
if command -v sha256sum >/dev/null 2>&1; then
	sha256sum -c SHA256SUMS.txt
else
	shasum -a 256 -c SHA256SUMS.txt
fi
