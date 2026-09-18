package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	certificateRolloutMaxEndpoints = 16
	certificateRolloutProbeWorkers = 4
)

type CertificateRolloutExpected struct {
	Source      string               `json:"source"`
	Value       string               `json:"value"`
	Fingerprint string               `json:"fingerprint,omitempty"`
	Status      string               `json:"status"`
	Problem     string               `json:"problem,omitempty"`
	TLS         *TLSEvidence         `json:"tls,omitempty"`
	Certificate *CertificateEvidence `json:"certificate,omitempty"`
}

type CertificateRolloutEndpoint struct {
	Target      string               `json:"target"`
	Host        string               `json:"host"`
	Port        string               `json:"port"`
	Status      string               `json:"status"`
	MatchStatus string               `json:"match_status"`
	Problem     string               `json:"problem,omitempty"`
	TLS         *TLSEvidence         `json:"tls,omitempty"`
	Certificate *CertificateEvidence `json:"certificate,omitempty"`
}

type CertificateRolloutSummary struct {
	Total      int `json:"total"`
	Matched    int `json:"matched"`
	Mismatched int `json:"mismatched"`
	Unknown    int `json:"unknown"`
	Healthy    int `json:"healthy"`
	Unhealthy  int `json:"unhealthy"`
}

type CertificateRolloutStory struct {
	StartedAt    time.Time                    `json:"started_at"`
	Status       string                       `json:"status"`
	Conclusion   string                       `json:"conclusion"`
	FirstProblem string                       `json:"first_problem,omitempty"`
	Expected     CertificateRolloutExpected   `json:"expected"`
	Endpoints    []CertificateRolloutEndpoint `json:"endpoints"`
	Summary      CertificateRolloutSummary    `json:"summary"`
	ScopeNotes   []string                     `json:"scope_notes,omitempty"`
}

var certificateRolloutTLSProbe = probeTLS

func BuildCertificateRolloutStory(ctx context.Context, expectedFingerprint, reference string, endpoints []string) (CertificateRolloutStory, error) {
	expectedFingerprint = strings.TrimSpace(expectedFingerprint)
	reference = strings.TrimSpace(reference)

	if (expectedFingerprint == "") == (reference == "") {
		return CertificateRolloutStory{}, errors.New("provide exactly one expected source: fingerprint or reference endpoint")
	}

	cleanEndpoints, err := validateCertificateRolloutEndpoints(endpoints)
	if err != nil {
		return CertificateRolloutStory{}, err
	}

	story := CertificateRolloutStory{
		StartedAt: time.Now().UTC(),
		Endpoints: make([]CertificateRolloutEndpoint, len(cleanEndpoints)),
		ScopeNotes: []string{
			"Only direct TLS endpoints explicitly supplied by the user are probed.",
			"Expected certificate source is either a SHA-256 fingerprint or one explicit reference TLS endpoint.",
			"Arbitrary certificate-file reads are intentionally not exposed because the no-private-key-read boundary cannot be guaranteed before opening an arbitrary file.",
			"HostSleuth does not renew, install, reload, replace, or otherwise modify certificates or TLS services.",
		},
	}

	if expectedFingerprint != "" {
		normalized, err := normalizeCertificateFingerprint(expectedFingerprint)
		if err != nil {
			return CertificateRolloutStory{}, err
		}
		story.Expected = CertificateRolloutExpected{
			Source:      "fingerprint",
			Value:       normalized,
			Fingerprint: normalized,
			Status:      "pass",
		}
	} else {
		host, _, err := validateRolloutTarget(reference)
		if err != nil {
			return CertificateRolloutStory{}, fmt.Errorf("reference endpoint: %w", err)
		}
		story.Expected = CertificateRolloutExpected{
			Source: "reference",
			Value:  reference,
		}
		evidence := certificateRolloutTLSProbe(ctx, reference, host)
		story.Expected.TLS = evidence
		if evidence != nil {
			story.Expected.Certificate = evidence.Certificate
		}
		if evidence == nil || evidence.HandshakeStatus != "pass" || evidence.Certificate == nil || evidence.Certificate.SHA256Fingerprint == "" {
			story.Expected.Status = "fail"
			story.Expected.Problem = "reference endpoint did not provide a usable served certificate"
			if evidence != nil && evidence.HandshakeError != "" {
				story.Expected.Problem += ": " + boundedEvidence(evidence.HandshakeError, 512)
			}
			story.Status = "unknown"
			story.FirstProblem = "expected-reference"
			story.Conclusion = "certificate rollout could not be compared because the reference endpoint did not provide a usable certificate"
			return story, nil
		}
		story.Expected.Fingerprint = evidence.Certificate.SHA256Fingerprint
		story.Expected.Status = "pass"
	}

	story.Endpoints = probeCertificateRolloutEndpoints(ctx, cleanEndpoints, story.Expected.Fingerprint)
	story.Summary = summarizeCertificateRollout(story.Endpoints)
	finalizeCertificateRolloutStory(&story)
	return story, nil
}

func validateCertificateRolloutEndpoints(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, errors.New("at least one endpoint is required")
	}
	if len(values) > certificateRolloutMaxEndpoints {
		return nil, fmt.Errorf("at most %d endpoints are allowed", certificateRolloutMaxEndpoints)
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, errors.New("endpoint must not be empty")
		}
		if _, _, err := validateRolloutTarget(value); err != nil {
			return nil, fmt.Errorf("endpoint %q: %w", value, err)
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate endpoint %q", value)
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out, nil
}

