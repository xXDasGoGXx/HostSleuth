package core

import (
	"bufio"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	certbotRenewalConfigLimit = 32
	certbotConfigBytesLimit   = 64 * 1024
	certbotPEMBytesLimit      = 2 * 1024 * 1024
	certbotSystemdOutputLimit = 4 * 1024
)

var (
	certbotRenewalDirectory = "/etc/letsencrypt/renewal"
	certbotExecutableLookup = findCertbotExecutable
	certbotSystemdLookup    = lookupCertbotSystemdUnit
)

func collectCertbotEvidence(ctx context.Context, host string, served *CertificateEvidence) *CertbotEvidence {
	evidence := &CertbotEvidence{Status: "not-installed"}
	if executable, ok := certbotExecutableLookup(); ok {
		evidence.Status = "installed"
		evidence.Executable = executable
	}

	paths, err := filepath.Glob(filepath.Join(certbotRenewalDirectory, "*.conf"))
	if err == nil {
		sort.Strings(paths)
		if len(paths) > certbotRenewalConfigLimit {
			paths = paths[:certbotRenewalConfigLimit]
		}
		for _, path := range paths {
			lineage, ok := readCertbotLineage(path, host, served)
			if !ok {
				continue
			}
			evidence.Lineages = append(evidence.Lineages, lineage)
		}
	}

	if len(evidence.Lineages) > 0 && evidence.Status == "not-installed" {
		evidence.Status = "evidence-found"
	}
	if evidence.Status == "not-installed" {
		return evidence
	}

	evidence.TimerState = certbotSystemdLookup(ctx, "certbot.timer")
	evidence.ServiceState = certbotSystemdLookup(ctx, "certbot.service")
	compareCertbotWithServed(evidence, served)
	return evidence
}

func readCertbotLineage(path, host string, served *CertificateEvidence) (CertbotLineageEvidence, bool) {
	content, err := readBoundedFile(path, certbotConfigBytesLimit)
	if err != nil {
		return CertbotLineageEvidence{}, false
	}
	values := parseSimpleINI(content)
	lineage := CertbotLineageEvidence{
		Name:            strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		RenewalConfig:   path,
		CertificatePath: strings.TrimSpace(values["cert"]),
		FullchainPath:   strings.TrimSpace(values["fullchain"]),
		RenewalState:    certbotRenewalSummary(values),
	}

	certificatePath := lineage.CertificatePath
	if certificatePath == "" {
		certificatePath = lineage.FullchainPath
	}
	if certificatePath != "" {
		if cert, err := readPEMCertificate(certificatePath); err == nil {
			lineage.Certificate = certificateEvidence(cert, time.Now().UTC())
			lineage.HostnameMatch = cert.VerifyHostname(host) == nil
			if served != nil && lineage.HostnameMatch {
				if sameCertificateFingerprint(lineage.Certificate, served) {
					lineage.ServedComparison = "matches served certificate"
				} else {
					lineage.ServedComparison = "differs from served certificate"
				}
			}
		}
	}
	return lineage, true
}

func compareCertbotWithServed(evidence *CertbotEvidence, served *CertificateEvidence) {
	if evidence == nil || served == nil {
		return
	}
	matching := make([]CertbotLineageEvidence, 0)
	for _, lineage := range evidence.Lineages {
		if lineage.HostnameMatch && lineage.Certificate != nil {
			matching = append(matching, lineage)
		}
	}
	if len(matching) == 0 {
		evidence.Comparison = "no readable Certbot lineage certificate matches the requested host"
		return
	}
	for _, lineage := range matching {
		if sameCertificateFingerprint(lineage.Certificate, served) {
			evidence.Comparison = "served certificate matches Certbot lineage " + lineage.Name
			return
		}
	}
	if len(matching) == 1 && certificateIsNewer(matching[0].Certificate, served) {
		evidence.StaleServedCert = true
		evidence.Comparison = "the matching Certbot certificate on disk is newer/different than the certificate this endpoint is serving"
		return
	}
	evidence.Comparison = "a matching Certbot certificate on disk differs from the served certificate, but the available evidence does not prove the served certificate is stale"
}

func certbotCheck(evidence *CertbotEvidence) Check {
	if evidence == nil {
		return Check{Name: "certbot", Status: "unknown", Evidence: "Certbot evidence was not collected"}
	}
	if evidence.Status == "not-installed" {
		return Check{Name: "certbot", Status: "unknown", Evidence: "Certbot was not detected on this native host"}
	}

	parts := []string{"Certbot " + evidence.Status}
	if evidence.Executable != "" {
		parts = append(parts, "executable="+evidence.Executable)
	}
	parts = append(parts, fmt.Sprintf("%d renewal lineage(s) discovered", len(evidence.Lineages)))
	if summary := certbotLineagesSummary(evidence.Lineages); summary != "" {
		parts = append(parts, "lineages: "+summary)
	}
	if evidence.TimerState != "" {
		parts = append(parts, "timer: "+evidence.TimerState)
	}
	if evidence.ServiceState != "" {
		parts = append(parts, "service: "+evidence.ServiceState)
	}
	if evidence.Comparison != "" {
		parts = append(parts, evidence.Comparison)
	}
	status := "pass"
	if evidence.StaleServedCert {
		status = "fail"
	} else if evidence.Comparison == "" {
		status = "unknown"
	}
	return Check{Name: "certbot", Status: status, Evidence: boundedEvidence(strings.Join(parts, "; "), 2048)}
}

