#!/bin/sh
set -e  # Exit on any error

# Set variables using uname
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
REPO="jeeftor/opnsense-lease-sync"
TEMP_DIR="/tmp/dhcp-adguard-sync-install"
BINARY_NAME="dhcp-adguard-sync"

# Fetch latest release version from GitHub API
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Check if version was successfully retrieved
if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

echo "Found version: ${VERSION}"

# Create temp directory
mkdir -p "${TEMP_DIR}"

# Construct download URL (using .tar)
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar"

# Download and extract to temp directory
echo "Downloading from: ${URL}"
curl -L -o "${TEMP_DIR}/dhcp-adguard-sync.tar" "$URL"
tar xf "${TEMP_DIR}/dhcp-adguard-sync.tar" -C "${TEMP_DIR}"

# Make the binary executable
chmod +x "${TEMP_DIR}/${BINARY_NAME}"

# Now run the binary's own install command
echo "Running service installation..."
"${TEMP_DIR}/${BINARY_NAME}" install "$@"

# Clean up
rm -rf "${TEMP_DIR}"

echo "Installation complete!"