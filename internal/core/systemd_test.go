package core

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSystemdFailureCheckPassesWhenSnapshotHasNoFailedServices(t *testing.T) {
	old := journalUnitLookup
	calls := 0
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		calls++
		return boundedCommandResult{}, nil
	}
	defer func() { journalUnitLookup = old }()

	check := systemdFailureCheck(context.Background(), Snapshot{Services: []ServiceInfo{{Name: "demo.service", Active: "active", Sub: "running"}}})
	if check.Status != "pass" || !strings.Contains(check.Evidence, "no failed systemd services") {
		t.Fatalf("unexpected check: %#v", check)
	}
	if calls != 0 {
		t.Fatalf("journal should not be queried without failed services, calls=%d", calls)
	}
}

func TestSystemdFailureCheckSurfacesCandidateAndRedactsJournal(t *testing.T) {
	old := journalUnitLookup
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "startup failed token=supersecret Authorization: Bearer abc.def.ghi\n"}, nil
	}
	defer func() { journalUnitLookup = old }()

	snap := Snapshot{Services: []ServiceInfo{{Name: "demo.service", Active: "failed", Sub: "failed"}}}
	check := systemdFailureCheck(context.Background(), snap)
	if check.Status != "unknown" || !strings.Contains(check.Evidence, "demo.service failed/failed") {
		t.Fatalf("unexpected check: %#v", check)
	}
	if !strings.Contains(check.Evidence, "[REDACTED]") {
		t.Fatalf("expected redaction marker, got %q", check.Evidence)
	}
	if strings.Contains(check.Evidence, "supersecret") || strings.Contains(check.Evidence, "abc.def.ghi") {
		t.Fatalf("sensitive journal value leaked: %q", check.Evidence)
	}
}

func TestSystemdFailureCheckReportsUnavailableJournal(t *testing.T) {
	old := journalUnitLookup
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "Failed to open journal: Permission denied\n"}, errors.New("exit status 1")
	}
	defer func() { journalUnitLookup = old }()

	snap := Snapshot{Services: []ServiceInfo{{Name: "demo.service", Active: "failed", Sub: "failed"}}}
	check := systemdFailureCheck(context.Background(), snap)
	if check.Status != "unknown" || !strings.Contains(check.Evidence, "Permission denied") {
		t.Fatalf("unexpected check: %#v", check)
	}
}

func TestSystemdFailureCheckLimitsCandidates(t *testing.T) {
	old := journalUnitLookup
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		return boundedCommandResult{}, nil
	}
	defer func() { journalUnitLookup = old }()

	snap := Snapshot{Services: []ServiceInfo{
		{Name: "z.service", Active: "failed", Sub: "failed"},
		{Name: "a.service", Active: "failed", Sub: "failed"},
		{Name: "b.service", Active: "failed", Sub: "failed"},
		{Name: "c.service", Active: "failed", Sub: "failed"},
	}}
	check := systemdFailureCheck(context.Background(), snap)
	if !strings.Contains(check.Evidence, "a.service") || !strings.Contains(check.Evidence, "b.service") || !strings.Contains(check.Evidence, "c.service") {
		t.Fatalf("expected first three sorted candidates, got %q", check.Evidence)
	}
	if strings.Contains(check.Evidence, "z.service") || !strings.Contains(check.Evidence, "+1 additional failed service") {
		t.Fatalf("candidate limit not reflected correctly: %q", check.Evidence)
	}
}

func TestDiagnoseIncludesSystemdFailureEvidenceForLocalNoListener(t *testing.T) {
	oldRoute := routeLookup
	oldFirewall := nftRulesetLookup
	oldJournal := journalUnitLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, nil
	}
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "process exited with status 1\n"}, nil
	}
	defer func() {
		routeLookup = oldRoute
		nftRulesetLookup = oldFirewall
		journalUnitLookup = oldJournal
	}()

	snap := Snapshot{Services: []ServiceInfo{{Name: "demo.service", Active: "failed", Sub: "failed"}}}
	d := Diagnose(context.Background(), "127.0.0.1:0", snap)
	for _, check := range d.Checks {
		if check.Name == "systemd-failures" {
			if check.Status != "unknown" || !strings.Contains(check.Evidence, "demo.service") {
				t.Fatalf("unexpected systemd check: %#v", check)
			}
			return
		}
	}
	t.Fatal("expected systemd failure evidence for local no-listener diagnosis")
}

func TestDiagnoseSkipsSystemdFailureEvidenceWhenListenerExists(t *testing.T) {
	oldRoute := routeLookup
	oldFirewall := nftRulesetLookup
	oldJournal := journalUnitLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{}, nil
	}
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		t.Fatal("journal should not be queried when a local listener exists")
		return boundedCommandResult{}, nil
	}
	defer func() {
		routeLookup = oldRoute
		nftRulesetLookup = oldFirewall
		journalUnitLookup = oldJournal
	}()

	snap := Snapshot{
		Listeners: []Listener{{Protocol: "tcp", Address: "127.0.0.1:0", Process: "demo"}},
		Services:  []ServiceInfo{{Name: "demo.service", Active: "failed", Sub: "failed"}},
	}
	d := Diagnose(context.Background(), "127.0.0.1:0", snap)
	for _, check := range d.Checks {
		if check.Name == "systemd-failures" {
			t.Fatalf("unexpected systemd failure check: %#v", check)
		}
	}
}
