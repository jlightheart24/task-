#!/usr/bin/env bash
set -euo pipefail

# Build Go core as an iOS XCFramework using gomobile bind.
# TODO: install gomobile and configure GOPATH before running.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT_DIR/build/ios"
PKG="taskpp/core/bind"

mkdir -p "$OUT_DIR"

# Example (requires gomobile):
# gomobile bind -target=ios -o "$OUT_DIR/Core.xcframework" "$PKG"

printf "TODO: gomobile bind -target=ios -o %s/Core.xcframework %s\n" "$OUT_DIR" "$PKG"
