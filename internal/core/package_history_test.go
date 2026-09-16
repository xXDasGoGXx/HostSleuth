package core

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseAPTHistory(t *testing.T) {
	text := `Start-Date: 2026-09-03  13:00:58
Commandline: apt-get dist-upgrade
Install: proxmox-headers-7.0.14-15-pve:amd64 (7.0.14-15, automatic)
Upgrade: docker-compose-plugin:amd64 (5.5.0-1~debian.13~trixie, 5.5.1-1~debian.13~trixie), proxmox-kernel-7.0:amd64 (7.0.14-14, 7.0.14-15)
Remove: proxmox-headers-7.0.14-14-pve:amd64 (7.0.14-14)
End-Date: 2026-09-03  13:02:21
`
	changes := parseAPTHistory(text)
	if len(changes) != 4 {
		t.Fatalf("expected 4 package changes, got %#v", changes)
	}
	assertPackageChange(t, changes[0], "install", "proxmox-headers-7.0.14-15-pve", "amd64", "", "7.0.14-15")
	assertPackageChange(t, changes[1], "remove", "proxmox-headers-7.0.14-14-pve", "amd64", "7.0.14-14", "")
	assertPackageChange(t, changes[2], "update", "docker-compose-plugin", "amd64", "5.5.0-1~debian.13~trixie", "5.5.1-1~debian.13~trixie")
	assertPackageChange(t, changes[3], "update", "proxmox-kernel-7.0", "amd64", "7.0.14-14", "7.0.14-15")
}

func TestParseDPKGLog(t *testing.T) {
	text := `2026-09-15 17:00:36 upgrade docker-ce-cli:amd64 5:29.8.0-1~debian.13~trixie 5:29.8.1-1~debian.13~trixie
2026-09-15 17:00:37 install demo-package:amd64 <none> 1.2.3-1
2026-09-15 17:00:38 remove old-package:all 4.5.6-1 <none>
2026-09-15 17:00:39 status installed demo-package:amd64 1.2.3-1
`
	changes := parseDPKGLog(text)
	if len(changes) != 3 {
		t.Fatalf("expected 3 package changes, got %#v", changes)
	}
	assertPackageChange(t, changes[0], "update", "docker-ce-cli", "amd64", "5:29.8.0-1~debian.13~trixie", "5:29.8.1-1~debian.13~trixie")
	assertPackageChange(t, changes[1], "install", "demo-package", "amd64", "", "1.2.3-1")
	assertPackageChange(t, changes[2], "remove", "old-package", "all", "4.5.6-1", "")
}

