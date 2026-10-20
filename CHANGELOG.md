# Changelog

All notable changes to **dmesg-watch** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.0-beta.3] - 2026-10-20

Automates systemd user service enablement and startup upon package installation across all distribution formats.

### ✨ Highlights
* **✨ Automatic Service Enablement**: Packages (`.deb`, `.rpm`) and installer scripts now automatically enable and start `dmesg-watch.service` under the active desktop user session immediately upon installation—eliminating manual `systemctl` setup.
* **✨ Declarative Systemd Preset**: Installed a system-wide user preset (`50-dmesg-watch.preset`) ensuring the service starts automatically on future desktop logins.
* **Native About Submenu & Website Launcher**: Converted the About item into a native tray submenu featuring application details and a "Website" action that opens the project repository in your default web browser via `xdg-open`.
* **Decluttered Top-Level Menu**: Relocated static version metadata into the About submenu and adopted standard desktop single-word actions (`About`, `Quit`).
* **Quick Drill-Down to Kernel Logs**: Added a "View Kernel Logs (journalctl)" action to the system tray menu to instantly open full kernel logs in your default terminal with continuous live tailing (`journalctl -k -f -n 50`).
* **Smart Terminal Emulator Detection**: Automatically detects Ghostty, KDE Plasma desktop terminal preferences (`kdeglobals`), `$TERMINAL`, `xdg-terminal-exec`, and common desktop terminals.
* **Configurable Viewer Command**: Added a `--viewer-cmd` CLI option allowing users to customize their log viewer command.

### 🛠️ Technical Notes
* ✨ Added `scripts/postinstall.sh` and `scripts/prerm.sh` to nfpm packaging configuration to handle `$SUDO_USER` session activation and clean removal.
* ✨ Bundled `/usr/lib/systemd/user-preset/50-dmesg-watch.preset` across `.deb`, `.rpm`, AUR, and standalone tarballs.
* ✨ Updated `make install` and `scripts/install.sh` to run `systemctl --user enable --now dmesg-watch.service` directly during installation.
* Implemented `launcher.OpenURL` with detached process isolation (`syscall.SysProcAttr{Setsid: true}`) for cross-desktop browser dispatch.

## [v0.3.0-beta.2] - 2026-10-20

Streamlines tray menu UX with native submenus, web browser repository launching, and decluttered top-level actions.

### ✨ Highlights
* **✨ Native About Submenu & Website Launcher**: Converted the About item into a native tray submenu featuring application details and a "Website" action that opens the project repository in your default web browser via `xdg-open`.
* **✨ Decluttered Top-Level Menu**: Relocated static version metadata into the About submenu and adopted standard desktop single-word actions (`About`, `Quit`).
* **Quick Drill-Down to Kernel Logs**: Added a "View Kernel Logs (journalctl)" action to the system tray menu to instantly open full kernel logs in your default terminal with continuous live tailing (`journalctl -k -f -n 50`).
* **Smart Terminal Emulator Detection**: Automatically detects Ghostty, KDE Plasma desktop terminal preferences (`kdeglobals`), `$TERMINAL`, `xdg-terminal-exec`, and common desktop terminals (`konsole`, `gnome-terminal`, `ptyxis`, `alacritty`, `xfce4-terminal`, etc.).
* **Configurable Viewer Command**: Added a `--viewer-cmd` CLI option allowing users to customize their log viewer command or attach alternative tools.

### 🛠️ Technical Notes
* ✨ Implemented `launcher.OpenURL` with detached process isolation (`syscall.SysProcAttr{Setsid: true}`) for cross-desktop browser dispatch.
* ✨ Replaced D-Bus notification dispatch for About dialogs, eliminating HTML markup and link formatting discrepancies on GNOME Shell.
* Implemented `internal/launcher` with detached process isolation (`syscall.SysProcAttr{Setsid: true}`) for independent terminal lifetimes.
* Added terminal fallback cascade validating executable existence on disk before dispatch.

## [v0.3.0-beta.1] - 2026-10-20

Introduces one-click terminal log inspection to seamlessly transition from tray visibility to active debugging.

### ✨ Highlights
* **✨ Quick Drill-Down to Kernel Logs**: Added a "View Kernel Logs (journalctl)" action to the system tray menu to instantly open full kernel logs in your default terminal with continuous live tailing (`journalctl -k -f -n 50`).
* **✨ Smart Terminal Emulator Detection**: Automatically detects Ghostty, KDE Plasma desktop terminal preferences (`kdeglobals`), `$TERMINAL`, `xdg-terminal-exec`, and common desktop terminals (`konsole`, `gnome-terminal`, `ptyxis`, `alacritty`, `xfce4-terminal`, etc.).
* **✨ Configurable Viewer Command**: Added a `--viewer-cmd` CLI option allowing users to customize their log viewer command or attach alternative tools.

### 🛠️ Technical Notes
* ✨ Implemented `internal/launcher` with detached process isolation (`syscall.SysProcAttr{Setsid: true}`) so opened terminal windows persist independently across daemon restarts.
* ✨ Added terminal fallback cascade validating executable existence on disk before dispatch.
* ✨ Integrated click event handling in `internal/tray` with asynchronous non-blocking process dispatch.

## [v0.2.0] - 2026-10-20

The first feature milestone of dmesg-watch, delivering a lightweight, resource-conscious Linux background daemon for real-time kernel error monitoring with native desktop tray integration.

### Highlights
* **Real-time Kernel Error Monitoring**: Streams kernel alerts in real time (`journalctl -k -f -n 0 -p 3`) without terminal polling or manual log inspection.
* **Quiet by Default**: Background tray dashboard tracks error rates across a rolling 60-second window, alerting only when error thresholds are crossed.
* **Clean Startup State**: Launches directly in a calm, zero-error state without replaying historical journal entries or triggering false-positive alerts upon startup.
* **Dynamic Calm/Alert Icon States**: Displays a subtle monochrome monitor outline when healthy (0 errors in 60s), transitioning to an amber warning icon during active faults.
* **Smart Notification Cooldown**: Desktop notification dispatch throttles bursts to prevent notification flooding while keeping the tray menu continuously updated.
* **Human-Readable Desktop Integration**: Native StatusNotifierItem (SNI) integration with clean, two-line hover tooltips (`dmesg watch`) and dynamic context menu error diagnostics.
* **Universal Static Binaries**: Pure Go implementation compiled statically (`CGO_ENABLED=0`) across `.deb`, `.rpm`, and standalone archives, with zero runtime library or GUI dependencies.
* **Seamless Systemd Service**: Complete systemd user service integration with automatic restart and desktop graphical session target binding.

### 🛠️ Technical Notes
* Implemented concurrency-safe sliding-window time queue in `internal/tracker` with sub-millisecond evaluation.
* Direct D-Bus desktop notifications via `org.freedesktop.Notifications` with clickable link support.
* Direct StatusNotifierItem export over session D-Bus with minimal memory footprint (~15-20MB RSS).
* Fixed systemd user unit `ExecStart` target in `.deb` and `.rpm` packages to point to `/usr/bin/dmesg-watch` while maintaining local path substitution for `make install`.
* Added `-n 0` to background `journalctl` process invocation, overriding systemd's default 10-line history buffer and jumping straight to the journal tail.
* Multi-architecture Linux packages (`amd64` and `arm64`) generated automatically via GoReleaser.

## [v0.1.0] - 2026-10-20

* Initial repository scaffold and build system setup.
