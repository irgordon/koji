#!/usr/bin/env sh
set -eu

db_path="${1:-${KOJI_DB_PATH:-/var/lib/koji/koji.db}}"
metadata_path="${2:-}"
expected_schema="${KOJI_EXPECTED_SCHEMA_VERSION:-0008_observability_metrics_read_capability}"

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

metadata_count() {
	name="$1"
	if [ -z "$metadata_path" ] || [ ! -f "$metadata_path" ]; then
		printf '%s\n' ""
		return
	fi
	sed -n "s/.*\"$name\": \([0-9][0-9]*\).*/\1/p" "$metadata_path" | sed -n '1p'
}

require_count_at_least() {
	table="$1"
	expected="$(metadata_count "$table")"
	if [ -z "$expected" ]; then
		return
	fi
	actual="$(sql_scalar "SELECT COUNT(*) FROM $table;")"
	if [ "$actual" -lt "$expected" ]; then
		fail "$table count $actual is lower than backup metadata $expected"
	fi
}

require_command sqlite3
if [ ! -f "$db_path" ]; then
	fail "missing database: $db_path"
fi

sqlite3 "$db_path" "PRAGMA integrity_check;" | grep '^ok$' >/dev/null
schema_version="$(sql_scalar "SELECT COALESCE(MAX(name), '') FROM schema_migrations;")"
if [ "$schema_version" != "$expected_schema" ]; then
	fail "database schema version $schema_version does not match expected $expected_schema"
fi

require_count_at_least users
require_count_at_least jobs
require_count_at_least audit_events
require_count_at_least user_capabilities

printf '%s\n' "restore verification passed"
