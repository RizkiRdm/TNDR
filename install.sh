#!/usr/bin/env bash
set -euo pipefail

# Simple installer for tendr CLI
# Detect OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  armv7l) ARCH=arm ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# GitHub repo
REPO="RizkiRdm/TNDR"
API_URL="https://api.github.com/repos/$REPO/releases/latest"

# Fetch latest release info
JSON=$(curl -fsSL "$API_URL")
TAG=$(echo "$JSON" | grep -Po '"tag_name":\s*"\K[^"]+')
if [ -z "$TAG" ]; then echo "Failed to get latest tag" >&2; exit 1; fi

# Find asset URL for this OS/ARCH
ASSET_URL=$(echo "$JSON" | grep -Po "browser_download_url":\s*\"\Khttps://[^"]+${OS}_${ARCH}[^"]+")
if [ -z "$ASSET_URL" ]; then echo "No binary for $OS/$ARCH" >&2; exit 1; fi

# Download binary
TMPFILE=$(mktemp)
curl -L -o "$TMPFILE" "$ASSET_URL"
chmod +x "$TMPFILE"

# Install location (default /usr/local/bin, fallback to $HOME/.local/bin)
if [ -w "/usr/local/bin" ]; then
  INSTALL_PATH="/usr/local/bin/tendr"
else
  mkdir -p "$HOME/.local/bin"
  INSTALL_PATH="$HOME/.local/bin/tendr"
fi
mv "$TMPFILE" "$INSTALL_PATH"

echo "tendr $TAG installed to $INSTALL_PATH"
