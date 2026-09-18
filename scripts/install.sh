#!/bin/sh
set -eu

REPO="xXDasGoGXx/HostSleuth"
INSTALL_DIR="/usr/local/bin"
STATE_DIR="/var/lib/hostsleuth"
UNIT_PATH="/etc/systemd/system/hostsleuth.service"

if [ "$(id -u)" -ne 0 ]; then
  echo "run as root" >&2
  exit 1
fi

for cmd in curl cut grep install sha256sum systemctl; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "required command not found: $cmd" >&2
    exit 1
  fi
done

case "$(uname -m)" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

VERSION="${HOSTSLEUTH_VERSION:-latest}"
ASSET="hostsleuth-linux-${ARCH}"
if [ "$VERSION" = "latest" ]; then
  RELEASE_BASE="https://github.com/${REPO}/releases/latest/download"
else
  RELEASE_BASE="https://github.com/${REPO}/releases/download/${VERSION}"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM
ASSET_PATH="${TMP_DIR}/${ASSET}"
SUMS_PATH="${TMP_DIR}/SHA256SUMS"

echo "Downloading HostSleuth (${ARCH}, ${VERSION})..."
curl -fL --retry 3 --proto '=https' --tlsv1.2 "${RELEASE_BASE}/${ASSET}" -o "$ASSET_PATH"
curl -fL --retry 3 --proto '=https' --tlsv1.2 "${RELEASE_BASE}/SHA256SUMS" -o "$SUMS_PATH"

EXPECTED="$(grep "  ${ASSET}\$" "$SUMS_PATH" | cut -d' ' -f1)"
if [ -z "$EXPECTED" ] || [ "${#EXPECTED}" -ne 64 ]; then
  echo "release checksum for ${ASSET} is missing or invalid" >&2
  exit 1
fi

echo "Verifying release checksum..."
printf '%s  %s\n' "$EXPECTED" "$ASSET" | (cd "$TMP_DIR" && sha256sum -c -)

install -m 0755 "$ASSET_PATH" "${INSTALL_DIR}/hostsleuth"
install -d -m 0700 "$STATE_DIR"

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
install -m 0644 "$ROOT_DIR/packaging/hostsleuth.service" "$UNIT_PATH"
systemctl daemon-reload
systemctl enable --now hostsleuth.service

echo "HostSleuth installed."
echo "Dashboard: http://127.0.0.1:8787"
echo "CLI: hostsleuth diagnose host:port"
