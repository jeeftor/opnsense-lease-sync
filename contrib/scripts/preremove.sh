#!/bin/sh
set -e

if [ -d /run/systemd/system ]; then
    systemctl stop dhcp-adguard-sync.service || true
    systemctl disable dhcp-adguard-sync.service || true
fi
