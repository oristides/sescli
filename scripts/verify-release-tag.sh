#!/usr/bin/env bash
# Exit 0 if TAG is strictly newer than the latest v* tag (version sort).
# Usage: ./scripts/verify-release-tag.sh v0.4.0
set -euo pipefail

TAG="${1:?usage: verify-release-tag.sh TAG}"

case "$TAG" in
  v[0-9]*.[0-9]*) ;;
  *)
    echo "$0: TAG must look like v0.3.1 (semver with v prefix)" >&2
    exit 1
    ;;
esac

latest="$(git tag -l 'v*' 2>/dev/null | sort -V | tail -1 || true)"
if [[ -z "$latest" ]]; then
  echo "OK: no existing v* tags; $TAG will be the first."
  exit 0
fi

if [[ "$TAG" == "$latest" ]]; then
  echo "Refusing: $TAG is already the latest tag ($latest)." >&2
  exit 1
fi

if [[ "$(printf '%s\n' "$latest" "$TAG" | sort -V | tail -1)" != "$TAG" ]]; then
  echo "Refusing: $TAG must be newer than latest tag $latest (version sort)." >&2
  exit 1
fi

echo "OK: $TAG is newer than latest release tag $latest"
