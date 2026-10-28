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

    # Reload systemd user daemon so it picks up any new unit file definitions
    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user daemon-reload 2>/dev/null || true

    # If the service is already running (upgrade scenario), seamlessly restart it;
    # otherwise enable and start it (fresh install scenario).
    if runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user is-active --quiet scurrytail.service 2>/dev/null; then
        runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
            systemctl --user restart scurrytail.service 2>/dev/null || true
    else
        runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
            systemctl --user enable --now scurrytail.service 2>/dev/null || true
    fi
done
