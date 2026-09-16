package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

func withServiceStoryMocks(t *testing.T, runtime boundedCommandResult, journal boundedCommandResult) {
	t.Helper()
	oldRuntime := serviceRuntimeLookup
	oldJournal := serviceJournalLookup
	oldCgroup := serviceCgroupPIDs
	serviceRuntimeLookup = func(context.Context, string) (boundedCommandResult, error) { return runtime, nil }
	serviceJournalLookup = func(context.Context, string) (boundedCommandResult, error) { return journal, nil }
	serviceCgroupPIDs = func(string) ([]int, error) { return nil, nil }
	t.Cleanup(func() {
		serviceRuntimeLookup = oldRuntime
		serviceJournalLookup = oldJournal
		serviceCgroupPIDs = oldCgroup
	})
}

func TestServiceStoryFailedServiceDetectsPortCollision(t *testing.T) {
	withServiceStoryMocks(t,
		boundedCommandResult{Output: "LoadState=loaded\nActiveState=failed\nSubState=failed\nUnitFileState=enabled\nResult=exit-code\nMainPID=123\nControlGroup=/system.slice/demo.service\nExecMainCode=1\nExecMainStatus=1\n"},
		boundedCommandResult{Output: "demo failed password=secret\n"},
	)

	snap := Snapshot{
		Services:  []ServiceInfo{{Name: "demo.service", Load: "loaded", Active: "failed", Sub: "failed"}},
		Listeners: []Listener{{Protocol: "tcp", Address: "0.0.0.0:2525", Process: `users:(("other",pid=999,fd=3))`}},
	}
	story := ServiceStoryFor(context.Background(), "demo", "127.0.0.1:2525", snap, nil)
	if story.Conclusion != "service is failed and the expected port is occupied by another process" || story.Confidence != "high" {
		t.Fatalf("unexpected outcome: %#v", story)
	}
	var sawPort, sawRedaction bool
	for _, check := range story.Checks {
		if check.Name == "service-port" && check.Status == "fail" && strings.Contains(check.Evidence, "other") {
			sawPort = true
		}
		if check.Name == "service-journal" && strings.Contains(check.Evidence, "[REDACTED]") && !strings.Contains(check.Evidence, "secret") {
			sawRedaction = true
		}
	}
	if !sawPort || !sawRedaction {
		t.Fatalf("expected collision and redacted journal checks: %#v", story.Checks)
	}
}

func TestServiceStoryActiveServiceOwnsExpectedReachablePort(t *testing.T) {
	withServiceStoryMocks(t,
		boundedCommandResult{Output: "LoadState=loaded\nActiveState=active\nSubState=running\nUnitFileState=enabled\nResult=success\nMainPID=123\nControlGroup=/system.slice/demo.service\nExecMainCode=0\nExecMainStatus=0\n"},
		boundedCommandResult{Output: "started successfully\n"},
	)
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldTLS := tlsProbeLookup
	routeLookup = func(context.Context, string) (string, error) { return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil }
	tcpConnect = func(context.Context, string) error { return nil }
	tlsProbeLookup = func(context.Context, string, string) *TLSEvidence { return nil }
	t.Cleanup(func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		tlsProbeLookup = oldTLS
	})

	snap := Snapshot{
		Services:  []ServiceInfo{{Name: "demo.service", Load: "loaded", Active: "active", Sub: "running"}},
		Listeners: []Listener{{Protocol: "tcp", Address: "127.0.0.1:8443", Process: `users:(("demo",pid=123,fd=3))`}},
	}
	story := ServiceStoryFor(context.Background(), "demo.service", "127.0.0.1:8443", snap, nil)
	if story.Conclusion != "service is active and the expected endpoint is reachable" || story.Confidence != "high" {
		t.Fatalf("unexpected outcome: %#v", story)
	}
	if len(story.Listeners) != 1 || story.Endpoint == nil || story.Endpoint.Conclusion != "target is reachable" {
		t.Fatalf("missing listener/endpoint evidence: %#v", story)
	}
}

func TestServiceStoryContextUsesDirectAnchorAndNearbyChanges(t *testing.T) {
	anchor := time.Date(2026, 9, 16, 18, 0, 0, 0, time.UTC)
	events := []Event{
		{At: anchor.Add(-10 * time.Minute), Category: "package", Summary: "package updated: openssl"},
		{At: anchor, Category: "service", Summary: "service changed: demo.service: active/running -> failed/failed"},
		{At: anchor.Add(5 * time.Minute), Category: "configuration", Summary: "configuration changed: /etc/hosts"},
		{At: anchor.Add(45 * time.Minute), Category: "package", Summary: "package updated: unrelated"},
	}
	got := serviceContextEvents(events, "demo.service", "")
	if len(got) != 3 {
		t.Fatalf("expected direct event plus two nearby changes, got %#v", got)
	}
	for _, event := range got {
		if strings.Contains(event.Summary, "unrelated") {
			t.Fatalf("far event should not be included: %#v", got)
		}
	}
}

func TestServiceStoryDockerModeStaysTruthfullyUnavailable(t *testing.T) {
	story := ServiceStoryFor(context.Background(), "demo.service", "", Snapshot{Mode: dockerDeploymentMode}, nil)
	if story.Conclusion != "service story is unavailable in Docker deployment mode" || len(story.Checks) != 1 || story.Checks[0].Status != "unknown" {
		t.Fatalf("unexpected docker-mode story: %#v", story)
	}
}

func TestNormalizeServiceUnitAddsServiceSuffixOnlyWhenNeeded(t *testing.T) {
	if got := normalizeServiceUnit("postfix"); got != "postfix.service" {
		t.Fatalf("unexpected normalized unit %q", got)
	}
	if got := normalizeServiceUnit("demo@one.service"); got != "demo@one.service" {
		t.Fatalf("unexpected explicit unit %q", got)
	}
}
