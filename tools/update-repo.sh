#!/bin/sh
# Script to update the OPNsense plugin repository

# Create directory structure if it doesn't exist
mkdir -p repo/FreeBSD:13:amd64/Latest

# Build the Go binary
make build

# Create a temporary directory for the plugin package
TEMP_DIR=$(mktemp -d)
cp -r opnsense-plugin $TEMP_DIR/os-dhcpadguardsync

# Copy the built binary to the plugin's files directory
mkdir -p $TEMP_DIR/os-dhcpadguardsync/src/opnsense/scripts
cp build/opnsense-lease-sync $TEMP_DIR/os-dhcpadguardsync/src/opnsense/scripts/
chmod +x $TEMP_DIR/os-dhcpadguardsync/src/opnsense/scripts/opnsense-lease-sync

# Create package
cd $TEMP_DIR
pkg create -M os-dhcpadguardsync

# Copy package to repo
cp $TEMP_DIR/os-dhcpadguardsync-*.txz ../repo/FreeBSD:13:amd64/Latest/

# Generate repository metadata
cd ../repo
pkg repo FreeBSD:13:amd64

# Clean up
rm -rf $TEMP_DIR

echo "Repository updated successfully!"
echo "Package is available at: repo/FreeBSD:13:amd64/Latest/"
echo "Repository metadata is at: repo/FreeBSD:13:amd64/packagesite.txz"
