# Release Workflow

This document details the branching model, versioning semantics, and automation mechanics for cutting releases of **dmesg-watch**.

---

## 🌿 Branching Strategy

The repository follows a trunk-based model with dedicated release-preparation branches:

- **`main`**: Protected branch representing verified production states. Every release commit on `main` is tagged.
- **`feat/*` / `fix/*`**: Short-lived branches for enhancements and bug fixes. Merged into `main` via reviewed Pull Requests.
- **`release/vX.Y.Z[-beta.N]`**: Short-lived branches cut from `main` to curate changelogs and prepare releases.

---

## 🚀 Step-by-Step Release Procedure

### 1. Cut the Release Branch
Ensure `main` is up to date, then branch off:
```bash
git checkout main && git pull
git checkout -b release/v0.1.0-beta.1
```

### 2. Update Version in Code
Update the default version string in `cmd/dmesg-watch/main.go`:
```go
var version = "v0.1.0-beta.1"
```

### 3. Curate Release Notes
Follow the procedure codified in [`.agents/skills/release-notes/SKILL.md`](../.agents/skills/release-notes/SKILL.md):
1. Inspect the commit delta:
   ```bash
   git log v0.1.0..HEAD --oneline --no-merges
   ```
2. Add a new release block at the top of [`CHANGELOG.md`](../CHANGELOG.md).
3. Ensure notes adhere to the 15–20 line limit, emphasize user outcomes, and isolate build/D-Bus changes under `### 🛠️ Technical Notes`.

### 4. Commit and Open Pull Request
Commit the release preparation:
```bash
git add cmd/dmesg-watch/main.go CHANGELOG.md
git commit -m "chore(release): prepare v0.1.0-beta.1"
git push -u origin release/v0.1.0-beta.1
```
Open a Pull Request targeting `main` on GitHub:
- **Title**: `Release v0.1.0-beta.1`
- **Body**: Paste the curated release notes from `CHANGELOG.md`.

---

## ⚡ Automated CI/CD Execution on PR Merge

When the `release/*` PR is merged into `main`, [`.github/workflows/release.yml`](../.github/workflows/release.yml) executes automatically:

1. **Tag Creation**: Extracts the tag from the branch name (`release/v0.1.0-beta.1` $\to$ `v0.1.0-beta.1`), creates an annotated Git tag on `main`, and pushes it.
2. **Note Extraction**: Automatically parses the top release section from `CHANGELOG.md`.
3. **GoReleaser**:
   - Cross-compiles multi-arch binaries (`linux/amd64`, `linux/arm64`).
   - Packages `.deb` and `.rpm` files with systemd user unit definitions.
   - Packages release tarballs and checksums.
4. **Publish**: Creates the GitHub Release with the extracted release notes, attaches all package assets, and marks it as a **Pre-release** if the tag contains `beta`, `alpha`, or `rc`.
5. **Pre-Release Retention**: Automatically prunes older historical pre-releases on GitHub, retaining a sliding window of the 3 most recent while preserving Git tags.
