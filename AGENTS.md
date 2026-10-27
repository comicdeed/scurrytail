# Agent Guidelines

Welcome to **ScurryTail**. This document defines architectural principles, component responsibilities, and operational mechanics for AI agents working in this codebase.

---

## 🏗️ Core Architecture & Boundaries

`scurrytail` is a resource-conscious, continuous Linux background daemon that monitors kernel errors in real time.

### Architectural Rules
- **Minimal Footprint**: Keep idle memory low (~15MB RSS). Do not introduce heavy graphical toolkits (Qt, GTK, Fyne windows).
- **Direct D-Bus IPC**: All tray and notification interactions use pure Go D-Bus bindings (`StatusNotifierItem` and `org.freedesktop.Notifications`).
- **Separation of Concerns**: Ingestion, time-window evaluation, and UI display are strictly decoupled via channels and thread-safe boundaries.

---

## 📂 What Goes Where

Avoid adding Go source files to the root. Organize new functionality into focused packages:

- **`cmd/`**: Entrypoint binaries.
  - `cmd/scurrytail/main.go` parses CLI flags, manages the root context lifecycle (`SIGINT`/`SIGTERM`), and wires domain packages together. Keep `main` focused on orchestration.
- **`internal/config`**: Declarative configuration and env var parsing.
  - Powered by `github.com/ardanlabs/conf/v3` with YAML file overlay (`yaml.WithData`). Resolves XDG/system paths, unmarshals scalar/array options, and binds CLI flags and `SCURRYTAIL_*` environment variables.
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
- **`make build`**: Builds the binary to `bin/scurrytail` with embedded Git version tags.
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
- **Workflow changes**: Run `actionlint .github/workflows/*.yml` **only if** `actionlint` is already installed natively on the host. Never spin up Docker containers or pull Nix closures over the network just to run `actionlint`. If uninstalled, perform a direct manual syntax review.
- **Targeted validation**: Never run Nix checks (`make nix-check`, `make nix-build`) for standard Go code or workflow edits; only invoke Nix targets when altering Nix expressions, Flake outputs, or Go module dependencies (`go.mod`).
- **Strict Nix Boundary Isolation**: Nix (`nix`, `make nix-*`, `nixos/nix` container) is strictly for Nix packaging and flake maintenance. Never reach for Nix to run generic developer utilities, linters, or formatters (e.g., `nix run nixpkgs#actionlint`). CI workflows and non-Nix code have zero dependency on Nix.

### 🐹 Go Toolchain & Version Management
- **Single Source of Truth**: `go.mod` is the single authority for the project's minimum and active Go version.
- **Automated Workflow Alignment**: Never hardcode static `go-version` strings in GitHub Actions. Always configure `actions/setup-go@v5` with `go-version-file: 'go.mod'` across all `.github/workflows/*.yml` files.
- **Docker Fallback Synchronization**: The `Makefile` dynamically dispatches through `$(GOLANG_IMAGE)` (defaulting to `golang:alpine`) when native `go` is missing from the host. When bumping `go.mod`, ensure the toolchain image satisfies the declared version.
- **Nixpkgs Toolchain Alignment**: When `go.mod` specifies a Go version (e.g., `go 1.27.2`), `nixpkgs` default `pkgs.go` and `pkgs.buildGoModule` may lag behind on an older stable release (e.g., `go 1.26.9`). To prevent build failures:
  - In `nix/package.nix`, select the explicit versioned builder (e.g., `buildGo127Module`) or override `buildGoModule.override { go = pkgs.go_1_27; }`.
  - In `flake.nix`, ensure `devShells` includes the matching versioned package (e.g., `pkgs.go_1_27`).

### ⚙️ Configuration Management & Verification
- **Unified Declarative Config**: Configuration is managed declaratively using `github.com/ardanlabs/conf/v3` and `yaml.WithData`.
- **Precedence Hierarchy**: Built-in Struct Defaults $\rightarrow$ YAML Config File $\rightarrow$ Environment Variables $\rightarrow$ Command-Line Flags.
- **Mandatory `--help` Verification**: Never guess or hardcode derived environment variable names or CLI flag syntax. Always compile the binary and inspect `./bin/scurrytail --help` to verify the exact generated environment variables, flags, and default values before updating documentation, unit tests, or systemd services.

### Nix Development & Docker Fallback
- The `Makefile` dynamically detects if native `nix` is installed on the host. If missing, it transparently executes through the `nixos/nix:latest` container.
- For complete operational details and instructions on updating `vendorHash` when Go dependencies change, consult the [Nix Development Guide](docs/nix-docker-guide.md).
- **Dependency `vendorHash` Protocol**: Because Nix enforces hermetic, offline builds, dependencies are downloaded in an isolated Fixed-Output Derivation (FOD) verified by `vendorHash`.
  - Any addition, update, or removal of Go dependencies in `go.mod`/`go.sum` alters the module tree and invalidates `vendorHash`.
  - To refresh: set `vendorHash = lib.fakeHash;` in `nix/package.nix`, run `make nix-build` to capture the calculated hash (`got: sha256-...`), and replace it before running tests or tagging releases.

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
  1. Update `CHANGELOG.md` and version constant in `cmd/scurrytail/main.go` on `main`.
  2. Commit: `git commit -am "chore(release): prepare vX.Y.Z-beta.N"`.
  3. Tag and push: `git tag -a vX.Y.Z-beta.N -m "Release vX.Y.Z-beta.N" && git push origin main --tags`.
  4. GitHub Actions `release.yml` triggers immediately on the tag push (`push: tags: 'v*'`), runs GoReleaser, packages artifacts, applies Convention 1 labels, and prunes older pre-releases.
- **Stable Releases**: Dedicated `release/*` branches and reviewed PRs can optionally be used for major production milestones (`vX.Y.Z`). For full procedural details, consult the [Release Workflow Guide](docs/release-workflow.md).
