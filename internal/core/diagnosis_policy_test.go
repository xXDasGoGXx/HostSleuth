package core

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDiagnosisPolicyReachableStopsBeforeSecondaryEvidence(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldTLS := tlsProbeLookup
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return nil }
	tlsProbeLookup = func(context.Context, string, string) *TLSEvidence {
		return &TLSEvidence{HandshakeStatus: "unknown", HandshakeError: "TLS probe unavailable in reachability policy test"}
	}
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		t.Fatal("firewall evidence must not run after successful TCP")
		return boundedCommandResult{}, nil
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		tlsProbeLookup = oldTLS
		nftRulesetLookup = oldFirewall
	}()

	d := Diagnose(context.Background(), "127.0.0.1:443", Snapshot{})
	assertCheckOrder(t, d, []string{"dns", "route", "tcp"})
	if d.Conclusion != "target is reachable" || d.Confidence != "high" {
		t.Fatalf("unexpected diagnosis: %#v", d)
	}
}

func TestDiagnosisPolicyLocalListenerPrecedesFirewallCandidate(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return errors.New("connection refused") }
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "table inet filter { chain input { type filter hook input priority filter; policy drop; tcp dport 4444 reject; } }"}, nil
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		nftRulesetLookup = oldFirewall
	}()

	snap := Snapshot{
		Listeners:  []Listener{{Protocol: "tcp", Address: "127.0.0.1:4444", Process: "demo"}},
		Containers: []ContainerInfo{{Name: "other", Status: "Up 1 minute", Ports: "80/tcp"}},
	}
	d := Diagnose(context.Background(), "127.0.0.1:4444", snap)
	assertCheckOrder(t, d, []string{"dns", "route", "tcp", "local-listener", "docker-port", "firewall"})
	if d.Confidence != "medium" || !strings.Contains(d.Conclusion, "snapshot shows a listener on the requested local address") {
		t.Fatalf("unexpected diagnosis: %#v", d)
	}
	if d.Checks[5].Status != "unknown" || !strings.Contains(d.Checks[5].Evidence, "candidate evidence") {
		t.Fatalf("expected conservative firewall candidate evidence, got %#v", d.Checks[5])
	}
}

func TestDiagnosisPolicyFailedServiceCandidateDoesNotOverrideNoListener(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldFirewall := nftRulesetLookup
	oldJournal := journalUnitLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return errors.New("connection refused") }
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) { return boundedCommandResult{}, nil }
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "demo process exited with status 1\n"}, nil
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		nftRulesetLookup = oldFirewall
		journalUnitLookup = oldJournal
	}()

	snap := Snapshot{
		Containers: []ContainerInfo{{Name: "other", Status: "Up 1 minute", Ports: "80/tcp"}},
		Services:   []ServiceInfo{{Name: "demo.service", Active: "failed", Sub: "failed"}},
	}
	d := Diagnose(context.Background(), "127.0.0.1:5555", snap)
	assertCheckOrder(t, d, []string{"dns", "route", "tcp", "local-listener", "docker-port", "firewall", "systemd-failures"})
	if d.Conclusion != "no process appears to be listening on the requested local address and port" || d.Confidence != "high" {
		t.Fatalf("failed-unit candidate should not replace stronger listener conclusion: %#v", d)
	}
	if d.Checks[6].Status != "unknown" || !strings.Contains(d.Checks[6].Evidence, "not proven related to port") {
		t.Fatalf("unexpected systemd candidate: %#v", d.Checks[6])
	}
}

func TestDiagnosisPolicyDockerBindMismatchGetsSpecificOutcome(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return errors.New("connection refused") }
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) { return boundedCommandResult{}, nil }
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		nftRulesetLookup = oldFirewall
	}()

	snap := Snapshot{Containers: []ContainerInfo{{
		Name:     "chaptarr",
		Status:   "Up 1 minute",
		Ports:    "192.168.2.181:8789->8789/tcp",
		Networks: "chaptarr_default",
	}}}
	d := Diagnose(context.Background(), "127.0.0.1:8789", snap)
	if d.Conclusion != "no process is listening on the requested local address; Docker publishes TCP/8789 only on a different host address" || d.Confidence != "high" {
		t.Fatalf("unexpected bind-mismatch diagnosis: %#v", d)
	}
}

func TestDiagnosisPolicyDirectDockerPublishConflictReducesConfidence(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return errors.New("connection refused") }
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) { return boundedCommandResult{}, nil }
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		nftRulesetLookup = oldFirewall
	}()

	snap := Snapshot{Containers: []ContainerInfo{{Name: "demo", Status: "Up 1 minute", Ports: "0.0.0.0:6000->6000/tcp"}}}
	d := Diagnose(context.Background(), "127.0.0.1:6000", snap)
	if d.Confidence != "medium" || !strings.Contains(d.Conclusion, "current TCP/listener evidence disagrees") {
		t.Fatalf("conflicting snapshot evidence should reduce confidence: %#v", d)
	}
}

func TestDiagnosisPolicyRemoteNoRouteSkipsFirewall(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "RTNETLINK answers: Network is unreachable\n", errors.New("exit status 2")
	}
	tcpConnect = func(context.Context, string) error { return errors.New("network unreachable") }
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		t.Fatal("firewall evidence should not run after a confirmed remote no-route result")
		return boundedCommandResult{}, nil
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		nftRulesetLookup = oldFirewall
	}()

	d := Diagnose(context.Background(), "203.0.113.10:443", Snapshot{})
	assertCheckOrder(t, d, []string{"dns", "route", "tcp"})
	if d.Confidence != "high" || !strings.Contains(d.Conclusion, "destination unreachable") {
		t.Fatalf("unexpected no-route diagnosis: %#v", d)
	}
}

func TestDiagnosisPolicyUnavailableOptionalEvidenceRemainsNeutral(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldFirewall := nftRulesetLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "Cannot open netlink socket: Address family not supported by protocol\n", errors.New("exit status 1")
	}
	tcpConnect = func(context.Context, string) error { return errors.New("connection refused") }
	nftRulesetLookup = func(context.Context) (boundedCommandResult, error) {
		return boundedCommandResult{Output: "Unable to initialize Netlink socket: Operation not permitted\n"}, errors.New("exit status 1")
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		nftRulesetLookup = oldFirewall
	}()

	d := Diagnose(context.Background(), "127.0.0.1:65534", Snapshot{})
	if d.Confidence != "high" || d.Conclusion != "no process appears to be listening on the requested local address and port" {
		t.Fatalf("optional unknown evidence must not weaken stronger no-listener evidence: %#v", d)
	}
	if checkByName(d, "route").Status != "unknown" || checkByName(d, "firewall").Status != "unknown" || checkByName(d, "docker-port").Status != "unknown" {
		t.Fatalf("expected optional evidence to remain unknown: %#v", d.Checks)
	}
}

func assertCheckOrder(t *testing.T, d Diagnosis, want []string) {
	t.Helper()
	got := make([]string, len(d.Checks))
	for i, check := range d.Checks {
		got[i] = check.Name
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected check order: got %v want %v", got, want)
	}
}

func checkByName(d Diagnosis, name string) Check {
	for _, check := range d.Checks {
		if check.Name == name {
			return check
		}
	}
	return Check{}
}
