#!/usr/bin/env bash
set -e

REPO="TrianaLab/remake"
API_URL="https://api.github.com/repos/$REPO/releases/latest"

detect_os_arch() {
  OS="$(uname | tr '[:upper:]' '[:lower:]')"
  case "$OS" in
    linux|darwin|mingw*|msys*|cygwin*)
      if [[ "$OS" == mingw* || "$OS" == msys* || "$OS" == cygwin* ]]; then
        OS="windows"
      fi
      ;;
    *)
      echo "Unsupported OS: $OS" >&2
      exit 1
      ;;
  esac

  ARCH="$(uname -m)"
  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
      echo "Unsupported architecture: $ARCH" >&2
      exit 1
      ;;
  esac

  EXT=""
  if [ "$OS" = "windows" ]; then
    EXT=".exe"
  fi

  echo "${OS}" "${ARCH}" "${EXT}"
}

fetch_latest_tag() {
  curl -sSL "$API_URL" \
    | grep -E '"tag_name":' \
    | sed -E 's/.*"([^"]+)".*/\1/'
}

find_install_dir() {
  IFS=':' read -r -a paths <<< "$PATH"
  for dir in "${paths[@]}"; do
    if [ -d "$dir" ] && [ -w "$dir" ]; then
      echo "$dir"
      return
    fi
  done
  echo ""
}

install_binary() {
  os="$1"; arch="$2"; ext="$3"; tag="$4"
  filename="remake_${os}_${arch}${ext}"
  url="https://github.com/$REPO/releases/download/$tag/$filename"

  tmpdir="$(mktemp -d)"
  target="$tmpdir/$filename"

  curl -sSL "$url" -o "$target"
  chmod +x "$target"

  install_dir="$(find_install_dir)"
  if [ -z "$install_dir" ]; then
    install_dir="$HOME/.local/bin"
    mkdir -p "$install_dir"
    echo "Warning: no writable dir in PATH, installing into $install_dir" >&2
  fi

  mv "$target" "$install_dir/remake${ext}"
  echo "Installed remake to $install_dir/remake${ext}"
}

main() {
  read -r os arch ext <<<"$(detect_os_arch)"
  tag="$(fetch_latest_tag)"
  install_binary "$os" "$arch" "$ext" "$tag"
}

main
