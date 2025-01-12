#!/bin/sh
set -e

if [ "$1" = "remove" ]; then
    if [ -d /run/systemd/system ]; then
        systemctl daemon-reload || true
    fi
fi