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

# Construct download URL (using .tar.gz)
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar.gz"
echo "Constructed download URL:"
echo "* ${URL}"
echo "--------------------------------"

# Download and extract to temp directory
echo "Downloading release package..."
curl -L -s -o "${TEMP_DIR}/dhcp-adguard-sync.tar.gz" "$URL"
echo "* Download complete"

echo "Extracting package..."
tar xfz "${TEMP_DIR}/dhcp-adguard-sync.tar.gz" -C "${TEMP_DIR}"
echo "* Extraction complete"
echo "--------------------------------"

# Make the binary executable
echo "Setting executable permissions..."
chmod +x "${TEMP_DIR}/${BINARY_NAME}"
echo "* Permissions set"
echo "--------------------------------"

# Prompt for AdGuard Home credentials
echo "AdGuard Home Configuration:"
echo "Please provide your AdGuard Home credentials for configuration:"
echo "--------------------------------"

printf "AdGuard Home Username: "
read -r ADGUARD_USERNAME

printf "AdGuard Home Password: "
# Hide password input if we have a proper TTY
if [ -t 0 ] && command -v stty >/dev/null 2>&1; then
    stty -echo 2>/dev/null
    read -r ADGUARD_PASSWORD
    stty echo 2>/dev/null
    echo ""
else
    # No TTY or stty not available, just read normally
    read -r ADGUARD_PASSWORD
fi

printf "AdGuard Home URL (default: 127.0.0.1:3000): "
read -r ADGUARD_URL
if [ -z "$ADGUARD_URL" ]; then
    ADGUARD_URL="127.0.0.1:3000"
fi

# Test AdGuard Home connection
echo "Testing AdGuard Home connection..."
SCHEME="http"
if echo "$ADGUARD_URL" | grep -q "443"; then
    SCHEME="https"
fi

# Simple connection test using curl
TEST_URL="${SCHEME}://${ADGUARD_URL}/control/status"
echo "* Testing connection to: $TEST_URL"

if curl -s --connect-timeout 5 --max-time 10 -u "${ADGUARD_USERNAME}:${ADGUARD_PASSWORD}" "$TEST_URL" >/dev/null 2>&1; then
    echo "* ✓ AdGuard Home connection successful"
else
    echo "* ✗ AdGuard Home connection failed"
    echo ""
    echo "This could mean:"
    echo "  - AdGuard Home is not running"
    echo "  - Wrong URL, username, or password"
    echo "  - Network connectivity issues"
    echo ""
    printf "Continue with installation anyway? (y/n): "
    read -r CONTINUE
    if [ "$CONTINUE" != "y" ] && [ "$CONTINUE" != "Y" ]; then
        echo "Installation cancelled."
        rm -rf "${TEMP_DIR}"
        exit 1
    fi
    echo "* Continuing with installation..."
fi

echo "--------------------------------"

# Now run the binary's own install command
echo "Running service installation..."

# First attempt to run the installer directly
if "${TEMP_DIR}/${BINARY_NAME}" install --username "$ADGUARD_USERNAME" --password "$ADGUARD_PASSWORD" --adguard-url "$ADGUARD_URL" "$@"; then
    echo "* Service installation complete"
else
    # If installation fails, it might be due to "text file busy" error
    RESULT=$?
    echo "* Initial installation attempt failed (error $RESULT)"

    # Check if binary already exists and get its version
    if [ -f "/usr/local/bin/${BINARY_NAME}" ]; then
        echo "* Checking existing installation..."
        CURRENT_VERSION=$(/usr/local/bin/${BINARY_NAME} version 2>/dev/null | grep "version" | awk '{print $3}')

        if [ -n "$CURRENT_VERSION" ]; then
            echo "* Current version: ${CURRENT_VERSION}"
            echo "* New version: ${VERSION}"
            echo "* Would you like to update? (y/n)"
            read -r ANSWER

            if [ "$ANSWER" != "y" ] && [ "$ANSWER" != "Y" ]; then
                echo "Installation cancelled."
                rm -rf "${TEMP_DIR}"
                exit 0
            fi

            # Stop the service if it's running to avoid "text file busy" error
            echo "* Stopping service before update..."
            /usr/sbin/service dhcp-adguard-sync stop 2>/dev/null
            sleep 2  # Give it time to fully stop

            # Try installation again after stopping the service
            echo "* Retrying installation after stopping service..."
            if "${TEMP_DIR}/${BINARY_NAME}" install --username "$ADGUARD_USERNAME" --password "$ADGUARD_PASSWORD" --adguard-url "$ADGUARD_URL" "$@"; then
                echo "* Service installation complete"
            else
                RESULT=$?
                echo "* Installation failed after stopping service (error $RESULT)"
                echo "* You may need to manually stop the service with: /usr/sbin/service dhcp-adguard-sync stop"
                echo "* Then try running the installer again"
                rm -rf "${TEMP_DIR}"
                exit $RESULT
            fi
        else
            echo "* Existing binary found but couldn't determine version"
            echo "* You may need to manually stop the service with: /usr/sbin/service dhcp-adguard-sync stop"
            echo "* Then try running the installer again"
            rm -rf "${TEMP_DIR}"
            exit $RESULT
        fi
    else
        echo "* Installation failed for unknown reason"
        rm -rf "${TEMP_DIR}"
        exit $RESULT
    fi
fi

echo "--------------------------------"


# Clean up
echo "Cleaning up temporary files..."
rm -rf "${TEMP_DIR}"
echo "* Cleanup complete"
echo "--------------------------------"

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
echo "4. (Optional) Edit configuration if needed:"
echo "   vi /usr/local/etc/dhcp-adguard-sync/config.yaml"
echo "--------------------------------"
