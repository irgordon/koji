#!/bin/sh
set -eu

wiki_dir="docs/wiki"
home="$wiki_dir/Home.md"

fail() {
	printf '%s\n' "docs validation failed: $1" >&2
	exit 1
}

require_page() {
	[ -f "$1" ] || fail "missing required page: $1"
}

require_link_target() {
	page="$1"
	target="$2"
	case "$target" in
		http://*|https://*|mailto:*|\#*) return ;;
	esac
	target=${target%%#*}
	[ -n "$target" ] || return
	case "$target" in
		/*) return ;;
	esac
	resolved="$(dirname "$page")/$target"
	[ -e "$resolved" ] || fail "stale local link in ${page#"$wiki_dir/"}: $target"
}

[ -d "$wiki_dir" ] || fail "missing docs/wiki"
[ -f "$home" ] || fail "missing docs/wiki/Home.md"

for section in "Operator Quick Start" "Developer Quick Start" "Architecture Overview" "Common Links"; do
	grep -q "^## $section" "$home" || fail "Home.md missing section: $section"
done

stale_product="my""panel"
stale_daemon="my""paneld"
stale_agent="my""panel-agent"
stale_etc="/etc/my""panel"
stale_var="/var/lib/my""panel"
stale_pattern="$stale_product|$stale_daemon|$stale_agent|$stale_etc|$stale_var"

if grep -RniE "$stale_pattern" docs "$wiki_dir" >/dev/null; then
	fail "stale pre-Koji terminology found"
fi

for required in \
	"$wiki_dir/Security/Authentication.md" \
	"$wiki_dir/Security/Sessions.md" \
	"$wiki_dir/Security/CSRF.md" \
	"$wiki_dir/Security/Capabilities.md" \
	"$wiki_dir/Security/Magic-Tokens.md" \
	"$wiki_dir/Security/Threat-Model.md" \
	"$wiki_dir/Security/Security-Review.md" \
	"$wiki_dir/Security/Audit.md" \
	"$wiki_dir/Operations/Installation.md" \
	"$wiki_dir/Operations/Configuration.md" \
	"$wiki_dir/Operations/Health-and-Readiness.md" \
	"$wiki_dir/Operations/Observability.md" \
	"$wiki_dir/Developer/Architectural-Inventory.md" \
	"$wiki_dir/Developer/Backend-Inventory.md" \
	"$wiki_dir/Developer/Frontend-Inventory.md" \
	"$wiki_dir/Developer/Phase-History.md" \
	"$wiki_dir/Developer/Repository-Layout.md" \
	"$wiki_dir/Developer/Testing-Strategy.md" \
	"$wiki_dir/Developer/Release-Workflow.md" \
	"$wiki_dir/Operations/User-Administration.md" \
	"$wiki_dir/Operations/Upgrade-Procedure.md" \
	"$wiki_dir/Operations/Backup-and-Recovery.md" \
	"$wiki_dir/Operations/Disaster-Recovery.md" \
	"$wiki_dir/Operations/Release-Rollback.md" \
	"$wiki_dir/Operations/Release-Candidate-Checklist.md" \
	"$wiki_dir/User-Guide/Overview-Page.md" \
	"$wiki_dir/User-Guide/Services-Page.md" \
	"$wiki_dir/User-Guide/Processes-Page.md" \
	"$wiki_dir/User-Guide/Jobs-Page.md" \
	"$wiki_dir/User-Guide/Activity-Page.md" \
	"$wiki_dir/User-Guide/Administration-Page.md" \
	"$wiki_dir/User-Guide/Settings-Page.md" \
	"$wiki_dir/Reference/Configuration-Reference.md" \
	"$wiki_dir/Reference/Capability-Reference.md" \
	"$wiki_dir/Reference/Job-State-Reference.md" \
	"$wiki_dir/Reference/Audit-Event-Reference.md" \
	"$wiki_dir/Reference/Metrics-Reference.md" \
	"$wiki_dir/Reference/Error-Code-Reference.md" \
	"$wiki_dir/Reference/API-Reference.md"; do
	require_page "$required"
done

find "$wiki_dir" -type f -name '*.md' | while IFS= read -r page; do
	case "$page" in
		"$home") continue ;;
	esac
	relative=${page#"$wiki_dir/"}
	if ! grep -q "($relative)" "$home"; then
		fail "wiki page is not linked from Home.md: $relative"
	fi
	if ! grep -q "(../Home.md)\|(Home.md)" "$page"; then
		fail "wiki page does not link back to Home: $relative"
	fi
	grep -Eo '\]\([^)]+\)' "$page" | sed 's/^](//; s/)$//' | while IFS= read -r target; do
		require_link_target "$page" "$target"
	done
done

for page in \
	"$wiki_dir/Architecture/Request-Flow.md" \
	"$wiki_dir/Architecture/Job-Lifecycle.md" \
	"$wiki_dir/Architecture/Agent-Architecture.md" \
	"$wiki_dir/Developer/Backend-Inventory.md" \
	"$wiki_dir/Reference/Job-State-Reference.md"; do
	grep -q '```mermaid' "$page" || fail "missing Mermaid diagram: $page"
done

printf '%s\n' "docs validation passed"
