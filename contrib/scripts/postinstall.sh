#!/bin/sh
set -e

# Create config directory if it doesn't exist
mkdir -p /etc/dhcp-adguard-sync

# Set proper permissions for config file
if [ -f /etc/dhcp-adguard-sync/config.yaml ]; then
    chmod 600 /etc/dhcp-adguard-sync/config.yaml
fi

# Reload systemd
if [ -d /run/systemd/system ]; then
    systemctl daemon-reload
    systemctl enable dhcp-adguard-sync.service || true
    systemctl restart dhcp-adguard-sync.service || true
fi