package core

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func TestTargetExpectsTLSFromPortOrProxyMetadata(t *testing.T) {
	if !targetExpectsTLS(Snapshot{}, "example.test", "443") {
		t.Fatal("TCP/443 should be treated as TLS")
	}
	snap := Snapshot{ReverseProxies: []ReverseProxyRoute{{
		Hostnames:   []string{"*.example.test"},
		ListenPorts: []string{"8443"},
		FrontendTLS: true,
	}}}
	if !targetExpectsTLS(snap, "api.example.test", "8443") {
		t.Fatal("proxy TLS metadata should mark nonstandard frontend ports as TLS")
	}
	if targetExpectsTLS(snap, "other.test", "8443") {
		t.Fatal("unrelated hostname must not inherit proxy TLS expectation")
	}
}

func TestDiagnosisTLSValidCertificate(t *testing.T) {
	now := time.Date(2026, 9, 15, 22, 0, 0, 0, time.UTC)
	leaf := testLeafForLoopback(now.Add(-time.Hour), now.Add(24*time.Hour))
	restore := stubTLSForDiagnosis(t, now, tlsProbeResult{
		Version:          tls.VersionTLS13,
		CipherSuite:      tls.TLS_AES_128_GCM_SHA256,
		PeerCertificates: []*x509.Certificate{leaf},
	}, nil, nil)
	defer restore()

	d := Diagnose(context.Background(), "127.0.0.1:443", Snapshot{})
	assertCheckOrder(t, d, []string{"dns", "route", "tcp", "tls-handshake", "tls-validity", "tls-hostname", "tls-trust"})
	if d.Conclusion != "target is reachable and the TLS certificate validates" || d.Confidence != "high" {
		t.Fatalf("unexpected TLS diagnosis: %#v", d)
	}
	if checkByName(d, "tls-handshake").Status != "pass" || !strings.Contains(checkByName(d, "tls-handshake").Evidence, "TLS 1.3") {
		t.Fatalf("unexpected handshake evidence: %#v", checkByName(d, "tls-handshake"))
	}
}

func TestTLSExpiredCertificateGetsSpecificConclusion(t *testing.T) {
	now := time.Date(2026, 9, 15, 22, 0, 0, 0, time.UTC)
	leaf := testLeafForLoopback(now.Add(-48*time.Hour), now.Add(-time.Hour))
	restore := stubTLSOnly(t, now, tlsProbeResult{Version: tls.VersionTLS13, CipherSuite: tls.TLS_AES_128_GCM_SHA256, PeerCertificates: []*x509.Certificate{leaf}}, nil, nil)
	defer restore()

	checks, conclusion, confidence := tlsDiagnosis(context.Background(), "127.0.0.1", "443")
	if conclusion != "TCP/TLS is reachable but the certificate is expired" || confidence != "high" {
		t.Fatalf("unexpected conclusion: %q %q", conclusion, confidence)
	}
	if checkNamed(checks, "tls-validity").Status != "fail" || !strings.Contains(checkNamed(checks, "tls-validity").Evidence, "expired") {
		t.Fatalf("unexpected validity check: %#v", checkNamed(checks, "tls-validity"))
	}
}

func TestTLSNotYetValidCertificateGetsSpecificConclusion(t *testing.T) {
	now := time.Date(2026, 9, 15, 22, 0, 0, 0, time.UTC)
	leaf := testLeafForLoopback(now.Add(time.Hour), now.Add(48*time.Hour))
	restore := stubTLSOnly(t, now, tlsProbeResult{Version: tls.VersionTLS12, CipherSuite: tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, PeerCertificates: []*x509.Certificate{leaf}}, nil, nil)
	defer restore()

	_, conclusion, confidence := tlsDiagnosis(context.Background(), "127.0.0.1", "443")
	if conclusion != "TCP/TLS is reachable but the certificate is not yet valid" || confidence != "high" {
		t.Fatalf("unexpected conclusion: %q %q", conclusion, confidence)
	}
}

