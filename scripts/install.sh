#!/bin/sh
set -e

PREFIX="${PREFIX:-$HOME/.local}"
SYSTEMD_DIR="${SYSTEMD_DIR:-$HOME/.config/systemd/user}"
BIN_DIR="$PREFIX/bin"

usage() {
    echo "Usage: $0 [--install | --uninstall]"
    echo "  --install   Install scurrytail binary and user service (default)"
    echo "  --uninstall Remove scurrytail binary and user service"
    exit 1
}

ACTION="install"
if [ "$1" = "--uninstall" ]; then
    ACTION="uninstall"
elif [ "$1" = "--install" ] || [ -z "$1" ]; then
    ACTION="install"
else
    usage
fi

if [ "$ACTION" = "install" ]; then
    echo "==> Verifying system prerequisites..."
    if ! command -v journalctl >/dev/null 2>&1; then
        echo "Error: 'journalctl' was not found on this system." >&2
        echo "scurrytail requires systemd-journald to stream kernel errors." >&2
        exit 1
    fi

    if [ ! -d /run/systemd/system ]; then
        echo "Error: systemd is not running as the active init system." >&2
        echo "scurrytail requires a running systemd user session." >&2
        exit 1
    fi

    echo "==> Installing scurrytail..."
    mkdir -p "$BIN_DIR"
    mkdir -p "$SYSTEMD_DIR"

    if [ -f "./bin/scurrytail" ]; then
        cp -f "./bin/scurrytail" "$BIN_DIR/scurrytail"
    elif [ -f "./scurrytail" ]; then
        cp -f "./scurrytail" "$BIN_DIR/scurrytail"
    else
        echo "Error: scurrytail binary not found in current directory or bin/" >&2
        exit 1
    fi
    chmod 755 "$BIN_DIR/scurrytail"

    SERVICE_SRC=""
    if [ -f "./systemd/scurrytail.service" ]; then
        SERVICE_SRC="./systemd/scurrytail.service"
    elif [ -f "./scurrytail.service" ]; then
        SERVICE_SRC="./scurrytail.service"
    else
        echo "Error: scurrytail.service not found" >&2
        exit 1
    fi
    sed "s|ExecStart=/usr/bin/scurrytail|ExecStart=$BIN_DIR/scurrytail|" "$SERVICE_SRC" > "$SYSTEMD_DIR/scurrytail.service"
    chmod 644 "$SYSTEMD_DIR/scurrytail.service"

    echo "==> Reloading systemd user daemon..."
    systemctl --user daemon-reload || true

    echo "==> Enabling and starting scurrytail user service..."
    systemctl --user enable --now scurrytail.service || true

    echo "==> Installation complete! scurrytail is active."
else
    echo "==> Stopping service if active..."
    systemctl --user stop scurrytail.service 2>/dev/null || true
    systemctl --user disable scurrytail.service 2>/dev/null || true

    echo "==> Removing installed files..."
    rm -f "$BIN_DIR/scurrytail"
    rm -f "$SYSTEMD_DIR/scurrytail.service"

    systemctl --user daemon-reload || true
    echo "==> scurrytail uninstalled."
fi
