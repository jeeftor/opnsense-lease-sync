#!/bin/sh
set -e  # Exit on any error

# Set variables using uname
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
REPO="jeeftor/opnsense-lease-sync"
INSTALL_DIR="/usr/local/bin"

# Fetch latest release version from GitHub API
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

echo "Installing dhcp-adguard-sync ${VERSION} for ${OS}_${ARCH}"

# Construct download URL
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar"

# Download and extract
echo "Downloading from: ${URL}"
curl -L -o /tmp/dhcp-adguard-sync.tar "$URL"
tar xf /tmp/dhcp-adguard-sync.tar -C "${INSTALL_DIR}"

# Set permissions
chmod +x "${INSTALL_DIR}/dhcp-adguard-sync"

# Clean up
rm /tmp/dhcp-adguard-sync.tar

echo "Installation complete! Run 'dhcp-adguard-sync --help' for usage information"