# Changelog

All notable changes to **ScurryTail** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.0-beta.16] - 2026-10-28

Runtime runner script generation for reliable terminal dispatch and scoped kernel priority filtering.

### ✨ Highlights
* **✨ Runtime Runner Script Generation**: System tray log viewer now generates an executable runner script (`$XDG_RUNTIME_DIR/scurrytail/view-logs.sh`) to launch viewer commands, completely resolving terminal CLI argument parsing quirks and character stripping across Ghostty, Konsole, GNOME Terminal, and other emulators.
* **✨ Full Interleaved Audit Stream Visibility**: Scoped kernel priority filtering to `_TRANSPORT=kernel PRIORITY=0 PRIORITY=1 PRIORITY=2 PRIORITY=3`, ensuring ScurryTail's `INFO`-level lifecycle and state telemetry entries are never filtered out by global priority flags.
* **Interleaved Audit Log Viewer**: Upgraded the system tray's **View Kernel Logs (journalctl)** action to automatically stream kernel errors interleaved with ScurryTail's state evaluations.

### 🛠️ Technical Notes
* ✨ Implemented `PrepareRunnerScript` in `internal/launcher/launcher.go` with atomic file generation in private user runtime tmpfs.
* ✨ Updated `internal/journal/stream.go` and `internal/launcher/launcher.go` default viewer command to use per-field priority filters.
* ✨ Updated unit tests in `internal/launcher/launcher_test.go` to verify runner script generation and execution specs.

## [v0.3.0-beta.15] - 2026-10-28

Interleaved kernel and ScurryTail audit correlation in tray log viewer, robust Makefile toolchain dispatch, and documentation refinements.

### ✨ Highlights
* **✨ Interleaved Audit Log Viewer**: Upgraded the system tray's **View Kernel Logs (journalctl)** action to automatically stream kernel errors interleaved with ScurryTail's state evaluations (`journalctl -b _TRANSPORT=kernel -p 3 + -b _SYSTEMD_USER_UNIT=scurrytail.service + -b SYSLOG_IDENTIFIER=scurrytail -n 50 -f`), providing instant cause-and-effect clarity.
* **✨ Subsystem Tag Preservation & Filtering**: Correlated viewer strictly displays error-level kernel messages (`-p 3`), preventing background Wi-Fi and Bluetooth info chatter from washing out hardware faults.
* **Pure-Go Syslog & Console Logging**: Integrated `codemedic/go-log` with dual support for system syslog (`/dev/log` with tag `scurrytail`) in background/daemon mode and direct console logging in interactive terminal sessions.
* **High-Signal State Auditing**: Emits lifecycle state transitions (`Passive` $\leftrightarrow$ `Active` $\leftrightarrow$ `NeedsAttention`) and threshold alerts to syslog/journald without duplicating raw kernel log messages.

### 🛠️ Technical Notes
* ✨ Configured `internal/journal/stream.go` and `internal/launcher/launcher.go` with unified default correlated journalctl viewer command.
* ✨ Updated `internal/tray/tray.go` with contextual tooltip indicating correlated audit stream under `journalctl`.
* ✨ Refined `Makefile` toolchain resolution to pair `GOFMT` directly with detected `GO_BIN`.

## [v0.3.0-beta.14] - 2026-10-28

Pure-Go syslog logging integration via codemedic/go-log, high-signal activity and state transition auditing, kernel-style batched debounce summaries, and non-blocking desktop notifications.

### ✨ Highlights
* **✨ Pure-Go Syslog & Console Logging**: Integrated `codemedic/go-log` with dual support for system syslog (`/dev/log` with tag `scurrytail`) in background/daemon mode and direct console logging in interactive terminal sessions.
* **✨ High-Signal State Auditing**: Emits lifecycle state transitions (`Passive` $\leftrightarrow$ `Active` $\leftrightarrow$ `NeedsAttention`) and threshold alerts to syslog/journald without duplicating raw kernel log messages.
* **✨ Enriched Wakeup Attribution**: Initial transition from clean state (`Passive -> Active`) identifies the specific triggering error category and message snippet.
* **✨ Kernel-Style Batched Debounce Summaries**: Replaced noisy per-message throttling with silent in-memory counting and single boundary summaries upon cooldown expiration.
* **✨ Priority Latching in Audit Trail**: Explicitly records when an alert threshold is met while a higher-urgency alert holds the tray latch.
* **Subsystem Error Categories**: Classify errors into dedicated categories (`storage`, `gpu`, `wifi`, `dock`, `usb`) with independent sliding windows, alert thresholds, and notification urgency levels (`critical`, `normal`, `low`).

