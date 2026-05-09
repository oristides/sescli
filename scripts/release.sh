#!/usr/bin/env sh
# Create an annotated semver tag locally and push it.
# GitHub Actions runs GoReleaser on tag push (see .github/workflows/release.yml).
#
# Usage:
#   ./scripts/release.sh v0.3.1
#
# Requires: git, remote branch up to date, tag must not exist.

set -eu

TAG="${1:-}"
case "$TAG" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  v[0-9]*.[0-9]*) ;;
  *)
    echo "usage: $0 v0.3.1   # semver tag starting with v" >&2
    exit 1
    ;;
esac

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "$0: tag $TAG already exists" >&2
  exit 1
fi

echo "Creating annotated tag $TAG on $(git branch --show-current) ($(git rev-parse --short HEAD))..."
git tag -a "$TAG" -m "Release $TAG"
echo "Pushing tag to origin..."
git push origin "$TAG"

echo "Done."
echo "GitHub → Actions → wait for the \"Release\" workflow."
echo "GitHub → Releases → confirm binaries and checksums."
