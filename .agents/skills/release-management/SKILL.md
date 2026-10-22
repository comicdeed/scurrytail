---
name: release-management
description: Comprehensive release lifecycle orchestrator for dmesg-watch with two operational modes Pre-Release (fast tag-push for betas and daily development) and Formal Release (cumulative milestone finalization for stable vX.Y.Z). Automates version bumps, changelog notes, validation gates, git mechanics, and GitHub Actions post-release verification.
---

# Release Management Skill

This skill defines the complete operational mechanics for cutting releases of **dmesg-watch** (`github.com/comicdeed/dmesg-watch`). It operates in two explicit modes:

- **Mode 1: Pre-Release (Fast Tag-Push)** — For rapid beta iterations, testing, and daily development builds.
- **Mode 2: Formal Release (Milestone Finalization)** — For major production milestones (`vX.Y.Z`).

---

## 🎯 Universal Quality & Packaging Gates

Before initiating any release in either mode, the working tree must satisfy the following gates:

1. **Working Tree Cleanliness**: `git status` must be clean, on `main`, and synchronized (`git pull`).
2. **Ecosystem-Native Formatting**: All Go sources must be formatted:
   ```bash
   gofmt -w -s .
   ```
3. **Validation & Race Detection**: Static checks and race tests must pass:
   ```bash
   make test && make vet
   ```
4. **Pure Static Binary Constraint**: The build must not require dynamic C libraries (`CGO_ENABLED=0`).
5. **No AI Attribution**: Changelogs, commit messages, and release notes must never include AI attribution footers or disclaimers.

---

## 🔢 Automated Version Bumping & SemVer Semantics

Version selection is derived systematically from **Conventional Commits** in git history, with an explicit confirmation gate.

### 1. Milestone Target Calculation
Inspect the commit log between the last stable tag (`LAST_STABLE`) and `HEAD`:

```bash
LAST_STABLE=$(git tag --sort=-creatordate | grep -v -E '(beta|alpha|rc)' | head -n 1)
git log "${LAST_STABLE}..HEAD" --oneline --no-merges
```

| Commit Markers in Delta | SemVer Classification | Target from `v0.2.0` | Initial Pre-Release Tag |
| :--- | :--- | :--- | :--- |
| Any `!` (`feat!:`, `fix!:`) or `BREAKING CHANGE:` | **MAJOR** (or SemVer §4 Minor in `0.y.z`) | `v1.0.0` | `v1.0.0-beta.1` |
| Any `feat:` without breaking markers | **MINOR** | `v0.3.0` | `v0.3.0-beta.1` |
| Only `fix:`, `perf:`, `refactor:`, `docs:`, `chore:` | **PATCH** | `v0.2.1` | `v0.2.1-beta.1` |

### 2. Handling Mid-Beta Cycle Breaking Changes
If a breaking change commit lands *while already in an active beta cycle*:

1. **Breaks Only an Unreleased Feature (Added in Earlier Beta)**:
   - Does **not** break previous stable release.
   - Retain milestone target; increment beta counter (e.g., `v0.3.0-beta.1` $\rightarrow$ `v0.3.0-beta.2`).
   - Add a `### ⚠️ Breaking Changes in Beta` section in `CHANGELOG.md` for beta testers.
2. **Breaks Contract with Previous Stable Release**:
   - Re-anchor milestone to next major (e.g., re-target `v0.3.0` $\rightarrow$ `v1.0.0`).
   - Reset beta sequence under new major target (e.g., `v1.0.0-beta.1`).
3. **Pre-1.0 SemVer Rule (0.y.z initial development, SemVer §4)**:
   - In `0.y.z`, bumping the minor version (`0.2.0` $\rightarrow$ `0.3.0`) is the conventional signal for breaking changes without prematurely triggering `v1.0.0`.

