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
    echo "==> Verifying system prerequisites..."
    if ! command -v journalctl >/dev/null 2>&1; then
        echo "Error: 'journalctl' was not found on this system." >&2
        echo "dmesg-watch requires systemd-journald to stream kernel errors." >&2
        exit 1
    fi

    if [ ! -d /run/systemd/system ]; then
        echo "Error: systemd is not running as the active init system." >&2
        echo "dmesg-watch requires a running systemd user session." >&2
        exit 1
    fi

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

    SERVICE_SRC=""
    if [ -f "./systemd/dmesg-watch.service" ]; then
        SERVICE_SRC="./systemd/dmesg-watch.service"
    elif [ -f "./dmesg-watch.service" ]; then
        SERVICE_SRC="./dmesg-watch.service"
    else
        echo "Error: dmesg-watch.service not found" >&2
        exit 1
    fi
    sed "s|ExecStart=/usr/bin/dmesg-watch|ExecStart=$BIN_DIR/dmesg-watch|" "$SERVICE_SRC" > "$SYSTEMD_DIR/dmesg-watch.service"
    chmod 644 "$SYSTEMD_DIR/dmesg-watch.service"

    echo "==> Reloading systemd user daemon..."
    systemctl --user daemon-reload || true

    echo "==> Enabling and starting dmesg-watch user service..."
    systemctl --user enable --now dmesg-watch.service || true

    echo "==> Installation complete! dmesg-watch is active."
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