### 🛠️ Technical Notes
* ✨ Added `github.com/codemedic/go-log v1.1.0` with zero external dependencies and pure-Go `log/syslog` backing (`CGO_ENABLED=0`).
* ✨ Decoupled desktop notification dispatch into an asynchronous background goroutine to eliminate D-Bus RPC head-of-line blocking during floods.
* ✨ Consolidated `Tracker.RecordAndAlert()` to acquire locks and evaluate prune cutoffs in a single atomic pass.
* ✨ Added `--log-level` (`SCURRYTAIL_LOG_LEVEL`) and `--log-target` (`SCURRYTAIL_LOG_TARGET`) configuration options.
* ✨ Refreshed hermetic Nix package derivation `vendorHash`.

## [v0.3.0-beta.13] - 2026-10-28

Subsystem pattern classification, benign error suppression, pure hierarchical YAML layering via conf.d/, and priority-latched desktop alerts.

### ✨ Highlights
* **✨ Subsystem Error Categories**: Classify errors into dedicated categories (`storage`, `gpu`, `wifi`, `dock`, `usb`) with independent sliding windows, alert thresholds, and notification urgency levels (`critical`, `normal`, `low`).
* **✨ Benign Noise Suppression**: Suppress recurring, harmless kernel chatter (such as buggy ACPI DSDT tables, KVM MSR reads, or firmware blips) via top-level or category `ignore` lists with zero lock contention.
* **✨ Layerable `conf.d/` Drop-ins**: Drop modular configuration files into `/etc/scurrytail/conf.d/*.yaml` and `~/.config/scurrytail/conf.d/*.yaml` with automatic lexical sorting, scalar overriding, and category pattern merging.
* **✨ Priority-Latching Safeguard**: Critical alerts (such as drive I/O errors) latch the tray menu snippet and dynamic tooltip, preventing transient lower-urgency events from masking urgent faults.
* **✨ Dmesg Subsystem Tag Preservation**: Fixed kernel log message normalization across `dmesg` and `journalctl`, ensuring driver subsystem prefixes (`EXT4-fs`, `iwlwifi`, `kvm`) match consistently across backends.

### 🛠️ Technical Notes
* ✨ Implemented lock-free two-tier pattern matching in `internal/tracker` with a zero-allocation `strings.Contains` fast-path for literal patterns and compiled RE2 for regexes.
* ✨ Added bounded circular buffers (`maxCapacity`) and copy-to-head slice truncation in sliding-window trackers, eliminating memory leaks and bounding memory during extreme error floods.
* ✨ Pre-allocated dynamic category menu items in `internal/tray` to prevent D-Bus `LayoutUpdated` storm loops and item leaks.
* ✨ Throttled tray UI updates to at most 4Hz (every 250ms) using a coalescing ticker in `cmd/scurrytail/main.go`.
* ✨ Mapped host user permissions (`UID:GID`) in Makefile Docker toolchain fallbacks to prevent creating root-owned build artifacts.

## [v0.3.0-beta.12] - 2026-10-27

Rebranding to ScurryTail with new agile squirrel mascot, updated CLI binary and services, and theme-adaptive documentation.

### ✨ Highlights
* **✨ Rebranded to ScurryTail**: Renamed the project from `dmesg-watch` to `ScurryTail` (`scurrytail`), capturing our nimble, featherweight daemon tailing fast-moving kernel logs.
* **✨ Agile Mascot Imagery**: Added dark and light mode mascot banners dynamically adapting to the user's desktop/GitHub theme.
* **✨ Unified `scurrytail` Command & Service**: Executable binary renamed to `scurrytail` with systemd unit `scurrytail.service` and preset `50-scurrytail.preset`.
* **✨ `SCURRYTAIL_*` Environment Namespace**: Migrated environment variables to `SCURRYTAIL_BACKEND`, `SCURRYTAIL_THRESHOLD`, etc., and primary config to `~/.config/scurrytail/config.yaml` with legacy fallback support.
* **Clean Attention De-escalation**: Fixed D-Bus status desynchronization so the tray icon cleanly stops flashing and returns to standard warning state when error rates subside below threshold.

### 🛠️ Technical Notes
* ✨ Updated Go module path to `github.com/comicdeed/scurrytail` and entrypoint to `cmd/scurrytail/main.go`.
* ✨ Updated Nix Flake outputs, package derivations (`pname = "scurrytail"`), and Home Manager/NixOS modules (`services.scurrytail`).
* ✨ Aligned GoReleaser configuration with `scurrytail` artifacts and packages.

## [v0.3.0-beta.11] - 2026-10-27

Robust StatusNotifierItem state tracking engine with clean attention de-escalation on error decay and user guide visual state flow.

