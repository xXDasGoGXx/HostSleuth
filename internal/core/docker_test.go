package core

import (
	"context"
	"strings"
	"testing"
)

func TestParseDockerPublishedIPv4Binding(t *testing.T) {
	binding, ok := parseDockerPortBinding("0.0.0.0:3001->3001/tcp")
	if !ok {
		t.Fatal("expected IPv4 published binding to parse")
	}
	if !binding.Published || binding.HostIP != "0.0.0.0" || binding.HostPort != "3001" || binding.ContainerPort != "3001" || binding.Protocol != "tcp" {
		t.Fatalf("unexpected binding: %#v", binding)
	}
}

func TestParseDockerPublishedIPv6Binding(t *testing.T) {
	binding, ok := parseDockerPortBinding("[::]:443->443/tcp")
	if !ok {
		t.Fatal("expected IPv6 published binding to parse")
	}
	if !binding.Published || binding.HostIP != "::" || binding.HostPort != "443" || binding.ContainerPort != "443" || binding.Protocol != "tcp" {
		t.Fatalf("unexpected binding: %#v", binding)
	}
}

func TestParseDockerInternalOnlyPort(t *testing.T) {
	binding, ok := parseDockerPortBinding("8192/tcp")
	if !ok {
		t.Fatal("expected internal-only port to parse")
	}
	if binding.Published || binding.ContainerPort != "8192" || binding.Protocol != "tcp" {
		t.Fatalf("unexpected binding: %#v", binding)
	}
}

func TestDockerPortCheckMatchesWildcardPublishAndNetworks(t *testing.T) {
	snap := Snapshot{Containers: []ContainerInfo{{
		Name:     "uptime-kuma",
		Status:   "Up 10 minutes (healthy)",
		Ports:    "0.0.0.0:3001->3001/tcp, [::]:3001->3001/tcp",
		Networks: "monitoring",
	}}}
	check := dockerPortCheck(snap, "3001", []string{"127.0.0.1"})
	if check.Status != "fail" {
		t.Fatalf("expected failed correlation after TCP failure, got %#v", check)
	}
	if !strings.Contains(check.Evidence, "uptime-kuma") || !strings.Contains(check.Evidence, "0.0.0.0:3001->3001/tcp") || !strings.Contains(check.Evidence, "networks=monitoring") {
		t.Fatalf("unexpected evidence: %q", check.Evidence)
	}
}

func TestDockerPortCheckDetectsBindAddressMismatch(t *testing.T) {
	snap := Snapshot{Containers: []ContainerInfo{{
		Name:     "chaptarr",
		Status:   "Up 10 minutes",
		Ports:    "192.168.2.181:8789->8789/tcp",
		Networks: "chaptarr_default",
	}}}
	check := dockerPortCheck(snap, "8789", []string{"127.0.0.1"})
	if check.Status != "fail" || !strings.Contains(check.Evidence, "different host address") || !strings.Contains(check.Evidence, "192.168.2.181:8789") {
		t.Fatalf("unexpected check: %#v", check)
	}
}

func TestDockerPortCheckSurfacesInternalOnlyExposure(t *testing.T) {
	snap := Snapshot{Containers: []ContainerInfo{{
		Name:     "flaresolverr",
		Status:   "Up 10 minutes",
		Ports:    "0.0.0.0:8191->8191/tcp, 8192/tcp",
		Networks: "flaresolverr_default",
	}}}
	check := dockerPortCheck(snap, "8192", []string{"127.0.0.1"})
	if check.Status != "unknown" || !strings.Contains(check.Evidence, "internally but not published") || !strings.Contains(check.Evidence, "flaresolverr") {
		t.Fatalf("unexpected check: %#v", check)
	}
}

func TestDockerPortCheckUnknownWithoutInventory(t *testing.T) {
	check := dockerPortCheck(Snapshot{}, "3000", []string{"127.0.0.1"})
	if check.Status != "unknown" || !strings.Contains(check.Evidence, "Docker may be absent or inaccessible") {
		t.Fatalf("unexpected check: %#v", check)
	}
}

func TestDockerPortCheckPassesWhenInventoryHasNoMatchingPort(t *testing.T) {
	snap := Snapshot{Containers: []ContainerInfo{{Name: "demo", Status: "Up 10 minutes", Ports: "80/tcp"}}}
	check := dockerPortCheck(snap, "65534", []string{"127.0.0.1"})
	if check.Status != "pass" {
		t.Fatalf("expected pass, got %#v", check)
	}
}

func TestLocalListenerForTargetRejectsDifferentBindAddress(t *testing.T) {
	snap := Snapshot{Listeners: []Listener{{Protocol: "tcp", Address: "192.168.2.181:8789", Process: "docker-proxy"}}}
	if _, ok := LocalListenerForTarget(snap, "8789", []string{"127.0.0.1"}); ok {
		t.Fatal("listener bound to 192.168.2.181 must not match loopback target")
	}
	if _, ok := LocalListenerForTarget(snap, "8789", []string{"192.168.2.181"}); !ok {
		t.Fatal("listener should match its exact target address")
	}
}

func TestLocalListenerForTargetMatchesFamilyWildcardAndIgnoresUDP(t *testing.T) {
	snap := Snapshot{Listeners: []Listener{
		{Protocol: "udp", Address: "0.0.0.0:3001"},
		{Protocol: "tcp", Address: "0.0.0.0:3001"},
		{Protocol: "tcp", Address: "[::]:443"},
	}}
	listener, ok := LocalListenerForTarget(snap, "3001", []string{"127.0.0.1"})
	if !ok || listener.Protocol != "tcp" {
		t.Fatalf("expected TCP IPv4 wildcard match, got %#v, %v", listener, ok)
	}
	if _, ok := LocalListenerForTarget(snap, "443", []string{"::1"}); !ok {
		t.Fatal("expected IPv6 wildcard match")
	}
	if _, ok := LocalListenerForTarget(snap, "443", []string{"127.0.0.1"}); ok {
		t.Fatal("IPv6 wildcard should not be treated as proven IPv4 listener evidence")
	}
}

func TestDiagnoseDistinguishesListenerAndDockerBindMismatch(t *testing.T) {
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

	snap := Snapshot{
		Listeners: []Listener{{Protocol: "tcp", Address: "192.168.2.181:0", Process: "docker-proxy"}},
		Containers: []ContainerInfo{{
			Name:     "demo",
			Status:   "Up 10 minutes",
			Ports:    "192.168.2.181:0->8080/tcp",
			Networks: "demo_default",
		}},
	}
	d := Diagnose(context.Background(), "127.0.0.1:0", snap)
	var listenerCheck, dockerCheck *Check
	for i := range d.Checks {
		switch d.Checks[i].Name {
		case "local-listener":
			listenerCheck = &d.Checks[i]
		case "docker-port":
			dockerCheck = &d.Checks[i]
		}
	}
	if listenerCheck == nil || listenerCheck.Status != "fail" {
		t.Fatalf("expected target-aware listener failure, got %#v", listenerCheck)
	}
	if dockerCheck == nil || dockerCheck.Status != "fail" || !strings.Contains(dockerCheck.Evidence, "different host address") {
		t.Fatalf("expected Docker bind mismatch, got %#v", dockerCheck)
	}
	if d.Conclusion != "no process appears to be listening on the requested local port" || d.Confidence != "high" {
		t.Fatalf("unexpected diagnosis: %#v", d)
	}
}
