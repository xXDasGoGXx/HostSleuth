package core

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInspectFileHashesMetadataAndExpectedChecksum(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.txt")
	if err := os.WriteFile(path, []byte("hostsleuth\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	first, err := InspectFile(context.Background(), path, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.SHA256) != 64 || len(first.SHA512) != 128 {
		t.Fatalf("missing hashes: %#v", first)
	}
	if first.Permissions == "" || first.Size != int64(len("hostsleuth\n")) {
		t.Fatalf("unexpected metadata: %#v", first)
	}
	checked, err := InspectFile(context.Background(), path, "sha256:"+strings.ToLower(first.SHA256))
	if err != nil {
		t.Fatal(err)
	}
	if checked.ChecksumStatus != "match" || checked.ExpectedAlgorithm != "sha256" {
		t.Fatalf("expected checksum match: %#v", checked)
	}
}

func TestCompareFilesUsesFingerprintNotFilename(t *testing.T) {
	dir := t.TempDir()
	left := filepath.Join(dir, "one")
	right := filepath.Join(dir, "two")
	if err := os.WriteFile(left, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(right, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}
	comparison, err := CompareFiles(context.Background(), left, right)
	if err != nil {
		t.Fatal(err)
	}
	if !comparison.SameSHA256 || comparison.Conclusion == "" {
		t.Fatalf("expected identical content: %#v", comparison)
	}
	if err := os.WriteFile(right, []byte("different"), 0o644); err != nil {
		t.Fatal(err)
	}
	comparison, err = CompareFiles(context.Background(), left, right)
	if err != nil {
		t.Fatal(err)
	}
	if comparison.SameSHA256 {
		t.Fatalf("expected different fingerprints: %#v", comparison)
	}
}

func TestInspectDNSLocalhostIsBoundedAndDeterministic(t *testing.T) {
	result, err := InspectDNS(context.Background(), "localhost")
	if err != nil {
		t.Fatal(err)
	}
	if result.Resolver != "system resolver" || len(result.Lookups) == 0 {
		t.Fatalf("unexpected DNS result: %#v", result)
	}
	if len(result.Records) > workbenchDNSRecordLimit {
		t.Fatalf("DNS result exceeded bound: %d", len(result.Records))
	}
	for i := 1; i < len(result.Records); i++ {
		previous, current := result.Records[i-1], result.Records[i]
		if previous.Type > current.Type || (previous.Type == current.Type && previous.Value > current.Value) {
			t.Fatalf("DNS records are not deterministic: %#v", result.Records)
		}
	}
}

func TestInspectHTTPShowsRedirectChainWithoutFetchingBody(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("expected HEAD, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer final.Close()
	start := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL+"/done", http.StatusFound)
	}))
	defer start.Close()

	result, err := InspectHTTP(context.Background(), start.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Redirected || len(result.Hops) != 2 || result.FinalStatus != http.StatusNoContent {
		t.Fatalf("unexpected redirect evidence: %#v", result)
	}
	if result.Hops[0].StatusCode != http.StatusFound || result.Hops[1].StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected hop order: %#v", result.Hops)
	}
}

func TestCertificateFileInspectionDoesNotAcceptPrivateKeyFirst(t *testing.T) {
	certPEM, keyPEM, _ := makeTestCertificate(t, "workbench.test", time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour))
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		t.Fatal(err)
	}
	inspection, err := InspectCertificateFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	if inspection.Certificate == nil || len(inspection.Certificate.SHA256Fingerprint) != 64 {
		t.Fatalf("missing certificate evidence: %#v", inspection)
	}
	mixed := filepath.Join(dir, "key-first.pem")
	if err := os.WriteFile(mixed, append(append([]byte{}, keyPEM...), certPEM...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := InspectCertificateFile(mixed); err == nil || !strings.Contains(err.Error(), "private-key") {
		t.Fatalf("expected private-key boundary rejection, got %v", err)
	}
}

func TestCompareCertificateFileToServedUsesExactFingerprint(t *testing.T) {
	certPEM, keyPEM, _ := makeTestCertificate(t, "workbench.test", time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour))
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
	path := filepath.Join(t.TempDir(), "cert.pem")
	if err := os.WriteFile(path, certPEM, 0o644); err != nil {
		t.Fatal(err)
	}
	comparison, err := CompareCertificateFileToServed(context.Background(), path, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if comparison.Status != "match" || !strings.Contains(comparison.Conclusion, "exactly matches") {
		t.Fatalf("expected exact fingerprint match: %#v", comparison)
	}
}
