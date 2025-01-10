#!/bin/sh
if [ -f /bin/systemctl ]; then
    systemctl stop dhcp-adguard-sync
    systemctl disable dhcp-adguard-sync
fi