func validateRolloutTarget(target string) (string, string, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(target))
	if err != nil {
		return "", "", errors.New("target must be in host:port form")
	}
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" {
		return "", "", errors.New("target host is required")
	}
	value, err := strconv.ParseUint(port, 10, 16)
	if err != nil || value == 0 {
		return "", "", errors.New("target port must be numeric and between 1 and 65535")
	}
	return host, port, nil
}

func normalizeCertificateFingerprint(value string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r == ':' || unicode.IsSpace(r):
			continue
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'a' && r <= 'f':
			b.WriteRune(r - ('a' - 'A'))
		case r >= 'A' && r <= 'F':
			b.WriteRune(r)
		default:
			return "", errors.New("expected fingerprint must contain only hexadecimal digits, optional colons, and whitespace")
		}
	}
	normalized := b.String()
	if len(normalized) != 64 {
		return "", errors.New("expected fingerprint must be exactly 32 bytes / 64 hexadecimal characters")
	}
	return normalized, nil
}

func probeCertificateRolloutEndpoints(ctx context.Context, targets []string, expectedFingerprint string) []CertificateRolloutEndpoint {
	results := make([]CertificateRolloutEndpoint, len(targets))
	type job struct {
		index  int
		target string
	}
	jobs := make(chan job)
	var wg sync.WaitGroup

	workers := certificateRolloutProbeWorkers
	if workers > len(targets) {
		workers = len(targets)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				results[item.index] = probeCertificateRolloutEndpoint(ctx, item.target, expectedFingerprint)
			}
		}()
	}

	for i, target := range targets {
		select {
		case <-ctx.Done():
			results[i] = CertificateRolloutEndpoint{
				Target:      target,
				Status:      "unknown",
				MatchStatus: "unknown",
				Problem:     "rollout probe context ended before this endpoint was tested",
			}
		case jobs <- job{index: i, target: target}:
		}
	}
	close(jobs)
	wg.Wait()
	return results
}

func probeCertificateRolloutEndpoint(ctx context.Context, target, expectedFingerprint string) CertificateRolloutEndpoint {
	host, port, _ := validateRolloutTarget(target)
	row := CertificateRolloutEndpoint{
		Target:      target,
		Host:        host,
		Port:        port,
		Status:      "unknown",
		MatchStatus: "unknown",
	}

	evidence := certificateRolloutTLSProbe(ctx, target, host)
	row.TLS = evidence
	if evidence != nil {
		row.Certificate = evidence.Certificate
	}
	if evidence == nil || evidence.HandshakeStatus != "pass" {
		row.Status = "fail"
		row.Problem = "TLS handshake did not complete"
		if evidence != nil && evidence.HandshakeError != "" {
			row.Problem += ": " + boundedEvidence(evidence.HandshakeError, 512)
		}
		return row
	}
	if evidence.Certificate == nil || evidence.Certificate.SHA256Fingerprint == "" {
		row.Status = "unknown"
		row.Problem = "TLS completed without usable served-certificate fingerprint evidence"
		return row
	}

	served := evidence.Certificate.SHA256Fingerprint
	if strings.EqualFold(served, expectedFingerprint) {
		row.MatchStatus = "match"
	} else {
		row.MatchStatus = "mismatch"
		row.Status = "fail"
		row.Problem = fmt.Sprintf("served fingerprint %s does not match expected %s", served, expectedFingerprint)
		return row
	}

	checks := []Check{
		tlsCertificateCheck(evidence),
		tlsHostnameCheck(evidence),
		tlsTrustCheck(evidence),
	}
	for _, check := range checks {
		if check.Status == "fail" {
			row.Status = "fail"
			row.Problem = check.Evidence
			return row
		}
	}
	for _, check := range checks {
		if check.Status == "unknown" {
			row.Status = "unknown"
			row.Problem = check.Evidence
			return row
		}
	}

	row.Status = "pass"
	return row
}

func summarizeCertificateRollout(rows []CertificateRolloutEndpoint) CertificateRolloutSummary {
	summary := CertificateRolloutSummary{Total: len(rows)}
	for _, row := range rows {
		switch row.MatchStatus {
		case "match":
			summary.Matched++
		case "mismatch":
			summary.Mismatched++
		default:
			summary.Unknown++
		}
		switch row.Status {
		case "pass":
			summary.Healthy++
		case "fail":
			summary.Unhealthy++
		}
	}
	return summary
}

func finalizeCertificateRolloutStory(story *CertificateRolloutStory) {
	if story == nil {
		return
	}
	for _, row := range story.Endpoints {
		if row.MatchStatus == "mismatch" {
			story.Status = "fail"
			story.FirstProblem = row.Target
			story.Conclusion = fmt.Sprintf("%d of %d endpoint(s) serve a certificate different from the expected fingerprint", story.Summary.Mismatched, story.Summary.Total)
			return
		}
	}

	for _, row := range story.Endpoints {
		if row.MatchStatus == "unknown" {
			story.Status = "unknown"
			story.FirstProblem = row.Target
			story.Conclusion = "certificate rollout verification is incomplete because one or more endpoint fingerprints could not be determined"
			return
		}
	}

	for _, row := range story.Endpoints {
		if row.Status == "fail" {
			story.Status = "fail"
			story.FirstProblem = row.Target
			story.Conclusion = "all served fingerprints match, but one or more endpoints have certificate validity, hostname, or trust failures"
			return
		}
	}

	for _, row := range story.Endpoints {
		if row.Status == "unknown" {
			story.Status = "unknown"
			story.FirstProblem = row.Target
			story.Conclusion = "all comparable fingerprints match, but one or more endpoint certificate checks are incomplete"
			return
		}
	}

	story.Status = "pass"
	story.Conclusion = fmt.Sprintf("all %d endpoint(s) serve the expected certificate and passed validity, hostname, and trust checks", story.Summary.Total)
}
