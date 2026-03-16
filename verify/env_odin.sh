#!/usr/bin/env bash
set -euo pipefail

check_odin_presence() {
    if ! command -v odin >/dev/null 2>&1; then
        echo "Error: Odin toolchain not found. Please install Odin." >&2
        exit 1
    fi
}

main() {
    check_odin_presence
    echo "Found Odin: $(odin version)"
}

main "$@"
