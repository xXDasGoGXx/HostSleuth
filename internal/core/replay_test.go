package core

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// replay_week.json is one real week from a home server, with every name,
// image and address replaced. It holds a reboot onto a new kernel, three
// all-container restarts, a container that took 15 hours to become healthy
// (with a crash in the middle), a health failure fixed by autoheal, a backup
// container still failing, two updates and a clean restart.
func loadReplayWeek(t *testing.T) (Snapshot, []Event) {
	t.Helper()
	b, err := os.ReadFile("testdata/replay_week.json")
	if err != nil {
		t.Fatal(err)
	}
	var fx struct {
		Snapshot Snapshot `json:"snapshot"`
		Events   []Event  `json:"events"`
	}
	if err := json.Unmarshal(b, &fx); err != nil {
		t.Fatal(err)
	}
	return fx.Snapshot, fx.Events
}

func TestReplayNarratesTheRecordedWeek(t *testing.T) {
	snap, events := loadReplayWeek(t)
	r := BuildReplay(snap, events)

	var texts []string
	kinds := map[string]int{}
	for i, e := range r.Events {
		texts = append(texts, e.Text)
		kinds[e.Kind]++
		if e.N != i+1 {
			t.Fatalf("event numbers not sequential: %+v", e)
		}
		if i > 0 && e.T < r.Events[i-1].T {
			t.Fatal("events not in time order")
		}
		if e.Down != nil && (len(e.Down) != 2 || e.Down[0] < e.T || e.Down[1] <= e.Down[0]) {
			t.Fatalf("outage window must start at or after the event: %+v", e)
		}
	}
	all := strings.Join(texts, "\n")
	for _, want := range []string{
		"The host rebooted (kernel 7.0.14-23-pve → 7.0.14-24-pve). Every container stopped and came back.",
		"All 8 containers with a health check restarted at the same moment. That pattern fits a Docker restart.",
		"torrent-vpn passed its health check for the first time since it restarted 15 hours earlier.",
		"torrent-vpn crashed with exit code 1 and Docker restarted it.",
		"torrent-vpn started failing its health check. Docker shows it started again 13 seconds later. An autoheal container runs on this host; it restarts unhealthy containers.",
		"backup-snapshot started failing its health check. It has not passed since.",
		"backup-upload passed its health check again after 10 hours.",
		"search updated 2026.10.15-b00212a83 → 2026.10.16-ff8f5ddcc.",
		"hostsleuth updated 1.2.0 → 1.3.0.",
		"books restarted (stopped cleanly, exit code 0).",
		"chat-ui restarted.",
	} {
		if !strings.Contains(all, want) {
			t.Errorf("missing narration %q\nall:\n%s", want, all)
		}
	}
	if kinds["reboot"] != 1 || kinds["docker"] != 3 || kinds["crash"] != 1 || kinds["slow"] != 1 || kinds["unhealthy"] != 7 || kinds["recovered"] != 5 {
		t.Fatalf("kinds = %v", kinds)
	}
	// The reboot is anchored at the boot time, not when it was noticed.
	for _, e := range r.Events {
		if e.Kind == "reboot" && time.UnixMilli(e.T).UTC().Format(time.RFC3339) != "2026-10-14T22:32:55Z" {
			t.Fatalf("reboot anchored at %v", time.UnixMilli(e.T).UTC())
		}
	}
	if v := r.Versions["search"]; v.From != "2026.10.15-b00212a83" || v.To != "2026.10.16-ff8f5ddcc" {
		t.Fatalf("search version history: %+v", v)
	}
	if r.Now-r.Start != int64(replayWindow/time.Millisecond) {
		t.Fatal("window is not seven days")
	}
}

func TestReplayPartsStacksAndConnectors(t *testing.T) {
	snap, events := loadReplayWeek(t)
	r := BuildReplay(snap, events)
	if len(r.Parts) != 25 {
		t.Fatalf("parts = %d", len(r.Parts))
	}
	if r.Stacks[0].ID != "arr" || len(r.Stacks[0].Parts) != 9 {
		t.Fatalf("largest stack first: %+v", r.Stacks[0])
	}
	flags := map[string][]ReplayFlag{}
	hc := map[string]bool{}
	for _, p := range r.Parts {
		flags[p.ID] = p.Flags
		hc[p.ID] = p.HealthChecked
	}
	cases := map[string]string{
		"proxy":           "all :80 :443 :8181",
		"torrent-vpn":     "LAN :8080 :8118",
		"search":          "local :8088",
		"indexer-cache":   "local :9697",
		"media-server":    "host network",
		"backup-snapshot": "no network",
	}
	for name, want := range cases {
		got := flags[name]
		if len(got) != 1 || got[0].Label != want {
			t.Errorf("%s connectors = %+v, want %q", name, got, want)
		}
	}
	if len(flags["browser"]) != 0 {
		t.Errorf("a container without published ports has no connectors: %+v", flags["browser"])
	}
	if !hc["backup-snapshot"] || hc["shows"] {
		t.Error("health-check detection wrong")
	}
	b, _ := json.Marshal(r)
	if strings.Contains(string(b), "192.0.2.10") {
		t.Fatal("replay must not send addresses, only bind classes")
	}
	if len(r.Trans["torrent-vpn"]) == 0 || r.Trans["torrent-vpn"][0][1] != "running (healthy)" {
		t.Fatalf("transitions: %+v", r.Trans["torrent-vpn"])
	}
}

func TestReplayWithoutContainers(t *testing.T) {
	r := BuildReplay(Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: time.Now(), Mode: "native"}, nil)
	if len(r.Parts) != 0 || len(r.Notes) == 0 || r.Events == nil || r.Stacks == nil {
		t.Fatalf("empty replay: %+v", r)
	}
	b, _ := json.Marshal(r)
	if !strings.Contains(string(b), `"events":[]`) || !strings.Contains(string(b), `"parts":[]`) {
		t.Fatalf("empty lists must encode as []: %s", b)
	}
}

func TestReplayRebootWithoutKernelChange(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	boot := now.Add(-time.Hour)
	r := BuildReplay(Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: now}, []Event{
		{At: boot.Add(20 * time.Second), Category: "system", Severity: "warning", Summary: "host reboot detected: current boot started " + boot.Format(time.RFC3339)},
		{At: now.Add(-8 * 24 * time.Hour), Category: "system", Severity: "warning", Summary: "host reboot detected: kernel boot identity changed"},
	})
	if len(r.Events) != 1 || r.Events[0].Text != "The host rebooted. Every container stopped and came back." || r.Events[0].T != boot.UnixMilli() {
		t.Fatalf("events: %+v", r.Events)
	}
}
