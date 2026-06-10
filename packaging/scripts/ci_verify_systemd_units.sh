#!/usr/bin/env sh
set -eu

rootfs_dir="${1:-build/rootfs}"
daemon_unit="$rootfs_dir/usr/lib/systemd/system/kojid.service"
agent_unit="$rootfs_dir/usr/lib/systemd/system/koji-agent.service"

require_unit() {
	if [ ! -s "$1" ]; then
		echo "missing or empty systemd unit: $1" >&2
		exit 1
	fi
}

require_line() {
	if ! grep -Fx "$2" "$1" >/dev/null; then
		echo "missing systemd line in $1: $2" >&2
		exit 1
	fi
}

reject_line() {
	if grep -E "$2" "$1" >/dev/null; then
		echo "forbidden systemd line in $1: $2" >&2
		exit 1
	fi
}

require_unit "$daemon_unit"
require_unit "$agent_unit"

require_line "$daemon_unit" "ExecStart=/usr/bin/kojid"
require_line "$daemon_unit" "WorkingDirectory=/"
require_line "$daemon_unit" "RuntimeDirectory=koji"

require_line "$agent_unit" "ExecStart=/usr/bin/koji-agent -config /etc/koji/agent.yaml"
require_line "$agent_unit" "WorkingDirectory=/"
require_line "$agent_unit" "RuntimeDirectory=koji"

reject_line "$daemon_unit" '^ExecStart=\./'
reject_line "$agent_unit" '^ExecStart=\./'
reject_line "$daemon_unit" '^WorkingDirectory=(/Users/|/home/|[^/])'
reject_line "$agent_unit" '^WorkingDirectory=(/Users/|/home/|[^/])'
