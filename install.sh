#!/usr/bin/env bash
set -euo pipefail

# broom install script
# Usage: curl -fsSL https://raw.githubusercontent.com/okolilemuel/broom/main/install.sh | bash

REPO="okolilemuel/broom"
BINARY="broom"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin)
    case "$ARCH" in
      arm64)  ASSET="${BINARY}-darwin-arm64" ;;
      x86_64) ASSET="${BINARY}-darwin-amd64" ;;
      *)
        echo "Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
    esac
    ;;
  Linux)
    case "$ARCH" in
      x86_64) ASSET="${BINARY}-linux-amd64" ;;
      aarch64|arm64) ASSET="${BINARY}-linux-arm64" ;;
      *)
        echo "Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# Get latest release tag
echo "Fetching latest release..."
TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG" ]; then
  echo "Could not determine latest release tag." >&2
  exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

echo "Downloading ${BINARY} ${TAG} for ${OS}/${ARCH}..."
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

curl -fsSL "$DOWNLOAD_URL" -o "$TMP"
chmod +x "$TMP"

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

echo "broom ${TAG} installed successfully!"
echo "Run: broom --help"
