package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

func moveTimeline(boot time.Time) []Event {
	ev := func(at time.Time, sev, summary string) Event {
		cat := "listener"
		if strings.HasPrefix(summary, "host reboot") {
			cat = "system"
		}
		return Event{SchemaVersion: 1, At: at, Category: cat, Severity: sev, Summary: summary}
	}
	first := boot.Add(15 * time.Second)
	second := first.Add(time.Minute)
	return []Event{
		// First post-boot snapshot: diffSet sorts "appeared" before
		// "disappeared", so a new port can precede the old one.
		ev(first, "warning", "host reboot detected: current boot started "+boot.Format(time.RFC3339)),
		ev(first, "info", "listener appeared: tcp 0.0.0.0:37925"),
		ev(first, "warning", "listener disappeared: tcp 0.0.0.0:60407"),
		ev(first, "warning", "listener disappeared: tcp [::]:42039"),
		ev(first, "warning", "listener disappeared: tcp [::]:42075"),
		ev(first, "warning", "listener disappeared: udp 127.0.0.1:769"),
		ev(first, "warning", "listener disappeared: tcp 127.0.0.1:36647"),
		ev(first, "warning", "listener disappeared: tcp 0.0.0.0:8096"), // fixed port: not explained
		// Next snapshot.
		ev(second, "info", "listener appeared: tcp [::]:33383"),
		ev(second, "info", "listener appeared: tcp [::]:46839"),
		ev(second, "info", "listener appeared: udp 127.0.0.1:772"),
		ev(second, "info", "listener appeared: tcp 0.0.0.0:3000"), // fixed port: no partner
	}
}

func TestDynamicPortMovesAfterRebootArePaired(t *testing.T) {
	boot := time.Date(2026, 10, 14, 22, 32, 55, 0, time.UTC)
	out := CollapseListenerBlips(moveTimeline(boot))

	var moved []Event
	remaining := map[string]bool{}
	for _, e := range out {
		if strings.Contains(e.Summary, "replaced by new ports") {
			moved = append(moved, e)
		}
		remaining[e.Summary] = true
	}
	if len(moved) != 1 {
		t.Fatalf("want one move summary, got %d: %+v", len(moved), out)
	}
	want := "after the host reboot, 4 dynamic-port listeners were replaced by new ports on the same address: tcp 0.0.0.0 60407 → 37925, tcp [::] 42039 → 33383, tcp [::] 42075 → 46839 and 1 more"
	if moved[0].Summary != want || moved[0].Severity != "info" {
		t.Fatalf("summary = %q (%s)\nwant      %q", moved[0].Summary, moved[0].Severity, want)
	}
	for _, gone := range []string{
		"listener appeared: tcp 0.0.0.0:37925", "listener disappeared: tcp 0.0.0.0:60407",
		"listener disappeared: tcp [::]:42039", "listener appeared: tcp [::]:33383",
		"listener disappeared: udp 127.0.0.1:769", "listener appeared: udp 127.0.0.1:772",
	} {
		if remaining[gone] {
			t.Errorf("paired event still listed: %s", gone)
		}
	}
	// No partner on the same address and class: still a real warning.
	for _, kept := range []string{
		"listener disappeared: tcp 127.0.0.1:36647",
		"listener disappeared: tcp 0.0.0.0:8096",
		"listener appeared: tcp 0.0.0.0:3000",
		"host reboot detected: current boot started 2026-10-14T22:32:55Z",
	} {
		if !remaining[kept] {
			t.Errorf("event wrongly folded away: %s", kept)
		}
	}
}

func TestDynamicPortMovesNeedARecordedReboot(t *testing.T) {
	at := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{At: at, Category: "listener", Severity: "warning", Summary: "listener disappeared: tcp [::]:42039"},
		{At: at.Add(time.Minute), Category: "listener", Severity: "info", Summary: "listener appeared: tcp [::]:33383"},
	}
	out := CollapseListenerBlips(events)
	if len(out) != 2 {
		t.Fatalf("a port change without a reboot must stay visible, got %+v", out)
	}
}

func TestDynamicPortMovesRespectClassAddressAndWindow(t *testing.T) {
	boot := time.Date(2026, 10, 14, 22, 32, 55, 0, time.UTC)
	reboot := Event{At: boot, Category: "system", Severity: "warning", Summary: "host reboot detected: current boot started " + boot.Format(time.RFC3339)}
	cases := map[string][2]string{
		"different class":   {"tcp 0.0.0.0:40000", "tcp 0.0.0.0:800"},
		"different address": {"tcp 127.0.0.1:40000", "tcp 0.0.0.0:41000"},
		"different proto":   {"tcp 0.0.0.0:40000", "udp 0.0.0.0:41000"},
	}
	for name, pair := range cases {
		out := CollapseListenerBlips([]Event{
			reboot,
			{At: boot, Category: "listener", Severity: "warning", Summary: "listener disappeared: " + pair[0]},
			{At: boot.Add(time.Minute), Category: "listener", Severity: "info", Summary: "listener appeared: " + pair[1]},
		})
		if len(out) != 3 {
			t.Errorf("%s: wrongly paired: %+v", name, out)
		}
	}
	late := CollapseListenerBlips([]Event{
		reboot,
		{At: boot, Category: "listener", Severity: "warning", Summary: "listener disappeared: tcp 0.0.0.0:40000"},
		{At: boot.Add(listenerBlipWindow + time.Minute), Category: "listener", Severity: "info", Summary: "listener appeared: tcp 0.0.0.0:41000"},
	})
	if len(late) != 3 {
		t.Errorf("pairing outside the window: %+v", late)
	}
}

// Reboot Story must not report a service that came back on a new
// dynamic port as "not recovered".
func TestRebootStoryDoesNotReportDynamicPortMovesAsUnrecovered(t *testing.T) {
	boot := time.Date(2026, 10, 14, 22, 32, 55, 0, time.UTC)
	snap := Snapshot{
		SchemaVersion: snapshotSchemaVersion,
		CapturedAt:    boot.Add(time.Hour),
		Mode:          dockerDeploymentMode,
		Host:          HostInfo{BootID: "726da9c0-b325-47c5-94a3-6d801ac4fd94", BootStartedAt: boot},
		Listeners: []Listener{
			{Protocol: "tcp", Address: "0.0.0.0:37925"},
			{Protocol: "tcp", Address: "[::]:33383"},
			{Protocol: "tcp", Address: "[::]:46839"},
			{Protocol: "udp", Address: "127.0.0.1:772"},
			{Protocol: "tcp", Address: "0.0.0.0:3000"},
		},
	}
	story := BuildRebootStory(context.Background(), snap, CollapseListenerBlips(MeaningfulEvents(moveTimeline(boot))))
	got := map[string]bool{}
	for _, issue := range story.RecoveryIssues {
		got[issue.Name] = true
	}
	for _, moved := range []string{"tcp 0.0.0.0:60407", "tcp [::]:42039", "tcp [::]:42075", "udp 127.0.0.1:769"} {
		if got[moved] {
			t.Errorf("moved listener reported as unrecovered: %s", moved)
		}
	}
	for _, missing := range []string{"tcp 127.0.0.1:36647", "tcp 0.0.0.0:8096"} {
		if !got[missing] {
			t.Errorf("real missing listener not reported: %s (issues %+v)", missing, story.RecoveryIssues)
		}
	}
}
