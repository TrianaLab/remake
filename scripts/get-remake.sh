#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="remake"
INSTALL_DIR="${REM_INSTALL_DIR:-/usr/local/bin}"
USE_SUDO="${REM_USE_SUDO:-true}"
VERIFY_CHECKSUM="${REM_VERIFY_CHECKSUM:-true}"
DESIRED_VERSION="${DESIRED_VERSION:-}"

has() { type "$1" &>/dev/null; }

init_platform() {
  OS="$(uname | tr '[:upper:]' '[:lower:]')"
  case "$OS" in
    mingw*|cygwin*) OS="windows";;
  esac
  ARCH="$(uname -m)"
  case "$ARCH" in
    x86_64) ARCH="amd64";;
    aarch64) ARCH="arm64";;
    armv7l) ARCH="arm";;
    i386) ARCH="386";;
  esac
}

check_prereqs() {
  if ! has curl && ! has wget; then
    echo "curl or wget is required" >&2
    exit 1
  fi
  if [ "$VERIFY_CHECKSUM" = "true" ] && ! has openssl; then
    echo "openssl is required" >&2
    exit 1
  fi
  if ! has tar; then
    echo "tar is required" >&2
    exit 1
  fi
}

run_as_root() {
  if [ "$EUID" -ne 0 ] && [ "$USE_SUDO" = "true" ]; then
    sudo "$@"
  else
    "$@"
  fi
}

get_latest_tag() {
  local url="https://get.remake.sh/remake-latest-version"
  if has curl; then
    TAG="$(curl -fsSL "$url")"
  else
    TAG="$(wget -qO- "$url")"
  fi
  if [[ ! $TAG =~ ^v[0-9] ]]; then
    echo "could not fetch latest version" >&2
    exit 1
  fi
}

download_and_verify() {
  local dist="${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz"
  local base="https://get.remake.sh"
  local archive_url="$base/$dist"
  local checksum_url="$archive_url.sha256"

  SCRIPT_TMPDIR="$(mktemp -d)"
  cd "$SCRIPT_TMPDIR"

  if has curl; then
    curl -fsSL "$checksum_url" -o checksum
    curl -fsSL "$archive_url" -o archive
  else
    wget -qO checksum "$checksum_url"
    wget -qO archive "$archive_url"
  fi

  if [ "$VERIFY_CHECKSUM" = "true" ]; then
    local sum exp
    sum="$(openssl sha256 archive | awk '{print $2}')"
    exp="$(cat checksum)"
    if [ "$sum" != "$exp" ]; then
      echo "checksum mismatch" >&2
      exit 1
    fi
  fi
}

install_binary() {
  tar xf "$SCRIPT_TMPDIR/archive" -C "$SCRIPT_TMPDIR"
  run_as_root cp "$SCRIPT_TMPDIR/$BINARY_NAME" "$INSTALL_DIR/"
}

cleanup() {
  [ -n "${SCRIPT_TMPDIR:-}" ] && rm -rf "$SCRIPT_TMPDIR"
}
trap cleanup EXIT

main() {
  init_platform
  check_prereqs

  if [ -z "$DESIRED_VERSION" ]; then
    get_latest_tag
  else
    TAG="$DESIRED_VERSION"
  fi

  download_and_verify
  install_binary
  "$INSTALL_DIR/$BINARY_NAME" version
}

main "$@"
