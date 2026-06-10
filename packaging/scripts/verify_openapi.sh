#!/bin/sh
set -eu

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

node packaging/scripts/generate_openapi_docs.mjs --out "$tmpdir/api" --wiki-out "$tmpdir/wiki-api-reference.md"

node -e 'JSON.parse(require("fs").readFileSync("docs/api/openapi.json", "utf8"));'
test -s docs/api/openapi.yaml
test -s docs/api/generated/API-Reference.md
test -s docs/api/generated/Endpoints.md
test -s docs/api/generated/Errors.md
test -s docs/wiki/Reference/API-Reference.md

if ! grep -q '^openapi: "3.1.0"' docs/api/openapi.yaml; then
	printf '%s\n' "openapi validation failed: docs/api/openapi.yaml missing OpenAPI 3.1.0 header" >&2
	exit 1
fi

if ! grep -q 'x-koji-capability' docs/api/openapi.yaml; then
	printf '%s\n' "openapi validation failed: missing capability metadata" >&2
	exit 1
fi

diff -u "$tmpdir/api/openapi.yaml" docs/api/openapi.yaml >/dev/null || {
	printf '%s\n' "openapi validation failed: docs/api/openapi.yaml is stale" >&2
	exit 1
}

diff -u "$tmpdir/api/generated/API-Reference.md" docs/api/generated/API-Reference.md >/dev/null || {
	printf '%s\n' "openapi validation failed: generated API reference is stale" >&2
	exit 1
}

diff -u "$tmpdir/api/generated/Endpoints.md" docs/api/generated/Endpoints.md >/dev/null || {
	printf '%s\n' "openapi validation failed: generated endpoint reference is stale" >&2
	exit 1
}

diff -u "$tmpdir/api/generated/Errors.md" docs/api/generated/Errors.md >/dev/null || {
	printf '%s\n' "openapi validation failed: generated error reference is stale" >&2
	exit 1
}

diff -u "$tmpdir/wiki-api-reference.md" docs/wiki/Reference/API-Reference.md >/dev/null || {
	printf '%s\n' "openapi validation failed: wiki API reference is stale" >&2
	exit 1
}

printf '%s\n' "openapi validation passed"
