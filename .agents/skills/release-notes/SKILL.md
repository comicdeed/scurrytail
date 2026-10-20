---
name: release-notes
description: Use when preparing or updating dmesg-watch release notes, CHANGELOG.md entries, or GitHub Release text; analyzes commit ranges for Beta vs Stable releases and produces concise, user-experience-first notes with technical changes in a separate section.
---

# Release Notes Generation Skill

This skill defines the standard procedure for analyzing git commits, filtering development noise, and generating concise, user-focused release notes in Markdown for **dmesg-watch** (`github.com/comicdeed/dmesg-watch`).

---

## 🎯 Core Principles

1. **User-Centric Focus**: Highlight tangible user benefits, error visibility, tray interactions, desktop notifications, and meaningful fixes. Exclude internal refactors, test additions, code review cleanup, and CI maintenance commits.
2. **Human, Experience-First Tone**: Write in clear, natural language that describes what users can now see, feel, or notice. Avoid robotic changelog phrasing, implementation jargon, and commit-message style fragments.
3. **Separate Technical Notes**: Place important packaging, D-Bus protocols, architecture, or systemd changes in a distinct section labeled **Technical Notes** after user-facing highlights.
4. **Strict Conciseness**: Maximum 15–20 lines total. Keep descriptions crisp and high-impact.
5. **Single Source of Truth**: `CHANGELOG.md` is the canonical source of truth, from which GitHub Releases and package descriptions are derived.

### Required Section Order & Living Draft Model

dmesg-watch uses a **Living Draft** release notes model: beta release notes serve as the evolving draft of the final target milestone release. Both Beta and Stable notes share the identical structural framework.

#### For Beta Releases (`vX.Y.Z-beta.N`)
* **Milestone Summary**: 1 sentence summarizing the milestone's primary focus across the cycle.
* **Highlights (with `✨` for new additions)**: The full milestone feature list where new additions delivered in the current beta build are prefixed with `✨` (e.g. `* ✨ Clickable project links in notifications...`).
* **Technical Notes (with `✨` for new additions)**: Architecture, D-Bus, systemd, or packaging changes (items new to this build prefixed with `✨`).

#### For Stable Releases (`vX.Y.Z`)
* **Finalizing Milestone Notes**: The accumulated beta draft is finalized by removing the `✨` indicator icons and transitioning to the formal milestone release summary.

---

## 🔍 Diff Scope Rules

### 1. For Beta Releases (`vX.Y.Z-beta.N`)
* **Cumulative Milestone Range**: `git log <last-stable-tag>..HEAD` $\rightarrow$ Inspects full cycle to maintain the cumulative draft.
* **Incremental Delta Range**: `git log <last-tag>..HEAD` $\rightarrow$ Identifies new items to tag with `✨`.
* **Objective**: Maintains a living draft of the full milestone notes, with new items in this beta tagged with `✨`.

### 2. For Stable Releases (`vX.Y.Z`)
* **Range**: `git log <last-stable-tag>..HEAD`
* **Objective**: Finalizes the cumulative milestone changelog by stripping intermediate `✨` markers.

---

## 🛠️ Step-by-Step Procedure

### 1. Determine Base Tag and Target Version
```bash
# For Beta: find the immediately preceding tag (e.g. v0.1.0-beta.1)
PREV_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")

# For Stable: find the latest stable tag (excluding prerelease qualifiers)
PREV_STABLE_TAG=$(git tag --sort=-creatordate | grep -v -E '(beta|alpha|rc)' | head -n 1)
```

### 2. Inspect Commit History
```bash
git log "${PREV_TAG}..HEAD" --oneline --no-merges
```

### 3. Categorize Changes
Group commits into high-signal categories:
* **`✨ Highlights`**: Major user-visible improvements, cleaner tray interactions, notification behavior.
* **`🔒 Reliability & Filtering`**: Error rate calculations, sliding-window accuracy, cooldown behavior.
* **`🛠️ Technical Notes`**: Packaging (`.deb`, `.rpm`, AUR, tarball), systemd user unit changes, Go dependencies.

*Omit all `test:`, `chore:`, `ci:`, and routine maintenance commits.*

---

## 📝 Canonical CHANGELOG.md Format

```markdown
# Changelog

All notable changes to **dmesg-watch** are documented in this file.

## [v0.1.0-beta.1] - 2026-10-20

The inaugural beta release of **dmesg-watch**, a lightweight Linux desktop background daemon that monitors kernel errors in real time.

### ✨ Highlights
* **Real-time Error Awareness**: Instantly flags kernel faults as they occur without manual terminal inspection.
* **Quiet by Default**: Background tray dashboard tracks error frequency across a rolling 60-second window, alerting only when issues surge.
* **Flood Protection**: Smart notification cooldown prevents banner spam when drivers or hardware emit error bursts.
* **Desktop Native**: Integrates into the system tray across KDE Plasma, GNOME, Sway, and Waybar.
* **At-a-Glance Dashboard**: Context menu presents live error counts, latest truncated message, and quick links to project information.

### 🛠️ Technical Notes
* Implemented in pure Go, maintaining a minimal idle memory footprint (~15MB RSS).
* Direct D-Bus Session Bus integration for StatusNotifierItem (SNI) and desktop notifications without external GUI dependencies.
* Packaged with systemd user unit (`graphical-session.target`), standalone tarballs, and automated `.deb` / `.rpm` distro packages.
```

---

## ✅ Quality Gate Before Finalizing Notes

Before publishing, verify all of the following:
* The first section is explicitly user experience driven.
* No internal code-refactor commits or trivial details appear in Highlights.
* All low-level packaging or architecture details are isolated under **Technical Notes**.
* Descriptions are outcome-focused and respect the 15–20 line total limit.
