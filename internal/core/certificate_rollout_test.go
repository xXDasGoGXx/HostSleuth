package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNormalizeCertificateFingerprint(t *testing.T) {
	value := "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"
	got, err := normalizeCertificateFingerprint(value)
	if err != nil {
		t.Fatal(err)
	}
	want := "AABBCCDDEEFF00112233445566778899AABBCCDDEEFF00112233445566778899"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}

	for _, bad := range []string{"abcd", strings.Repeat("G", 64), strings.Repeat("A", 66)} {
		if _, err := normalizeCertificateFingerprint(bad); err == nil {
			t.Fatalf("expected invalid fingerprint %q", bad)
		}
	}
}

func TestCertificateRolloutRequiresExactlyOneExpectedSource(t *testing.T) {
	endpoints := []string{"one.example:443"}
	if _, err := BuildCertificateRolloutStory(t.Context(), "", "", endpoints); err == nil {
		t.Fatal("expected missing-source validation failure")
	}
	if _, err := BuildCertificateRolloutStory(t.Context(), strings.Repeat("A", 64), "ref.example:443", endpoints); err == nil {
		t.Fatal("expected conflicting-source validation failure")
	}
}

func TestCertificateRolloutEndpointValidationIsBoundedAndDeterministic(t *testing.T) {
	if _, err := validateCertificateRolloutEndpoints(nil); err == nil {
		t.Fatal("expected missing endpoint failure")
	}
	if _, err := validateCertificateRolloutEndpoints([]string{"example.test:https"}); err == nil {
		t.Fatal("expected numeric-port failure")
	}
	if _, err := validateCertificateRolloutEndpoints([]string{"EXAMPLE.test:443", "example.TEST:443"}); err == nil {
		t.Fatal("expected duplicate endpoint failure")
	}
	tooMany := make([]string, certificateRolloutMaxEndpoints+1)
	for i := range tooMany {
		tooMany[i] = "example.test:443"
	}
	if _, err := validateCertificateRolloutEndpoints(tooMany); err == nil {
		t.Fatal("expected endpoint limit failure")
	}
}

func TestCertificateRolloutFindsMismatchBeforeHealthProblems(t *testing.T) {
	old := certificateRolloutTLSProbe
	defer func() { certificateRolloutTLSProbe = old }()

	expected := strings.Repeat("A", 64)
	certificateRolloutTLSProbe = func(_ context.Context, target, _ string) *TLSEvidence {
		switch target {
		case "good.example:443":
			return rolloutTLSEvidence(expected, "pass", "pass")
		case "old.example:443":
			return rolloutTLSEvidence(strings.Repeat("B", 64), "pass", "pass")
		default:
			return nil
		}
	}

	story, err := BuildCertificateRolloutStory(t.Context(), expected, "", []string{
		"good.example:443",
		"old.example:443",
	})
	if err != nil {
		t.Fatal(err)
	}
	if story.Status != "fail" || story.FirstProblem != "old.example:443" {
		t.Fatalf("unexpected story: %#v", story)
	}
	if story.Summary.Matched != 1 || story.Summary.Mismatched != 1 || story.Summary.Unknown != 0 {
		t.Fatalf("unexpected summary: %#v", story.Summary)
	}
	if story.Endpoints[0].MatchStatus != "match" || story.Endpoints[1].MatchStatus != "mismatch" {
		t.Fatalf("unexpected endpoint matches: %#v", story.Endpoints)
	}
}

func TestCertificateRolloutSeparatesMatchFromCertificateHealth(t *testing.T) {
	old := certificateRolloutTLSProbe
	defer func() { certificateRolloutTLSProbe = old }()

	expected := strings.Repeat("C", 64)
	certificateRolloutTLSProbe = func(_ context.Context, _, _ string) *TLSEvidence {
		return rolloutTLSEvidence(expected, "fail", "pass")
	}

	story, err := BuildCertificateRolloutStory(t.Context(), expected, "", []string{"wrong-name.example:443"})
	if err != nil {
		t.Fatal(err)
	}
	if story.Endpoints[0].MatchStatus != "match" {
		t.Fatalf("fingerprint should match: %#v", story.Endpoints[0])
	}
	if story.Endpoints[0].Status != "fail" || story.Status != "fail" {
		t.Fatalf("hostname failure must remain visible: %#v", story)
	}
	if !strings.Contains(story.Conclusion, "fingerprints match") {
		t.Fatalf("unexpected conclusion: %q", story.Conclusion)
	}
}

