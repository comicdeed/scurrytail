#!/bin/sh
set -e

# Verify journalctl command exists
if ! command -v journalctl >/dev/null 2>&1; then
    echo "ERROR: 'journalctl' was not found on this system." >&2
    echo "dmesg-watch requires systemd-journald to stream kernel errors in real time." >&2
    exit 1
fi

# Verify systemd is the active booted init system (standard sd_booted check)
if [ ! -d /run/systemd/system ]; then
    echo "ERROR: systemd is not running as the active init system." >&2
    echo "dmesg-watch requires a running systemd user session." >&2
    exit 1
fi

exit 0
