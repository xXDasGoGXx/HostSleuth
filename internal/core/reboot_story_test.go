package core

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestParseProcStatBootStartedAt(t *testing.T) {
	got, ok := parseProcStatBootStartedAt("btime 1789513490")
	if !ok {
		t.Fatal("expected btime to parse")
	}
	if got.Unix() != 1789513490 || got.Location() != time.UTC {
		t.Fatalf("unexpected boot time: %v", got)
	}
	if _, ok := parseProcStatBootStartedAt("cpu 1 2 3"); ok {
		t.Fatal("non-btime line must not parse")
	}
}

func TestDiffSnapshotsDetectsBootIdentityChange(t *testing.T) {
	started := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	at := started.Add(2 * time.Minute)
	oldSnap := Snapshot{SchemaVersion: rebootStorySchemaVersion, Host: HostInfo{BootID: "old-boot"}}
	newSnap := Snapshot{SchemaVersion: rebootStorySchemaVersion, CapturedAt: at, Host: HostInfo{BootID: "new-boot", BootStartedAt: started}}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) != 1 || events[0].Category != "system" || events[0].Severity != "warning" || !strings.Contains(events[0].Summary, "host reboot detected") {
		t.Fatalf("expected deterministic reboot event, got %#v", events)
	}
}

func TestDiffSnapshotsDoesNotInventRebootDuringSchemaBaseline(t *testing.T) {
	oldSnap := Snapshot{SchemaVersion: configFingerprintSchemaVersion, Host: HostInfo{}}
	newSnap := Snapshot{SchemaVersion: rebootStorySchemaVersion, Host: HostInfo{BootID: "new-boot", BootStartedAt: time.Now().UTC()}}
	for _, event := range DiffSnapshots(oldSnap, newSnap) {
		if strings.Contains(event.Summary, "reboot") {
			t.Fatalf("schema upgrade must not become reboot evidence: %#v", event)
		}
	}
}

func TestRebootStoryCorrelatesPostBootProblemsWithoutClaimingCause(t *testing.T) {
	oldLookup := previousBootJournalLookup
	previousBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "previous boot ended with ordinary bounded evidence\n"}, nil
	}
	t.Cleanup(func() { previousBootJournalLookup = oldLookup })

	boot := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	snap := Snapshot{
		Host: HostInfo{BootID: "boot-1", BootStartedAt: boot},
		Services: []ServiceInfo{
			{Name: "healthy.service", Active: "active", Sub: "running"},
			{Name: "broken.service", Active: "failed", Sub: "failed"},
		},
	}
	events := []Event{
		{At: boot.Add(-5 * time.Minute), Category: "package", Severity: "info", Summary: "package updated: kernel"},
		{At: boot.Add(3 * time.Minute), Category: "listener", Severity: "warning", Summary: "listener disappeared: tcp 0.0.0.0:443"},
		{At: boot.Add(20 * time.Minute), Category: "container", Severity: "warning", Summary: "outside window"},
	}
	story := BuildRebootStory(context.Background(), snap, events)
	if len(story.FailedServices) != 1 || story.FailedServices[0].Name != "broken.service" {
		t.Fatalf("expected current failed service: %#v", story.FailedServices)
	}
	if len(story.RelatedEvents) != 2 || len(story.ProblemEvents) != 1 {
		t.Fatalf("unexpected bounded boot context: related=%#v problem=%#v", story.RelatedEvents, story.ProblemEvents)
	}
	if story.PreviousBoot.Status != "available" {
		t.Fatalf("expected bounded previous journal evidence: %#v", story.PreviousBoot)
	}
	if !strings.Contains(story.CauseAssessment, "no reboot cause") || strings.Contains(strings.ToLower(story.Conclusion), "caused") {
		t.Fatalf("story must not invent reboot cause: %#v", story)
	}
}

func TestRebootStoryJournalUnavailableStaysUnknown(t *testing.T) {
	oldLookup := previousBootJournalLookup
	previousBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, exec.ErrNotFound
	}
	t.Cleanup(func() { previousBootJournalLookup = oldLookup })

	story := BuildRebootStory(context.Background(), Snapshot{Host: HostInfo{BootID: "boot", BootStartedAt: time.Now().UTC()}}, nil)
	if story.PreviousBoot.Status != "unknown" || !strings.Contains(story.PreviousBoot.Assessment, "unknown") {
		t.Fatalf("unavailable journal must remain unknown: %#v", story.PreviousBoot)
	}
}

func TestPreviousBootJournalErrorDoesNotBecomeShutdownClassification(t *testing.T) {
	oldLookup := previousBootJournalLookup
	previousBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, errors.New("permission denied")
	}
	t.Cleanup(func() { previousBootJournalLookup = oldLookup })

	story := BuildRebootStory(context.Background(), Snapshot{Host: HostInfo{BootID: "boot", BootStartedAt: time.Now().UTC()}}, nil)
	if story.PreviousBoot.Status != "unknown" || !strings.Contains(story.PreviousBoot.Evidence, "permission denied") {
		t.Fatalf("expected truthful permission boundary: %#v", story.PreviousBoot)
	}
	if strings.Contains(strings.ToLower(story.PreviousBoot.Assessment), "abnormal") && !strings.Contains(strings.ToLower(story.PreviousBoot.Assessment), "unknown") {
		t.Fatalf("journal error must not imply abnormal shutdown: %#v", story.PreviousBoot)
	}
}

