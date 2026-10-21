# Agent Guidelines

Welcome to **dmesg-watch**. This document defines architectural principles, component responsibilities, and operational mechanics for AI agents working in this codebase.

---

## 🏗️ Core Architecture & Boundaries

`dmesg-watch` is a resource-conscious, continuous Linux background daemon that monitors kernel errors in real time.

### Architectural Rules
- **Minimal Footprint**: Keep idle memory low (~15MB RSS). Do not introduce heavy graphical toolkits (Qt, GTK, Fyne windows).
- **Direct D-Bus IPC**: All tray and notification interactions use pure Go D-Bus bindings (`StatusNotifierItem` and `org.freedesktop.Notifications`).
- **Separation of Concerns**: Ingestion, time-window evaluation, and UI display are strictly decoupled via channels and thread-safe boundaries.

---

## 📂 What Goes Where

Avoid adding Go source files to the root. Organize new functionality into focused packages:

- **`cmd/`**: Entrypoint binaries.
  - `cmd/dmesg-watch/main.go` parses CLI flags, manages the root context lifecycle (`SIGINT`/`SIGTERM`), and wires domain packages together. Keep `main` focused on orchestration.
- **`internal/config`**: YAML and environment configuration.
  - Resolves XDG and system config paths (`~/.config/dmesg-watch/config.yaml`, `/etc/dmesg-watch/config.yaml`), unmarshals scalar/array options, and merges CLI flag overrides.
- **`internal/journal`**: Kernel log streaming.
  - Manages swappable kernel error streaming backends (`journalctl -k -f -n 0 -p 3` and `dmesg -w -l err,crit,alert,emerg`), backend auto-detection, and syslog/uptime/ISO timestamp normalization.
- **`internal/launcher`**: Process execution and terminal detection.
  - Launches external viewers in detached processes; implements terminal emulator detection cascade (Ghostty, KDE preferences, `$TERMINAL`, `xdg-terminal-exec`, etc.).
- **`internal/tracker`**: Rate monitoring and state machine.
  - Maintains the concurrency-safe sliding-window queue of error timestamps. Evaluates rate thresholds, sliding-window expirations, and notification cooldowns.
- **`internal/notify`**: Desktop notification dispatcher.
  - Formats and sends rate-threshold alerts directly over the D-Bus session bus with urgency levels and cooldown suppression.
- **`internal/tray`**: System tray interface.
  - Publishes the `StatusNotifierItem` interface, embeds default icon assets, dynamically updates menu labels (e.g. `View Kernel Logs (<backend>)`), and routes user clicks.
- **`nix/`**: Nix expressions and declarative modules.
  - Contains package derivation (`package.nix`), Home Manager module (`hm-module.nix`), and NixOS system module (`nixos-module.nix`).
- **`systemd/`**: User session service unit configurations (`graphical-session.target`).
- **`scripts/`**: Portable installation and maintenance scripts for standalone archives.
- **`docs/`**: Detailed procedural and architectural documentation.
- **`.agents/`**: Coding agent skills and curation playbooks.

---

## 🛠️ Key Developer Mechanics

### Make Targets
- **`make build`**: Builds the binary to `bin/dmesg-watch` with embedded Git version tags.
- **`make fmt`**: Formats all Go source files using `gofmt` (using host toolchain or auto Docker fallback).
- **`make test`**: Runs unit tests with race detection (`go test -v -race ./...`).
- **`make vet`**: Runs Go static analysis (`go vet ./...`).
- **`make nix-build`**: Compiles the Nix package derivation (using host `nix` or auto Docker fallback).
- **`make nix-check`**: Validates Flake schema, dev shells, and NixOS/Home Manager modules.
- **`make nix-fmt`**: Formats all Nix source files using `nixpkgs-fmt`.
- **`make dist`**: Packages a standalone `.tar.gz` distribution archive with service units and installer script.
- **`make install` / `make uninstall`**: Installs or removes binary and systemd user unit in `$HOME/.local/bin` and `$HOME/.config/systemd/user`.
- **`make service-enable` / `make service-status`**: Controls and inspects the live user systemd service.

