#!/bin/sh
set -eu

if [ "$(id -u)" -ne 0 ]; then
  echo "run as root" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required for this source installer." >&2
  exit 1
fi

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

go build -trimpath -o /usr/local/bin/hostsleuth ./cmd/hostsleuth
install -d -m 700 /var/lib/hostsleuth
install -m 644 packaging/hostsleuth.service /etc/systemd/system/hostsleuth.service
systemctl daemon-reload
systemctl enable --now hostsleuth.service

echo "HostSleuth installed."
echo "Dashboard: http://127.0.0.1:8787"
echo "CLI: hostsleuth diagnose host:port"
