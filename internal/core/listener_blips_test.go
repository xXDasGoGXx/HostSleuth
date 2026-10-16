package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func blipEvent(at time.Time, kind, key string) Event {
	sev := "info"
	if kind == "disappeared" {
		sev = "warning"
	}
	return Event{SchemaVersion: 1, At: at, Category: "listener", Severity: sev, Summary: "listener " + kind + ": " + key}
}

// The real shape recorded on an OMV Docker host (window of the newest 100
// events, sanitized): the first post-boot snapshot records the reboot plus
// listeners not yet back; the next snapshot records them again. Random-port
// services return on different ports and must stay visible. Later a
// redeploy blips one listener, and one loopback listener leaves for good.
func TestCollapseListenerBlipsOnRecordedRebootTimeline(t *testing.T) {
	b, err := os.ReadFile("testdata/omv_reboot_timeline.json")
	if err != nil {
		t.Fatal(err)
	}
	var events []Event
	if err := json.Unmarshal(b, &events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 101 {
		t.Fatalf("fixture has %d events", len(events))
	}

	got := CollapseListenerBlips(events)

	var reboot, wave, redeploy, stillGone, moved []Event
	appeared, disappeared := 0, 0
	for _, e := range got {
		switch {
		case strings.HasPrefix(e.Summary, "host reboot detected"):
			reboot = append(reboot, e)
		case strings.HasPrefix(e.Summary, "34 listeners went away after the host reboot and were back within 61s: tcp 0.0.0.0:5902, tcp 0.0.0.0:5903, tcp 0.0.0.0:5904 and 31 more"):
			wave = append(wave, e)
		case e.Summary == "listener went away and was back within 61s: tcp 192.0.2.10:8787":
			redeploy = append(redeploy, e)
		case strings.HasPrefix(e.Summary, "after the host reboot, 2 dynamic-port listeners were replaced by new ports on the same address: tcp 127.0.0.1 "):
			moved = append(moved, e)
		case e.Summary == "listener disappeared: tcp 127.0.0.1:35099":
			stillGone = append(stillGone, e)
		case strings.HasPrefix(e.Summary, "listener appeared: "):
			appeared++
		case strings.HasPrefix(e.Summary, "listener disappeared: "):
			disappeared++
		}
	}
	if len(reboot) != 1 || len(wave) != 1 || len(redeploy) != 1 || len(stillGone) != 1 {
		t.Fatalf("reboot=%d wave=%d redeploy=%d gone=%d\n%+v", len(reboot), len(wave), len(redeploy), len(stillGone), got)
	}
	// 8 random-port listeners did not return on the same port. Two pair
	// with new loopback ports on the same address; the rest have no partner
	// inside this 100-event slice (their new ports were recorded earlier in
	// the same snapshot, before the slice starts).
	if len(moved) != 1 {
		t.Fatalf("want one dynamic-port move summary, got %d\n%+v", len(moved), got)
	}
	if disappeared != 6 || appeared != 6 {
		t.Fatalf("unpaired listener events: disappeared=%d appeared=%d", disappeared, appeared)
	}
	if wave[0].Severity != "info" || !wave[0].At.Equal(reboot[0].At) {
		t.Fatalf("reboot wave summary: %+v", wave[0])
	}
	if redeploy[0].Severity != "warning" || stillGone[0].Severity != "warning" {
		t.Fatalf("blip without reboot and a lasting outage are warnings: %+v %+v", redeploy[0], stillGone[0])
	}
	if len(got) != 101-34*2+1-1-2*2+1 {
		t.Fatalf("collapsed to %d events", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i].At.Before(got[i-1].At) {
			t.Fatal("collapsing reordered events")
		}
	}
}

func TestCollapseListenerBlipsKeepsPartialReturns(t *testing.T) {
	t0 := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	events := []Event{
		blipEvent(t0, "disappeared", "tcp 0.0.0.0:22"),
		blipEvent(t0, "disappeared", "tcp 0.0.0.0:443"),
		blipEvent(t0, "disappeared", "tcp 0.0.0.0:9000"),
		blipEvent(t0.Add(30*time.Second), "appeared", "tcp 0.0.0.0:22"),
		blipEvent(t0.Add(4*time.Minute), "appeared", "tcp 0.0.0.0:443"),
		blipEvent(t0.Add(20*time.Minute), "appeared", "tcp 0.0.0.0:9000"), // too late
		blipEvent(t0.Add(21*time.Minute), "appeared", "tcp 0.0.0.0:8080"), // new, never left
	}
	got := CollapseListenerBlips(events)
	want := []string{
		"2 listeners went away and were back within 4m: tcp 0.0.0.0:22, tcp 0.0.0.0:443",
		"listener disappeared: tcp 0.0.0.0:9000",
		"listener appeared: tcp 0.0.0.0:9000",
		"listener appeared: tcp 0.0.0.0:8080",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d events: %+v", len(got), got)
	}
	for i, w := range want {
		if got[i].Summary != w {
			t.Errorf("event %d = %q, want %q", i, got[i].Summary, w)
		}
	}
	if got[0].Severity != "warning" || !got[0].At.Equal(t0) {
		t.Fatalf("summary event: %+v", got[0])
	}
}

func TestCollapseListenerBlipsPairsEachReturnOnce(t *testing.T) {
	t0 := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	key := "tcp 0.0.0.0:22"
	events := []Event{
		blipEvent(t0, "disappeared", key),
		blipEvent(t0.Add(time.Minute), "appeared", key),
		blipEvent(t0.Add(2*time.Minute), "disappeared", key),
		blipEvent(t0.Add(3*time.Minute), "appeared", key),
		blipEvent(t0.Add(4*time.Minute), "disappeared", key), // still down
	}
	got := CollapseListenerBlips(events)
	if len(got) != 3 {
		t.Fatalf("got %+v", got)
	}
	if got[0].Summary != "listener went away and was back within 60s: "+key || got[1].Summary != got[0].Summary {
		t.Fatalf("each outage should be its own blip: %+v", got)
	}
	if got[2].Summary != "listener disappeared: "+key {
		t.Fatalf("current outage must stay visible: %+v", got[2])
	}
}

func TestCollapseListenerBlipsLeavesOtherEventsAlone(t *testing.T) {
	t0 := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{At: t0, Category: "container", Severity: "warning", Summary: "container changed: app: running (healthy) -> running (unhealthy)"},
		blipEvent(t0.Add(time.Minute), "appeared", "tcp 0.0.0.0:22"),
		{At: t0.Add(2 * time.Minute), Category: "service", Severity: "warning", Summary: "service changed: ssh.service: active/running -> failed/failed"},
	}
	got := CollapseListenerBlips(events)
	if len(got) != len(events) {
		t.Fatalf("nothing to collapse, got %+v", got)
	}
	for i := range events {
		if got[i] != events[i] {
			t.Fatalf("event %d changed: %+v", i, got[i])
		}
	}
}

func TestReadEventsSurfacesRebootHiddenByListenerWave(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	boot := time.Date(2026, 10, 14, 22, 33, 8, 0, time.UTC)
	events := []Event{{At: boot, Category: "system", Severity: "warning", Summary: "host reboot detected: current boot started 2026-10-14T22:32:55Z"}}
	for i := 0; i < 120; i++ {
		events = append(events, blipEvent(boot, "disappeared", fmt.Sprintf("tcp 0.0.0.0:%d", 1000+i)))
	}
	for i := 0; i < 120; i++ {
		events = append(events, blipEvent(boot.Add(time.Minute), "appeared", fmt.Sprintf("tcp 0.0.0.0:%d", 1000+i)))
	}
	if err := store.AppendEvents(events); err != nil {
		t.Fatal(err)
	}
	got, err := store.ReadEvents(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !strings.HasPrefix(got[0].Summary, "host reboot detected") || !strings.HasPrefix(got[1].Summary, "120 listeners went away after the host reboot") {
		t.Fatalf("got %+v", got)
	}
	detailed, err := store.ReadDetailedEvents(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(detailed) != 241 {
		t.Fatalf("detailed view should keep every recorded listener event, got %d", len(detailed))
	}
}
