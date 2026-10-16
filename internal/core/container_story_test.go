package core

import (
	"strings"
	"testing"
	"time"
)

func containerStoryFixture() (Snapshot, []Event) {
	now := time.Date(2026, 10, 15, 22, 0, 0, 0, time.UTC)
	snap := Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: now, Mode: dockerDeploymentMode, Containers: []ContainerInfo{
		{ID: "a1", Name: "media", Image: "example/media:latest", ImageVersion: "1.42.1", Status: "Up 2 hours", StartedAt: "2026-10-15T20:00:00Z", ImageID: "sha256:m"},
		{ID: "b2", Name: "vpn", Image: "example/vpn:5", Status: "Up 4 minutes (healthy)", RestartCount: 6, StartedAt: "2026-10-15T21:56:00Z", ImageID: "sha256:v"},
		{ID: "c3", Name: "backup", Image: "example/backup:1", Status: "Up 3 days (unhealthy)", HealthFailingStreak: 4, StartedAt: "2026-10-12T22:00:00Z", ImageID: "sha256:b"},
		{ID: "d4", Name: "quiet", Image: "example/quiet:2", Status: "Up 9 days", StartedAt: "2026-10-06T22:00:00Z", ImageID: "sha256:q"},
	}}
	at := func(d time.Duration) time.Time { return now.Add(-d) }
	ev := func(t time.Time, cat, sev, s string) Event {
		return Event{At: t, Category: cat, Severity: sev, Summary: s}
	}
	events := []Event{
		ev(at(8*24*time.Hour), "container", "warning", "container restarted: vpn (crashed, exit code 1)"), // outside 7 days
		ev(at(30*time.Hour), "container", "warning", "container restarted: vpn (killed for running out of memory)"),
		ev(at(30*time.Hour), "listener", "warning", "listener disappeared: tcp 192.0.2.10:8080"),
		ev(at(30*time.Hour-time.Minute), "listener", "info", "listener appeared: tcp 192.0.2.10:8080"),
		ev(at(5*time.Hour), "container", "warning", "container restarted: vpn (crashed, exit code 2)"),
		ev(at(4*time.Minute), "container", "warning", "container restarted: vpn (crashed, exit code 2)"),
		ev(at(4*time.Minute), "container", "info", "container changed: vpn: running (healthy) -> running (unhealthy)"),
		ev(at(3*time.Minute), "container", "info", "container changed: vpn: running (unhealthy) -> running (healthy)"),
		ev(at(3*time.Hour), "listener", "warning", "listener disappeared: tcp 192.0.2.10:9999"), // not near any vpn change
		ev(at(2*time.Hour), "container", "info", "container updated: media: example/media:latest, version 1.41.0 → 1.42.1"),
		ev(at(90*time.Minute), "container", "info", "container restarted: media (stopped by SIGTERM, exit code 143)"),
		ev(at(10*time.Hour), "container", "info", "container changed: backup: running (healthy) -> running (unhealthy)"),
	}
	return snap, events
}

func TestContainerEventNameParsesEveryShape(t *testing.T) {
	cases := map[string][2]string{
		"container restarted: vpn (crashed, exit code 2)":                  {"vpn", "restart"},
		"container updated: media: example/media:latest, now version 2":    {"media", "update"},
		"container recreated: media (same image)":                          {"media", "recreate"},
		"container changed: vpn: running (healthy) -> running (unhealthy)": {"vpn", "state"},
		"container appeared: new-app (running)":                            {"new-app", "appeared"},
		"container disappeared: old-app (was running)":                     {"old-app", "disappeared"},
	}
	for summary, want := range cases {
		name, kind := containerEventName(Event{Category: "container", Summary: summary})
		if name != want[0] || kind != want[1] {
			t.Errorf("%q -> %q/%q, want %q/%q", summary, name, kind, want[0], want[1])
		}
	}
	if name, _ := containerEventName(Event{Category: "listener", Summary: "container restarted: x (y)"}); name != "" {
		t.Error("non-container category parsed")
	}
}

func TestContainerListPutsTroubleFirst(t *testing.T) {
	snap, events := containerStoryFixture()
	list := BuildContainerList(snap, events)
	var order []string
	for _, c := range list.Containers {
		order = append(order, c.Name+":"+c.Attention)
	}
	if got := strings.Join(order, " "); got != "backup:unhealthy vpn:crash-looping media: quiet:" {
		t.Fatalf("order = %s", got)
	}
	vpn := list.Containers[1]
	if vpn.Restarts7d != 3 || vpn.Restarts24h != 2 || vpn.Crashes7d != 3 || vpn.OOMKills7d != 1 || vpn.Unhealthy7d != 1 || vpn.RestartCount != 6 {
		t.Fatalf("vpn summary: %+v", vpn)
	}
	media := list.Containers[2]
	if media.Restarts7d != 1 || media.Crashes7d != 0 || !strings.Contains(media.LastUpdate, "1.41.0 → 1.42.1") || media.ImageVersion != "1.42.1" {
		t.Fatalf("media summary: %+v", media)
	}
}

