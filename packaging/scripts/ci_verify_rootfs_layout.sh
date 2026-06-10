#!/usr/bin/env sh
set -eu

rootfs_dir="${1:-build/rootfs}"

require_path() {
	if [ ! -e "$rootfs_dir/$1" ]; then
		echo "missing rootfs path: $1" >&2
		exit 1
	fi
}

require_dir() {
	if [ ! -d "$rootfs_dir/$1" ]; then
		echo "missing rootfs directory: $1" >&2
		exit 1
	fi
}

require_path usr/bin/kojid
require_path usr/bin/koji-agent
require_dir usr/share/koji/dist
require_dir etc/koji
require_dir usr/lib/systemd/system
require_dir var/lib/koji

if grep -R -E '/Users/|/home/|Documents/Projects|godzilla|zuki' "$rootfs_dir/etc" "$rootfs_dir/usr/lib/systemd/system" >/dev/null; then
	echo "developer-local runtime path found in rootfs text files" >&2
	exit 1
fi
