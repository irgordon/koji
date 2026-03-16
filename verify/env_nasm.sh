#!/usr/bin/env bash
set -euo pipefail

check_nasm_presence() {
    if ! command -v nasm >/dev/null 2>&1; then
        echo "Error: NASM not found. Please install NASM." >&2
        exit 1
    fi
}

main() {
    check_nasm_presence
    echo "Found NASM: $(nasm -v)"
}

main "$@"
