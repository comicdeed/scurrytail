# Changelog

All notable changes to **dmesg-watch** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.0-beta.8] - 2026-10-22

Unified declarative configuration powered by ardanlabs/conf/v3 with full environment variable support and Go 1.27 toolchain alignment.

### ✨ Highlights
* **✨ Declarative Configuration (`conf/v3`)**: Full declarative configuration parsing across CLI flags, environment variables (`DMESG_WATCH_*`), YAML files (`~/.config/dmesg-watch/config.yaml`), and built-in defaults with automatic `--help` generation.
* **✨ Full Environment Variable Support**: All configuration options (`DMESG_WATCH_BACKEND`, `DMESG_WATCH_CONFIG`, `DMESG_WATCH_THRESHOLD`, `DMESG_WATCH_WINDOW`, `DMESG_WATCH_COOLDOWN`, `DMESG_WATCH_VIEWER_CMD`) can now be passed via environment variables.
* **✨ Streamlined README**: Redesigned project storefront with screenshot visual anchor and high-signal, benefit-first highlights.
* **Pure Home Manager Module**: Declarative Home Manager module (`homeManagerModules.default`) configuring `~/.config/dmesg-watch/config.yaml` and launching the systemd user service under any Linux desktop.
* **NixOS System Module**: System-wide NixOS module (`nixosModules.default`) providing automatic desktop session integration (`graphical-session.target`).
* **Ad-Hoc Flake Runner**: Directly run `dmesg-watch` on any Nix-enabled system via `nix run github:codemedic/dmesg-watch`.
* **Dynamic Docker Nix Toolchain**: `Makefile` dynamically detects host `nix` or automatically falls back to containerized execution via `nixos/nix:latest`.

### 🛠️ Technical Notes
* ✨ Integrated `github.com/ardanlabs/conf/v3` with `yaml.WithData` enforcing strict hierarchy: Struct Defaults $\rightarrow$ YAML Config File $\rightarrow$ Environment Variables $\rightarrow$ CLI Flags.
* ✨ Implemented custom `conf.Setter` on `StringList` for CLI/env var parsing while maintaining YAML array/scalar backward compatibility.
* ✨ Switched Nix builder to `buildGo127Module`, refreshed `vendorHash`, and updated `flake.nix` devShells to `go_1_27`.
* ✨ Automated CI Go toolchain alignment dynamically via `actions/setup-go@v5` with `go-version-file: 'go.mod'`.
* ✨ Codified mandatory `--help` verification and Nix Go toolchain alignment rules into `AGENTS.md`.

## [v0.3.0-beta.7] - 2026-10-21

Adds complete declarative Nix and NixOS support for Home Manager, NixOS system configuration, and ad-hoc Flake runners.

### ✨ Highlights
* **✨ Pure Home Manager Module**: Added declarative Home Manager module (`homeManagerModules.default`) configuring `~/.config/dmesg-watch/config.yaml` and launching the systemd user service under any Linux desktop.
* **✨ NixOS System Module**: Added system-wide NixOS module (`nixosModules.default`) providing automatic desktop session integration (`graphical-session.target`) and system-wide default configuration.
* **✨ Ad-Hoc Flake Runner**: Directly run `dmesg-watch` on any Nix-enabled system via `nix run github:codemedic/dmesg-watch`.
* **✨ Dynamic Docker Nix Toolchain**: `Makefile` dynamically detects host `nix` or automatically falls back to containerized execution via `nixos/nix:latest` for `make nix-build`, `make nix-check`, and `make nix-fmt`.
* **YAML Configuration File**: Added native support for `~/.config/dmesg-watch/config.yaml` with flexible array support for `viewer_cmd`, allowing background service users to customize settings without editing systemd unit files.
* **Dynamic Tray Log Label**: Fixed tray menu to dynamically reflect the active streaming backend (`View Kernel Logs (dmesg)` vs `View Kernel Logs (journalctl)`).
* **Swappable Streaming Backends**: Added `-backend` flag (`auto`, `journalctl`, `dmesg`) to stream directly via `dmesg -W` on non-systemd or embedded systems with automatic viewer command synchronization.
* **Dedicated User Guide**: Created `docs/user-guide.md` providing high-signal, plain-English guidance on icon states, configuration, permissions, and service management.

### 🛠️ Technical Notes
* ✨ Implemented `nix/package.nix` with static `CGO_ENABLED=0` build and `makeWrapper` for runtime binary wrapping (`systemd`, `util-linux`, `xdg-utils`).
* ✨ Implemented `nix/hm-module.nix` and `nix/nixos-module.nix` with typed option definitions and YAML serialization.
* ✨ Added `docs/nix-docker-guide.md` and updated `AGENTS.md` with targeted validation workflows.
* ✨ Aligned Go toolchain to `go 1.27.2` with automated `go-version-file` synchronization across CI workflows.

## [v0.3.0-beta.6] - 2026-10-21

Adds YAML configuration file support, swappable streaming backends (-backend), dynamic tray labels, and dedicated user-guide documentation.

