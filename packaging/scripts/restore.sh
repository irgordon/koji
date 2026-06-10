#!/usr/bin/env sh
set -eu

backup_source="${1:-}"
db_path="${KOJI_DB_PATH:-/var/lib/koji/koji.db}"
config_dir="${KOJI_CONFIG_DIR:-/etc/koji}"
verify_script="${KOJI_VERIFY_RESTORE_SCRIPT:-$(dirname "$0")/verify_restore.sh}"

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

fail() {
	echo "$1" >&2
	exit 1
}

extract_backup() {
	source="$1"
	tmpdir="$2"
	if [ -d "$source" ]; then
		printf '%s\n' "$source"
		return
	fi
	case "$source" in
		*.tar.gz|*.tgz)
			tar -xzf "$source" -C "$tmpdir"
			find "$tmpdir" -mindepth 1 -maxdepth 1 -type d | sed -n '1p'
			;;
		*)
			fail "backup must be a directory or .tar.gz archive"
			;;
	esac
}

require_backup_file() {
	if [ ! -f "$backup_dir/$1" ]; then
		fail "backup missing required file: $1"
	fi
}

validate_backup() {
	require_backup_file "database/koji.db"
	require_backup_file "config/koji.yaml"
	require_backup_file "config/agent.yaml"
	require_backup_file "metadata.json"
	sqlite3 "$backup_dir/database/koji.db" "PRAGMA integrity_check;" | grep '^ok$' >/dev/null
}

restore_files() {
	mkdir -p "$(dirname "$db_path")" "$config_dir"
	cp "$backup_dir/database/koji.db" "$db_path"
	cp "$backup_dir/config/koji.yaml" "$config_dir/koji.yaml"
	cp "$backup_dir/config/agent.yaml" "$config_dir/agent.yaml"
}

if [ -z "$backup_source" ]; then
	fail "usage: restore.sh BACKUP_DIR_OR_TAR_GZ"
fi

require_command sqlite3
require_command tar
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
backup_dir="$(extract_backup "$backup_source" "$tmpdir")"
if [ -z "$backup_dir" ] || [ ! -d "$backup_dir" ]; then
	fail "could not locate backup directory"
fi
validate_backup
restore_files
"$verify_script" "$db_path" "$backup_dir/metadata.json"
