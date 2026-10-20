#!/bin/sh
set -e

# If uninstalled via sudo, stop and disable the user service for the active desktop session
if [ -n "$SUDO_USER" ] && [ "$SUDO_USER" != "root" ]; then
    SUDO_UID=$(id -u "$SUDO_USER" 2>/dev/null || true)
    if [ -n "$SUDO_UID" ]; then
        USER_RUNTIME="/run/user/$SUDO_UID"
        if [ -d "$USER_RUNTIME/systemd" ]; then
            runuser -u "$SUDO_USER" -- env XDG_RUNTIME_DIR="$USER_RUNTIME" \
                systemctl --user stop dmesg-watch.service 2>/dev/null || true
            runuser -u "$SUDO_USER" -- env XDG_RUNTIME_DIR="$USER_RUNTIME" \
                systemctl --user disable dmesg-watch.service 2>/dev/null || true
        fi
    fi
fi
