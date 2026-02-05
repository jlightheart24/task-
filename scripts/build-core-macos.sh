#!/usr/bin/env bash
set -euo pipefail

# Build Go core as a macOS XCFramework using gomobile bind.
# TODO: install gomobile and configure GOPATH before running.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT_DIR/build/macos"
PKG="taskpp/core/bind"

mkdir -p "$OUT_DIR"

# Example (requires gomobile):
# gomobile bind -target=macos -o "$OUT_DIR/Core.xcframework" "$PKG"

printf "TODO: gomobile bind -target=macos -o %s/Core.xcframework %s\n" "$OUT_DIR" "$PKG"