### ✨ Highlights
* **✨ Clean Attention De-escalation**: Fixed D-Bus status desynchronization so the tray icon cleanly stops flashing and returns to standard warning state when error rates subside below threshold.
* **✨ State Flow Documentation**: Added native Mermaid UML state machine diagram and UI behavior breakdown to the user guide under "How the Icon Behaves".
* **Quiet Background Presence**: Tray icon remains hidden by default when 0 errors are recorded (`Passive` StatusNotifierItem status), leaving the desktop panel clean and distraction-free.
* **Fault-Driven Dynamic Appearance**: Tray icon emerges into view (`Active` status) with an amber warning icon when kernel errors occur within the rolling window.
* **Desktop Attention on Threshold Breaches**: Requests desktop environment attention (`NeedsAttention` status) when kernel error counts reach or exceed the configured alert threshold.

### 🛠️ Technical Notes
* ✨ Refactored tray state machine in `internal/tray/tray.go` to use an explicit 3-state enum (`statusPassive`, `statusActive`, `statusNeedsAttention`), eliminating boolean permutations and duplicate D-Bus signal emissions.
* ✨ Centralized state transitions in `setStatusLocked()` ensuring proper call sequencing (`Show()` before `RequestAttention()` on direct bursts from hidden state).

## [v0.3.0-beta.10] - 2026-10-26

Tray icon is hidden by default on clean systems, appearing on faults and requesting desktop attention when error thresholds are crossed.

### ✨ Highlights
* **✨ Quiet Background Presence**: Tray icon remains hidden by default when 0 errors are recorded (`Passive` StatusNotifierItem status), leaving the desktop panel clean and distraction-free.
* **✨ Fault-Driven Dynamic Appearance**: Tray icon emerges into view (`Active` status) with an amber warning icon when kernel errors occur within the rolling window.
* **✨ Desktop Attention on Threshold Breaches**: Requests desktop environment attention (`NeedsAttention` status) when kernel error counts reach or exceed the configured alert threshold.
* **Conventional Distro Package Naming**: Adopted GoReleaser/NFPM conventional naming for Linux packages, removing redundant `linux` substrings and adhering strictly to Debian (`_amd64.deb`, `_arm64.deb`) and RedHat/RPM (`-1.x86_64.rpm`, `-1.aarch64.rpm`) conventions.
* **Declarative Configuration (`conf/v3`)**: Full declarative configuration parsing across CLI flags, environment variables (`DMESG_WATCH_*`), YAML files (`~/.config/dmesg-watch/config.yaml`), and built-in defaults with automatic `--help` generation.

### 🛠️ Technical Notes
* ✨ Switched `fyne.io/systray` dependency to `github.com/codemedic/fyne-io-systray` fork implementing `Show()`, `Hide()`, and `RequestAttention()` methods.
* ✨ Synchronized desktop attention requests with burst notification dispatches in `cmd/dmesg-watch/main.go`.
* ✨ Updated Nix package `vendorHash` in `nix/package.nix` for the replaced dependency closure.
* ✨ Pruned legacy RPM asset label matchers in release workflow.

## [v0.3.0-beta.9] - 2026-10-22

Adopts distro-native conventional package naming for deb and rpm packages and tightens developer workflow rules.

### ✨ Highlights
* **✨ Conventional Distro Package Naming**: Adopted GoReleaser/NFPM conventional naming for Linux packages, removing redundant `linux` substrings and adhering strictly to Debian (`_amd64.deb`, `_arm64.deb`) and RedHat/RPM (`-1.x86_64.rpm`, `-1.aarch64.rpm`) conventions.
* **✨ Clean Standalone Archive Names**: Streamlined standalone tarball naming to `dmesg-watch-<version>-<arch>.tar.gz`.
* **Declarative Configuration (`conf/v3`)**: Full declarative configuration parsing across CLI flags, environment variables (`DMESG_WATCH_*`), YAML files (`~/.config/dmesg-watch/config.yaml`), and built-in defaults with automatic `--help` generation.
* **Full Environment Variable Support**: All configuration options (`DMESG_WATCH_BACKEND`, `DMESG_WATCH_CONFIG`, `DMESG_WATCH_THRESHOLD`, `DMESG_WATCH_WINDOW`, `DMESG_WATCH_COOLDOWN`, `DMESG_WATCH_VIEWER_CMD`) can be passed via environment variables.
* **Pure Home Manager Module**: Declarative Home Manager module (`homeManagerModules.default`) configuring `~/.config/dmesg-watch/config.yaml` and launching the systemd user service under any Linux desktop.
* **NixOS System Module**: System-wide NixOS module (`nixosModules.default`) providing automatic desktop session integration (`graphical-session.target`).

### 🛠️ Technical Notes
* ✨ Configured `conventional_file_name: true` in `.goreleaser.yaml` and aligned `Makefile`'s `dist` target.
* ✨ Updated GitHub release asset labeling in `release.yml` to support `x86_64.rpm` and `aarch64.rpm`.
* ✨ Tightened workflow linting rules and strictly codified Nix boundary isolation in `AGENTS.md`.

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
