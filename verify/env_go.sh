#!/usr/bin/env bash
set -euo pipefail

readonly TARGET_OS="linux"
readonly REQUIRED_ARCHS=("amd64" "arm64")

check_go_presence() {
    if ! command -v go >/dev/null 2>&1; then
        echo "Error: Go toolchain not found." >&2
        exit 1
    fi
}

verify_cross_compile() {
    local arch="$1"
    # macOS mktemp fix: requires a template or -t
    local temp_dir
    temp_dir=$(mktemp -d -t koji-go-verify)
    local source_file="$temp_dir/main.go"
    local output_file="$temp_dir/verify_go_${arch}.bin"

    trap 'rm -rf "$temp_dir"' EXIT

    cat > "$source_file" <<'SOURCE'
package main
func main() {}
SOURCE

    if GOOS="$TARGET_OS" GOARCH="$arch" go build -o "$output_file" "$source_file" >/dev/null 2>&1; then
        echo "Go: Cross-compile for ${arch}/${TARGET_OS}: SUCCESS"
    else
        echo "Error: Go failed to cross-compile for ${arch}/${TARGET_OS}." >&2
        exit 1
    fi
}

main() {
    check_go_presence
    echo "Found Go: $(go version)"
    for arch in "${REQUIRED_ARCHS[@]}"; do
        verify_cross_compile "$arch"
    done
}

main "$@"