func TestPreviousBootJournalClassificationRequiresDirectEvidence(t *testing.T) {
	oldLookup := previousBootJournalLookup
	t.Cleanup(func() { previousBootJournalLookup = oldLookup })

	cases := []struct {
		name       string
		output     string
		wantStatus string
	}{
		{name: "panic", output: "kernel: Kernel panic - not syncing: fatal test\n", wantStatus: "abnormal"},
		{name: "orderly", output: "systemd[1]: Reached target Power-Off.\nsystemd-shutdown[1]: Syncing filesystems and block devices.\n", wantStatus: "orderly"},
		{name: "bounded but inconclusive", output: "ordinary previous boot evidence\n", wantStatus: "available"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			previousBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
				return boundedCommandResult{Output: tc.output}, nil
			}
			got := collectPreviousBootEvidence(context.Background())
			if got.Status != tc.wantStatus {
				t.Fatalf("status=%q want %q: %#v", got.Status, tc.wantStatus, got)
			}
			if tc.wantStatus == "available" && !strings.Contains(got.Assessment, "unknown") {
				t.Fatalf("inconclusive bounded evidence must stay unknown: %#v", got)
			}
		})
	}
}

func TestRebootStoryRecoveryIssuesRequireCurrentAgreement(t *testing.T) {
	oldPrevious := previousBootJournalLookup
	oldCurrent := currentBootJournalLookup
	previousBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, exec.ErrNotFound
	}
	currentBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, exec.ErrNotFound
	}
	t.Cleanup(func() {
		previousBootJournalLookup = oldPrevious
		currentBootJournalLookup = oldCurrent
	})

	boot := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	snap := Snapshot{
		Host: HostInfo{BootID: "boot-1", BootStartedAt: boot},
		Services: []ServiceInfo{
			{Name: "recovered.service", Active: "active", Sub: "running"},
		},
		Containers: []ContainerInfo{
			{Name: "broken-container", Status: "Exited (1) 2 minutes ago"},
		},
	}
	events := []Event{
		{At: boot.Add(time.Minute), Category: "listener", Severity: "warning", Summary: "listener disappeared: tcp 0.0.0.0:443"},
		{At: boot.Add(2 * time.Minute), Category: "service", Severity: "warning", Summary: "service changed: recovered.service: active/running -> failed/failed"},
		{At: boot.Add(3 * time.Minute), Category: "container", Severity: "warning", Summary: "container changed: broken-container: running -> exited (1)"},
		{At: boot.Add(4 * time.Minute), Category: "package", Severity: "info", Summary: "package updated: linux-image"},
		{At: boot.Add(5 * time.Minute), Category: "configuration", Severity: "info", Summary: "configuration changed: /etc/example.conf"},
	}

	story := BuildRebootStory(context.Background(), snap, events)
	if len(story.RecoveryIssues) != 2 {
		t.Fatalf("expected listener and container recovery issues only: %#v", story.RecoveryIssues)
	}
	for _, issue := range story.RecoveryIssues {
		if issue.Kind == "service" {
			t.Fatalf("recovered service must not remain a recovery issue: %#v", story.RecoveryIssues)
		}
	}
	if len(story.ContextEvents) != 2 {
		t.Fatalf("expected package/configuration context: %#v", story.ContextEvents)
	}
}

func TestRebootStoryRecoveredListenerNotReported(t *testing.T) {
	boot := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	snap := Snapshot{
		Host:      HostInfo{BootID: "boot-1", BootStartedAt: boot},
		Listeners: []Listener{{Protocol: "tcp", Address: "0.0.0.0:443"}},
	}
	events := []Event{
		{At: boot.Add(time.Minute), Category: "listener", Severity: "warning", Summary: "listener disappeared: tcp 0.0.0.0:443"},
	}
	if got := unresolvedRecoveryIssues(snap, events); len(got) != 0 {
		t.Fatalf("listener present in current snapshot must not be reported as still missing: %#v", got)
	}
}

func TestCurrentBootJournalUnavailableStaysUnknown(t *testing.T) {
	oldLookup := currentBootJournalLookup
	currentBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, errors.New("permission denied")
	}
	t.Cleanup(func() { currentBootJournalLookup = oldLookup })

	got := collectCurrentBootEvidence(context.Background())
	if got.Status != "unknown" || !strings.Contains(got.Evidence, "permission denied") {
		t.Fatalf("unavailable current journal must remain unknown: %#v", got)
	}
}

func TestJournalPermissionDiagnosticStaysUnknown(t *testing.T) {
	oldPrevious := previousBootJournalLookup
	oldCurrent := currentBootJournalLookup
	diagnostic := "No journal files were opened due to insufficient permissions.\n"
	previousBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: diagnostic}, nil
	}
	currentBootJournalLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: diagnostic}, nil
	}
	t.Cleanup(func() {
		previousBootJournalLookup = oldPrevious
		currentBootJournalLookup = oldCurrent
	})

	previous := collectPreviousBootEvidence(context.Background())
	if previous.Status != "unknown" || !strings.Contains(previous.Evidence, "insufficient permissions") {
		t.Fatalf("permission-only previous journal output must stay unknown: %#v", previous)
	}
	current := collectCurrentBootEvidence(context.Background())
	if current.Status != "unknown" || !strings.Contains(current.Evidence, "insufficient permissions") {
		t.Fatalf("permission-only current journal output must stay unknown: %#v", current)
	}
}
