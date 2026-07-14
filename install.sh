#!/usr/bin/env bash
#
# install.sh — One-line installer for racore-cli
# Usage: curl -fsSL https://raw.githubusercontent.com/yingcaihuang/Racore-cli/main/install.sh | bash
#

set -e

REPO="yingcaihuang/Racore-cli"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Error: Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin"; ARCH="all" ;;
  *) echo "Error: Unsupported OS: $OS. Use the MSI installer on Windows."; exit 1 ;;
esac

# Get latest version
echo "Fetching latest version..."
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Error: Could not determine latest version"
  exit 1
fi

echo "Installing racore-cli v${VERSION} for ${OS}/${ARCH}..."

# Download and extract
FILENAME="racore-cli_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "${TMP_DIR}/${FILENAME}"

echo "Extracting..."
tar -xzf "${TMP_DIR}/${FILENAME}" -C "$TMP_DIR"

# Install
echo "Installing to ${INSTALL_DIR}/racore-cli..."
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP_DIR}/racore-cli" "${INSTALL_DIR}/racore-cli"
else
  sudo mv "${TMP_DIR}/racore-cli" "${INSTALL_DIR}/racore-cli"
fi

chmod +x "${INSTALL_DIR}/racore-cli"

echo ""
echo "racore-cli v${VERSION} installed successfully!"
echo "   Run 'racore-cli --version' to verify."
