#!/bin/sh
set -e  # Exit on any error

echo "Starting installation process..."
echo "--------------------------------"

# Set variables using uname
echo "Detecting system information..."
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
echo "* Operating System: ${OS}"
echo "* Architecture: ${ARCH}"

REPO="jeeftor/opnsense-lease-sync"
TEMP_DIR="/tmp/dhcp-adguard-sync-install"
BINARY_NAME="dhcp-adguard-sync"

echo "* Using repository: ${REPO}"
echo "* Temporary directory: ${TEMP_DIR}"
echo "* Binary name: ${BINARY_NAME}"
echo "--------------------------------"

# Fetch latest release version from GitHub API
echo "Fetching latest release information from GitHub..."
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Check if version was successfully retrieved
if [ -z "$VERSION" ]; then
    echo "ERROR: Failed to fetch latest version"
    exit 1
fi

echo "* Found version: ${VERSION}"
echo "--------------------------------"

# Create temp directory
echo "Creating temporary directory..."
mkdir -p "${TEMP_DIR}"
echo "* Created: ${TEMP_DIR}"

# Construct download URL (using .tar)
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar"
echo "Constructed download URL:"
echo "* ${URL}"
echo "--------------------------------"

# Download and extract to temp directory
echo "Downloading release package..."
curl -L -s -o "${TEMP_DIR}/dhcp-adguard-sync.tar" "$URL"
echo "* Download complete"

echo "Extracting package..."
tar xf "${TEMP_DIR}/dhcp-adguard-sync.tar" -C "${TEMP_DIR}"
echo "* Extraction complete"
echo "--------------------------------"

# Make the binary executable
echo "Setting executable permissions..."
chmod +x "${TEMP_DIR}/${BINARY_NAME}"
echo "* Permissions set"
echo "--------------------------------"

# Now run the binary's own install command
echo "Running service installation..."
"${TEMP_DIR}/${BINARY_NAME}" install "$@"
echo "* Service installation complete"
echo "--------------------------------"

# Clean up
echo "Cleaning up temporary files..."
rm -rf "${TEMP_DIR}"
echo "* Cleanup complete"
echo "--------------------------------"

echo "Installation complete!"
echo ""
echo "NEXT STEPS:"
echo "1. Edit the configuration file:"
echo "   sudo nano /usr/local/etc/dhcp-adguard-sync/config.yaml"
echo "   sudo vim /usr/local/etc/dhcp-adguard-sync/config.yaml"
echo ""
echo "2. Start the service:"
echo "   sudo service dhcp-adguard-sync start"
echo "--------------------------------"