### Formatting & Linting
- **Go changes**: Always format modified Go files using `make fmt` (or `gofmt -w -s .`) before running tests.
- **Nix changes**: Format modified Nix files using `make nix-fmt` only when `.nix` files or `flake.nix` are touched.
- **Workflow changes**: Verify GitHub Actions workflow syntax using `actionlint .github/workflows/*.yml` only when `.github/workflows` is modified.
- **Targeted validation**: Do not run Nix checks (`make nix-check`, `make nix-build`) for standard Go code edits; only invoke Nix targets when altering Nix expressions, Flake outputs, or Go module dependencies (`go.mod`).

### Nix Development & Docker Fallback
- The `Makefile` dynamically detects if native `nix` is installed on the host. If missing, it transparently executes through the `nixos/nix:latest` container.
- For complete operational details and instructions on updating `vendorHash` when Go dependencies change, consult the [Nix Development Guide](docs/nix-docker-guide.md).

### Local CI Testing
- Run `act pull_request` to execute the full GitHub Actions pipeline locally inside Docker before pushing changes.

---

## 📖 Documentation Standards & Target Audience Tuning

When writing or updating user-facing documentation (`README.md`, `docs/user-guide.md`):

- **Target Audience**: Tech-savvy Linux desktop users, system administrators, and hardware tinkerers.
- **Core Principle: "Less is More"**: Provide maximum signal per sentence. Eliminate marketing buzzwords, conversational filler, and unnecessary adjectives.
- **Outcome-First Explanations**: Explain what the user sees, what the tray icons mean, and how to operate the tool. Avoid leaking developer internals (channels, goroutines, D-Bus signatures) into end-user docs.
- **Clear Separation**: Maintain architectural principles and internal mechanics in `AGENTS.md`; maintain operational instructions in `docs/user-guide.md` and `README.md`.

### 📄 README.md Lifecycle & Maintenance Rules

`README.md` is strictly the project storefront. Keep it under ~50 lines following this fixed layout:
1. **Title & One-Liner Hook**: Calm, confident description of what the tool does (e.g. ~15MB RAM footprint, intelligent burst flood dampening). Avoid theatrical melodrama (*"dead silent until errors strike"*).
2. **Visual Anchor**: Screenshot preview (`docs/assets/screenshot.png`).
3. **`## Highlights`**: Exactly 5 bullet points with emoji prefixes. Lead with the user benefit, follow with the technical mechanism.
4. **`## Quick Start`**: Exactly 3 concise one-liners (Distro Packages with release link, Nix/NixOS runner, Source build).
5. **Handoff**: A single link pointing to `docs/user-guide.md` for deep configuration and troubleshooting.

#### Feature Additions & Modification Protocol
- **New CLI Flags or Config Options**: Add to `docs/user-guide.md` only. **Never** add raw flag tables or YAML schemas to `README.md`.
- **New Major Capabilities**: Fully document in `docs/user-guide.md`. Only update `README.md` if the change replaces one of the top 5 flagship capabilities; maintain the strict 5-bullet ceiling in `## Highlights`.
- **Packaging Changes**: Only update `Quick Start` in `README.md` if a primary installation path changes.

---

## 🌿 Git & Release Workflows

- **Branching Model**: Trunk-based development. Development occurs directly on `main` or short-lived feature branches.
- **Commit Messages**: Use **Conventional Commits** (`feat:`, `fix:`, `chore:`, `refactor:`, `test:`, `docs:`). Never stage or commit changes without explicit user directives.
- **Release Management**: When preparing any release (Pre-Release or Formal Release), consult the [Release Management Skill](.agents/skills/release-management/SKILL.md) which orchestrates version bumps, changelog notes, validation gates, git mechanics, and post-publish verification.
- **Fast Tag-Push Releases (Betas, Pre-releases, Daily Builds)**:
  1. Update `CHANGELOG.md` and version constant in `cmd/dmesg-watch/main.go` on `main`.
  2. Commit: `git commit -am "chore(release): prepare vX.Y.Z-beta.N"`.
  3. Tag and push: `git tag -a vX.Y.Z-beta.N -m "Release vX.Y.Z-beta.N" && git push origin main --tags`.
  4. GitHub Actions `release.yml` triggers immediately on the tag push (`push: tags: 'v*'`), runs GoReleaser, packages artifacts, applies Convention 1 labels, and prunes older pre-releases.
- **Stable Releases**: Dedicated `release/*` branches and reviewed PRs can optionally be used for major production milestones (`vX.Y.Z`). For full procedural details, consult the [Release Workflow Guide](docs/release-workflow.md).
