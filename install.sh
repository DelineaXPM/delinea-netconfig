#!/bin/sh
# Installation script for delinea-netconfig
# Usage: curl -sfL https://raw.githubusercontent.com/DelineaXPM/delinea-netconfig/main/install.sh | sh

set -e

# Detect OS
OS=$(uname -s)
ARCH=$(uname -m)

# Map OS to GoReleaser naming
case "$OS" in
    Linux)
        OS="Linux"
        ;;
    Darwin)
        OS="Darwin"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Map architecture to GoReleaser naming
case "$ARCH" in
    x86_64)
        ARCH="x86_64"
        ;;
    amd64)
        ARCH="x86_64"
        ;;
    arm64)
        ARCH="arm64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    i386|i686)
        ARCH="i386"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest version from GitHub API
# Optional: set GITHUB_TOKEN env var to authenticate (required for private repos,
# also helps avoid API rate limits)
echo "Fetching latest release..."
if [ -n "$GITHUB_TOKEN" ]; then
    API_RESPONSE=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/repos/DelineaXPM/delinea-netconfig/releases/latest)
else
    API_RESPONSE=$(curl -s https://api.github.com/repos/DelineaXPM/delinea-netconfig/releases/latest)
fi
LATEST_VERSION=$(echo "$API_RESPONSE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo ""
    echo "Error: Failed to fetch latest version from GitHub API."
    echo ""
    echo "API response was:"
    echo "$API_RESPONSE" | head -5
    echo ""
    echo "Possible causes:"
    echo "  - The repository may not be publicly accessible"
    echo "  - No releases have been published yet"
    echo "  - GitHub API rate limit exceeded (try again in a few minutes)"
    echo ""
    echo "Manual install: download a release from https://github.com/DelineaXPM/delinea-netconfig/releases"
    exit 1
fi

echo "Latest version: $LATEST_VERSION"

# Construct download URL
BINARY_NAME="delinea-netconfig_${LATEST_VERSION#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/DelineaXPM/delinea-netconfig/releases/download/${LATEST_VERSION}/${BINARY_NAME}"

echo "Downloading from: $DOWNLOAD_URL"

# Download to temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

cd "$TMP_DIR"

# Download and extract
if command -v curl >/dev/null 2>&1; then
    curl -LO "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget "$DOWNLOAD_URL"
else
    echo "Neither curl nor wget found. Please install one of them."
    exit 1
fi

# Extract
tar -xzf "$BINARY_NAME"

# Determine install location
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    echo "No write permission to $INSTALL_DIR. Trying with sudo..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo mv delinea-netconfig "$INSTALL_DIR/"
else
    mkdir -p "$INSTALL_DIR"
    mv delinea-netconfig "$INSTALL_DIR/"
fi

# Make executable
if [ -w "$INSTALL_DIR/delinea-netconfig" ]; then
    chmod +x "$INSTALL_DIR/delinea-netconfig"
else
    sudo chmod +x "$INSTALL_DIR/delinea-netconfig"
fi

echo ""
echo "✓ delinea-netconfig installed successfully to $INSTALL_DIR/delinea-netconfig"
echo ""
echo "Run 'delinea-netconfig --help' to get started"
echo "Run 'delinea-netconfig version' to verify installation"
echo ""
echo "Shell completion:"
echo "  Bash:  delinea-netconfig completion bash > /etc/bash_completion.d/delinea-netconfig"
echo "  Zsh:   delinea-netconfig completion zsh > \"\${fpath[1]}/_delinea-netconfig\""
echo "  Fish:  delinea-netconfig completion fish > ~/.config/fish/completions/delinea-netconfig.fish"
