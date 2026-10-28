#!/bin/sh
set -e

# Only stop and disable on complete package removal.
# On upgrade (RPM passes 1, DEB passes upgrade), do not stop or disable the service.
if [ "$1" = "upgrade" ] || [ "$1" = "1" ]; then
    exit 0
fi

# Stop and disable scurrytail across active desktop sessions
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
        systemctl --user stop scurrytail.service 2>/dev/null || true
    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user disable scurrytail.service 2>/dev/null || true
    runuser -u "$uname" -- env XDG_RUNTIME_DIR="$runtime" \
        systemctl --user daemon-reload 2>/dev/null || true
done

systemctl --global daemon-reload 2>/dev/null || true
