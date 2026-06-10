#!/bin/sh
set -eu

fail() {
	printf '%s\n' "code quality validation failed: $1" >&2
	exit 1
}

require_file() {
	[ -f "$1" ] || fail "missing required file: $1"
}

check_no_match() {
	pattern="$1"
	shift
	if rg -n "$pattern" "$@" >/dev/null; then
		rg -n "$pattern" "$@" >&2 || true
		fail "unexpected match for pattern: $pattern"
	fi
}

check_max_lines() {
	file="$1"
	max="$2"
	lines=$(wc -l < "$file" | tr -d ' ')
	[ "$lines" -le "$max" ] || fail "$file has $lines lines, limit is $max"
}

require_file "docs/phases/PHASE-0039-code-quality-audit-complexity-reduction.md"
require_file "docs/wiki/Developer/Code-Quality-Audit.md"

check_max_lines "web/src/App.tsx" 1500
check_max_lines "web/src/api.ts" 700
check_max_lines "web/src/AdminView.tsx" 450
check_max_lines "internal/http/handlers_admin.go" 350
check_max_lines "internal/identity/store.go" 250
check_max_lines "internal/identity/capabilities.go" 160
check_max_lines "internal/identity/lockout.go" 160
check_max_lines "internal/identity/magic_tokens.go" 140

check_no_match ": any|as any" web/src
check_no_match "catch\\s*\\([^)]*\\)\\s*\\{\\s*\\}" web/src
check_no_match "exec\\.Command|CommandContext" internal/http internal/agent internal/system
check_no_match "systemctl" internal/http internal/agent internal/system internal/jobs

if ! rg -n "ServiceManager\\s*=\\s*\"systemctl\"" internal/platform/command >/dev/null; then
	fail "expected systemctl ownership marker in internal/platform/command"
fi

printf '%s\n' "code quality validation passed"
