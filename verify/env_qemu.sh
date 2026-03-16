#!/usr/bin/env bash
set -euo pipefail

readonly QEMU_BINARIES=("qemu-system-x86_64" "qemu-system-aarch64")

check_qemu_binary() {
    local binary="$1"
    if ! command -v "$binary" >/dev/null 2>&1; then
        echo "Error: Required emulator not found: ${binary}" >&2
        exit 1
    fi
    local version
    version="$("$binary" --version | head -n 1)"
    echo "${binary}: Found (${version})"
}

main() {
    for binary in "${QEMU_BINARIES[@]}"; do
        check_qemu_binary "$binary"
    done
    echo "QEMU: All required emulators found."
}

main "$@"
