package core

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"
)

type tlsProbeResult struct {
	Version          uint16
	CipherSuite      uint16
	PeerCertificates []*x509.Certificate
}

var tlsProbe = func(ctx context.Context, address, serverName string) (tlsProbeResult, error) {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	raw, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return tlsProbeResult{}, err
	}
	defer raw.Close()

	// Verification is deliberately performed below as separate deterministic
	// checks so HostSleuth can distinguish time validity, hostname, and trust
	// failures instead of collapsing them into one handshake error.
	conn := tls.Client(raw, &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: true, // #nosec G402 -- manual verification follows.
	})
	handshakeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := conn.HandshakeContext(handshakeCtx); err != nil {
		return tlsProbeResult{}, err
	}
	state := conn.ConnectionState()
	return tlsProbeResult{
		Version:          state.Version,
		CipherSuite:      state.CipherSuite,
		PeerCertificates: state.PeerCertificates,
	}, nil
}

var tlsChainVerify = func(leaf *x509.Certificate, peers []*x509.Certificate, currentTime time.Time) error {
	intermediates := x509.NewCertPool()
	for _, cert := range peers[1:] {
		intermediates.AddCert(cert)
	}
	_, err := leaf.Verify(x509.VerifyOptions{
		Intermediates: intermediates,
		CurrentTime:   currentTime,
	})
	return err
}

var tlsCurrentTime = func() time.Time { return time.Now().UTC() }

func targetExpectsTLS(snap Snapshot, host, port string) bool {
	if port == "443" {
		return true
	}
	for _, route := range snap.ReverseProxies {
		if !route.FrontendTLS || !stringSliceContains(route.ListenPorts, port) {
			continue
		}
		for _, hostname := range route.Hostnames {
			if proxyFrontendHostnameMatches(hostname, host) {
				return true
			}
		}
	}
	return false
}

func proxyFrontendHostnameMatches(pattern, host string) bool {
	pattern = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(pattern)), ".")
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if pattern == "" || host == "" {
		return false
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		return host != strings.TrimPrefix(suffix, ".") && strings.HasSuffix(host, suffix)
	}
	return pattern == host
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func tlsDiagnosis(ctx context.Context, host, port string) ([]Check, string, string) {
	result, err := tlsProbe(ctx, net.JoinHostPort(host, port), host)
	if err != nil {
		return []Check{{
			Name:     "tls-handshake",
			Status:   "fail",
			Evidence: boundedEvidence(err.Error(), 512),
		}}, "TCP is reachable but the TLS handshake failed", "high"
	}

	checks := []Check{{
		Name:     "tls-handshake",
		Status:   "pass",
		Evidence: tlsVersionName(result.Version) + "; cipher=" + tls.CipherSuiteName(result.CipherSuite),
	}}
	if len(result.PeerCertificates) == 0 || result.PeerCertificates[0] == nil {
		checks = append(checks, Check{Name: "tls-certificate", Status: "fail", Evidence: "TLS peer returned no leaf certificate"})
		return checks, "TCP/TLS is reachable but the peer returned no certificate", "high"
	}

	leaf := result.PeerCertificates[0]
	now := tlsCurrentTime()
	validity := Check{
		Name:   "tls-validity",
		Status: "pass",
		Evidence: fmt.Sprintf("certificate valid %s through %s",
			leaf.NotBefore.UTC().Format(time.RFC3339), leaf.NotAfter.UTC().Format(time.RFC3339)),
	}
	validityConclusion := ""
	if now.Before(leaf.NotBefore) {
		validity.Status = "fail"
		validity.Evidence = "certificate is not valid until " + leaf.NotBefore.UTC().Format(time.RFC3339)
		validityConclusion = "TCP/TLS is reachable but the certificate is not yet valid"
	} else if now.After(leaf.NotAfter) {
		validity.Status = "fail"
		validity.Evidence = "certificate expired at " + leaf.NotAfter.UTC().Format(time.RFC3339)
		validityConclusion = "TCP/TLS is reachable but the certificate is expired"
	}
	checks = append(checks, validity)

	hostname := Check{Name: "tls-hostname", Status: "pass", Evidence: "certificate covers " + host}
	hostnameConclusion := ""
	if err := leaf.VerifyHostname(host); err != nil {
		hostname.Status = "fail"
		hostname.Evidence = boundedEvidence(err.Error(), 512)
		hostnameConclusion = "TCP/TLS is reachable but the certificate does not match the requested host"
	}
	checks = append(checks, hostname)

	chainTime := now
	if now.Before(leaf.NotBefore) {
		chainTime = leaf.NotBefore.Add(time.Second)
	} else if now.After(leaf.NotAfter) {
		chainTime = leaf.NotAfter.Add(-time.Second)
	}
	trust := Check{Name: "tls-trust", Status: "pass", Evidence: "certificate chain validates against the system trust store"}
	trustConclusion := ""
	if err := tlsChainVerify(leaf, result.PeerCertificates, chainTime); err != nil {
		trust.Status = "fail"
		trust.Evidence = boundedEvidence(err.Error(), 512)
		trustConclusion = "TCP/TLS is reachable but the certificate chain is not trusted by the system trust store"
	}
	checks = append(checks, trust)

	switch {
	case validityConclusion != "":
		return checks, validityConclusion, "high"
	case hostnameConclusion != "":
		return checks, hostnameConclusion, "high"
	case trustConclusion != "":
		return checks, trustConclusion, "high"
	default:
		return checks, "target is reachable and the TLS certificate validates", "high"
	}
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("TLS version 0x%04x", version)
	}
}
