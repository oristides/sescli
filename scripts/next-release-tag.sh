#!/usr/bin/env bash
# Print the next sequential release tag after the latest v* tag (semver-style).
# Usage:
#   ./scripts/next-release-tag.sh [patch|minor|major]
# Default bump is patch. If there are no v* tags, prints v0.1.0.
set -euo pipefail

BUMP="${1:-patch}"
case "$BUMP" in
  patch | minor | major) ;;
  *)
    echo "usage: $0 patch|minor|major" >&2
    exit 1
    ;;
esac

latest="$(git tag -l 'v*' 2>/dev/null | sort -V | tail -1 || true)"
if [[ -z "$latest" ]]; then
  echo "v0.1.0"
  exit 0
fi

ver="${latest#v}"
parts=()
IFS='.' read -ra parts <<< "$ver"
major="${parts[0]:-0}"
minor="${parts[1]:-0}"
patch="${parts[2]:-0}"

case "$BUMP" in
  patch)
    patch=$((patch + 1))
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
esac

echo "v${major}.${minor}.${patch}"