func certbotLineagesSummary(lineages []CertbotLineageEvidence) string {
	items := make([]string, 0, len(lineages))
	for _, lineage := range lineages {
		fields := []string{lineage.Name}
		if lineage.RenewalConfig != "" {
			fields = append(fields, "config="+lineage.RenewalConfig)
		}
		if lineage.RenewalState != "" {
			fields = append(fields, lineage.RenewalState)
		}
		if lineage.Certificate != nil {
			fields = append(fields, "valid_until="+lineage.Certificate.ValidUntil.Format(time.RFC3339))
		}
		if lineage.HostnameMatch {
			fields = append(fields, "hostname-match")
		}
		if lineage.ServedComparison != "" {
			fields = append(fields, lineage.ServedComparison)
		}
		items = append(items, strings.Join(fields, ", "))
	}
	return boundedEvidence(strings.Join(items, " | "), 1024)
}

func certificateComparisonCheck(evidence *CertbotEvidence) Check {
	if evidence == nil || evidence.Comparison == "" {
		return Check{Name: "certificate-comparison", Status: "unknown", Evidence: "no matching local Certbot certificate was available for served-certificate comparison"}
	}
	status := "unknown"
	if evidence.StaleServedCert {
		status = "fail"
	} else if strings.Contains(evidence.Comparison, "served certificate matches") {
		status = "pass"
	}
	return Check{Name: "certificate-comparison", Status: status, Evidence: boundedEvidence(evidence.Comparison, 1024)}
}

func findCertbotExecutable() (string, bool) {
	if path, err := exec.LookPath("certbot"); err == nil {
		return path, true
	}
	for _, path := range []string{"/usr/bin/certbot", "/usr/local/bin/certbot", "/snap/bin/certbot"} {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return path, true
		}
	}
	return "", false
}

func lookupCertbotSystemdUnit(ctx context.Context, unit string) string {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("systemctl", "/usr/bin/systemctl", "/bin/systemctl")
	result, err := runBoundedCommand(
		lookupCtx,
		certbotSystemdOutputLimit,
		command,
		"show",
		unit,
		"--no-pager",
		"--property=LoadState,ActiveState,SubState,UnitFileState,NextElapseUSecRealtime",
	)
	output := boundedEvidence(result.Output, 1024)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "systemctl unavailable"
		}
		if output != "" {
			return "unavailable: " + output
		}
		return "unavailable: " + boundedEvidence(err.Error(), 384)
	}
	if output == "" {
		return "no unit state returned"
	}
	return output
}

func parseSimpleINI(content string) map[string]string {
	values := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "[") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			values[key] = value
		}
	}
	return values
}

func certbotRenewalSummary(values map[string]string) string {
	keys := []string{"renew_before_expiry", "authenticator", "installer", "server", "preferred_chain"}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if value := strings.TrimSpace(values[key]); value != "" {
			parts = append(parts, key+"="+value)
		}
	}
	return boundedEvidence(strings.Join(parts, "; "), 512)
}

func readBoundedFile(path string, maxBytes int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader := io.LimitReader(file, maxBytes+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxBytes {
		return "", fmt.Errorf("file exceeds %d byte evidence limit", maxBytes)
	}
	return string(data), nil
}

func readPEMCertificate(path string) (*x509.Certificate, error) {
	content, err := readBoundedFile(path, certbotPEMBytesLimit)
	if err != nil {
		return nil, err
	}
	rest := []byte(content)
	for len(rest) > 0 {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		rest = remaining
		if block.Type != "CERTIFICATE" {
			continue
		}
		return x509.ParseCertificate(block.Bytes)
	}
	return nil, errors.New("no certificate PEM block found")
}

func sameCertificateFingerprint(a, b *CertificateEvidence) bool {
	return a != nil && b != nil && a.SHA256Fingerprint != "" && strings.EqualFold(a.SHA256Fingerprint, b.SHA256Fingerprint)
}

func certificateIsNewer(local, served *CertificateEvidence) bool {
	if local == nil || served == nil || sameCertificateFingerprint(local, served) {
		return false
	}
	const tolerance = time.Minute
	return local.ValidFrom.After(served.ValidFrom.Add(tolerance)) || local.ValidUntil.After(served.ValidUntil.Add(tolerance))
}
