package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDiffSnapshotsDetectsServiceAndListenerChanges(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Services:   []ServiceInfo{{Name: "demo.service", Active: "active", Sub: "running"}},
		Listeners:  []Listener{{Protocol: "tcp", Address: "0.0.0.0:8080"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Services:   []ServiceInfo{{Name: "demo.service", Active: "failed", Sub: "failed"}},
	}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(events))
	}
}

func TestDiffSnapshotsIgnoresContainerUptimeOnlyChanges(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 5 minutes (healthy)"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 6 minutes (healthy)"}},
	}
	if events := DiffSnapshots(oldSnap, newSnap); len(events) != 0 {
		t.Fatalf("expected uptime-only container change to be ignored, got %#v", events)
	}
}

func TestDiffSnapshotsIgnoresContainerRestartTimerChurn(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Restarting (1) 5 seconds ago"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Restarting (1) 15 seconds ago"}},
	}
	if events := DiffSnapshots(oldSnap, newSnap); len(events) != 0 {
		t.Fatalf("expected restart-timer-only container change to be ignored, got %#v", events)
	}
}

func TestDiffSnapshotsDetectsContainerHealthChange(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 5 minutes (healthy)"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 6 minutes (unhealthy)"}},
	}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) != 1 {
		t.Fatalf("expected one health-change event, got %#v", events)
	}
	if events[0].Severity != "warning" {
		t.Fatalf("expected warning severity, got %q", events[0].Severity)
	}
	if events[0].Summary != "container changed: demo: running (healthy) -> running (unhealthy)" {
		t.Fatalf("unexpected summary: %q", events[0].Summary)
	}
}

func TestDiffSnapshotsDetectsContainerExitWithoutAgeNoise(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 5 minutes"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Exited (1) 3 seconds ago"}},
	}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) != 1 {
		t.Fatalf("expected one exit event, got %#v", events)
	}
	if events[0].Severity != "warning" || !strings.Contains(events[0].Summary, "running -> exited (1)") {
		t.Fatalf("unexpected exit event: %#v", events[0])
	}
}

func TestDiagnoseRejectsInvalidTarget(t *testing.T) {
	d := Diagnose(context.Background(), "not-a-host-port", Snapshot{})
	if d.Conclusion != "invalid target" {
		t.Fatalf("unexpected conclusion: %q", d.Conclusion)
	}
	if d.Confidence != "high" {
		t.Fatalf("unexpected confidence: %q", d.Confidence)
	}
}

func TestRouteCheckPassesKernelEvidence(t *testing.T) {
	old := routeLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "203.0.113.10 via 192.0.2.1 dev eth0 src 192.0.2.20 uid 1000\n", nil
	}
	defer func() { routeLookup = old }()

	check := routeCheck(context.Background(), "203.0.113.10")
	if check.Status != "pass" {
		t.Fatalf("expected pass, got %#v", check)
	}
	if !strings.Contains(check.Evidence, "via 192.0.2.1 dev eth0") {
		t.Fatalf("unexpected evidence: %q", check.Evidence)
	}
}

func TestRouteCheckTreatsExecutionRestrictionAsUnknown(t *testing.T) {
	old := routeLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "Cannot open netlink socket: Address family not supported by protocol\n", errors.New("exit status 1")
	}
	defer func() { routeLookup = old }()

	check := routeCheck(context.Background(), "203.0.113.10")
	if check.Status != "unknown" {
		t.Fatalf("expected unknown, got %#v", check)
	}
}

func TestRouteCheckDetectsKernelUnreachable(t *testing.T) {
	old := routeLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "RTNETLINK answers: Network is unreachable\n", errors.New("exit status 2")
	}
	defer func() { routeLookup = old }()

	check := routeCheck(context.Background(), "203.0.113.10")
	if check.Status != "fail" {
		t.Fatalf("expected fail, got %#v", check)
	}
}

