#!/bin/sh
set -e

# CoinMarketCap CLI installer
# Usage: curl -sSfL https://raw.githubusercontent.com/openCMC/CoinMarketCap-CLI/main/install.sh | sh

REPO="openCMC/CoinMarketCap-CLI"
BINARY="cmc"
if [ -n "${INSTALL_DIR:-}" ]; then
  INSTALL_DIR="$INSTALL_DIR"
elif [ -n "${HOME:-}" ]; then
  INSTALL_DIR="${HOME}/.local/bin"
else
  INSTALL_DIR="/usr/local/bin"
fi

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) echo "unsupported"; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported"; exit 1 ;;
  esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

echo "Detected: ${OS}/${ARCH}"

# Get latest release tag
LATEST=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Error: could not determine latest release"
  exit 1
fi

echo "Latest release: ${LATEST}"

VERSION="${LATEST#v}"
EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

FILENAME="${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${LATEST}/checksums.txt"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading ${URL}..."
curl -sSfL "$URL" -o "${TMPDIR}/${FILENAME}"

echo "Downloading checksums..."
curl -sSfL "$CHECKSUMS_URL" -o "${TMPDIR}/checksums.txt"

echo "Verifying checksum..."
EXPECTED=$(grep "${FILENAME}" "${TMPDIR}/checksums.txt" | awk '{print $1}')
if [ -z "$EXPECTED" ]; then
  echo "Error: checksum not found for ${FILENAME}"
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "${TMPDIR}/${FILENAME}" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "${TMPDIR}/${FILENAME}" | awk '{print $1}')
else
  echo "Error: no sha256sum or shasum found — cannot verify download integrity"
  exit 1
fi

if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Error: checksum mismatch!"
  echo "  Expected: ${EXPECTED}"
  echo "  Actual:   ${ACTUAL}"
  exit 1
fi
echo "Checksum verified."

echo "Extracting..."
if [ "$EXT" = "tar.gz" ]; then
  tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"
else
  unzip -q "${TMPDIR}/${FILENAME}" -d "$TMPDIR"
fi

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
USE_SUDO=0
if mkdir -p "$INSTALL_DIR" 2>/dev/null && [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  sudo mkdir -p "$INSTALL_DIR"
  sudo mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  USE_SUDO=1
fi
if [ "$USE_SUDO" -eq 1 ]; then
  sudo chmod +x "${INSTALL_DIR}/${BINARY}"
else
  chmod +x "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed ${BINARY} ${LATEST} to ${INSTALL_DIR}/${BINARY}"
if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
  echo "Note: add ${HOME}/.local/bin to PATH if cmc is not found."
fi
