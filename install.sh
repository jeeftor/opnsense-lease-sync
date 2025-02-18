#!/bin/sh
set -e  # Exit on any error

# Define colors
NORMAL="\033[0m"
BOLD="\033[1m"
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
BLUE="\033[34m"
MAGENTA="\033[35m"
CYAN="\033[36m"

echo "${BOLD}Starting installation process...${NORMAL}"
echo "${BLUE}--------------------------------${NORMAL}"

# Set variables using uname
echo "${CYAN}Detecting system information...${NORMAL}"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
echo "${YELLOW}* Operating System: ${NORMAL}${OS}"
echo "${YELLOW}* Architecture: ${NORMAL}${ARCH}"

REPO="jeeftor/opnsense-lease-sync"
TEMP_DIR="/tmp/dhcp-adguard-sync-install"
BINARY_NAME="dhcp-adguard-sync"

echo "${YELLOW}* Using repository: ${NORMAL}${REPO}"
echo "${YELLOW}* Temporary directory: ${NORMAL}${TEMP_DIR}"
echo "${YELLOW}* Binary name: ${NORMAL}${BINARY_NAME}"
echo "${BLUE}--------------------------------${NORMAL}"

# Fetch latest release version from GitHub API
echo "${CYAN}Fetching latest release information from GitHub...${NORMAL}"
VERSION=$(curl -s -S "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Check if version was successfully retrieved
if [ -z "$VERSION" ]; then
    echo "${RED}ERROR: Failed to fetch latest version${NORMAL}"
    exit 1
fi

echo "${YELLOW}* Found version: ${NORMAL}${VERSION}"
echo "${BLUE}--------------------------------${NORMAL}"

# Create temp directory
echo "${CYAN}Creating temporary directory...${NORMAL}"
mkdir -p "${TEMP_DIR}"
echo "${YELLOW}* Created: ${NORMAL}${TEMP_DIR}"

# Construct download URL (using .tar)
URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar"
echo "${CYAN}Constructed download URL:${NORMAL}"
echo "${YELLOW}* ${NORMAL}${URL}"
echo "${BLUE}--------------------------------${NORMAL}"

# Download and extract to temp directory
echo "${CYAN}Downloading release package...${NORMAL}"
curl -L -s -S -o "${TEMP_DIR}/dhcp-adguard-sync.tar" "$URL"
echo "${YELLOW}* Download complete${NORMAL}"

echo "${CYAN}Extracting package...${NORMAL}"
tar xf "${TEMP_DIR}/dhcp-adguard-sync.tar" -C "${TEMP_DIR}"
echo "${YELLOW}* Extraction complete${NORMAL}"
echo "${BLUE}--------------------------------${NORMAL}"

# Make the binary executable
echo "${CYAN}Setting executable permissions...${NORMAL}"
chmod +x "${TEMP_DIR}/${BINARY_NAME}"
echo "${YELLOW}* Permissions set${NORMAL}"
echo "${BLUE}--------------------------------${NORMAL}"

# Now run the binary's own install command
echo "${CYAN}Running service installation...${NORMAL}"
"${TEMP_DIR}/${BINARY_NAME}" install "$@"
echo "${YELLOW}* Service installation complete${NORMAL}"
echo "${BLUE}--------------------------------${NORMAL}"

# Clean up
echo "${CYAN}Cleaning up temporary files...${NORMAL}"
rm -rf "${TEMP_DIR}"
echo "${YELLOW}* Cleanup complete${NORMAL}"
echo "${BLUE}--------------------------------${NORMAL}"

echo "${GREEN}${BOLD}Installation complete!${NORMAL}"
echo ""
echo "${MAGENTA}NEXT STEPS:${NORMAL}"
echo "${BOLD}1. Edit the configuration file:${NORMAL}"
echo "   ${CYAN}sudo nano /usr/local/etc/dhcp-adguard-sync/config.yaml${NORMAL}"
echo ""
echo "${BOLD}2. Start the service:${NORMAL}"
echo "   ${CYAN}sudo service dhcp-adguard-sync start${NORMAL}"
echo "${BLUE}--------------------------------${NORMAL}"