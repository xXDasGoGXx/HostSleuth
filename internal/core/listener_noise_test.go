package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestIsRoutineListener(t *testing.T) {
	cases := []struct {
		proto, addr string
		routine     bool
	}{
		// Interface churn observed on a Docker host.
		{"udp", "[fe80::10af:b9ff:fe77:6b3c]%veth11a5660:3702", true},
		{"tcp", "[fe80::2491:20ff:feac:88ab]%vethcfe4f16:5357", true},
		{"udp", "[ff02::c]%br-3942559617ab:3702", true},
		{"udp", "239.255.255.250:3702", true},
		{"udp", "239.255.255.250:1900", true},
		{"udp", "169.254.10.2:5000", true},
		{"udp", "172.18.0.1:5353", true},
		{"udp", "172.18.0.1:3702", true},
		{"tcp", "172.18.0.1:5357", true},
		{"udp", "192.168.2.181:5353", true},
		// Ephemeral UDP client sockets.
		{"udp", "*:45591", true},
		{"udp", "192.168.2.181:51108", true},
		{"udp", "[::]:42320", true},
		{"udp", "127.0.0.1:57370", true},
		// Real services stay meaningful.
		{"tcp", "0.0.0.0:22", false},
		{"tcp", "[::]:443", false},
		{"tcp", "192.168.2.181:8787", false},
		{"tcp", "0.0.0.0:37925", false}, // TCP LISTEN is always a server
		{"tcp", "[::]:33383", false},
		{"udp", "0.0.0.0:5353", false}, // discovery daemon itself
		{"udp", "[::]:5355", false},
		{"udp", "127.0.0.1:5353", false},
		{"udp", "127.0.0.53%lo:53", false},
		{"udp", "0.0.0.0:111", false},
		{"udp", "0.0.0.0:32410", false}, // below ephemeral range
		{"udp", "0.0.0.0:51820", false}, // WireGuard
		{"udp", "[::]:41641", false},    // Tailscale
		{"udp", "10.0.3.1:53", false},
		{"tcp", "[fc42:5009:ba4b:5ab0::1]:53", false},
		// Unparseable input is never hidden.
		{"tcp", "garbage", false},
		{"udp", "host.example:5353", false},
		{"udp", ":5353", false},
	}
	for _, tc := range cases {
		if got := IsRoutineListener(tc.proto, tc.addr); got != tc.routine {
			t.Errorf("IsRoutineListener(%q, %q) = %v, want %v", tc.proto, tc.addr, got, tc.routine)
		}
	}
}

func TestIsRoutineEventOnlyMatchesListenerChurn(t *testing.T) {
	routine := Event{Category: "listener", Summary: "listener appeared: udp [ff02::c]%veth0:3702"}
	if !IsRoutineEvent(routine) {
		t.Fatal("expected multicast listener event to be routine")
	}
	for _, e := range []Event{
		{Category: "listener", Summary: "listener disappeared: tcp 0.0.0.0:22"},
		{Category: "container", Summary: "container changed: x: running (healthy) -> running (unhealthy)"},
		{Category: "system", Summary: "host reboot detected: current boot started 2026-10-14T22:32:55Z"},
		{Category: "listener", Summary: "listener appeared: udp"},
		{Category: "package", Summary: "listener appeared: udp *:45591"},
	} {
		if IsRoutineEvent(e) {
			t.Errorf("event wrongly classified as routine: %+v", e)
		}
	}
}

