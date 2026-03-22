#!/bin/bash
set -e

PLUGIN_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="$PLUGIN_DIR/bin"
mkdir -p "$BIN_DIR"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac

REPO="yaronya/backstory"
VERSION=$(curl -sf "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"v\(.*\)".*/\1/' 2>/dev/null)

if [ -n "$VERSION" ]; then
  echo "Downloading backstory v${VERSION}..."
  for BIN in backstory backstory-mcp; do
    URL="https://github.com/$REPO/releases/download/v${VERSION}/${BIN}-${OS}-${ARCH}"
    if curl -sfL "$URL" -o "$BIN_DIR/$BIN"; then
      chmod +x "$BIN_DIR/$BIN"
      echo "  Downloaded $BIN"
    fi
  done
fi

if [ ! -f "$BIN_DIR/backstory" ]; then
  echo "No release found. Building from source..."
  if command -v go &>/dev/null; then
    PROJ_DIR="$PLUGIN_DIR/.."
    go build -C "$PROJ_DIR" -o "$BIN_DIR/backstory" ./cmd/backstory
    go build -C "$PROJ_DIR" -o "$BIN_DIR/backstory-mcp" ./cmd/backstory-mcp
    echo "  Built from source"
  else
    echo "Error: Go not installed and no pre-built binary available."
    exit 1
  fi
fi

echo ""
echo "Backstory installed!"
echo "Set BACKSTORY_REPO to your decisions repo path to get started."
