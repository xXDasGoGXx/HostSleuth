package core

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCollectConfigFingerprintsStoresOnlyFingerprintMetadata(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "etc", "hosts")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("127.0.0.1 localhost\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOSTSLEUTH_CONFIG_ROOT", root)
	values := collectConfigFingerprints()
	if len(values) != len(defaultConfigFingerprintPaths) {
		t.Fatalf("expected %d fingerprints, got %#v", len(defaultConfigFingerprintPaths), values)
	}

	var hosts ConfigFingerprint
	for _, value := range values {
		if value.Path == "/etc/hosts" {
			hosts = value
			break
		}
	}
	if hosts.State != "present" {
		t.Fatalf("expected /etc/hosts present, got %#v", hosts)
	}
	wantHash := sha256.Sum256(content)
	if hosts.Fingerprint != hex.EncodeToString(wantHash[:]) {
		t.Fatalf("unexpected fingerprint: %q", hosts.Fingerprint)
	}
	if hosts.SizeBytes != int64(len(content)) {
		t.Fatalf("unexpected size: %d", hosts.SizeBytes)
	}
	if hosts.Fingerprint == string(content) {
		t.Fatal("fingerprint must not contain configuration contents")
	}

	for i := 1; i < len(values); i++ {
		if values[i-1].Path > values[i].Path {
			t.Fatalf("fingerprints are not sorted: %#v", values)
		}
	}
}

func TestCollectConfigFingerprintsRepresentsMissingFiles(t *testing.T) {
	t.Setenv("HOSTSLEUTH_CONFIG_ROOT", t.TempDir())
	values := collectConfigFingerprints()
	if len(values) != len(defaultConfigFingerprintPaths) {
		t.Fatalf("expected all default paths, got %#v", values)
	}
	for _, value := range values {
		if value.State != "missing" || value.Fingerprint != "" || value.SizeBytes != 0 {
			t.Fatalf("expected missing metadata only, got %#v", value)
		}
	}
}

func TestDiffSnapshotsBaselinesConfigurationAcrossSchemaUpgradeAndKeepsPackages(t *testing.T) {
	firstAt := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	secondAt := firstAt.Add(time.Minute)
	firstPackage := PackageChange{At: firstAt, Action: "update", Name: "demo", Architecture: "amd64", FromVersion: "1", ToVersion: "2"}
	secondPackage := PackageChange{At: secondAt, Action: "install", Name: "new-demo", Architecture: "amd64", ToVersion: "1"}

	events := DiffSnapshots(
		Snapshot{SchemaVersion: packageHistorySchemaVersion, PackageChanges: []PackageChange{firstPackage}},
		Snapshot{
			SchemaVersion:      configFingerprintSchemaVersion,
			CapturedAt:         secondAt.Add(time.Minute),
			PackageChanges:     []PackageChange{firstPackage, secondPackage},
			ConfigFingerprints: []ConfigFingerprint{{Path: "/etc/hosts", State: "present", Fingerprint: "abc"}},
		},
	)

	packageEvents := 0
	configEvents := 0
	for _, event := range events {
		switch event.Category {
		case "package":
			packageEvents++
		case "configuration":
			configEvents++
		}
	}
	if packageEvents != 1 {
		t.Fatalf("expected one package event across schema upgrade, got %#v", events)
	}
	if configEvents != 0 {
		t.Fatalf("expected configuration baseline across schema upgrade, got %#v", events)
	}
}

func TestDiffConfigFingerprintsChangedAppearedDisappeared(t *testing.T) {
	at := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	oldValues := []ConfigFingerprint{
		{Path: "/etc/fstab", State: "present", Fingerprint: "old"},
		{Path: "/etc/hosts", State: "present", Fingerprint: "same"},
		{Path: "/etc/nftables.conf", State: "present", Fingerprint: "gone"},
		{Path: "/etc/ssh/sshd_config", State: "missing"},
	}
	newValues := []ConfigFingerprint{
		{Path: "/etc/fstab", State: "present", Fingerprint: "new"},
		{Path: "/etc/hosts", State: "present", Fingerprint: "same"},
		{Path: "/etc/nftables.conf", State: "missing"},
		{Path: "/etc/ssh/sshd_config", State: "present", Fingerprint: "appeared"},
	}

	events := diffConfigFingerprints(at, oldValues, newValues)
	if len(events) != 3 {
		t.Fatalf("expected three configuration events, got %#v", events)
	}
	want := []struct {
		summary  string
		severity string
	}{
		{"configuration changed: /etc/fstab", "info"},
		{"configuration disappeared: /etc/nftables.conf", "warning"},
		{"configuration appeared: /etc/ssh/sshd_config", "info"},
	}
	for i, expected := range want {
		if events[i].Category != "configuration" || events[i].Summary != expected.summary || events[i].Severity != expected.severity || !events[i].At.Equal(at) {
			t.Fatalf("event %d = %#v, want summary=%q severity=%q", i, events[i], expected.summary, expected.severity)
		}
	}
}

func TestDiffConfigFingerprintsIgnoresUnreadableTransitions(t *testing.T) {
	at := time.Now().UTC()
	cases := []struct {
		old ConfigFingerprint
		new ConfigFingerprint
	}{
		{ConfigFingerprint{Path: "/etc/hosts", State: "present", Fingerprint: "a"}, ConfigFingerprint{Path: "/etc/hosts", State: "unreadable"}},
		{ConfigFingerprint{Path: "/etc/hosts", State: "unreadable"}, ConfigFingerprint{Path: "/etc/hosts", State: "present", Fingerprint: "b"}},
		{ConfigFingerprint{Path: "/etc/hosts", State: "unreadable"}, ConfigFingerprint{Path: "/etc/hosts", State: "missing"}},
	}
	for _, tc := range cases {
		if events := diffConfigFingerprints(at, []ConfigFingerprint{tc.old}, []ConfigFingerprint{tc.new}); len(events) != 0 {
			t.Fatalf("expected unreadable transition to stay quiet, got %#v", events)
		}
	}
}

func TestDiffSnapshotsSchemaThreeEmitsOnlyRealConfigurationChange(t *testing.T) {
	at := time.Now().UTC()
	oldSnap := Snapshot{
		SchemaVersion: configFingerprintSchemaVersion,
		ConfigFingerprints: []ConfigFingerprint{
			{Path: "/etc/hosts", State: "present", Fingerprint: "one"},
		},
	}
	newSnap := Snapshot{
		SchemaVersion: configFingerprintSchemaVersion,
		CapturedAt:    at,
		ConfigFingerprints: []ConfigFingerprint{
			{Path: "/etc/hosts", State: "present", Fingerprint: "two"},
		},
	}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) != 1 || events[0].Category != "configuration" || events[0].Summary != "configuration changed: /etc/hosts" {
		t.Fatalf("unexpected events: %#v", events)
	}
}
