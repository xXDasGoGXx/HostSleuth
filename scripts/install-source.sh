#!/bin/sh
set -eu

if [ "$(id -u)" -ne 0 ]; then
  echo "run as root" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Go 1.24+ is required for source installation." >&2
  exit 1
fi

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

go test ./...
go build -trimpath -o /usr/local/bin/hostsleuth ./cmd/hostsleuth
install -d -m 0700 /var/lib/hostsleuth
install -m 0644 packaging/hostsleuth.service /etc/systemd/system/hostsleuth.service
systemctl daemon-reload
systemctl enable --now hostsleuth.service

echo "HostSleuth installed from source."
