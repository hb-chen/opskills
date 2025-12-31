#!/bin/bash

# Install KubeKey tool
# Usage: ./install_kubekey.sh [VERSION]
# Example: ./install_kubekey.sh v3.0.0

set -e

VERSION="${1:-latest}"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="kk"

echo "Installing KubeKey ${VERSION}..."

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Get latest version if not specified
if [ "$VERSION" = "latest" ]; then
    echo "Fetching latest version..."
    VERSION=$(curl -s https://api.github.com/repos/kubesphere/kubekey/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    echo "Latest version: $VERSION"
fi

# Remove 'v' prefix if present
VERSION_NUM="${VERSION#v}"

# Construct download URL
DOWNLOAD_URL="https://github.com/kubesphere/kubekey/releases/download/${VERSION}/kubekey-${VERSION_NUM}-${OS}-${ARCH}.tar.gz"

echo "Downloading from: $DOWNLOAD_URL"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
cd "$TMP_DIR"
curl -L -o kubekey.tar.gz "$DOWNLOAD_URL" || {
    echo "Failed to download KubeKey"
    echo "Please check the version and try again"
    exit 1
}

tar -xzf kubekey.tar.gz

# Find the kk binary
if [ -f "kk" ]; then
    KK_BINARY="./kk"
elif [ -f "kubekey-${VERSION_NUM}-${OS}-${ARCH}/kk" ]; then
    KK_BINARY="./kubekey-${VERSION_NUM}-${OS}-${ARCH}/kk"
else
    echo "Could not find kk binary in archive"
    exit 1
fi

# Install binary
echo "Installing to $INSTALL_DIR/$BINARY_NAME..."
if [ -w "$INSTALL_DIR" ]; then
    cp "$KK_BINARY" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Requires sudo to install to $INSTALL_DIR"
    sudo cp "$KK_BINARY" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Verify installation
if command -v kk &> /dev/null; then
    echo "✓ KubeKey installed successfully"
    echo ""
    echo "Version:"
    kk version
else
    echo "Installation completed but kk command not found in PATH"
    exit 1
fi

