package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeDNSResolverSpecs(t *testing.T) {
	specs, err := normalizeDNSResolverSpecs([]string{"pihole-a=192.168.2.5", "2001:db8::53", "public=1.1.1.1:53"})
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(specs))
	}
	if specs[0].Label != "pihole-a" || specs[0].Server != "192.168.2.5:53" {
		t.Fatalf("unexpected first spec: %#v", specs[0])
	}
	if specs[1].Server != "[2001:db8::53]:53" {
		t.Fatalf("unexpected IPv6 resolver: %#v", specs[1])
	}
	if specs[2].Server != "1.1.1.1:53" {
		t.Fatalf("unexpected explicit port normalization: %#v", specs[2])
	}
}

func TestNormalizeDNSResolverSpecsRejectsUnsafeInputs(t *testing.T) {
	tests := [][]string{
		{"dns.example.test"},
		{"public=1.1.1.1:5353"},
		{"bad label=1.1.1.1"},
		{"a=1.1.1.1", "b=1.1.1.1"},
		{"same=1.1.1.1", "same=8.8.8.8"},
		{"1.1.1.1", "8.8.8.8", "9.9.9.9", "208.67.222.222", "8.8.4.4"},
	}
	for _, values := range tests {
		if _, err := normalizeDNSResolverSpecs(values); err == nil {
			t.Fatalf("expected resolver validation failure for %#v", values)
		}
	}
}

func TestDNSAddressScope(t *testing.T) {
	cases := map[string]string{
		"127.0.0.1":   "loopback",
		"192.168.1.2": "private",
		"169.254.1.2": "link-local",
		"1.1.1.1":     "global",
		"::1":         "loopback",
		"fd00::1":     "private",
	}
	for raw, want := range cases {
		if got := dnsAddressScope(parseTestIP(t, raw)); got != want {
			t.Fatalf("%s: got %s want %s", raw, got, want)
		}
	}
}

func TestCompareDNSResolverEvidenceAgree(t *testing.T) {
	results := []DNSResolverEvidence{
		{Label: "system", Status: "pass", Addresses: []DNSAddressEvidence{{Address: "192.168.2.10", Scope: "private"}}},
		{Label: "pihole", Status: "pass", Addresses: []DNSAddressEvidence{{Address: "192.168.2.10", Scope: "private"}}},
	}
	status, conclusion, differences, split := compareDNSResolverEvidence(results)
	if status != "agree" || len(differences) != 0 || split {
		t.Fatalf("unexpected comparison: %s %s %#v split=%v", status, conclusion, differences, split)
	}
}

func TestCompareDNSResolverEvidenceDetectsSplitViewPattern(t *testing.T) {
	results := []DNSResolverEvidence{
		{Label: "system", Status: "pass", Addresses: []DNSAddressEvidence{{Address: "192.168.2.10", Scope: "private"}}},
		{Label: "public", Status: "pass", Addresses: []DNSAddressEvidence{{Address: "203.0.113.10", Scope: "global"}}},
	}
	status, conclusion, differences, split := compareDNSResolverEvidence(results)
	if status != "diverge" || len(differences) != 1 || !split {
		t.Fatalf("unexpected comparison: %s %s %#v split=%v", status, conclusion, differences, split)
	}
	if !strings.Contains(conclusion, "consistent with split-view DNS") {
		t.Fatalf("expected bounded split-view hint, got %q", conclusion)
	}
}

func TestCompareDNSResolverEvidencePartialFailure(t *testing.T) {
	results := []DNSResolverEvidence{
		{Label: "system", Status: "pass", Addresses: []DNSAddressEvidence{{Address: "192.168.2.10", Scope: "private"}}},
		{Label: "other", Status: "fail", Error: "timeout"},
	}
	status, _, differences, split := compareDNSResolverEvidence(results)
	if status != "partial" || len(differences) != 0 || split {
		t.Fatalf("unexpected comparison: %s %#v split=%v", status, differences, split)
	}
}

func TestInspectDNSDetectiveUsesSystemAndExplicitResolvers(t *testing.T) {
	oldLookup := dnsDetectiveLookup
	oldPath := dnsResolvConfPath
	t.Cleanup(func() {
		dnsDetectiveLookup = oldLookup
		dnsResolvConfPath = oldPath
	})

	temp := t.TempDir()
	dnsResolvConfPath = filepath.Join(temp, "resolv.conf")
	if err := os.WriteFile(dnsResolvConfPath, []byte("nameserver 192.168.2.5\nsearch lab.example\noptions ndots:2 timeout:1 edns0\n"), 0600); err != nil {
		t.Fatal(err)
	}

	dnsDetectiveLookup = func(_ context.Context, spec dnsResolverSpec, name string) DNSResolverEvidence {
		if name != "app.example" {
			t.Fatalf("unexpected name %q", name)
		}
		address := "192.168.2.10"
		if spec.Label == "public" {
			address = "203.0.113.10"
		}
		return DNSResolverEvidence{
			Label: spec.Label,
			Server: spec.Server,
			Status: "pass",
			DurationMS: 2,
			Addresses: []DNSAddressEvidence{{Address: address, Family: "IPv4", Scope: map[bool]string{true: "global", false: "private"}[spec.Label == "public"]}},
		}
	}

	result, err := InspectDNSDetective(context.Background(), "app.example", []string{"public=1.1.1.1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "diverge" || !result.SplitViewHint {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(result.Resolvers) != 2 || result.Resolvers[0].Label != "system" || result.Resolvers[1].Label != "public" {
		t.Fatalf("unexpected resolver ordering: %#v", result.Resolvers)
	}
	if len(result.Runtime.Nameservers) != 1 || result.Runtime.Nameservers[0] != "192.168.2.5" {
		t.Fatalf("unexpected runtime config: %#v", result.Runtime)
	}
	if len(result.Runtime.Options) != 2 {
		t.Fatalf("expected bounded resolver options, got %#v", result.Runtime.Options)
	}
}

func TestInspectDNSDetectiveSingleResolver(t *testing.T) {
	oldLookup := dnsDetectiveLookup
	t.Cleanup(func() { dnsDetectiveLookup = oldLookup })
	dnsDetectiveLookup = func(_ context.Context, spec dnsResolverSpec, _ string) DNSResolverEvidence {
		return DNSResolverEvidence{Label: spec.Label, Server: spec.Server, Status: "pass", DurationMS: time.Millisecond.Milliseconds()}
	}
	result, err := InspectDNSDetective(context.Background(), "localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "single" {
		t.Fatalf("expected single status, got %#v", result)
	}
}

func parseTestIP(t *testing.T, raw string) []byte {
	t.Helper()
	ip := parseIPForTest(raw)
	if ip == nil {
		t.Fatalf("invalid test IP %s", raw)
	}
	return ip
}