func TestDiffSnapshotsIgnoresListenerChurn(t *testing.T) {
	at := time.Date(2026, 10, 15, 0, 56, 8, 0, time.UTC)
	old := Snapshot{SchemaVersion: 4, Listeners: []Listener{
		{Protocol: "tcp", Address: "0.0.0.0:22"},
		{Protocol: "tcp", Address: "[fe80::10af:b9ff:fe77:6b3c]%veth11a5660:5357"},
		{Protocol: "udp", Address: "*:45902"},
		{Protocol: "tcp", Address: "0.0.0.0:8096"},
	}}
	cur := Snapshot{SchemaVersion: 4, CapturedAt: at, Listeners: []Listener{
		{Protocol: "tcp", Address: "0.0.0.0:22"},
		{Protocol: "tcp", Address: "[fe80::2491:20ff:feac:88ab]%vethcfe4f16:5357"},
		{Protocol: "udp", Address: "*:45591"},
		{Protocol: "udp", Address: "172.25.0.1:5353"},
	}}
	events := DiffSnapshots(old, cur)
	if len(events) != 1 {
		t.Fatalf("expected only the real listener change, got %d: %+v", len(events), events)
	}
	if events[0].Summary != "listener disappeared: tcp 0.0.0.0:8096" || events[0].Severity != "warning" {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestReadEventsSkipsRecordedChurnBeforeLimit(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	base := time.Date(2026, 10, 14, 22, 34, 0, 0, time.UTC)
	reboot := Event{At: base, Category: "system", Severity: "warning", Summary: "host reboot detected: current boot started 2026-10-14T22:32:55Z"}
	events := []Event{reboot}
	for i := 0; i < 150; i++ {
		events = append(events, Event{At: base.Add(time.Minute), Category: "listener", Severity: "info", Summary: fmt.Sprintf("listener appeared: udp [fe80::%x]%%veth%d:3702", i+1, i)})
	}
	health := Event{At: base.Add(2 * time.Hour), Category: "container", Severity: "warning", Summary: "container changed: backblaze-snapshotter: running (healthy) -> running (unhealthy)"}
	events = append(events, health)
	if err := store.AppendEvents(events); err != nil {
		t.Fatal(err)
	}

	got, err := store.ReadEvents(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Summary != reboot.Summary || got[1].Summary != health.Summary {
		t.Fatalf("meaningful history not preserved: %+v", got)
	}

	raw, err := store.ReadAllEvents(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 100 || raw[len(raw)-1].Summary != health.Summary {
		t.Fatalf("raw timeline should keep newest 100 recorded events, got %d", len(raw))
	}
}

func TestReadEventsEmptyTimelineIsEmptyList(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	got, err := store.ReadEvents(100)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(got)
	if string(b) != "[]" {
		t.Fatalf("expected [], got %s", b)
	}
}

func TestEventTimelineCompactsWhenLarge(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	pad := strings.Repeat("x", 200)

	// Write well past the compaction threshold without triggering it, as an
	// older version without compaction would have.
	f, err := os.OpenFile(store.EventsPath(), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(f)
	total := 0
	for size := 0; size <= eventFileCompactBytes+(1<<20); total++ {
		e := Event{SchemaVersion: 1, At: base.Add(time.Duration(total) * time.Second), Category: "container", Severity: "info", Summary: fmt.Sprintf("container changed: c%d: %s", total, pad)}
		if total%2 == 1 {
			e = Event{SchemaVersion: 1, At: e.At, Category: "listener", Severity: "info", Summary: fmt.Sprintf("listener appeared: udp *:%d", 40000+total%20000)}
		}
		b, _ := json.Marshal(e)
		size += len(b) + 1
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
	_ = f.Close()

	last := Event{At: base.Add(time.Duration(total) * time.Second), Category: "system", Severity: "warning", Summary: "host reboot detected: kernel boot identity changed"}
	if err := store.AppendEvents([]Event{last}); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(store.EventsPath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() > eventFileCompactBytes {
		t.Fatalf("timeline not compacted: %d bytes", info.Size())
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("compacted timeline permissions = %o, want 600", info.Mode().Perm())
	}
	raw, err := store.ReadAllEvents(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != eventFileKeep {
		t.Fatalf("compacted timeline kept %d events, want %d", len(raw), eventFileKeep)
	}
	for _, e := range raw {
		if IsRoutineEvent(e) {
			t.Fatalf("compaction kept routine churn: %+v", e)
		}
	}
	if raw[len(raw)-1].Summary != last.Summary {
		t.Fatalf("newest event lost during compaction: %+v", raw[len(raw)-1])
	}
	for i := 1; i < len(raw); i++ {
		if raw[i].At.Before(raw[i-1].At) {
			t.Fatal("compaction reordered events")
		}
	}
	if _, err := os.Stat(store.EventsPath() + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("compaction left a temporary file behind")
	}
}
