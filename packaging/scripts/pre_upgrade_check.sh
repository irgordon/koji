#!/usr/bin/env sh
set -eu

db_path="${KOJI_DB_PATH:-/var/lib/koji/koji.db}"
config_dir="${KOJI_CONFIG_DIR:-/etc/koji}"
target_schema="${KOJI_TARGET_SCHEMA_VERSION:-0008_observability_metrics_read_capability}"

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

sql_scalar() {
	sqlite3 "$db_path" "$1"
}

schema_table_exists() {
	sql_scalar "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations';"
}

current_schema() {
	if [ "$(schema_table_exists)" = "0" ]; then
		printf '%s\n' ""
		return
	fi
	sql_scalar "SELECT COALESCE(MAX(name), '') FROM schema_migrations;"
}

applied_migrations() {
	if [ "$(schema_table_exists)" = "0" ]; then
		printf '%s\n' "0"
		return
	fi
	sql_scalar "SELECT COUNT(*) FROM schema_migrations;"
}

print_report() {
	status="$1"
	reason="$2"
	current="$3"
	applied="$4"
	cat <<EOF
{
  "currentSchema": "$current",
  "targetSchema": "$target_schema",
  "appliedMigrations": $applied,
  "status": "$status",
  "reason": "$reason",
  "backupRequired": true
}
EOF
}

require_command sqlite3

if [ ! -f "$db_path" ]; then
	fail "database missing: $db_path"
fi
if [ ! -f "$config_dir/koji.yaml" ]; then
	fail "daemon config missing: $config_dir/koji.yaml"
fi
if [ ! -f "$config_dir/agent.yaml" ]; then
	fail "agent config missing: $config_dir/agent.yaml"
fi

sqlite3 "$db_path" "PRAGMA integrity_check;" | grep '^ok$' >/dev/null
current="$(current_schema)"
applied="$(applied_migrations)"

if [ "$current" \> "$target_schema" ]; then
	print_report "future_schema_detected" "database_schema_newer_than_release" "$current" "$applied"
	exit 1
fi
if [ "$current" = "$target_schema" ]; then
	print_report "ok" "schema_current" "$current" "$applied"
	exit 0
fi

print_report "migration_required" "backup_required_before_upgrade" "$current" "$applied"
