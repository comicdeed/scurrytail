# dmesg-watch

A lightweight, resource-efficient Linux desktop background daemon that monitors kernel errors in real time, displays live metrics in the system tray via StatusNotifierItem (SNI), and fires desktop notifications when high error bursts occur.

Engineered in Go to provide minimal background memory consumption (~15MB RSS), zero complex GUI dependencies, and seamless integration under systemd user sessions.

---

## Features

- **Resource Efficient**: Pure Go implementation using direct D-Bus communication without heavy GUI platform runtimes.
- **Real-Time Streaming**: Attaches directly to `journalctl -k -f -n 0 -p 3 --output short-precise`.
- **Sliding-Window Metrics**: Tracks error frequency dynamically across a rolling 60-second window.
- **Flood Protection**: Notification cooldown ensures burst errors do not spam notification banners while the tray menu continuously updates.
- **Quick Drill-Down**: One-click tray menu action to open full kernel logs in your default terminal (`journalctl -k -f -n 50`), with automatic detection for Ghostty, KDE desktop preferences, `$TERMINAL`, `xdg-terminal-exec`, and common Linux terminals.
- **Cross-Desktop Tray**: Implements the StatusNotifierItem (SNI) protocol supported natively by KDE Plasma, GNOME (with AppIndicator), Sway, Waybar, and other Wayland/X11 bars.

---

## Prerequisites

Your Linux user account must have permissions to inspect kernel logs without root privilege. On most distributions (Debian, Ubuntu, Arch, Fedora):

```bash
sudo usermod -aG systemd-journal $USER
sudo usermod -aG adm $USER
```
*(Log out and log back in for group changes to apply.)*

---

## Installation

### Option 1: Build & Install from Source (Makefile)

```bash
# Clone the repository
git clone https://github.com/comicdeed/dmesg-watch.git
cd dmesg-watch

# Build, install, and automatically start user service
make install
```

To control or inspect the service:
```bash
make service-status
make service-disable
make service-enable
```

To uninstall:
```bash
make uninstall
```

### Option 2: Go Install

```bash
go install github.com/comicdeed/dmesg-watch/cmd/dmesg-watch@latest
```

---

## CLI Options

```text
Usage of dmesg-watch:
  -cooldown duration
        Notification suppression cooldown period (default 45s)
  -threshold int
        Number of kernel errors per window to trigger an alert (default 5)
  -version
        Print application version and exit
  -viewer-cmd string
        Command to view kernel logs in terminal (default "journalctl -k -f -n 50")
  -window duration
        Rolling evaluation window duration (default 1m0s)
```

---

## Architecture

- `cmd/dmesg-watch`: Application entrypoint, CLI flag parsing, signal orchestration.
- `internal/journal`: Manages `journalctl` background streaming and message prefix normalization.
- `internal/launcher`: Detached process launcher and smart terminal emulator detection cascade.
- `internal/tracker`: Sliding-window queue of error timestamps, rate threshold calculations, and cooldown state.
- `internal/notify`: Pure Go D-Bus client for `org.freedesktop.Notifications`.
- `internal/tray`: StatusNotifierItem tray menu with live metric updates and embedded fallback icon.
- `systemd/`: User session service unit for `graphical-session.target`.

---

## License

MIT License.
