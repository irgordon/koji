#!/usr/bin/env sh
set -eu

db_path="${KOJI_DB_PATH:-/var/lib/koji/koji.db}"
expected_schema="${KOJI_EXPECTED_SCHEMA_VERSION:-0009_identity_magic_tokens}"
observability_url="${KOJI_OBSERVABILITY_URL:-}"

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

require_table_readable() {
	sql_scalar "SELECT COUNT(*) FROM $1;" >/dev/null
}

require_command sqlite3
if [ ! -f "$db_path" ]; then
	fail "database missing: $db_path"
fi

sqlite3 "$db_path" "PRAGMA integrity_check;" | grep '^ok$' >/dev/null
schema_version="$(sql_scalar "SELECT COALESCE(MAX(name), '') FROM schema_migrations;")"
if [ "$schema_version" != "$expected_schema" ]; then
	fail "database schema version $schema_version does not match expected $expected_schema"
fi

require_table_readable users
require_table_readable jobs
require_table_readable audit_events
require_table_readable capabilities
require_table_readable user_capabilities

if [ -n "$observability_url" ]; then
	require_command curl
	curl --fail --silent "$observability_url" >/dev/null
fi

cat <<EOF
{
  "schema": "$schema_version",
  "status": "ok",
  "usersReadable": true,
  "jobsReadable": true,
  "auditReadable": true,
  "capabilitiesReadable": true,
  "observabilityChecked": $(if [ -n "$observability_url" ]; then printf true; else printf false; fi)
}
EOF
