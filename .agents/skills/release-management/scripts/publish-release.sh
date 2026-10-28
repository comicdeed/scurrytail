#!/usr/bin/env bash
#
# publish-release.sh
# Automated release pipeline for ScurryTail (Betas & Final Releases).
#
# Usage:
#   ./.agents/skills/release-management/scripts/publish-release.sh <tag>
#
# Examples:
#   ./.agents/skills/release-management/scripts/publish-release.sh v0.3.0-beta.15
#   ./.agents/skills/release-management/scripts/publish-release.sh v0.3.0

set -euo pipefail

TAG="${1:-}"

if [[ -z "$TAG" ]]; then
  echo "Error: Release tag argument missing." >&2
  echo "Usage: $0 <tag> (e.g. v0.3.0-beta.15 or v0.3.0)" >&2
  exit 1
fi

# Validate tag format (vX.Y.Z or vX.Y.Z-prerelease.N)
if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
  echo "Error: Tag '$TAG' is not valid SemVer (expected format: vX.Y.Z or vX.Y.Z-beta.N)." >&2
  exit 1
fi

# Ensure working directory is repo root
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "$REPO_ROOT" ]]; then
  echo "Error: Must be run inside a git repository." >&2
  exit 1
fi
cd "$REPO_ROOT"

# Ensure we are on main branch
CURRENT_BRANCH="$(git branch --show-current)"
if [[ "$CURRENT_BRANCH" != "main" ]]; then
  echo "Error: Releases must be cut from 'main' (currently on '$CURRENT_BRANCH')." >&2
  exit 1
fi

# Check that version constant in cmd/scurrytail/main.go matches the tag
MAIN_GO="cmd/scurrytail/main.go"
if [[ ! -f "$MAIN_GO" ]]; then
  echo "Error: Could not locate $MAIN_GO." >&2
  exit 1
fi

if ! grep -q "version = \"$TAG\"" "$MAIN_GO"; then
  echo "Error: Version mismatch! $MAIN_GO does not declare: version = \"$TAG\"" >&2
  echo "Please update $MAIN_GO before running this release script." >&2
  exit 1
fi

# Check that CHANGELOG.md contains the tag header
CHANGELOG="CHANGELOG.md"
if [[ ! -f "$CHANGELOG" ]]; then
  echo "Error: Could not locate $CHANGELOG." >&2
  exit 1
fi

if ! grep -q "## \[$TAG\]" "$CHANGELOG"; then
  echo "Error: Changelog mismatch! $CHANGELOG does not contain header: ## [$TAG]" >&2
  echo "Please curate $CHANGELOG before running this release script." >&2
  exit 1
fi

# Check that git tag does not already exist
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "Error: Git tag '$TAG' already exists locally or remotely." >&2
  exit 1
fi

echo "=================================================="
echo " Starting ScurryTail Release Pipeline: $TAG"
echo "=================================================="

# 1. Quality Gates
echo "==> [1/4] Running Quality Gates (make fmt, make test, make vet)..."
make fmt
make test
make vet

# 2. Git Commit & Tag
echo "==> [2/4] Staging, committing, and tagging $TAG..."
git add -A
# Commit only if there are changes staged
if ! git diff --cached --quiet; then
  git commit -m "chore(release): prepare $TAG"
fi
git tag -a "$TAG" -m "Release $TAG"

# 3. Push to Remote
echo "==> [3/4] Pushing commit and tags to origin main..."
git push origin main --tags

# 4. Watch GitHub Actions & Verify
echo "==> [4/4] Watching GitHub Actions release workflow..."
# Brief sleep to allow GitHub to register the pushed tag and trigger release.yml
sleep 5

RUN_ID=""
for attempt in {1..10}; do
  RUN_ID=$(gh run list --repo comicdeed/scurrytail --workflow=release.yml --limit 1 --json databaseId,headBranch --jq ".[0].databaseId" 2>/dev/null || true)
  if [[ -n "$RUN_ID" ]]; then
    break
  fi
  sleep 2
done

if [[ -n "$RUN_ID" ]]; then
  echo "Watching workflow run $RUN_ID..."
  gh run watch --repo comicdeed/scurrytail "$RUN_ID"
fi

echo ""
echo "=================================================="
echo " Release $TAG Published Successfully!"
echo "=================================================="
gh release view "$TAG" --repo comicdeed/scurrytail
