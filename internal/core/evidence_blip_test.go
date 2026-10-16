package core

import (
	"strings"
	"testing"
	"time"
)

// Collapsed listener-blip summaries are a newer event-text shape than the
// per-socket lines the Evidence Bundle was first tested with. They list
// several addresses in one line, so prove the bundle still redacts every
// one of them.
func TestEvidenceBundleRedactsCollapsedListenerBlipSummaries(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	now := time.Date(2026, 10, 15, 22, 0, 0, 0, time.UTC)
	if err := store.SaveSnapshot(Snapshot{
		SchemaVersion: snapshotSchemaVersion,
		CapturedAt:    now,
		Mode:          dockerDeploymentMode,
		Host:          HostInfo{Hostname: "nas.internal.example", OS: "Debian GNU/Linux 13"},
		Interfaces:    []InterfaceInfo{{Name: "enp1s0", Addresses: []string{"192.168.2.181/24"}, State: "up"}},
	}); err != nil {
		t.Fatal(err)
	}

	boot := now.Add(-time.Hour)
	sockets := []string{
		"tcp 192.168.2.181:8787",
		"tcp 192.168.2.181:9117",
		"tcp [2001:db8:5::17]:443",
		"tcp 10.44.3.9:5432",
		"udp 192.168.2.181:111",
	}
	events := []Event{{At: boot, Category: "system", Severity: "warning", Summary: "host reboot detected: current boot started " + boot.Format(time.RFC3339)}}
	for _, s := range sockets {
		events = append(events, Event{At: boot, Category: "listener", Severity: "warning", Summary: "listener disappeared: " + s})
	}
	for _, s := range sockets {
		events = append(events, Event{At: boot.Add(time.Minute), Category: "listener", Severity: "info", Summary: "listener appeared: " + s})
	}
	if err := store.AppendEvents(events); err != nil {
		t.Fatal(err)
	}

	// Precondition: the bundle reads the collapsed view, so the summary
	// line under test really is the multi-address shape.
	read, err := store.ReadEvents(EvidenceEventLimit)
	if err != nil {
		t.Fatal(err)
	}
	collapsed := false
	for _, e := range read {
		if strings.Contains(e.Summary, "listeners went away after the host reboot") {
			collapsed = true
		}
	}
	if !collapsed {
		t.Fatalf("expected a collapsed blip summary, got %+v", read)
	}

	bundle, err := buildEvidenceBundle(store, "v-test", now)
	if err != nil {
		t.Fatal(err)
	}
	out := string(evidenceFileBytes(bundle.Files))
	for _, secret := range []string{"192.168.2.181", "2001:db8:5::17", "10.44.3.9", "nas.internal.example"} {
		if strings.Contains(out, secret) {
			t.Errorf("evidence bundle leaked %q", secret)
		}
	}
	if !strings.Contains(out, "went away after the host reboot") || !strings.Contains(out, "tcp ipv4-") {
		t.Error("collapsed summary missing from bundle")
	}
}

func TestEvidenceBundleRedactsContainerUpdateAndRestartEvents(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	now := time.Date(2026, 10, 15, 22, 0, 0, 0, time.UTC)
	if err := store.SaveSnapshot(Snapshot{
		SchemaVersion: snapshotSchemaVersion, CapturedAt: now, Mode: dockerDeploymentMode,
		Host: HostInfo{Hostname: "nas.internal.example"},
		Containers: []ContainerInfo{{ID: "abcdef123456", Name: "family-photos", Image: "registry.internal.example/team/family-photos:latest",
			Status: "Up 5 minutes", ImageID: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", ImageVersion: "2.4.1", StartedAt: "2026-10-15T21:55:00Z"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendEvents([]Event{
		{At: now.Add(-time.Hour), Category: "container", Severity: "info", Summary: "container updated: family-photos: registry.internal.example/team/family-photos:latest, version 2.4.0 → 2.4.1"},
		{At: now.Add(-5 * time.Minute), Category: "container", Severity: "warning", Summary: "container restarted: family-photos (killed for running out of memory)"},
	}); err != nil {
		t.Fatal(err)
	}
	bundle, err := buildEvidenceBundle(store, "v-test", now)
	if err != nil {
		t.Fatal(err)
	}
	out := string(evidenceFileBytes(bundle.Files))
	for _, secret := range []string{"family-photos", "registry.internal.example", "nas.internal.example", "0123456789abcdef"} {
		if strings.Contains(out, secret) {
			t.Errorf("evidence bundle leaked %q", secret)
		}
	}
	for _, kept := range []string{"version 2.4.0 → 2.4.1", "killed for running out of memory", "2.4.1"} {
		if !strings.Contains(out, kept) {
			t.Errorf("useful evidence %q missing from bundle", kept)
		}
	}
}
