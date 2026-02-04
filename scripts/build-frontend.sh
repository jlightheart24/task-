#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
ASSETS_DIR="$ROOT_DIR/cmd/desktop/assets"

cd "$FRONTEND_DIR"

if [ ! -d node_modules ]; then
  echo "node_modules not found. Run npm install first." >&2
  exit 1
fi

npm run build

rm -rf "$ASSETS_DIR"
mkdir -p "$ASSETS_DIR"

cp -R "$FRONTEND_DIR/dist/." "$ASSETS_DIR/"

echo "Copied frontend build to $ASSETS_DIR"
