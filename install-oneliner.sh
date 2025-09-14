#!/bin/sh
# One-liner installation script for DHCP AdGuard Sync
# Usage: fetch -o - https://raw.githubusercontent.com/jeeftor/opnsense-lease-sync/master/install-oneliner.sh | sh

# Set variables
REPO="jeeftor/opnsense-lease-sync"
BINARY_NAME="dhcp-adguard-sync"
LATEST_VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

echo "Installing DHCP AdGuard Sync version ${LATEST_VERSION}..."

# Download binary directly to temporary location
fetch -o /tmp/${BINARY_NAME} "https://github.com/${REPO}/releases/latest/download/dhcp-adguard-sync_freebsd_amd64_${LATEST_VERSION}"
chmod +x /tmp/${BINARY_NAME}

# Run the installer
echo "Running installer..."
/tmp/${BINARY_NAME} install "$@"

# Clean up
rm /tmp/${BINARY_NAME}

echo "Installation complete! Access the plugin at Services > DHCP AdGuard Sync"
echo "Start the service with: service dhcp-adguard-sync start"
echo "Enable at boot with: service dhcp-adguard-sync enable"
