#!/usr/bin/env bash
set -e
BINARY_NAME="remake"
INSTALL_DIR=${REM_REMAKE_INSTALL_DIR:-"/usr/local/bin"}
USE_SUDO=${REM_USE_SUDO:-"true"}
DEBUG=${REM_DEBUG:-"false"}
VERIFY_CHECKSUM=${REM_VERIFY_CHECKSUM:-"true"}
HAS_CURL=$(type "curl" &> /dev/null && echo true || echo false)
HAS_WGET=$(type "wget" &> /dev/null && echo true || echo false)
HAS_OPENSSL=$(type "openssl" &> /dev/null && echo true || echo false)
HAS_TAR=$(type "tar" &> /dev/null && echo true || echo false)
initArch(){ ARCH=$(uname -m); case $ARCH in x86_64) ARCH="amd64";; aarch64) ARCH="arm64";; armv7l) ARCH="arm";; i386) ARCH="386";; *) ARCH="$ARCH";; esac; }
initOS(){ OS=$(uname|tr '[:upper:]' '[:lower:]'); case "$OS" in mingw*|cygwin*) OS="windows";; esac; }
runAsRoot(){ if [ $EUID -ne 0 -a "$USE_SUDO" = "true" ]; then sudo "$@"; else "$@"; fi; }
verifySupported(){ supported="darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 linux-arm linux-386 windows-amd64 windows-arm64"; if ! echo "$supported" | grep -qw "${OS}-${ARCH}"; then echo "No binary for ${OS}-${ARCH}."; exit 1; fi; if [ "$HAS_CURL" != "true" ] && [ "$HAS_WGET" != "true" ]; then echo "curl or wget required"; exit 1; fi; if [ "$VERIFY_CHECKSUM" = "true" ] && [ "$HAS_OPENSSL" != "true" ]; then echo "openssl required for checksum"; exit 1; fi; if [ "$HAS_TAR" != "true" ]; then echo "tar required"; exit 1; fi; }
checkDesiredVersion(){ if [ -z "$DESIRED_VERSION" ]; then url="https://get.remake.sh/remake-latest-version"; if [ "$HAS_CURL" = "true" ]; then TAG=$(curl -LsS "$url"); elif [ "$HAS_WGET" = "true" ]; then TAG=$(wget -qO- "$url"); fi; if [[ ! $TAG =~ ^v[0-9] ]]; then echo "Failed to get latest version"; exit 1; fi; else TAG=$DESIRED_VERSION; fi; }
checkInstalledVersion(){ if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then ver=$("$INSTALL_DIR/$BINARY_NAME" version | awk '{print $NF}'); if [ "$ver" = "$TAG" ]; then echo "$BINARY_NAME $ver already installed"; exit 0; fi; fi; }
downloadFile(){ DIST_FILE="${BINARY_NAME}-${TAG}-${OS}-${ARCH}.tar.gz"; DOWNLOAD_URL="https://get.remake.sh/$DIST_FILE"; CHECKSUM_URL="$DOWNLOAD_URL.sha256"; TMPDIR=$(mktemp -d); cd "$TMPDIR"; if [ "$HAS_CURL" = "true" ]; then curl -SsL "$CHECKSUM_URL" -o checksum; curl -SsL "$DOWNLOAD_URL" -o archive; elif [ "$HAS_WGET" = "true" ]; then wget -qO checksum "$CHECKSUM_URL"; wget -qO archive "$DOWNLOAD_URL"; fi; if [ "$VERIFY_CHECKSUM" = "true" ]; then sum=$(openssl sha256 archive | awk '{print $2}'); exp=$(cat checksum); if [ "$sum" != "$exp" ]; then echo "Checksum mismatch"; exit 1; fi; fi; echo "$TMPDIR"; }
installFile(){ tmp="$1"; cd "$tmp"; tar xf archive; runAsRoot cp "$BINARY_NAME" "$INSTALL_DIR/"; }
cleanup(){ [ -n "$TMPDIR" ] && rm -rf "$TMPDIR"; }
failTrap(){ code=$?; echo "Installation failed"; cleanup; exit $code; }
testVersion(){ "$INSTALL_DIR/$BINARY_NAME" version; }
trap failTrap EXIT
set -u
while [[ $# -gt 0 ]]; do case $1 in --version|-v) shift; DESIRED_VERSION=$1;; --no-sudo) USE_SUDO="false";; --help|-h) echo "Usage: [--version vX.Y.Z] [--no-sudo]"; exit;; *) exit 1;; esac; shift; done
initArch
initOS
verifySupported
checkDesiredVersion
checkInstalledVersion
TMPDIR=$(downloadFile)
installFile "$TMPDIR"
testVersion
cleanup
trap - EXIT
