package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const npmProxyFixture = `
# ------------------------------------------------------------
# example.test api.example.test
# ------------------------------------------------------------
server {
    set $forward_scheme http;
    set $server "192.168.2.181";
    set $port 8789;

    listen 80;
    listen [::]:80;
    listen 443 ssl;
    listen [::]:443 ssl;

    server_name example.test api.example.test;
    proxy_set_header Authorization "Bearer do-not-store-this-secret";
    proxy_pass $forward_scheme://$server:$port;
}
`

func TestParseNPMProxyConfigSafeFieldsOnly(t *testing.T) {
	route, ok := parseNPMProxyConfig(npmProxyFixture)
	if !ok {
		t.Fatal("expected NPM proxy config to parse")
	}
	if route.BackendScheme != "http" || route.BackendHost != "192.168.2.181" || route.BackendPort != "8789" {
		t.Fatalf("unexpected backend: %#v", route)
	}
	if strings.Join(route.Hostnames, ",") != "api.example.test,example.test" {
		t.Fatalf("unexpected hostnames: %#v", route.Hostnames)
	}
	if strings.Join(route.ListenPorts, ",") != "80,443" {
		t.Fatalf("unexpected listen ports: %#v", route.ListenPorts)
	}
	if !route.FrontendTLS {
		t.Fatal("expected ssl listen directive to mark the frontend as TLS")
	}
	encoded, err := json.Marshal(route)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "do-not-store-this-secret") || strings.Contains(string(encoded), "Authorization") {
		t.Fatalf("route retained unsafe config content: %s", encoded)
	}
}

func TestCollectNPMRoutesFromDataDirRejectsSymlink(t *testing.T) {
	dataDir := t.TempDir()
	proxyDir := filepath.Join(dataDir, "nginx", "proxy_host")
	if err := os.MkdirAll(proxyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proxyDir, "1.conf"), []byte(npmProxyFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.conf")
	if err := os.WriteFile(outside, []byte(strings.ReplaceAll(npmProxyFixture, "8789", "9999")), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(proxyDir, "2.conf")); err != nil {
		t.Fatal(err)
	}

	routes := collectNPMRoutesFromDataDir(ContainerInfo{Name: "jc21-npm"}, dataDir)
	if len(routes) != 1 {
		t.Fatalf("expected only regular config file, got %#v", routes)
	}
	if routes[0].RouteID != "1" || routes[0].Container != "jc21-npm" || routes[0].Provider != "nginx-proxy-manager" {
		t.Fatalf("unexpected route identity: %#v", routes[0])
	}
}

func TestCollectReverseProxiesDiscoversNPMDataMount(t *testing.T) {
	dataDir := t.TempDir()
	proxyDir := filepath.Join(dataDir, "nginx", "proxy_host")
	if err := os.MkdirAll(proxyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proxyDir, "7.conf"), []byte(npmProxyFixture), 0o600); err != nil {
		t.Fatal(err)
	}

	oldLookup := dockerMountLookup
	dockerMountLookup = func(context.Context, string) ([]dockerMount, error) {
		return []dockerMount{{Type: "volume", Source: dataDir, Destination: "/data"}}, nil
	}
	defer func() { dockerMountLookup = oldLookup }()

	routes := collectReverseProxies(context.Background(), []ContainerInfo{{
		ID:    "abc123",
		Name:  "jc21-npm",
		Image: "jc21/nginx-proxy-manager:latest",
	}})
	if len(routes) != 1 || routes[0].RouteID != "7" {
		t.Fatalf("unexpected routes: %#v", routes)
	}
}

func TestCollectReverseProxiesDegradesWhenInspectUnavailable(t *testing.T) {
	oldLookup := dockerMountLookup
	dockerMountLookup = func(context.Context, string) ([]dockerMount, error) {
		return nil, errors.New("docker unavailable")
	}
	defer func() { dockerMountLookup = oldLookup }()

	routes := collectReverseProxies(context.Background(), []ContainerInfo{{
		ID:    "abc123",
		Name:  "jc21-npm",
		Image: "jc21/nginx-proxy-manager:latest",
	}})
	if len(routes) != 0 {
		t.Fatalf("expected unavailable proxy evidence to degrade to empty inventory, got %#v", routes)
	}
}

func TestReverseProxyBackendCheckMatchesResolvedIP(t *testing.T) {
	snap := Snapshot{ReverseProxies: []ReverseProxyRoute{{
		Provider:      "nginx-proxy-manager",
		Container:     "jc21-npm",
		RouteID:       "12",
		Hostnames:     []string{"books.example.test"},
		ListenPorts:   []string{"80", "443"},
		BackendScheme: "http",
		BackendHost:   "192.168.2.181",
		BackendPort:   "8789",
	}}}
	check, ok := reverseProxyBackendCheck(snap, "backend.example.test", "8789", []string{"192.168.2.181"})
	if !ok || check.Name != "reverse-proxy" || check.Status != "unknown" {
		t.Fatalf("unexpected check: %#v, %v", check, ok)
	}
	if !strings.Contains(check.Evidence, "books.example.test") || !strings.Contains(check.Evidence, "192.168.2.181:8789") {
		t.Fatalf("unexpected evidence: %q", check.Evidence)
	}
}

func TestDiagnoseAddsProxyDependencyBeforeWeakCandidates(t *testing.T) {
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

	snap := Snapshot{ReverseProxies: []ReverseProxyRoute{{
		Provider:      "nginx-proxy-manager",
		Container:     "jc21-npm",
		RouteID:       "12",
		Hostnames:     []string{"books.example.test"},
		ListenPorts:   []string{"443"},
		BackendScheme: "http",
		BackendHost:   "127.0.0.1",
		BackendPort:   "0",
	}}}
	d := Diagnose(context.Background(), "127.0.0.1:0", snap)

	order := map[string]int{}
	for i, check := range d.Checks {
		order[check.Name] = i
	}
	proxyIndex, proxyOK := order["reverse-proxy"]
	firewallIndex, firewallOK := order["firewall"]
	systemdIndex, systemdOK := order["systemd-failures"]
	if !proxyOK || !firewallOK || !systemdOK {
		t.Fatalf("expected proxy/firewall/systemd checks, got %#v", d.Checks)
	}
	if !(proxyIndex < firewallIndex && proxyIndex < systemdIndex) {
		t.Fatalf("expected proxy dependency before weak candidates, order=%#v", order)
	}
	if d.Confidence != "high" {
		t.Fatalf("proxy candidate must not weaken stronger no-listener evidence: %#v", d)
	}
}
