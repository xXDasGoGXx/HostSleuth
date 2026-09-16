package core

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"
)

const (
	tlsProbeTimeout      = 4 * time.Second
	tlsChainSubjectLimit = 8
	tlsSANLimit          = 64
)

var tlsProbeLookup = probeTLS

func probeTLS(ctx context.Context, address, host string) *TLSEvidence {
	probeCtx, cancel := context.WithTimeout(ctx, tlsProbeTimeout)
	defer cancel()

	dialer := net.Dialer{Timeout: tlsProbeTimeout}
	raw, err := dialer.DialContext(probeCtx, "tcp", address)
	if err != nil {
		return &TLSEvidence{
			HandshakeStatus: "unknown",
			HandshakeError:  boundedEvidence(err.Error(), 512),
			ServerName:      host,
		}
	}
	defer raw.Close()

	// This probe intentionally separates protocol handshake evidence from
	// certificate validation. No application data is sent. Hostname and trust
	// are verified explicitly below so HostSleuth can explain which validation
	// step failed instead of collapsing everything into one TLS error.
	conn := tls.Client(raw, &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true, // diagnostic capture only; explicit verification follows
	})
	if err := conn.HandshakeContext(probeCtx); err != nil {
		return &TLSEvidence{
			ProbeConnected:  true,
			HandshakeStatus: "fail",
			HandshakeError:  boundedEvidence(err.Error(), 512),
			ServerName:      host,
		}
	}
	defer conn.Close()

	state := conn.ConnectionState()
	evidence := &TLSEvidence{
		ProbeConnected:  true,
		HandshakeStatus: "pass",
		ServerName:      host,
		Protocol:        tlsVersionName(state.Version),
		CipherSuite:     tls.CipherSuiteName(state.CipherSuite),
	}
	if len(state.PeerCertificates) == 0 {
		evidence.HandshakeStatus = "unknown"
		evidence.HandshakeError = "TLS handshake completed without a peer certificate"
		return evidence
	}

	now := time.Now().UTC()
	leaf := state.PeerCertificates[0]
	evidence.Certificate = certificateEvidence(leaf, now)
	evidence.ChainSubjects = boundedChainSubjects(state.PeerCertificates)

	if err := leaf.VerifyHostname(host); err != nil {
		evidence.HostnameStatus = "fail"
		evidence.HostnameEvidence = boundedEvidence(err.Error(), 512)
	} else {
		evidence.HostnameStatus = "pass"
		evidence.HostnameEvidence = "certificate matches " + host
	}

	intermediates := x509.NewCertPool()
	for _, cert := range state.PeerCertificates[1:] {
		intermediates.AddCert(cert)
	}
	chains, err := leaf.Verify(x509.VerifyOptions{
		Intermediates: intermediates,
		CurrentTime:   now,
	})
	if err != nil {
		evidence.TrustStatus = "fail"
		evidence.TrustEvidence = boundedEvidence(err.Error(), 512)
	} else {
		evidence.TrustStatus = "pass"
		chainLength := 0
		if len(chains) > 0 {
			chainLength = len(chains[0])
		}
		evidence.TrustEvidence = fmt.Sprintf("certificate chain verified against this host's trust store (%d certificate(s) in verified chain)", chainLength)
	}

	return evidence
}

func certificateEvidence(cert *x509.Certificate, now time.Time) *CertificateEvidence {
	if cert == nil {
		return nil
	}
	sum := sha256.Sum256(cert.Raw)
	return &CertificateEvidence{
		Subject:           cert.Subject.String(),
		Issuer:            cert.Issuer.String(),
		Serial:            strings.ToUpper(cert.SerialNumber.Text(16)),
		SANs:              certificateSANs(cert),
		ValidFrom:         cert.NotBefore.UTC(),
		ValidUntil:        cert.NotAfter.UTC(),
		DaysRemaining:     int(cert.NotAfter.Sub(now) / (24 * time.Hour)),
		SHA256Fingerprint: strings.ToUpper(hex.EncodeToString(sum[:])),
	}
}

func certificateSANs(cert *x509.Certificate) []string {
	if cert == nil {
		return nil
	}
	values := make([]string, 0, len(cert.DNSNames)+len(cert.IPAddresses)+len(cert.EmailAddresses)+len(cert.URIs))
	values = append(values, cert.DNSNames...)
	for _, ip := range cert.IPAddresses {
		values = append(values, ip.String())
	}
	values = append(values, cert.EmailAddresses...)
	for _, uri := range cert.URIs {
		values = append(values, uri.String())
	}
	sort.Strings(values)
	values = uniqueStrings(values)
	if len(values) > tlsSANLimit {
		values = values[:tlsSANLimit]
	}
	return values
}

