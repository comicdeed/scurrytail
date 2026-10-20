#!/bin/sh
set -e

# Reload global systemd user presets for future logins
systemctl --global daemon-reload 2>/dev/null || true

# Detect active desktop user sessions (supports GUI Software Centers, Discover, PackageKit, and CLI)
for runtime in /run/user/*; do
    [ -d "$runtime/systemd" ] || continue
    uid="${runtime##*/}"
    case "$uid" in
        ''|*[!0-9]*) continue ;;
    esac
    [ "$uid" -ge 1000 ] || continue

    uname=$(id -nu "$uid" 2>/dev/null || true)
    [ -n "$uname" ] || continue

    # Activate immediately for this user's live desktop session
    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user daemon-reload 2>/dev/null || true
    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user enable --now dmesg-watch.service 2>/dev/null || true
done
