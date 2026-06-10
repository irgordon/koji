#!/bin/sh
set -eu

wiki_dir="docs/wiki"
home="$wiki_dir/Home.md"

fail() {
	printf '%s\n' "docs validation failed: $1" >&2
	exit 1
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
done

for page in \
	"$wiki_dir/Architecture/Request-Flow.md" \
	"$wiki_dir/Architecture/Job-Lifecycle.md" \
	"$wiki_dir/Architecture/Agent-Architecture.md"; do
	grep -q '```mermaid' "$page" || fail "missing Mermaid diagram: $page"
done

printf '%s\n' "docs validation passed"
