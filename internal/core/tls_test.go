package core

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProbeTLSSeparatesHandshakeHostnameAndTrust(t *testing.T) {
	certPEM, keyPEM, cert := makeTestCertificate(t, "example.test", time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour))
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{pair}})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		if tlsConn, ok := conn.(*tls.Conn); ok {
			_ = tlsConn.Handshake()
		}
	}()

	evidence := probeTLS(context.Background(), listener.Addr().String(), "example.test")
	if evidence == nil || evidence.HandshakeStatus != "pass" {
		t.Fatalf("expected successful TLS handshake, got %#v", evidence)
	}
	if evidence.HostnameStatus != "pass" {
		t.Fatalf("expected hostname match, got %#v", evidence)
	}
	if evidence.TrustStatus != "fail" {
		t.Fatalf("self-signed test certificate should not be trusted by system roots, got %#v", evidence)
	}
	if evidence.Certificate == nil || evidence.Certificate.SHA256Fingerprint == "" {
		t.Fatalf("expected certificate fingerprint, got %#v", evidence)
	}
	if evidence.Certificate.Subject != cert.Subject.String() {
		t.Fatalf("unexpected subject: %q", evidence.Certificate.Subject)
	}
}

func TestCertificateEvidenceIncludesBoundedMetadata(t *testing.T) {
	_, _, cert := makeTestCertificate(t, "tls.example.test", time.Now().Add(-time.Hour), time.Now().Add(48*time.Hour))
	evidence := certificateEvidence(cert, time.Now().UTC())
	if evidence == nil {
		t.Fatal("expected certificate evidence")
	}
	if len(evidence.SHA256Fingerprint) != 64 {
		t.Fatalf("expected 64-character SHA-256 fingerprint, got %q", evidence.SHA256Fingerprint)
	}
	if !containsString(evidence.SANs, "tls.example.test") {
		t.Fatalf("expected DNS SAN, got %v", evidence.SANs)
	}
	if evidence.DaysRemaining < 1 || evidence.DaysRemaining > 2 {
		t.Fatalf("unexpected remaining lifetime: %d days", evidence.DaysRemaining)
	}
}

