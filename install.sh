#!/bin/sh
set -e  # Exit on any error

# Set variables using uname
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
REPO="jeeftor/opnsense-lease-sync"
INSTALL_DIR="/usr/local/bin"

# Fetch latest release version from GitHub API
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Check if version was successfully retrieved
if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

echo "Found version: ${VERSION}"

# Construct download URL (using .tar)
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar"

# Download and extract
echo "Downloading from: ${URL}"
curl -L -o /tmp/dhcp-adguard-sync.tar "$URL"

# Create install directory if it doesn't exist
mkdir -p "${INSTALL_DIR}"

# Extract directly to install directory
echo "Installing to ${INSTALL_DIR}..."
tar xf /tmp/dhcp-adguard-sync.tar -C "${INSTALL_DIR}"

# Set executable permissions
chmod +x "${INSTALL_DIR}/dhcp-adguard-sync"

# Clean up
rm /tmp/dhcp-adguard-sync.tar

echo "Installation complete! The dhcp-adguard-sync binary is now available in ${INSTALL_DIR}"