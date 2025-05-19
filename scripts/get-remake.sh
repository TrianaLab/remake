#!/usr/bin/env bash

: ${BINARY_NAME:="remake"}
: ${USE_SUDO:="true"}
: ${DEBUG:="false"}
: ${VERIFY_CHECKSUM:="true"}
: ${REMAKE_INSTALL_DIR:="/usr/local/bin"}
DESIRED_VERSION="${DESIRED_VERSION:-}"

set -euo pipefail

HAS_CURL="$(type curl &>/dev/null && echo true || echo false)"
HAS_WGET="$(type wget &>/dev/null && echo true || echo false)"
HAS_OPENSSL="$(type openssl &>/dev/null && echo true || echo false)"
HAS_TAR="$(type tar &>/dev/null && echo true || echo false)"

initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*)   ARCH="armv5" ;;
    armv6*)   ARCH="armv6" ;;
    armv7*)   ARCH="arm"   ;;
    aarch64)  ARCH="arm64" ;;
    x86)      ARCH="386"   ;;
    x86_64)   ARCH="amd64" ;;
    i686|i386) ARCH="386"  ;;
  esac
}

initOS() {
  OS=$(uname | tr '[:upper:]' '[:lower:]')
  case "$OS" in
    mingw*|cygwin*) OS="windows" ;;
  esac
}

runAsRoot() {
  if [ "$EUID" -ne 0 ] && [ "$USE_SUDO" = "true" ]; then
    sudo "$@"
  else
    "$@"
  fi
}

verifySupported() {
  local sup="darwin-amd64 darwin-arm64 linux-386 linux-amd64 linux-arm linux-arm64 linux-ppc64le linux-s390x linux-riscv64 windows-amd64 windows-arm64"
  if ! echo "$sup" | grep -qw "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    exit 1
  fi
  if [ "$HAS_CURL" != "true" ] && [ "$HAS_WGET" != "true" ]; then
    echo "curl or wget is required"
    exit 1
  fi
  if [ "$VERIFY_CHECKSUM" = "true" ] && [ "$HAS_OPENSSL" != "true" ]; then
    echo "openssl is required"
    exit 1
  fi
  if [ "$HAS_TAR" != "true" ]; then
    echo "tar is required"
    exit 1
  fi
}

checkDesiredVersion() {
  if [ -z "$DESIRED_VERSION" ]; then
    local url="https://get.remake.sh/remake-latest-version"
    if [ "$HAS_CURL" = "true" ]; then
      DESIRED_VERSION=$(curl -fsSL "$url")
    else
      DESIRED_VERSION=$(wget -qO- "$url")
    fi
    if [[ ! "$DESIRED_VERSION" =~ ^v[0-9] ]]; then
      echo "Could not fetch latest version"
      exit 1
    fi
  fi
}

checkInstalledVersion() {
  if [ -x "${REMAKE_INSTALL_DIR}/${BINARY_NAME}" ]; then
    local v=$("${REMAKE_INSTALL_DIR}/${BINARY_NAME}" version | awk '{print $NF}')
    if [ "$v" = "$DESIRED_VERSION" ]; then
      echo "remake $v already installed"
      exit 0
    fi
  fi
}

downloadFile() {
  TMPDIR=$(mktemp -d "/tmp/remake.XXXXXX")
  local dist="${BINARY_NAME}-${DESIRED_VERSION}-${OS}-${ARCH}.tar.gz"
  local url="https://get.remake.sh/${dist}"
  local sumurl="${url}.sha256"
  if [ "$HAS_CURL" = "true" ]; then
    curl -fsSL "$sumurl" -o "$TMPDIR/checksum"
    curl -fsSL "$url"   -o "$TMPDIR/archive"
  else
    wget -qO "$TMPDIR/checksum" "$sumurl"
    wget -qO "$TMPDIR/archive"  "$url"
  fi
}

verifyFile() {
  if [ "$VERIFY_CHECKSUM" = "true" ]; then
    local sum exp
    sum=$(openssl sha256 "$TMPDIR/archive" | awk '{print $2}')
    exp=$(cat "$TMPDIR/checksum")
    if [ "$sum" != "$exp" ]; then
      echo "checksum mismatch"
      exit 1
    fi
  fi
}

installFile() {
  tar xf "$TMPDIR/archive" -C "$TMPDIR"
  runAsRoot cp "$TMPDIR/$BINARY_NAME" "$REMAKE_INSTALL_DIR/$BINARY_NAME"
}

testVersion() {
  "$REMAKE_INSTALL_DIR/$BINARY_NAME" version
}

help() {
  echo "Usage: [--version vX.Y.Z] [--no-sudo]"
  exit 0
}

cleanup() {
  rm -rf "${TMPDIR:-}"
}

fail_trap() {
  cleanup
  echo "Failed to install $BINARY_NAME"
  exit 1
}

trap fail_trap EXIT

while [[ $# -gt 0 ]]; do
  case $1 in
    --version|-v)
      shift
      DESIRED_VERSION=$1
      [[ "$DESIRED_VERSION" != v* ]] && DESIRED_VERSION="v${DESIRED_VERSION}"
      ;;
    --no-sudo)
      USE_SUDO="false"
      ;;
    --help|-h)
      help
      ;;
    *)
      exit 1
      ;;
  esac
  shift
done

initArch
initOS
verifySupported
checkDesiredVersion
checkInstalledVersion
downloadFile
verifyFile
installFile
testVersion
trap - EXIT