### 3. Pre-Release Tag Counter Logic
```bash
PREV_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")
```
* **New Milestone**: If `PREV_TAG` belongs to a previous milestone or stable release (e.g., `v0.2.0`), start at `.1`: `v0.3.0-beta.1`.
* **Same Milestone**: If `PREV_TAG` is an existing beta of the same target version (e.g., `v0.3.0-beta.1`), increment the beta counter: `v0.3.0-beta.2`.

### 4. Smart Confirmation Gate
Before applying the bump or creating tags, the agent must present the automated analysis and request confirmation:
* Summary of commit delta.
* Detected SemVer bump level (MAJOR / MINOR / PATCH).
* Recommended tag name.
* Option for user override (e.g. promoting a minor bump to `v1.0.0` major release).

---

## ⚡ Mode 1: Pre-Release Workflow (Betas & Daily Builds)

Use this mode when releasing candidate builds (`vX.Y.Z-beta.N`, `vX.Y.Z-alpha.N`, or `vX.Y.Z-rc.N`).

### 1. Identify Target Tag & Delta Scope
```bash
# Determine previous tag
PREV_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")

# Inspect incremental delta since previous tag
git log "${PREV_TAG}..HEAD" --oneline --no-merges
```

### 2. Update Version in Code
Update the default fallback version string in `cmd/dmesg-watch/main.go`:
```go
var version = "v0.2.0-beta.N"
```

### 3. Curate Living Draft Notes in `CHANGELOG.md`
Pre-releases maintain an evolving **Living Draft** of the milestone. Add a new section at the top of `CHANGELOG.md`:

```markdown
## [vX.Y.Z-beta.N] - YYYY-MM-DD

<1 sentence summarizing the milestone's primary focus>

### ✨ Highlights
* **<New Feature 1>**: <Outcome-first description prefixing items new to this build with ✨>.
* **<Existing Milestone Feature>**: <Previous milestone achievements carry forward without ✨>.

### 🛠️ Technical Notes
* <Architecture, D-Bus, systemd, or packaging changes (new items prefixed with ✨)>.
```

**Notes Rules:**
- Maximum 15–20 lines total.
- Outcome-first language describing what the user sees or feels.
- Keep technical details (D-Bus properties, systemd paths, static compilation) strictly under `### 🛠️ Technical Notes`.

### 4. Direct Commit, Tag, and Push
Execute the fast tag-push sequence directly on `main`:
```bash
git add cmd/dmesg-watch/main.go CHANGELOG.md internal/
git commit -m "chore(release): prepare vX.Y.Z-beta.N"
git tag -a vX.Y.Z-beta.N -m "Release vX.Y.Z-beta.N"
git push origin main --tags
```

### 5. Mode 1 Verification Checklist
- [ ] **GitHub Actions Run**: Run `gh run watch $(gh run list --workflow=release.yml --limit 1 --json databaseId --jq '.[0].databaseId')`.
- [ ] **Pre-Release Flag**: Verify `gh release view vX.Y.Z-beta.N` shows `Pre-release`.
- [ ] **Convention 1 Labels**: Verify assets match standard labels:
  - `.deb` (amd64) $\rightarrow$ `Debian / Ubuntu (x86_64)`
  - `.deb` (arm64) $\rightarrow$ `Debian / Ubuntu (aarch64 / ARM64)`
  - `.rpm` (x86_64) $\rightarrow$ `Fedora / RHEL / openSUSE (x86_64)`
  - `.rpm` (aarch64) $\rightarrow$ `Fedora / RHEL / openSUSE (aarch64 / ARM64)`
  - `.tar.gz` (amd64) $\rightarrow$ `Standalone Archive (x86_64)`
  - `.tar.gz` (arm64) $\rightarrow$ `Standalone Archive (aarch64 / ARM64)`
  - `checksums.txt` $\rightarrow$ `SHA256 Checksums`
- [ ] **Sliding Window Pruning**: Verify `gh release list` retains exactly the 3 most recent pre-releases.

---

## 🏛️ Mode 2: Formal Release Workflow (Stable Milestones)

Use this mode when releasing major production milestones (`vX.Y.Z`).

