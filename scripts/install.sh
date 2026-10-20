#!/bin/sh
set -e

PREFIX="${PREFIX:-$HOME/.local}"
SYSTEMD_DIR="${SYSTEMD_DIR:-$HOME/.config/systemd/user}"
BIN_DIR="$PREFIX/bin"

usage() {
    echo "Usage: $0 [--install | --uninstall]"
    echo "  --install   Install dmesg-watch binary and user service (default)"
    echo "  --uninstall Remove dmesg-watch binary and user service"
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
    echo "==> Installing dmesg-watch..."
    mkdir -p "$BIN_DIR"
    mkdir -p "$SYSTEMD_DIR"

    if [ -f "./bin/dmesg-watch" ]; then
        cp -f "./bin/dmesg-watch" "$BIN_DIR/dmesg-watch"
    elif [ -f "./dmesg-watch" ]; then
        cp -f "./dmesg-watch" "$BIN_DIR/dmesg-watch"
    else
        echo "Error: dmesg-watch binary not found in current directory or bin/" >&2
        exit 1
    fi
    chmod 755 "$BIN_DIR/dmesg-watch"

    if [ -f "./systemd/dmesg-watch.service" ]; then
        cp -f "./systemd/dmesg-watch.service" "$SYSTEMD_DIR/dmesg-watch.service"
    elif [ -f "./dmesg-watch.service" ]; then
        cp -f "./dmesg-watch.service" "$SYSTEMD_DIR/dmesg-watch.service"
    else
        echo "Error: dmesg-watch.service not found" >&2
        exit 1
    fi

    echo "==> Reloading systemd user daemon..."
    systemctl --user daemon-reload || true

    echo "==> Installation complete!"
    echo "To enable and start immediately, run:"
    echo "  systemctl --user enable --now dmesg-watch.service"
else
    echo "==> Stopping service if active..."
    systemctl --user stop dmesg-watch.service 2>/dev/null || true
    systemctl --user disable dmesg-watch.service 2>/dev/null || true

    echo "==> Removing installed files..."
    rm -f "$BIN_DIR/dmesg-watch"
    rm -f "$SYSTEMD_DIR/dmesg-watch.service"

    systemctl --user daemon-reload || true
    echo "==> dmesg-watch uninstalled."
fi
