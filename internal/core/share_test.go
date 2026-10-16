package core

import (
	"strings"
	"testing"
	"time"
)

func TestRedactForSharingReplacesHostIdentifiers(t *testing.T) {
	snap := Snapshot{
		SchemaVersion: snapshotSchemaVersion,
		CapturedAt:    time.Date(2026, 10, 15, 22, 0, 0, 0, time.UTC),
		Host:          HostInfo{Hostname: "nas-basement"},
		Interfaces:    []InterfaceInfo{{Name: "enp36s0f0", Addresses: []string{"192.168.2.181/24"}}},
		Containers:    []ContainerInfo{{ID: "abc123", Name: "qbittorrent-vpn", Image: "ghcr.io/example/qbittorrent:release-5.2.3"}},
		Services:      []ServiceInfo{{Name: "media-sync.service"}},
	}
	in := strings.Join([]string{
		"HostSleuth diagnosis for nas-basement:8080",
		"[FAILED] TCP: connect 192.168.2.181:8080 refused",
		"Container qbittorrent-vpn restarted (crashed, exit code 2)",
		"Upstream https://media.example.org/api?token=abc123secret",
		"Service media-sync.service failed; config /srv/private/media/config.yml",
		"password=hunter2",
	}, "\n")
	out := RedactForSharing(snap, nil, in)
	for _, leak := range []string{"nas-basement", "192.168.2.181", "qbittorrent-vpn", "media.example.org", "abc123secret", "media-sync", "/srv/private", "hunter2"} {
		if strings.Contains(out, leak) {
			t.Errorf("shared text leaked %q:\n%s", leak, out)
		}
	}
	for _, kept := range []string{"[FAILED] TCP: connect", ":8080 refused", "crashed, exit code 2", "HostSleuth diagnosis for"} {
		if !strings.Contains(out, kept) {
			t.Errorf("useful evidence %q lost:\n%s", kept, out)
		}
	}
	if strings.Count(out, "\n") != strings.Count(in, "\n") {
		t.Error("line structure changed")
	}
}
