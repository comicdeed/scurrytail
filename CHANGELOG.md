# Changelog

All notable changes to **ScurryTail** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.2-beta.1] - 2026-10-28

Maintenance release addressing child process leaks, D-Bus session shutdown hangs, log streaming resilience, runner script security, and YAML config layering.

### ✨ Highlights
* **✨ Zero Zombie Process Leaks**: Terminal viewer and browser subprocesses are now cleanly reaped in the background, eliminating `<defunct>` process leaks over long desktop sessions.
* **✨ Unblocked Session Shutdown**: Decoupled operating system signal handling into a dedicated listener, ensuring instant daemon shutdown even when desktop D-Bus tray services stall.
* **✨ Reliable YAML Log Settings**: Restored full precedence for `log_level` and `log_target` defined in YAML configuration and `conf.d/` drop-ins.
* **✨ Complete Category Menus**: All categories defined via flat `patterns` rules are dynamically registered in the tray context submenu.
* **✨ Hardened Runner Script & D-Bus Safety**: Runner scripts securely shell-quote multi-word arguments, runtime tmp directories verify ownership, and UTF-8 truncations are rune-safe across all D-Bus IPC strings.

### 🛠️ Technical Notes
* ✨ Reaped terminal and browser child processes asynchronously with `cmd.Wait()` in `internal/launcher`.
* ✨ Streamlined tray state transitions with unified `CalculateStatus` in `internal/tray`, eliminating duplicate state machine logic in `cmd/scurrytail`.
* ✨ Added 5-second `CallWithContext` timeout and bounded concurrency semaphore (capacity 3) in `internal/notify`.
* ✨ Expanded `internal/journal` scanner token buffer to 1 MiB and protected bracketed kernel tags (`[drm:...]`) in `CleanKernelMessage`.
* ✨ Added service hardening (`MemoryMax=64M`, `LockPersonality=true`, `RestrictRealtime=true`) across systemd units and Nix modules.
* ✨ Added comprehensive unit test suites for `notify`, `tray`, and rule precedence in `tracker`.

## [v0.3.1] - 2026-10-28

Documentation alignment for standard double-hyphen command-line flags.

### Highlights
* **Standardized CLI Flag Documentation**: Fixed outdated single-hyphen flag references across `README.md` and `docs/user-guide.md` (`--backend`, `--threshold`, `--window`, `--viewer-cmd`) to strictly match POSIX conventions and the application's `--help` output.

## [v0.3.0] - 2026-10-28

Major milestone release rebranding to ScurryTail, introducing subsystem error classification, modular conf.d/ configuration, silent background operation with attention-requesting tray alerts, pure-Go syslog auditing, and correlated terminal log inspection.

### Highlights
* **Rebranded to ScurryTail**: Complete evolution to ScurryTail (`scurrytail`) with an agile squirrel mascot, updated desktop integration, unified CLI, and session services.
* **Subsystem Error Classification**: Classifies kernel errors into targeted categories (`storage`, `gpu`, `wifi`, `dock`, `usb`) with independent sliding evaluation windows, alert thresholds, and configurable notification urgency levels (`critical`, `normal`, `low`).
* **Benign Hardware Noise Suppression**: Suppress recurring, harmless kernel chatter (such as buggy ACPI DSDT tables, KVM MSR reads, or firmware blips) via top-level or category `ignore` lists with zero lock contention.
* **Modular `conf.d/` Configuration Drop-ins**: Drop modular configuration files into `/etc/scurrytail/conf.d/*.yaml` and `~/.config/scurrytail/conf.d/*.yaml` with automatic lexical sorting, scalar overriding, and category pattern merging.
* **Silent by Default with Attention States**: Completely hidden from the desktop panel when healthy (`Passive`), emerging as an amber warning icon during active faults (`Active`), and flashing/pulsing (`NeedsAttention`) with persistent desktop alerts during high-priority bursts.
* **Priority-Latching Safeguard**: High-urgency alerts (such as drive I/O errors or GPU falloffs) latch the tray menu snippet and dynamic tooltip, ensuring transient background noise cannot mask critical hardware degradation.
* **Interleaved Audit Log Viewer**: Tray menu's **View Kernel Logs** opens a dedicated terminal streaming kernel errors chronologically interleaved with ScurryTail's state transitions, threshold triggers, and cooldown resolutions.
* **Universal Terminal Dispatch via Runtime Runner**: Viewer commands generate a runtime runner script in memory-backed user tmpfs (`$XDG_RUNTIME_DIR/scurrytail/view-logs.sh`), guaranteeing reliable dispatch across Ghostty, Konsole, GNOME Terminal, Alacritty, Kitty, and WezTerm without CLI argument stripping.
* **Pure-Go Syslog & Console Logging**: Integrated `codemedic/go-log` with dual support for system syslog (`/dev/log` with tag `scurrytail`) in daemon mode and human-readable timestamps in interactive terminal sessions.
* **Kernel-Style Batched Debounce Summaries**: Silent in-memory counting during notification cooldowns with concise boundary summaries upon cooldown expiration, eliminating flood spam.

### 🛠️ Technical Notes
* Implemented lock-free two-tier pattern matching in `internal/tracker` with a zero-allocation `strings.Contains` fast-path for literal patterns and compiled RE2 for regexes.
* Added bounded circular buffers (`maxCapacity`) and copy-to-head slice truncation in sliding-window trackers, preventing memory leaks during extreme error floods.
* Pre-allocated dynamic category menu items in `internal/tray` to eliminate D-Bus `LayoutUpdated` storm loops and item leaks.
* Adopted `github.com/ardanlabs/conf/v3` with YAML file overlay (`yaml.WithData`) for declarative hierarchical configuration, env vars (`SCURRYTAIL_*`), and CLI flag parsing.
* Decoupled desktop notification dispatch into an asynchronous background goroutine to eliminate D-Bus RPC head-of-line blocking.
* Packaged with declarative systemd user preset (`50-scurrytail.preset`) and automated service enablement across `.deb`, `.rpm`, and standalone archives.
* Added Nix package derivation, Nix Flake outputs, and declarative Home Manager and NixOS system modules (`services.scurrytail`).

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