func TestCollectCertbotEvidenceDetectsStaleServedCertificateOnlyWhenNewer(t *testing.T) {
	temp := t.TempDir()
	renewal := filepath.Join(temp, "renewal")
	live := filepath.Join(temp, "live", "example.test")
	if err := os.MkdirAll(renewal, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(live, 0o755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	_, _, oldCert := makeTestCertificate(t, "example.test", now.Add(-60*24*time.Hour), now.Add(30*24*time.Hour))
	newPEM, _, _ := makeTestCertificate(t, "example.test", now.Add(-time.Hour), now.Add(89*24*time.Hour))
	certPath := filepath.Join(live, "cert.pem")
	if err := os.WriteFile(certPath, newPEM, 0o644); err != nil {
		t.Fatal(err)
	}
	config := "cert = " + certPath + "\nfullchain = " + certPath + "\n[renewalparams]\nauthenticator = webroot\n"
	if err := os.WriteFile(filepath.Join(renewal, "example.test.conf"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	oldDir := certbotRenewalDirectory
	oldExec := certbotExecutableLookup
	oldSystemd := certbotSystemdLookup
	certbotRenewalDirectory = renewal
	certbotExecutableLookup = func() (string, bool) { return "/usr/bin/certbot", true }
	certbotSystemdLookup = func(context.Context, string) string { return "LoadState=loaded ActiveState=active" }
	defer func() {
		certbotRenewalDirectory = oldDir
		certbotExecutableLookup = oldExec
		certbotSystemdLookup = oldSystemd
	}()

	served := certificateEvidence(oldCert, now)
	evidence := collectCertbotEvidence(context.Background(), "example.test", served)
	if evidence == nil || !evidence.StaleServedCert {
		t.Fatalf("expected deterministic stale-served-certificate evidence, got %#v", evidence)
	}
	if !strings.Contains(evidence.Comparison, "newer/different") {
		t.Fatalf("unexpected comparison: %q", evidence.Comparison)
	}
}

func TestCompareCertbotWithServedDoesNotClaimStaleOnDifferenceAlone(t *testing.T) {
	now := time.Now().UTC()
	_, _, servedCert := makeTestCertificate(t, "example.test", now.Add(-time.Hour), now.Add(90*24*time.Hour))
	_, _, localCert := makeTestCertificate(t, "example.test", now.Add(-48*time.Hour), now.Add(30*24*time.Hour))
	served := certificateEvidence(servedCert, now)
	local := certificateEvidence(localCert, now)
	evidence := &CertbotEvidence{Lineages: []CertbotLineageEvidence{{
		Name:          "example.test",
		HostnameMatch: true,
		Certificate:   local,
	}}}
	compareCertbotWithServed(evidence, served)
	if evidence.StaleServedCert {
		t.Fatalf("difference alone must not prove stale service: %#v", evidence)
	}
	if !strings.Contains(evidence.Comparison, "does not prove") {
		t.Fatalf("expected conservative comparison wording, got %q", evidence.Comparison)
	}
}

func TestDiagnosisReportsTLSHostnameMismatchWithoutOverridingTCPReachability(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldTLS := tlsProbeLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return nil }
	tlsProbeLookup = func(context.Context, string, string) *TLSEvidence {
		return &TLSEvidence{
			ProbeConnected:   true,
			HandshakeStatus:  "pass",
			Protocol:         "TLS 1.3",
			CipherSuite:      "TLS_AES_128_GCM_SHA256",
			HostnameStatus:   "fail",
			HostnameEvidence: "certificate is valid for other.example.test, not 127.0.0.1",
			TrustStatus:      "pass",
			TrustEvidence:    "verified test chain",
			Certificate: &CertificateEvidence{
				Subject:           "CN=other.example.test",
				Issuer:            "CN=Test CA",
				Serial:            "01",
				SANs:              []string{"other.example.test"},
				ValidFrom:         time.Now().Add(-time.Hour),
				ValidUntil:        time.Now().Add(24 * time.Hour),
				DaysRemaining:     1,
				SHA256Fingerprint: strings.Repeat("A", 64),
			},
		}
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		tlsProbeLookup = oldTLS
	}()

	d := Diagnose(context.Background(), "127.0.0.1:443", Snapshot{Mode: dockerDeploymentMode})
	if d.Conclusion != "TCP and TLS are reachable, but the served certificate does not match the requested host" || d.Confidence != "high" {
		t.Fatalf("unexpected diagnosis: %#v", d)
	}
	if checkByName(d, "tcp").Status != "pass" || checkByName(d, "tls-hostname").Status != "fail" {
		t.Fatalf("expected separate transport and hostname evidence, got %#v", d.Checks)
	}
}

func TestDiagnosisDoesNotCallFreshTLSDialFailureATLSFailure(t *testing.T) {
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldTLS := tlsProbeLookup
	routeLookup = func(context.Context, string) (string, error) {
		return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil
	}
	tcpConnect = func(context.Context, string) error { return nil }
	tlsProbeLookup = func(context.Context, string, string) *TLSEvidence {
		return &TLSEvidence{HandshakeStatus: "unknown", HandshakeError: "fresh diagnostic connection was refused"}
	}
	defer func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		tlsProbeLookup = oldTLS
	}()

	d := Diagnose(context.Background(), "127.0.0.1:443", Snapshot{Mode: dockerDeploymentMode})
	if d.Conclusion != "target is reachable" {
		t.Fatalf("a contradictory second-dial failure must not erase successful TCP evidence: %#v", d)
	}
	if check := checkByName(d, "tls-handshake"); check.Name != "" {
		t.Fatalf("unavailable second TLS connection should remain structured evidence, not a false failed check: %#v", check)
	}
}

func makeTestCertificate(t *testing.T, dnsName string, notBefore, notAfter time.Time) ([]byte, []byte, *x509.Certificate) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: dnsName},
		DNSNames:              []string{dnsName},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, cert
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
