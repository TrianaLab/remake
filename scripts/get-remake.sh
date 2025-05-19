#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="remake"
REPO="TrianaLab/remake"
DEFAULT_INSTALL_DIR="/usr/local/bin"
INSTALL_DIR="${REM_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
USE_SUDO="${REM_USE_SUDO:-true}"
VERIFY_CHECKSUM="${REM_VERIFY_CHECKSUM:-true}"

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
    echo "openssl is required for checksum verification" >&2
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
  url="https://get.remake.sh/remake-latest-version"
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
  DIST="${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz"
  BASE_URL="https://get.remake.sh"
  ARCHIVE_URL="$BASE_URL/$DIST"
  CHECK_URL="$ARCHIVE_URL.sha256"

  TMPDIR="$(mktemp -d)"
  cd "$TMPDIR"

  if has curl; then
    curl -fsSL "$CHECK_URL" -o checksum
    curl -fsSL "$ARCHIVE_URL" -o archive
  else
    wget -qO checksum "$CHECK_URL"
    wget -qO archive "$ARCHIVE_URL"
  fi

  if [ "$VERIFY_CHECKSUM" = "true" ]; then
    sum="$(openssl sha256 archive | awk '{print $2}')"
    exp="$(cat checksum)"
    if [ "$sum" != "$exp" ]; then
      echo "checksum mismatch" >&2
      exit 1
    fi
  fi

  echo "$TMPDIR"
}

install_binary() {
  tmp="$1"
  cd "$tmp"
  tar xf archive
  run_as_root cp "$BINARY_NAME" "$INSTALL_DIR/"
}

cleanup() {
  [ -n "${TMPDIR:-}" ] && rm -rf "$TMPDIR"
}

trap 'cleanup' EXIT

main() {
  init_platform
  check_prereqs
  get_latest_tag
  TMPDIR="$(download_and_verify)"
  install_binary "$TMPDIR"
  "$INSTALL_DIR/$BINARY_NAME" version
}

main "$@"