### 1. Identify Milestone Scope
```bash
# Find the preceding stable tag (excluding alpha/beta/rc)
LAST_STABLE=$(git tag --sort=-creatordate | grep -v -E '(beta|alpha|rc)' | head -n 1)

# Inspect full cumulative milestone delta
git log "${LAST_STABLE}..HEAD" --oneline --no-merges
```

### 2. Update Version in Code
Update the default fallback version string in `cmd/dmesg-watch/main.go`:
```go
var version = "vX.Y.Z"
```

### 3. Finalize Milestone Notes in `CHANGELOG.md`
Consolidate the living draft into the final milestone release notes:
1. Strip all temporary `✨` markers from the entire feature and technical list.
2. Ensure the milestone summary reflects the complete, verified release.
3. Place `## [vX.Y.Z] - YYYY-MM-DD` at the top of `CHANGELOG.md`.
4. Squash and replace the intermediate beta blocks into the single consolidated milestone entry, keeping prior stable milestone releases below for clean, noise-free packaging notes.

```markdown
## [vX.Y.Z] - YYYY-MM-DD

The first feature milestone of dmesg-watch, delivering a lightweight, resource-conscious Linux background daemon for real-time kernel error monitoring with native desktop tray integration.

### Highlights
* **Real-time Kernel Error Monitoring**: Streams kernel alerts in real time (`journalctl -k -f -n 0 -p 3`) without terminal polling or manual log inspection.
* **Quiet by Default**: Background tray dashboard tracks error rates across a rolling 60-second window, alerting only when error thresholds are crossed.
* **Dynamic Calm/Alert Icon States**: Displays a subtle monochrome monitor icon when healthy (0 errors in 60s), transitioning to an amber warning icon during active faults.
* **Smart Notification Cooldown**: Desktop notification dispatch throttles bursts to prevent notification flooding while keeping the tray menu continuously updated.
* **Human-Readable Desktop Integration**: Native StatusNotifierItem (SNI) integration with clean, two-line hover tooltips and dynamic context menu error diagnostics.
* **Universal Static Binaries**: Pure Go implementation compiled statically (`CGO_ENABLED=0`) across `.deb`, `.rpm`, and standalone archives, with zero runtime library or GUI dependencies.
* **Seamless Systemd Service**: Complete systemd user service integration with automatic restart and desktop graphical session target binding.

### 🛠️ Technical Notes
* Implemented concurrency-safe sliding-window time queue in `internal/tracker` with sub-millisecond evaluation.
* Direct D-Bus desktop notifications via `org.freedesktop.Notifications` with clickable link support.
* Direct StatusNotifierItem export over session D-Bus with minimal memory footprint (~15-20MB RSS).
* Multi-architecture Linux packages (`amd64` and `arm64`) generated automatically via GoReleaser.
```

### 4. Git Mechanics
- **Direct Tag-Push (Standard)**:
  ```bash
  git add cmd/dmesg-watch/main.go CHANGELOG.md
  git commit -m "chore(release): prepare vX.Y.Z"
  git tag -a vX.Y.Z -m "Release vX.Y.Z"
  git push origin main --tags
  ```
- **Release Branch + PR (Optional for multi-person review)**:
  ```bash
  git checkout -b release/vX.Y.Z
  git commit -m "chore(release): prepare vX.Y.Z"
  git push -u origin release/vX.Y.Z
  gh pr create --title "Release vX.Y.Z" --body "..."
  gh pr merge --admin --squash
  ```

### 5. Mode 2 Verification Checklist
- [ ] **GitHub Actions Run**: Monitor workflow to completion (`success`).
- [ ] **Latest Release Flag**: Verify `gh release view vX.Y.Z` displays as the **Latest Release** (`isPrerelease: false`).
- [ ] **Convention 1 Labels**: Verify all assets have human-readable labels applied.
- [ ] **Stable Preservation**: Verify the pre-release pruner does **not** touch `vX.Y.Z` or previous stable releases.
- [ ] **Local Synchronize**: Ensure local `main` is clean and aligned with `origin/main`.