func boundedChainSubjects(certs []*x509.Certificate) []string {
	limit := len(certs)
	if limit > tlsChainSubjectLimit {
		limit = tlsChainSubjectLimit
	}
	out := make([]string, 0, limit)
	for _, cert := range certs[:limit] {
		out = append(out, cert.Subject.String())
	}
	return out
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	out := values[:0]
	var previous string
	for i, value := range values {
		if i == 0 || value != previous {
			out = append(out, value)
			previous = value
		}
	}
	return out
}

func tlsCheck(evidence *TLSEvidence, port string) Check {
	if evidence == nil {
		return Check{Name: "tls-handshake", Status: "unknown", Evidence: "TLS evidence was not available"}
	}
	if evidence.HandshakeStatus != "pass" {
		status := "unknown"
		prefix := "TLS handshake did not complete; this endpoint may not use TLS"
		if likelyTLSPort(port) {
			status = "fail"
			prefix = "TLS handshake failed on a commonly TLS-enabled port"
		}
		detail := evidence.HandshakeError
		if detail == "" {
			detail = "no additional handshake error was returned"
		}
		return Check{Name: "tls-handshake", Status: status, Evidence: prefix + ": " + detail}
	}
	return Check{
		Name:     "tls-handshake",
		Status:   "pass",
		Evidence: boundedEvidence(strings.TrimSpace(evidence.Protocol+" "+evidence.CipherSuite), 512),
	}
}

func tlsHostnameCheck(evidence *TLSEvidence) Check {
	if evidence == nil || evidence.Certificate == nil {
		return Check{Name: "tls-hostname", Status: "unknown", Evidence: "no peer certificate was available for hostname validation"}
	}
	status := evidence.HostnameStatus
	if status != "pass" && status != "fail" {
		status = "unknown"
	}
	return Check{Name: "tls-hostname", Status: status, Evidence: boundedEvidence(evidence.HostnameEvidence, 512)}
}

func tlsTrustCheck(evidence *TLSEvidence) Check {
	if evidence == nil || evidence.Certificate == nil {
		return Check{Name: "tls-trust", Status: "unknown", Evidence: "no peer certificate was available for trust validation"}
	}
	status := evidence.TrustStatus
	if status != "pass" && status != "fail" {
		status = "unknown"
	}
	return Check{Name: "tls-trust", Status: status, Evidence: boundedEvidence(evidence.TrustEvidence, 512)}
}

func likelyTLSPort(port string) bool {
	switch port {
	case "443", "465", "636", "853", "993", "995", "8443", "8883", "9443":
		return true
	default:
		return false
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

func certificateValidityProblem(cert *CertificateEvidence, now time.Time) string {
	if cert == nil {
		return ""
	}
	if now.Before(cert.ValidFrom) {
		return "the served certificate is not valid yet"
	}
	if now.After(cert.ValidUntil) {
		return "the served certificate has expired"
	}
	return ""
}

func tlsCertificateCheck(evidence *TLSEvidence) Check {
	if evidence == nil || evidence.Certificate == nil {
		return Check{Name: "tls-certificate", Status: "unknown", Evidence: "no served certificate was available"}
	}
	cert := evidence.Certificate
	status := "pass"
	validity := "certificate is currently within its validity window"
	if problem := certificateValidityProblem(cert, time.Now().UTC()); problem != "" {
		status = "fail"
		validity = problem
	}
	sans := strings.Join(cert.SANs, ", ")
	if sans == "" {
		sans = "none reported"
	}
	return Check{
		Name:   "tls-certificate",
		Status: status,
		Evidence: boundedEvidence(fmt.Sprintf(
			"subject=%s; issuer=%s; serial=%s; SANs=%s; valid=%s to %s; days_remaining=%d; sha256=%s; %s",
			cert.Subject,
			cert.Issuer,
			cert.Serial,
			sans,
			cert.ValidFrom.Format(time.RFC3339),
			cert.ValidUntil.Format(time.RFC3339),
			cert.DaysRemaining,
			cert.SHA256Fingerprint,
			validity,
		), 4096),
	}
}
