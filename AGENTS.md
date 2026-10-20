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
- **`internal/journal`**: Subprocess streaming.
  - Handles `journalctl -k -f -n 0 -p 3` execution, stdout pipe scanning, and syslog/kernel header normalization.
- **`internal/tracker`**: Rate monitoring and state machine.
  - Maintains the concurrency-safe sliding-window queue of error timestamps. Evaluates rate thresholds, sliding-window expirations, and notification cooldowns.
- **`internal/notify`**: Desktop notification dispatcher.
  - Formats and sends alerts directly over the D-Bus session bus, supporting HTML anchor tags for clickable links.
- **`internal/tray`**: System tray interface.
  - Publishes the `StatusNotifierItem` interface, embeds default icon assets, dynamically updates menu labels, and routes user clicks.
- **`systemd/`**: User session service unit configurations (`graphical-session.target`).
- **`scripts/`**: Portable installation and maintenance scripts for standalone archives.
- **`docs/`**: Detailed procedural and architectural documentation.
- **`.agents/`**: Coding agent skills and curation playbooks.

---

## 🛠️ Key Developer Mechanics

### Make Targets
- **`make build`**: Builds the binary to `bin/dmesg-watch` with embedded Git version tags.
- **`make test`**: Runs unit tests with race detection (`go test -v -race ./...`).
- **`make vet`**: Runs Go static analysis (`go vet ./...`).
- **`make dist`**: Packages a standalone `.tar.gz` distribution archive with service units and installer script.
- **`make install` / `make uninstall`**: Installs or removes binary and systemd user unit in `$HOME/.local/bin` and `$HOME/.config/systemd/user`.
- **`make service-enable` / `make service-status`**: Controls and inspects the live user systemd service.

### Formatting & Linting
- Always format modified files using `gofmt -w -s .` before running tests.
- Verify GitHub Actions workflow syntax using `actionlint .github/workflows/*.yml`.

### Local CI Testing
- Run `act pull_request` to execute the full GitHub Actions pipeline locally inside Docker before pushing changes.

---

## 🌿 Git & Release Workflows

- **Branching Model**: Trunk-based development. Development occurs directly on `main` or short-lived feature branches.
- **Commit Messages**: Use **Conventional Commits** (`feat:`, `fix:`, `chore:`, `refactor:`, `test:`, `docs:`). Never stage or commit changes without explicit user directives.
- **Curating Release Notes**: When preparing a release, consult the [Release Notes Skill](.agents/skills/release-notes/SKILL.md) to author concise, outcome-first notes in [CHANGELOG.md](CHANGELOG.md).
- **Fast Tag-Push Releases (Betas, Pre-releases, Daily Builds)**:
  1. Update `CHANGELOG.md` and version constant in `cmd/dmesg-watch/main.go` on `main`.
  2. Commit: `git commit -am "chore(release): prepare vX.Y.Z-beta.N"`.
  3. Tag and push: `git tag -a vX.Y.Z-beta.N -m "Release vX.Y.Z-beta.N" && git push origin main --tags`.
  4. GitHub Actions `release.yml` triggers immediately on the tag push (`push: tags: 'v*'`), runs GoReleaser, packages artifacts, applies Convention 1 labels, and prunes older pre-releases.
- **Stable Releases**: Dedicated `release/*` branches and reviewed PRs can optionally be used for major production milestones (`vX.Y.Z`). For full procedural details, consult the [Release Workflow Guide](docs/release-workflow.md).