func TestTLSHostnameMismatchGetsSpecificConclusion(t *testing.T) {
	now := time.Date(2026, 9, 15, 22, 0, 0, 0, time.UTC)
	leaf := &x509.Certificate{
		NotBefore:   now.Add(-time.Hour),
		NotAfter:    now.Add(24 * time.Hour),
		IPAddresses: []net.IP{net.ParseIP("192.0.2.10")},
	}
	restore := stubTLSOnly(t, now, tlsProbeResult{Version: tls.VersionTLS13, CipherSuite: tls.TLS_AES_128_GCM_SHA256, PeerCertificates: []*x509.Certificate{leaf}}, nil, nil)
	defer restore()

	checks, conclusion, confidence := tlsDiagnosis(context.Background(), "127.0.0.1", "443")
	if conclusion != "TCP/TLS is reachable but the certificate does not match the requested host" || confidence != "high" {
		t.Fatalf("unexpected conclusion: %q %q", conclusion, confidence)
	}
	if checkNamed(checks, "tls-hostname").Status != "fail" {
		t.Fatalf("unexpected hostname check: %#v", checkNamed(checks, "tls-hostname"))
	}
}

func TestTLSTrustFailureGetsSpecificConclusion(t *testing.T) {
	now := time.Date(2026, 9, 15, 22, 0, 0, 0, time.UTC)
	leaf := testLeafForLoopback(now.Add(-time.Hour), now.Add(24*time.Hour))
	restore := stubTLSOnly(t, now, tlsProbeResult{Version: tls.VersionTLS13, CipherSuite: tls.TLS_AES_128_GCM_SHA256, PeerCertificates: []*x509.Certificate{leaf}}, nil, errors.New("x509: certificate signed by unknown authority"))
	defer restore()

	checks, conclusion, confidence := tlsDiagnosis(context.Background(), "127.0.0.1", "443")
	if conclusion != "TCP/TLS is reachable but the certificate chain is not trusted by the system trust store" || confidence != "high" {
		t.Fatalf("unexpected conclusion: %q %q", conclusion, confidence)
	}
	if checkNamed(checks, "tls-trust").Status != "fail" || !strings.Contains(checkNamed(checks, "tls-trust").Evidence, "unknown authority") {
		t.Fatalf("unexpected trust check: %#v", checkNamed(checks, "tls-trust"))
	}
}

func TestTLSHandshakeFailurePreservesTCPReachabilityDistinction(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldProbe := tlsProbe
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return nil }
	tlsProbe = func(context.Context, string, string) (tlsProbeResult, error) {
		return tlsProbeResult{}, errors.New("remote error: tls: handshake failure")
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		tlsProbe = oldProbe
	}()

	d := Diagnose(context.Background(), "127.0.0.1:443", Snapshot{})
	assertCheckOrder(t, d, []string{"dns", "route", "tcp", "tls-handshake"})
	if checkByName(d, "tcp").Status != "pass" || checkByName(d, "tls-handshake").Status != "fail" {
		t.Fatalf("expected TCP pass plus TLS failure: %#v", d.Checks)
	}
	if d.Conclusion != "TCP is reachable but the TLS handshake failed" || d.Confidence != "high" {
		t.Fatalf("unexpected diagnosis: %#v", d)
	}
}

func testLeafForLoopback(notBefore, notAfter time.Time) *x509.Certificate {
	return &x509.Certificate{
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
}

func stubTLSForDiagnosis(t *testing.T, now time.Time, result tlsProbeResult, probeErr, chainErr error) func() {
	t.Helper()
	oldRoute := routeLookup
	oldTCP := tcpConnect
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return nil }
	restoreTLS := stubTLSOnly(t, now, result, probeErr, chainErr)
	return func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		restoreTLS()
	}
}

func stubTLSOnly(t *testing.T, now time.Time, result tlsProbeResult, probeErr, chainErr error) func() {
	t.Helper()
	oldProbe := tlsProbe
	oldChain := tlsChainVerify
	oldNow := tlsCurrentTime
	tlsProbe = func(context.Context, string, string) (tlsProbeResult, error) { return result, probeErr }
	tlsChainVerify = func(*x509.Certificate, []*x509.Certificate, time.Time) error { return chainErr }
	tlsCurrentTime = func() time.Time { return now }
	return func() {
		tlsProbe = oldProbe
		tlsChainVerify = oldChain
		tlsCurrentTime = oldNow
	}
}

func checkNamed(checks []Check, name string) Check {
	for _, check := range checks {
		if check.Name == name {
			return check
		}
	}
	return Check{}
}
