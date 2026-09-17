package core

import (
	"strings"
	"testing"
)

func TestEvidenceFreeformPreservesRedactedURLShape(t *testing.T) {
	source := evidenceSource{Events: []Event{{
		Category: "system",
		Summary:  "request https://alice:pw@private.example/path?q=secret",
	}}}
	r := newEvidenceRedactor(source)
	out := r.freeform(source.Events[0].Summary)
	if !strings.Contains(out, "https://host-") || !strings.Contains(out, "/[redacted]") {
		t.Fatalf("redacted URL lost useful structure: %q", out)
	}
	for _, forbidden := range []string{"alice:pw", "private.example", "q=secret"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("URL redaction leaked %q in %q", forbidden, out)
		}
	}
}

func TestEvidenceFreeformPreservesIPv6CIDRShape(t *testing.T) {
	r := newEvidenceRedactor(evidenceSource{})
	out := r.freeform("route 2001:db8:abcd::/64 via gateway")
	if !strings.Contains(out, "ipv6-") || !strings.Contains(out, "/64") || strings.Contains(out, "path-") {
		t.Fatalf("IPv6 CIDR structure was not preserved: %q", out)
	}
	if strings.Contains(out, "2001:db8:abcd::") {
		t.Fatalf("IPv6 address survived redaction: %q", out)
	}
}