func TestResolveCommandFallsBackToExecutableCandidate(t *testing.T) {
	candidate := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(candidate, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := resolveCommand("hostsleuth-command-that-should-not-exist", candidate)
	if got != candidate {
		t.Fatalf("expected %q, got %q", candidate, got)
	}
}

func TestFirewallCheckEmptyRuleset(t *testing.T) {
	old := nftRulesetLookup
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, nil
	}
	defer func() { nftRulesetLookup = old }()

	check := firewallCheck(context.Background(), "22")
	if check.Status != "pass" || check.Evidence != "nftables ruleset is empty" {
		t.Fatalf("unexpected firewall check: %#v", check)
	}
}

func TestFirewallCheckTreatsRestrictedReadAsUnknown(t *testing.T) {
	old := nftRulesetLookup
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "netlink: Error: cache initialization failed: Operation not permitted\n"}, errors.New("exit status 1")
	}
	defer func() { nftRulesetLookup = old }()

	check := firewallCheck(context.Background(), "22")
	if check.Status != "unknown" || !strings.Contains(check.Evidence, "Operation not permitted") {
		t.Fatalf("unexpected firewall check: %#v", check)
	}
}

func TestFirewallCheckSurfacesCandidateEvidence(t *testing.T) {
	old := nftRulesetLookup
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: `table inet filter {
	chain input {
		type filter hook input priority filter; policy drop;
		tcp dport { 22, 443 } accept
		tcp dport 8080 reject
	}
}`}, nil
	}
	defer func() { nftRulesetLookup = old }()

	check := firewallCheck(context.Background(), "8080")
	if check.Status != "unknown" {
		t.Fatalf("expected conservative unknown status, got %#v", check)
	}
	if !strings.Contains(check.Evidence, "policy drop") || !strings.Contains(check.Evidence, "tcp dport 8080 reject") {
		t.Fatalf("expected base policy and matching port candidate, got %q", check.Evidence)
	}
}

func TestFirewallCheckMarksTruncatedScan(t *testing.T) {
	old := nftRulesetLookup
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "table inet filter { chain input { type filter hook input priority filter; policy accept; } }", Truncated: true}, nil
	}
	defer func() { nftRulesetLookup = old }()

	check := firewallCheck(context.Background(), "443")
	if !strings.Contains(check.Evidence, "truncated at 64 KiB") {
		t.Fatalf("expected truncation warning, got %q", check.Evidence)
	}
}

func TestDiagnoseIncludesRouteEvidence(t *testing.T) {
	oldRoute := routeLookup
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, nil
	}
	defer func() {
		routeLookup = oldRoute
		nftRulesetLookup = oldFirewall
	}()

	d := Diagnose(context.Background(), "127.0.0.1:0", Snapshot{})
	for _, check := range d.Checks {
		if check.Name == "route" {
			if check.Status != "pass" || !strings.Contains(check.Evidence, "dev lo") {
				t.Fatalf("unexpected route check: %#v", check)
			}
			return
		}
	}
	t.Fatal("expected route check in diagnosis")
}

func TestDiagnoseIncludesFirewallEvidenceOnTCPFailure(t *testing.T) {
	oldRoute := routeLookup
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "table inet filter { chain input { type filter hook input priority filter; policy drop; tcp dport 65534 reject; } }"}, nil
	}
	defer func() {
		routeLookup = oldRoute
		nftRulesetLookup = oldFirewall
	}()

	d := Diagnose(context.Background(), "127.0.0.1:65534", Snapshot{})
	for _, check := range d.Checks {
		if check.Name == "firewall" {
			if check.Status != "unknown" || !strings.Contains(check.Evidence, "candidate evidence") {
				t.Fatalf("unexpected firewall check: %#v", check)
			}
			return
		}
	}
	t.Fatal("expected firewall check after TCP failure")
}

func TestStorePersistsSchemaAndReturnsEmptyEvents(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	if err := store.SaveSnapshot(Snapshot{Host: HostInfo{Hostname: "test-host"}}); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != 1 {
		t.Fatalf("expected snapshot schema 1, got %d", got.SchemaVersion)
	}
	events, err := store.ReadEvents(10)
	if err != nil {
		t.Fatal(err)
	}
	if events == nil || len(events) != 0 {
		t.Fatalf("expected empty non-nil events slice, got %#v", events)
	}
}

func TestStoreAddsEventSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	if err := store.AppendEvents([]Event{{Summary: "test"}}); err != nil {
		t.Fatal(err)
	}
	events, err := store.ReadEvents(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].SchemaVersion != 1 {
		t.Fatalf("unexpected events: %#v", events)
	}
}

func TestLegacyPersistedStateSurvivesCurrentVersionWrite(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	oldAt := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)
	newAt := oldAt.Add(24 * time.Hour)

	legacySnapshot := `{
  "schema_version": 1,
  "captured_at": "2025-12-01T10:00:00Z",
  "host": {
    "hostname": "legacy-host",
    "os": "Legacy Linux",
    "kernel": "6.1.0",
    "architecture": "amd64"
  },
  "listeners": [
    {"protocol":"tcp","address":"127.0.0.1:8787","process":"hostsleuth"}
  ]
}`
	if err := os.WriteFile(store.SnapshotPath(), []byte(legacySnapshot), 0o600); err != nil {
		t.Fatal(err)
	}
	legacyEvent := `{"schema_version":1,"at":"2025-12-01T10:00:00Z","category":"system","severity":"info","summary":"legacy event retained"}` + "\n"
	if err := os.WriteFile(store.EventsPath(), []byte(legacyEvent), 0o600); err != nil {
		t.Fatal(err)
	}
	legacyAudit := `{"schema_version":1,"at":"2025-12-01T10:00:00Z","phase":"completed","action_id":"service.restart","target":"legacy.service","status":"success","summary":"legacy audit retained"}` + "\n"
	if err := os.WriteFile(store.ActionsPath(), []byte(legacyAudit), 0o600); err != nil {
		t.Fatal(err)
	}

	previous, err := store.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if previous.SchemaVersion != 1 || previous.Host.Hostname != "legacy-host" || len(previous.Listeners) != 1 {
		t.Fatalf("legacy snapshot did not load intact: %#v", previous)
	}

	current := previous
	current.SchemaVersion = snapshotSchemaVersion
	current.CapturedAt = newAt
	current.Mode = dockerDeploymentMode
	current.Host.BootID = "current-boot"
	current.Host.BootStartedAt = newAt.Add(-time.Hour)
	if err := store.SaveSnapshot(current); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendEvents([]Event{{
		At:       newAt,
		Category: "system",
		Severity: "info",
		Summary:  "current event appended",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendActionAudit(ActionAudit{
		At:       newAt,
		Phase:    "completed",
		ActionID: actionServiceReloadID,
		Target:   "current.service",
		Status:   "success",
		Summary:  "current audit appended",
	}); err != nil {
		t.Fatal(err)
	}

	gotSnapshot, err := store.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if gotSnapshot.SchemaVersion != snapshotSchemaVersion || gotSnapshot.Host.Hostname != "legacy-host" || gotSnapshot.Mode != dockerDeploymentMode {
		t.Fatalf("current snapshot write lost compatible state: %#v", gotSnapshot)
	}
	if !gotSnapshot.CapturedAt.Equal(newAt) || gotSnapshot.Host.BootID != "current-boot" {
		t.Fatalf("current snapshot fields were not persisted: %#v", gotSnapshot)
	}

	events, err := store.ReadEvents(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("expected legacy + current events, got %#v", events)
	}
	if events[0].Summary != "legacy event retained" || !events[0].At.Equal(oldAt) {
		t.Fatalf("legacy event was not retained: %#v", events)
	}
	if events[1].Summary != "current event appended" || events[1].SchemaVersion != 1 {
		t.Fatalf("current event append was not preserved: %#v", events)
	}

	audits, err := store.ReadActionAudits(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(audits) != 2 {
		t.Fatalf("expected legacy + current audits, got %#v", audits)
	}
	if audits[0].Summary != "legacy audit retained" || audits[0].ActionID != actionServiceRestartID {
		t.Fatalf("legacy audit was not retained: %#v", audits)
	}
	if audits[1].Summary != "current audit appended" || audits[1].ActionID != actionServiceReloadID {
		t.Fatalf("current audit append was not preserved: %#v", audits)
	}
}