func TestContainerStoryConclusionsAndNearbyContext(t *testing.T) {
	snap, events := containerStoryFixture()

	vpn := BuildContainerStory(snap, events, "vpn")
	if !vpn.Found || vpn.Status != "warning" {
		t.Fatalf("vpn story: %+v", vpn)
	}
	if want := "vpn is running now. It restarted 3 times in 7 days (3 crashes, 1 out-of-memory kill). Docker reported it unhealthy once."; vpn.Conclusion != want {
		t.Fatalf("vpn conclusion:\n%q\nwant\n%q", vpn.Conclusion, want)
	}
	if len(vpn.Timeline) != 5 {
		t.Fatalf("vpn timeline (7-day window): %d %+v", len(vpn.Timeline), vpn.Timeline)
	}
	if len(vpn.Nearby) != 2 || !strings.Contains(vpn.Nearby[0].Summary, ":8080") {
		t.Fatalf("nearby listener context: %+v", vpn.Nearby)
	}

	backup := BuildContainerStory(snap, events, "backup")
	if backup.Status != "fail" || backup.Conclusion != "backup is running but Docker reports it unhealthy, failing 4 health checks in a row." {
		t.Fatalf("backup: %s %q", backup.Status, backup.Conclusion)
	}

	media := BuildContainerStory(snap, events, "media")
	if media.Status != "pass" || media.Conclusion != "media is running and healthy. It restarted once in 7 days. None were crashes. Its image was updated in the last 7 days (example/media:latest, version 1.41.0 \u2192 1.42.1)." {
		t.Fatalf("media: %s %q", media.Status, media.Conclusion)
	}

	quiet := BuildContainerStory(snap, events, "quiet")
	if quiet.Status != "pass" || !strings.Contains(quiet.Conclusion, "no restarts, crashes or health problems") {
		t.Fatalf("quiet: %+v", quiet)
	}

	gone := BuildContainerStory(snap, events, "nope")
	if gone.Found || gone.Status != "unknown" {
		t.Fatalf("missing container: %+v", gone)
	}
}

func TestContainerStoryMatchesPublishedPortsOnly(t *testing.T) {
	now := time.Date(2026, 10, 15, 22, 0, 0, 0, time.UTC)
	snap := Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: now, Containers: []ContainerInfo{
		{ID: "a1", Name: "media", Image: "example/media:latest", Status: "Up 5 hours (healthy)", Ports: "0.0.0.0:8096->8096/tcp, [::]:8096->8096/tcp", StartedAt: "x"},
	}}
	u := now.Add(-5 * time.Hour)
	boot := now.Add(-3 * 24 * time.Hour)
	events := []Event{
		{At: boot, Category: "container", Severity: "info", Summary: "container changed: media: running (health: starting) -> running (healthy)"},
		{At: boot, Category: "listener", Severity: "info", Summary: "listener appeared: tcp 0.0.0.0:443"},
		{At: boot, Category: "listener", Severity: "info", Summary: "listener appeared: tcp 0.0.0.0:8096"},
		{At: u, Category: "container", Severity: "info", Summary: "container updated: media: example/media:latest, version 1 → 2"},
		{At: u, Category: "listener", Severity: "warning", Summary: "listener disappeared: tcp 0.0.0.0:8096"},
		{At: u, Category: "listener", Severity: "warning", Summary: "listener disappeared: tcp 0.0.0.0:8123"},
		{At: u.Add(time.Minute), Category: "listener", Severity: "info", Summary: "listener appeared: tcp 0.0.0.0:8096"},
	}
	story := BuildContainerStory(snap, events, "media")
	if len(story.Nearby) != 2 {
		t.Fatalf("want only the update's own published-port changes, got %+v", story.Nearby)
	}
	for _, e := range story.Nearby {
		if !strings.HasSuffix(e.Summary, ":8096") || e.At.Before(u) {
			t.Fatalf("unexpected nearby event: %+v", e)
		}
	}
	if !strings.Contains(strings.Join(story.Notes, " "), "ports this container publishes") {
		t.Fatalf("notes: %v", story.Notes)
	}
}
