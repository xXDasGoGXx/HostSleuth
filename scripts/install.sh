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

for cmd in curl install systemctl; do
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
if [ "$VERSION" = "latest" ]; then
  RELEASE_URL="https://github.com/${REPO}/releases/latest/download/hostsleuth-linux-${ARCH}"
else
  RELEASE_URL="https://github.com/${REPO}/releases/download/${VERSION}/hostsleuth-linux-${ARCH}"
fi

TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT INT TERM

echo "Downloading HostSleuth (${ARCH}, ${VERSION})..."
curl -fL --retry 3 --proto '=https' --tlsv1.2 "$RELEASE_URL" -o "$TMP"
install -m 0755 "$TMP" "${INSTALL_DIR}/hostsleuth"
install -d -m 0700 "$STATE_DIR"

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
install -m 0644 "$ROOT_DIR/packaging/hostsleuth.service" "$UNIT_PATH"
systemctl daemon-reload
systemctl enable --now hostsleuth.service

echo "HostSleuth installed."
echo "Dashboard: http://127.0.0.1:8787"
echo "CLI: hostsleuth diagnose host:port"
