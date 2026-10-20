#!/bin/sh
set -e

# Reload global systemd user presets so the service is enabled on desktop login
systemctl --global daemon-reload 2>/dev/null || true

# If installed via sudo, activate the user service immediately for the active desktop session
if [ -n "$SUDO_USER" ] && [ "$SUDO_USER" != "root" ]; then
    SUDO_UID=$(id -u "$SUDO_USER" 2>/dev/null || true)
    if [ -n "$SUDO_UID" ]; then
        USER_RUNTIME="/run/user/$SUDO_UID"
        if [ -d "$USER_RUNTIME/systemd" ]; then
            runuser -u "$SUDO_USER" -- env XDG_RUNTIME_DIR="$USER_RUNTIME" \
                systemctl --user daemon-reload 2>/dev/null || true
            runuser -u "$SUDO_USER" -- env XDG_RUNTIME_DIR="$USER_RUNTIME" \
                systemctl --user enable --now dmesg-watch.service 2>/dev/null || true
        fi
    fi
fi
