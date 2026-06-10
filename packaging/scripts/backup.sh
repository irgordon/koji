#!/usr/bin/env sh
set -eu

backup_root="${1:-build/backups}"
db_path="${KOJI_DB_PATH:-/var/lib/koji/koji.db}"
config_dir="${KOJI_CONFIG_DIR:-/etc/koji}"
koji_version="${KOJI_VERSION:-unknown}"
format_version="1"

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

require_file() {
	if [ ! -f "$1" ]; then
		echo "missing required file: $1" >&2
		exit 1
	fi
}

timestamp() {
	date -u '+%Y%m%d-%H%M%S'
}

sql_scalar() {
	sqlite3 "$db_path" "$1"
}

schema_version() {
	sql_scalar "SELECT COALESCE(MAX(name), '') FROM schema_migrations;"
}

table_count() {
	sql_scalar "SELECT COUNT(*) FROM $1;"
}

write_metadata() {
	metadata_path="$1"
	cat >"$metadata_path" <<EOF
{
  "timestamp": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')",
  "koji_version": "$koji_version",
  "database_schema_version": "$(schema_version)",
  "backup_format_version": "$format_version",
  "counts": {
    "users": $(table_count users),
    "jobs": $(table_count jobs),
    "audit_events": $(table_count audit_events),
    "user_capabilities": $(table_count user_capabilities)
  }
}
EOF
}

create_backup() {
	artifact_dir="$backup_root/koji-backup-$(timestamp)"
	archive_path="$artifact_dir.tar.gz"

	mkdir -p "$artifact_dir/database" "$artifact_dir/config"
	sqlite3 "$db_path" ".backup '$artifact_dir/database/koji.db'"
	cp "$config_dir/koji.yaml" "$artifact_dir/config/koji.yaml"
	cp "$config_dir/agent.yaml" "$artifact_dir/config/agent.yaml"
	write_metadata "$artifact_dir/metadata.json"
	tar -czf "$archive_path" -C "$backup_root" "$(basename "$artifact_dir")"
	printf '%s\n' "$archive_path"
}

require_command sqlite3
require_command tar
require_file "$db_path"
require_file "$config_dir/koji.yaml"
require_file "$config_dir/agent.yaml"
mkdir -p "$backup_root"
create_backup
