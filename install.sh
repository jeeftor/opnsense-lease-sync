#!/bin/sh
set -e  # Exit on any error

# Colors for better visibility
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

echo "${CYAN}🚀 Starting DHCP AdGuard Sync installation...${NC}"
echo "${BLUE}================================================${NC}"

# Set variables using uname
echo "${WHITE}📋 Detecting system information...${NC}"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
echo "${GREEN}  ✓ Operating System: ${YELLOW}${OS}${NC}"
echo "${GREEN}  ✓ Architecture: ${YELLOW}${ARCH}${NC}"

REPO="jeeftor/opnsense-lease-sync"
TEMP_DIR="/tmp/dhcp-adguard-sync-install"
BINARY_NAME="dhcp-adguard-sync"

echo "${GREEN}  ✓ Using repository: ${YELLOW}${REPO}${NC}"
echo "${GREEN}  ✓ Temporary directory: ${YELLOW}${TEMP_DIR}${NC}"
echo "${GREEN}  ✓ Binary name: ${YELLOW}${BINARY_NAME}${NC}"
echo "${BLUE}------------------------------------------------${NC}"

# Fetch latest release version from GitHub API
echo "${WHITE}📡 Fetching latest release information from GitHub...${NC}"
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Check if version was successfully retrieved
if [ -z "$VERSION" ]; then
    echo "${RED}❌ ERROR: Failed to fetch latest version${NC}"
    exit 1
fi

echo "${GREEN}  ✓ Found version: ${YELLOW}${VERSION}${NC}"
echo "${BLUE}------------------------------------------------${NC}"

# Create temp directory
echo "${WHITE}📁 Creating temporary directory...${NC}"
mkdir -p "${TEMP_DIR}"
echo "${GREEN}  ✓ Created: ${YELLOW}${TEMP_DIR}${NC}"

# Construct download URL (using .tar.gz)
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar.gz"
echo "${WHITE}🔗 Constructing download URL...${NC}"
echo "${GREEN}  ✓ URL: ${YELLOW}${URL}${NC}"
echo "${BLUE}------------------------------------------------${NC}"

# Download and extract to temp directory
echo "${WHITE}⬇️  Downloading release package...${NC}"
curl -L -s -o "${TEMP_DIR}/dhcp-adguard-sync.tar.gz" "$URL"
echo "${GREEN}  ✓ Download complete${NC}"

echo "${WHITE}📦 Extracting package...${NC}"
tar xfz "${TEMP_DIR}/dhcp-adguard-sync.tar.gz" -C "${TEMP_DIR}"
echo "${GREEN}  ✓ Extraction complete${NC}"
echo "${BLUE}------------------------------------------------${NC}"

# Make the binary executable
echo "${WHITE}🔧 Setting executable permissions...${NC}"
chmod +x "${TEMP_DIR}/${BINARY_NAME}"
echo "${GREEN}  ✓ Permissions set${NC}"
echo "${BLUE}------------------------------------------------${NC}"

# Check for existing configuration
CONFIG_PATH="/usr/local/etc/dhcp-adguard-sync/config.yaml"
if [ -f "$CONFIG_PATH" ]; then
    echo "${YELLOW}⚠️  Existing configuration found at: ${CYAN}$CONFIG_PATH${NC}"
    echo "${YELLOW}   Installation will preserve existing configuration.${NC}"
    echo "${BLUE}------------------------------------------------${NC}"
fi

# Check if this is an interactive session
if [ -t 0 ]; then
    # Interactive mode - we can prompt for input
    echo "🔐 AdGuard Home Configuration"
    echo "Please provide your AdGuard Home credentials:"
    echo "--------------------------------"

    printf "👤 AdGuard Home Username: "
    read -r ADGUARD_USERNAME

    printf "🔑 AdGuard Home Password: "
    # Hide password input if we have a proper TTY
    if command -v stty >/dev/null 2>&1; then
        stty -echo 2>/dev/null
        read -r ADGUARD_PASSWORD
        stty echo 2>/dev/null
        echo ""
    else
        read -r ADGUARD_PASSWORD
    fi

    printf "🌐 AdGuard Home URL (default: 127.0.0.1:3000): "
    read -r ADGUARD_URL
    if [ -z "$ADGUARD_URL" ]; then
        ADGUARD_URL="127.0.0.1:3000"
    fi

    echo "--------------------------------"
    echo "🧪 Testing AdGuard Home connection..."

    SCHEME="http"
    if echo "$ADGUARD_URL" | grep -q "443"; then
        SCHEME="https"
    fi

    TEST_URL="${SCHEME}://${ADGUARD_URL}/control/status"
    echo "* Testing: $TEST_URL"

    if curl -s --connect-timeout 5 --max-time 10 -u "${ADGUARD_USERNAME}:${ADGUARD_PASSWORD}" "$TEST_URL" >/dev/null 2>&1; then
        echo "✅ AdGuard Home connection successful"
    else
        echo "❌ AdGuard Home connection failed"
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
    echo "🚀 Running service installation..."

    # Install with provided credentials
    if "${TEMP_DIR}/${BINARY_NAME}" install --username "$ADGUARD_USERNAME" --password "$ADGUARD_PASSWORD" --adguard-url "$ADGUARD_URL" "$@"; then
else
    # Non-interactive mode (piped from curl)
    echo "⚠️  Non-interactive installation detected"
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
    echo "🚀 Running service-only installation..."

    # Install without credentials (will need manual config)
    if "${TEMP_DIR}/${BINARY_NAME}" install "$@"; then
        echo "* Service installation complete"
    else
        echo "* Installation failed - you may need to configure manually"
        echo "* Edit config file: /usr/local/etc/dhcp-adguard-sync/config.yaml"
    fi
fi



# Clean up
echo "🧹 Cleaning up temporary files..."
rm -rf "${TEMP_DIR}"
echo "✅ Cleanup complete"
echo "================================================"

echo "🎉 Installation complete!"
echo ""
echo "📋 NEXT STEPS:"
echo "1. 🚀 Start the service:"
echo "   service dhcp-adguard-sync start"
echo ""
echo "2. 📊 Check service status:"
echo "   service dhcp-adguard-sync status"
echo ""
echo "3. 📝 View logs:"
echo "   tail -f /var/log/dhcp-adguard-sync.log"
echo ""
if [ ! -t 0 ]; then
    echo "4. ⚙️  Edit configuration (required for non-interactive install):"
    echo "   vi /usr/local/etc/dhcp-adguard-sync/config.yaml"
    echo "   - Set ADGUARD_USERNAME and ADGUARD_PASSWORD"
    echo "   - Set ADGUARD_URL if not using default"
    echo ""
fi
echo "================================================"
