#!/bin/sh
if [ -f /bin/systemctl ]; then
    systemctl daemon-reload
    systemctl enable dhcp-adguard-sync
    systemctl start dhcp-adguard-sync
fi