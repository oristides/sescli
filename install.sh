#!/usr/bin/env sh
# One-line install for sescli + sesmcp.
# Usage: curl -fsSL https://raw.githubusercontent.com/oristides/sescli/main/install.sh | sh
#
# INSTALL_DIR="$HOME/bin" curl -fsSL ... | sh
# REF=main curl -fsSL ... | sh     # git branch or tag when compiling from source (default: main)

set -eu

REPO="oristides/sescli"
REPO_URL="https://github.com/${REPO}.git"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
GIT_REF="${REF:-main}"

# Asset suffix must match GoReleaser naming, e.g. sescli_v1.2.3_linux_amd64.tar.gz / sesmcp_..._windows_amd64.zip
REL_OS=""
_RAW_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
_RAW_ARCH=$(uname -m)

case "$_RAW_ARCH" in
  x86_64|amd64) REL_ARCH=amd64 ;;
  aarch64|arm64) REL_ARCH=arm64 ;;
  i386|i686) REL_ARCH=386 ;;
  *)
    echo "install.sh: unsupported architecture: ${_RAW_ARCH}" >&2
    exit 1
    ;;
esac

case "$_RAW_OS" in
  linux) REL_OS=linux ;;
  darwin) REL_OS=darwin ;;
  mingw*|msys*|cygwin*) REL_OS=windows ;;
  *)
    echo "install.sh: unsupported OS: ${_RAW_OS}" >&2
    exit 1
    ;;
esac

if [ "$REL_OS" = "windows" ]; then
  RELEASE_SUFFIX="_${REL_OS}_${REL_ARCH}.zip"
else
  RELEASE_SUFFIX="_${REL_OS}_${REL_ARCH}.tar.gz"
fi

warn_path() {
  if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    echo "" >&2
    echo "Add to PATH: export PATH=\"\$PATH:${INSTALL_DIR}\"" >&2
  fi
}

download_url_for() {
  _name="$1"
  curl -s "https://api.github.com/repos/${REPO}/releases/latest" |
    grep "browser_download_url.*${_name}.*${RELEASE_SUFFIX}" |
    head -n1 |
    cut -d '"' -f4
}

install_from_release() {
  mkdir -p "$INSTALL_DIR"

  downloaded=0
  for bin in sescli sesmcp; do
    _url=$(download_url_for "$bin")
    if [ -z "$_url" ]; then
      continue
    fi
    echo "install.sh: downloading ${bin} (release)..." >&2
    _tmp=$(mktemp "${TMPDIR:-/tmp}/${bin}.XXXXXX")
    curl -fsSL -o "$_tmp" "$_url"
    if [ "$REL_OS" = "windows" ]; then
      unzip -q -o "$_tmp" -d "$INSTALL_DIR"
      chmod +x "$INSTALL_DIR/${bin}.exe" 2>/dev/null || true
    else
      tar -xzf "$_tmp" -C "$INSTALL_DIR"
      chmod +x "$INSTALL_DIR/${bin}" 2>/dev/null || true
    fi
    rm -f "$_tmp"
    downloaded=$((downloaded + 1))
  done

  [ "$downloaded" -gt 0 ]
}

install_from_source_go() {
  if ! command -v git >/dev/null 2>&1; then
    echo "install.sh: git not found; cannot compile from source." >&2
    return 1
  fi
  if ! command -v go >/dev/null 2>&1; then
    echo "install.sh: go not found; cannot compile from source." >&2
    return 1
  fi

  _work=$(mktemp -d "${TMPDIR:-/tmp}/sescli-install.XXXXXX")
  # shellcheck disable=SC2064
  trap 'rm -rf "$_work"' EXIT

  echo "install.sh: cloning ${REPO_URL} (ref: ${GIT_REF}) and running go install..." >&2
  if ! git clone --depth 1 --branch "$GIT_REF" "$REPO_URL" "$_work/sescli" 2>/dev/null; then
    if ! git clone --depth 1 "$REPO_URL" "$_work/sescli"; then
      return 1
    fi
    if ! git -C "$_work/sescli" checkout "$GIT_REF"; then
      echo "install.sh: could not checkout ref ${GIT_REF}" >&2
      return 1
    fi
  fi

  mkdir -p "$INSTALL_DIR"
  (
    cd "$_work/sescli" || exit 1
    export GOBIN="$INSTALL_DIR"
    go install ./cmd/sescli ./cmd/sesmcp
  )

  if [ "$REL_OS" != "windows" ]; then
    chmod +x "$INSTALL_DIR/sescli" "$INSTALL_DIR/sesmcp" 2>/dev/null || true
  fi
  return 0
}

if install_from_release; then
  echo "" >&2
  echo "install.sh: installed release binaries → ${INSTALL_DIR}" >&2
  warn_path
  echo "" >&2
  echo "Next: sescli help   ·   sesmcp (stdio MCP — see skills/references/MCP.md)" >&2
  exit 0
fi

echo "install.sh: no matching release for *${RELEASE_SUFFIX}; compiling from source..." >&2

if install_from_source_go; then
  echo "" >&2
  echo "install.sh: built from source → ${INSTALL_DIR}" >&2
  warn_path
  echo "" >&2
  echo "Next: sescli help   ·   sesmcp (stdio MCP — see skills/references/MCP.md)" >&2
  exit 0
fi

echo "" >&2
echo "Could not install. Options:" >&2
echo "  • Publish a GitHub Release (CI: push a v1.2.3 tag) with archives like sescli_*${RELEASE_SUFFIX} and sesmcp_*${RELEASE_SUFFIX}." >&2
echo "  • Install Git + Go (version per go.mod) and re-run this script." >&2
echo "  • Clone manually: git clone ${REPO_URL} && cd sescli && go install ./cmd/sescli ./cmd/sesmcp" >&2
echo "See: https://github.com/${REPO}/releases" >&2
exit 1
