#!/bin/sh
set -eu

if [ "$(id -u)" -ne 0 ]; then
  echo "run as root" >&2
  exit 1
fi

systemctl disable --now hostsleuth.service 2>/dev/null || true
rm -f /etc/systemd/system/hostsleuth.service
rm -f /usr/local/bin/hostsleuth
systemctl daemon-reload

echo "HostSleuth binary and service removed."
echo "State was preserved in /var/lib/hostsleuth. Remove it manually if you no longer need the recorded history."