### ✨ Highlights
* **✨ YAML Configuration File**: Added native support for `~/.config/dmesg-watch/config.yaml` with flexible array support for `viewer_cmd`, allowing background service users to customize settings without editing systemd unit files.
* **✨ Dynamic Tray Log Label**: Fixed tray menu to dynamically reflect the active streaming backend (`View Kernel Logs (dmesg)` vs `View Kernel Logs (journalctl)`).
* **✨ Swappable Streaming Backends**: Added `-backend` flag (`auto`, `journalctl`, `dmesg`) to stream directly via `dmesg -W` on non-systemd or embedded systems with automatic viewer command synchronization.
* **✨ Dedicated User Guide**: Created `docs/user-guide.md` providing high-signal, plain-English guidance on icon states, configuration, permissions, and service management.
* **Safe Non-Systemd Abort**: Added declarative package dependency on `systemd` and a `preinstall` verification hook to halt `.deb` / `.rpm` installation on non-systemd systems (e.g. Devuan, antiX, SysVinit, OpenRC) before any files are unpacked.
* **Automatic Service Enablement**: Packages (`.deb`, `.rpm`) and installer scripts now automatically enable and start `dmesg-watch.service` under the active desktop user session immediately upon installation.

### 🛠️ Technical Notes
* ✨ Implemented `internal/config` supporting XDG config discovery and YAML unmarshaling with `gopkg.in/yaml.v3`.
* ✨ Updated `internal/launcher` with `LaunchArgs` and direct argument list execution.
* ✨ Fixed GoReleaser v2 nfpm packaging schema (`dependencies:`).
* ✨ Codified "Less is More" documentation guidelines in `AGENTS.md`.

## [v0.3.0-beta.5] - 2026-10-21

Adds multi-layer systemd and journalctl prerequisite validation to safely abort installation on non-systemd systems.

### ✨ Highlights
* **✨ Safe Non-Systemd Abort**: Added declarative package dependency on `systemd` and a `preinstall` verification hook to halt `.deb` / `.rpm` installation on non-systemd systems (e.g. Devuan, antiX, SysVinit, OpenRC) before any files are unpacked.
* **✨ Installer Prerequisite Verification**: Added prerequisite checks to `scripts/install.sh` and startup checks to `cmd/dmesg-watch/main.go` verifying that `journalctl` and a running systemd instance are present.
* **GUI Software Center Support**: Enhanced package lifecycle hooks to discover active desktop sessions via `/run/user/*` instead of relying on `$SUDO_USER`, enabling automatic service activation when installing through KDE Discover, GNOME Software, or PackageKit.
* **Automatic Service Enablement**: Packages (`.deb`, `.rpm`) and installer scripts now automatically enable and start `dmesg-watch.service` under the active desktop user session immediately upon installation.
* **Declarative Systemd Preset**: Installed a system-wide user preset (`50-dmesg-watch.preset`) ensuring the service starts automatically on future desktop logins.
* **Native About Submenu & Website Launcher**: Converted the About item into a native tray submenu featuring application details and a "Website" action that opens the project repository in your default web browser via `xdg-open`.
* **Decluttered Top-Level Menu**: Relocated static version metadata into the About submenu and adopted standard desktop single-word actions (`About`, `Quit`).
* **Quick Drill-Down to Kernel Logs**: Added a "View Kernel Logs (journalctl)" action to the system tray menu to instantly open full kernel logs in your default terminal with continuous live tailing (`journalctl -k -f -n 50`).
* **Smart Terminal Emulator Detection**: Automatically detects Ghostty, KDE Plasma desktop terminal preferences (`kdeglobals`), `$TERMINAL`, `xdg-terminal-exec`, and common desktop terminals.

### 🛠️ Technical Notes
* ✨ Implemented `scripts/preinstall.sh` validating `command -v journalctl` and the standard `sd_booted` directory check (`/run/systemd/system`).
* ✨ Added `depends: [systemd]` to nfpm package definition in `.goreleaser.yaml`.
* ✨ Added runtime `exec.LookPath("journalctl")` guard in `cmd/dmesg-watch/main.go`.

## [v0.3.0-beta.4] - 2026-10-20

Adds dynamic active-session discovery to ensure automatic service activation works under GUI package managers (Discover, GNOME Software, PackageKit).

### ✨ Highlights
* **✨ GUI Software Center Support**: Enhanced package lifecycle hooks to discover active desktop sessions via `/run/user/*` instead of relying on `$SUDO_USER`, enabling automatic service activation when installing through KDE Discover, GNOME Software, or PackageKit.
* **Automatic Service Enablement**: Packages (`.deb`, `.rpm`) and installer scripts now automatically enable and start `dmesg-watch.service` under the active desktop user session immediately upon installation.
* **Declarative Systemd Preset**: Installed a system-wide user preset (`50-dmesg-watch.preset`) ensuring the service starts automatically on future desktop logins.
* **Native About Submenu & Website Launcher**: Converted the About item into a native tray submenu featuring application details and a "Website" action that opens the project repository in your default web browser via `xdg-open`.
* **Decluttered Top-Level Menu**: Relocated static version metadata into the About submenu and adopted standard desktop single-word actions (`About`, `Quit`).
* **Quick Drill-Down to Kernel Logs**: Added a "View Kernel Logs (journalctl)" action to the system tray menu to instantly open full kernel logs in your default terminal with continuous live tailing (`journalctl -k -f -n 50`).
* **Smart Terminal Emulator Detection**: Automatically detects Ghostty, KDE Plasma desktop terminal preferences (`kdeglobals`), `$TERMINAL`, `xdg-terminal-exec`, and common desktop terminals.
* **Configurable Viewer Command**: Added a `--viewer-cmd` CLI option allowing users to customize their log viewer command.

### 🛠️ Technical Notes
* ✨ Refactored `scripts/postinstall.sh` and `scripts/prerm.sh` to iterate through `/run/user/*` to target active systemd user managers without dependency on terminal sudo environment variables.
* Bundled `/usr/lib/systemd/user-preset/50-dmesg-watch.preset` across `.deb`, `.rpm`, AUR, and standalone tarballs.

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
