#!/bin/sh
set -e  # Exit on any error

echo "Starting DHCP AdGuard Sync installation..."
echo "================================================"

# Set variables using uname
echo "Detecting system information..."
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
echo "  * Operating System: ${OS}"
echo "  * Architecture: ${ARCH}"

REPO="jeeftor/opnsense-lease-sync"
TEMP_DIR="/tmp/dhcp-adguard-sync-install"
BINARY_NAME="dhcp-adguard-sync"

echo "  * Using repository: ${REPO}"
echo "  * Temporary directory: ${TEMP_DIR}"
echo "  * Binary name: ${BINARY_NAME}"
echo "------------------------------------------------"

# Fetch latest release version from GitHub API
echo "Fetching latest release information from GitHub..."
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Check if version was successfully retrieved
if [ -z "$VERSION" ]; then
    echo "ERROR: Failed to fetch latest version"
    exit 1
fi

echo "  * Found version: ${VERSION}"
echo "------------------------------------------------"

# Create temp directory
echo "Creating temporary directory..."
mkdir -p "${TEMP_DIR}"
echo "  * Created: ${TEMP_DIR}"

# Construct download URL (using .tar.gz)
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar.gz"
echo "Constructing download URL..."
echo "  * URL: ${URL}"
echo "------------------------------------------------"

# Download and extract to temp directory
echo "Downloading release package..."
curl -L -s -o "${TEMP_DIR}/dhcp-adguard-sync.tar.gz" "$URL"
echo "  * Download complete"

echo "Extracting package..."
tar xfz "${TEMP_DIR}/dhcp-adguard-sync.tar.gz" -C "${TEMP_DIR}"
echo "  * Extraction complete"
echo "------------------------------------------------"

# Make the binary executable
echo "Setting executable permissions..."
chmod +x "${TEMP_DIR}/${BINARY_NAME}"
echo "  * Permissions set"
echo "------------------------------------------------"

# Check for existing configuration
CONFIG_PATH="/usr/local/etc/dhcp-adguard-sync/config.yaml"
if [ -f "$CONFIG_PATH" ]; then
    echo "WARNING: Existing configuration found at: $CONFIG_PATH"
    echo "         Installation will preserve existing configuration."
    echo "------------------------------------------------"
fi

# Check if this is an interactive session
if [ -t 0 ]; then
    # Interactive mode - we can prompt for input
    echo "AdGuard Home Configuration"
    echo "Please provide your AdGuard Home credentials:"
    echo "--------------------------------"

    printf "AdGuard Home Username: "
    read -r ADGUARD_USERNAME

    printf "AdGuard Home Password: "
    # Hide password input if we have a proper TTY
    if command -v stty >/dev/null 2>&1; then
        stty -echo 2>/dev/null
        read -r ADGUARD_PASSWORD
        stty echo 2>/dev/null
        echo ""
    else
        read -r ADGUARD_PASSWORD
    fi

    printf "AdGuard Home URL (default: 127.0.0.1:3000): "
    read -r ADGUARD_URL
    if [ -z "$ADGUARD_URL" ]; then
        ADGUARD_URL="127.0.0.1:3000"
    fi

    echo "--------------------------------"
    echo "Testing AdGuard Home connection..."

    SCHEME="http"
    if echo "$ADGUARD_URL" | grep -q "443"; then
        SCHEME="https"
    fi

    TEST_URL="${SCHEME}://${ADGUARD_URL}/control/status"
    echo "* Testing: $TEST_URL"

    if curl -s --connect-timeout 5 --max-time 10 -u "${ADGUARD_USERNAME}:${ADGUARD_PASSWORD}" "$TEST_URL" >/dev/null 2>&1; then
        echo "SUCCESS: AdGuard Home connection successful"
    else
        echo "ERROR: AdGuard Home connection failed"
        echo ""
        echo "This could mean:"
        echo "  - AdGuard Home is not running"
        echo "  - Wrong URL, username, or password"
        echo "  - Network connectivity issues"
        echo ""
        printf "Continue anyway? (y/n): "
        read -r CONTINUE
        if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
            echo "Installation cancelled."
            rm -rf "${TEMP_DIR}"
            exit 1
        fi
    fi

    echo "--------------------------------"
    echo "Running service installation..."

    # Install with provided credentials
    if "${TEMP_DIR}/${BINARY_NAME}" install --username "$ADGUARD_USERNAME" --password "$ADGUARD_PASSWORD" --adguard-url "$ADGUARD_URL" "$@"; then
        echo "* Service installation complete"
    else
        echo "* Installation failed"
        rm -rf "${TEMP_DIR}"
        exit 1
    fi
else
    # Non-interactive mode (piped from curl)
    echo "WARNING: Non-interactive installation detected"
    echo ""
    echo "When piping from curl, credentials cannot be prompted interactively."
    echo ""
    echo "OPTIONS:"
    echo "1. Download and run manually:"
    echo "   curl -L -o install.sh https://raw.githubusercontent.com/jeeftor/opnsense-lease-sync/master/install.sh"
    echo "   chmod +x install.sh"
    echo "   ./install.sh"
    echo ""
    echo "2. Install service-only (you'll need to edit config manually):"
    echo "   Continuing with service-only installation..."
    echo ""

    echo "--------------------------------"
    echo "Running service-only installation..."

    # Install without credentials (will need manual config)
    if "${TEMP_DIR}/${BINARY_NAME}" install "$@"; then
        echo "* Service installation complete"
    else
        echo "* Installation failed - you may need to configure manually"
        echo "* Edit config file: /usr/local/etc/dhcp-adguard-sync/config.yaml"
    fi
fi

# Clean up
echo "Cleaning up temporary files..."
rm -rf "${TEMP_DIR}"
echo "Cleanup complete"
echo "================================================"

echo "Installation complete!"
echo ""
echo "NEXT STEPS:"
echo "1. Start the service:"
echo "   service dhcp-adguard-sync start"
echo ""
echo "2. Check service status:"
echo "   service dhcp-adguard-sync status"
echo ""
echo "3. View logs:"
echo "   tail -f /var/log/dhcp-adguard-sync.log"
echo ""
if [ ! -t 0 ]; then
    echo "4. Edit configuration (required for non-interactive install):"
    echo "   vi /usr/local/etc/dhcp-adguard-sync/config.yaml"
    echo "   - Set ADGUARD_USERNAME and ADGUARD_PASSWORD"
    echo "   - Set ADGUARD_URL if not using default"
    echo ""
fi
echo "================================================"
