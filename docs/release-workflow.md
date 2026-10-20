# Release Workflow

This document details the branching model, versioning semantics, and automation mechanics for cutting releases of **dmesg-watch**.

---

## 🌿 Branching Strategy

The repository follows a trunk-based model with two release pathways:

1. **Fast Tag-Push (Primary / Default)**: For pre-releases, betas, and daily development. Fast, atomic, and ceremony-free. Directly commits to `main`, tags, and pushes.
2. **Release Branch + PR (Optional)**: Reserved for major production milestones (`v1.0.0`) or when collaborative pre-release sign-off is needed.

---

## 🚀 Fast Tag-Push Release Procedure (Betas & Daily Builds)

For rapid iteration and beta releases, use this streamlined, ceremony-free workflow:

### 1. Ensure `main` is Clean and Updated
```bash
git checkout main && git pull
```

### 2. Update Version in Code
Update the default version string in `cmd/dmesg-watch/main.go`:
```go
var version = "v0.2.0-beta.6"
```

### 3. Curate Release Notes
Follow the procedure codified in [`.agents/skills/release-notes/SKILL.md`](../.agents/skills/release-notes/SKILL.md):
1. Inspect the commit delta:
   ```bash
   git log $(git describe --tags --abbrev=0)..HEAD --oneline --no-merges
   ```
2. Add a new release block at the top of [`CHANGELOG.md`](../CHANGELOG.md).

### 4. Commit, Tag, and Push
Commit and push directly with an annotated tag:
```bash
git add cmd/dmesg-watch/main.go CHANGELOG.md
git commit -m "chore(release): prepare v0.2.0-beta.6"
git tag -a v0.2.0-beta.6 -m "Release v0.2.0-beta.6"
git push origin main --tags
```

---

## 🏛️ Release Branch + PR Workflow (Major Milestones)

For major stable milestones (`vX.Y.Z`) requiring team sign-offs:

### 1. Cut the Release Branch
```bash
git checkout main && git pull
git checkout -b release/v1.0.0
```

### 2. Update Version & Notes
Update `cmd/dmesg-watch/main.go` and add the milestone section to `CHANGELOG.md`.

### 3. Commit and Open Pull Request
```bash
git add cmd/dmesg-watch/main.go CHANGELOG.md
git commit -m "chore(release): prepare v1.0.0"
git push -u origin release/v1.0.0
```
Open and squash-merge the PR targeting `main`. The bot automatically creates the tag on merge.

---

## ⚡ Automated CI/CD Execution

Regardless of whether the release is triggered by **direct tag push** (`push: tags: 'v*'`) or **PR merge** (`release/*` into `main`), [`.github/workflows/release.yml`](../.github/workflows/release.yml) executes identically:

1. **Tag Creation**: Extracts the tag from the branch name (`release/v0.1.0-beta.1` $\to$ `v0.1.0-beta.1`), creates an annotated Git tag on `main`, and pushes it.
2. **Note Extraction**: Automatically parses the top release section from `CHANGELOG.md`.
3. **GoReleaser**:
   - Cross-compiles multi-arch binaries (`linux/amd64`, `linux/arm64`).
   - Packages `.deb` and `.rpm` files with systemd user unit definitions.
   - Packages release tarballs and checksums.
4. **Publish**: Creates the GitHub Release with the extracted release notes, attaches all package assets, and marks it as a **Pre-release** if the tag contains `beta`, `alpha`, or `rc`.
5. **Asset Labeling**: Automatically applies human-friendly labels to all release assets (e.g. `Debian / Ubuntu (x86_64)`, `Fedora / RHEL / openSUSE (x86_64)`, `Standalone Archive (x86_64)`).
6. **Pre-Release Retention**: Automatically prunes older historical pre-releases on GitHub, retaining a sliding window of the 3 most recent while preserving Git tags.
