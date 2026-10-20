#!/bin/sh
set -e

# Stop and disable dmesg-watch across active desktop sessions
for runtime in /run/user/*; do
    [ -d "$runtime/systemd" ] || continue
    uid="${runtime##*/}"
    case "$uid" in
        ''|*[!0-9]*) continue ;;
    esac
    [ "$uid" -ge 1000 ] || continue

    uname=$(id -nu "$uid" 2>/dev/null || true)
    [ -n "$uname" ] || continue

    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user stop dmesg-watch.service 2>/dev/null || true
    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user disable dmesg-watch.service 2>/dev/null || true
    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user daemon-reload 2>/dev/null || true
done

systemctl --global daemon-reload 2>/dev/null || true