func TestCertificateRolloutUsesUnknownWhenFingerprintCannotBeObserved(t *testing.T) {
	old := certificateRolloutTLSProbe
	defer func() { certificateRolloutTLSProbe = old }()

	certificateRolloutTLSProbe = func(_ context.Context, _, _ string) *TLSEvidence {
		return &TLSEvidence{HandshakeStatus: "fail", HandshakeError: "connection reset"}
	}
	story, err := BuildCertificateRolloutStory(t.Context(), strings.Repeat("D", 64), "", []string{"down.example:443"})
	if err != nil {
		t.Fatal(err)
	}
	if story.Status != "unknown" || story.Endpoints[0].MatchStatus != "unknown" {
		t.Fatalf("unobserved fingerprint must be unknown: %#v", story)
	}
}

func TestCertificateRolloutReferenceEndpointBecomesExpectedFingerprint(t *testing.T) {
	old := certificateRolloutTLSProbe
	defer func() { certificateRolloutTLSProbe = old }()

	expected := strings.Repeat("E", 64)
	certificateRolloutTLSProbe = func(_ context.Context, target, _ string) *TLSEvidence {
		switch target {
		case "reference.example:443", "edge-a.example:443", "edge-b.example:443":
			return rolloutTLSEvidence(expected, "pass", "pass")
		default:
			return nil
		}
	}

	story, err := BuildCertificateRolloutStory(t.Context(), "", "reference.example:443", []string{
		"edge-a.example:443",
		"edge-b.example:443",
	})
	if err != nil {
		t.Fatal(err)
	}
	if story.Expected.Source != "reference" || story.Expected.Fingerprint != expected {
		t.Fatalf("unexpected expected source: %#v", story.Expected)
	}
	if story.Status != "pass" || story.Summary.Matched != 2 || story.Summary.Healthy != 2 {
		t.Fatalf("unexpected story: %#v", story)
	}
}

func TestCertificateRolloutStopsWhenReferenceCertificateIsUnavailable(t *testing.T) {
	old := certificateRolloutTLSProbe
	defer func() { certificateRolloutTLSProbe = old }()

	calls := 0
	certificateRolloutTLSProbe = func(_ context.Context, target, _ string) *TLSEvidence {
		calls++
		if target != "reference.example:443" {
			t.Fatalf("target endpoints must not be probed without a reference fingerprint: %s", target)
		}
		return &TLSEvidence{HandshakeStatus: "fail", HandshakeError: "TLS unavailable"}
	}

	story, err := BuildCertificateRolloutStory(t.Context(), "", "reference.example:443", []string{"edge.example:443"})
	if err != nil {
		t.Fatal(err)
	}
	if story.Status != "unknown" || story.FirstProblem != "expected-reference" || calls != 1 {
		t.Fatalf("unexpected story/calls: story=%#v calls=%d", story, calls)
	}
}

func rolloutTLSEvidence(fingerprint, hostnameStatus, trustStatus string) *TLSEvidence {
	now := time.Now().UTC()
	return &TLSEvidence{
		ProbeConnected:   true,
		HandshakeStatus:  "pass",
		ServerName:       "example.test",
		Protocol:         "TLS 1.3",
		CipherSuite:      "TLS_AES_128_GCM_SHA256",
		HostnameStatus:   hostnameStatus,
		HostnameEvidence: map[string]string{"pass": "certificate matches host", "fail": "certificate hostname mismatch"}[hostnameStatus],
		TrustStatus:      trustStatus,
		TrustEvidence:    map[string]string{"pass": "certificate chain verified", "fail": "certificate chain is not trusted"}[trustStatus],
		Certificate: &CertificateEvidence{
			Subject:           "CN=example.test",
			Issuer:            "CN=Test CA",
			Serial:            "01",
			SANs:              []string{"example.test"},
			ValidFrom:         now.Add(-time.Hour),
			ValidUntil:        now.Add(30 * 24 * time.Hour),
			DaysRemaining:     30,
			SHA256Fingerprint: fingerprint,
		},
	}
}