func TestCollectPackageChangesUsesAptGzipFallback(t *testing.T) {
	root := t.TempDir()
	aptDir := filepath.Join(root, "apt")
	if err := os.MkdirAll(aptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(aptDir, "history.log.1.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := gzip.NewWriter(f)
	_, err = zw.Write([]byte("Start-Date: 2026-09-12  10:12:39\nInstall: unrar-free:amd64 (1:0.3.1-1)\nEnd-Date: 2026-09-12  10:12:40\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOSTSLEUTH_PACKAGE_LOG_ROOT", root)
	changes := collectPackageChanges()
	if len(changes) != 1 {
		t.Fatalf("expected one apt fallback change, got %#v", changes)
	}
	assertPackageChange(t, changes[0], "install", "unrar-free", "amd64", "", "1:0.3.1-1")
}

func TestCollectPackageChangesPrefersDPKGWhenAvailable(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "apt"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "apt", "history.log"), []byte("Start-Date: 2026-09-12  10:12:39\nInstall: apt-only:amd64 (1.0)\nEnd-Date: 2026-09-12  10:12:40\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "dpkg.log"), []byte("2026-09-12 10:12:40 install dpkg-change:amd64 <none> 2.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOSTSLEUTH_PACKAGE_LOG_ROOT", root)
	changes := collectPackageChanges()
	if len(changes) != 1 || changes[0].Name != "dpkg-change" {
		t.Fatalf("expected dpkg history to be authoritative when available, got %#v", changes)
	}
}

func TestCollectPackageChangesQuietWhenLogsUnavailable(t *testing.T) {
	t.Setenv("HOSTSLEUTH_PACKAGE_LOG_ROOT", t.TempDir())
	if changes := collectPackageChanges(); len(changes) != 0 {
		t.Fatalf("expected no changes when package logs are unavailable, got %#v", changes)
	}
}

func TestDiffSnapshotsEmitsOnlyNewPackageChanges(t *testing.T) {
	firstAt := time.Date(2026, 9, 15, 17, 0, 36, 0, time.UTC)
	secondAt := firstAt.Add(time.Second)
	first := PackageChange{At: firstAt, Action: "update", Name: "docker-ce-cli", Architecture: "amd64", FromVersion: "1", ToVersion: "2"}
	second := PackageChange{At: secondAt, Action: "install", Name: "demo", Architecture: "amd64", ToVersion: "3"}

	events := DiffSnapshots(
		Snapshot{SchemaVersion: snapshotSchemaVersion, PackageChanges: []PackageChange{first}},
		Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: secondAt.Add(time.Minute), PackageChanges: []PackageChange{first, second}},
	)
	if len(events) != 1 {
		t.Fatalf("expected one package event, got %#v", events)
	}
	if events[0].Category != "package" || events[0].Severity != "info" {
		t.Fatalf("unexpected event metadata: %#v", events[0])
	}
	if !events[0].At.Equal(secondAt) {
		t.Fatalf("expected original package timestamp %s, got %s", secondAt, events[0].At)
	}
	if events[0].Summary != "package installed: demo:amd64 3" {
		t.Fatalf("unexpected event summary: %q", events[0].Summary)
	}
}

func TestDiffSnapshotsBaselinesPackageHistoryAcrossSchemaUpgrade(t *testing.T) {
	change := PackageChange{
		At:           time.Date(2026, 9, 15, 17, 0, 36, 0, time.UTC),
		Action:       "update",
		Name:         "docker-ce",
		Architecture: "amd64",
		FromVersion:  "1",
		ToVersion:    "2",
	}
	events := DiffSnapshots(
		Snapshot{SchemaVersion: 1},
		Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: change.At.Add(time.Minute), PackageChanges: []PackageChange{change}},
	)
	for _, event := range events {
		if event.Category == "package" {
			t.Fatalf("expected first schema-2 capture to baseline package history, got %#v", events)
		}
	}
}

func TestPackageChangeSummary(t *testing.T) {
	cases := []struct {
		change PackageChange
		want   string
	}{
		{PackageChange{Action: "install", Name: "demo", ToVersion: "1.0"}, "package installed: demo 1.0"},
		{PackageChange{Action: "update", Name: "demo", Architecture: "amd64", FromVersion: "1.0", ToVersion: "1.1"}, "package updated: demo:amd64 1.0 -> 1.1"},
		{PackageChange{Action: "remove", Name: "demo", FromVersion: "1.1"}, "package removed: demo 1.1"},
	}
	for _, tc := range cases {
		if got := packageChangeSummary(tc.change); got != tc.want {
			t.Fatalf("packageChangeSummary(%#v) = %q, want %q", tc.change, got, tc.want)
		}
	}
}

func assertPackageChange(t *testing.T, got PackageChange, action, name, arch, fromVersion, toVersion string) {
	t.Helper()
	if got.Action != action || got.Name != name || got.Architecture != arch || got.FromVersion != fromVersion || got.ToVersion != toVersion {
		t.Fatalf("unexpected package change: got %#v", got)
	}
	if got.At.IsZero() {
		t.Fatal("expected package timestamp")
	}
	if strings.TrimSpace(got.Name) == "" {
		t.Fatal("expected package name")
	}
}
