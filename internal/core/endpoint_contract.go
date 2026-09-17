package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	EndpointTLSIgnore    = "ignore"
	EndpointTLSPresent   = "present"
	EndpointTLSVerified  = "verified"
	EndpointTLSForbidden = "forbidden"
)

type EndpointContract struct {
	Name              string   `json:"name,omitempty"`
	Target            string   `json:"target"`
	ExpectedAddresses []string `json:"expected_addresses,omitempty"`
	TLS               string   `json:"tls,omitempty"`
	Service           string   `json:"service,omitempty"`
	Container         string   `json:"container,omitempty"`
}

type EndpointContractCheck struct {
	Name        string `json:"name"`
	Expectation string `json:"expectation"`
	Status      string `json:"status"`
	Observed    string `json:"observed"`
}

type EndpointContractEvaluation struct {
	Contract       EndpointContract        `json:"contract"`
	StartedAt      time.Time               `json:"started_at"`
	Status         string                  `json:"status"`
	Conclusion     string                  `json:"conclusion"`
	FirstMismatch string                  `json:"first_mismatch,omitempty"`
	Checks         []EndpointContractCheck `json:"checks"`
	Diagnosis      Diagnosis               `json:"diagnosis"`
}

func NormalizeEndpointContract(contract EndpointContract) (EndpointContract, error) {
	contract.Name = strings.TrimSpace(contract.Name)
	contract.Target = strings.TrimSpace(contract.Target)
	contract.Service = normalizeServiceUnit(contract.Service)
	contract.Container = strings.TrimPrefix(strings.TrimSpace(contract.Container), "/")
	contract.TLS = strings.ToLower(strings.TrimSpace(contract.TLS))
	if contract.TLS == "" {
		contract.TLS = EndpointTLSIgnore
	}
	switch contract.TLS {
	case EndpointTLSIgnore, EndpointTLSPresent, EndpointTLSVerified, EndpointTLSForbidden:
	default:
		return EndpointContract{}, fmt.Errorf("tls expectation must be one of: ignore, present, verified, forbidden")
	}

	host, port, err := net.SplitHostPort(contract.Target)
	if err != nil || strings.TrimSpace(host) == "" {
		return EndpointContract{}, errors.New("target must be host:port")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return EndpointContract{}, errors.New("target port must be between 1 and 65535")
	}

	addresses := make([]string, 0, len(contract.ExpectedAddresses))
	seen := map[string]struct{}{}
	for _, raw := range contract.ExpectedAddresses {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		ip := net.ParseIP(raw)
		if ip == nil {
			return EndpointContract{}, fmt.Errorf("expected address %q is not an IP address", raw)
		}
		normalized := ip.String()
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		addresses = append(addresses, normalized)
	}
	sort.Strings(addresses)
	contract.ExpectedAddresses = addresses
	return contract, nil
}

func EvaluateEndpointContract(ctx context.Context, contract EndpointContract, snap Snapshot) (EndpointContractEvaluation, error) {
	normalized, err := NormalizeEndpointContract(contract)
	if err != nil {
		return EndpointContractEvaluation{}, err
	}

	evaluation := EndpointContractEvaluation{
		Contract:  normalized,
		StartedAt: time.Now().UTC(),
	}
	evaluation.Diagnosis = Diagnose(ctx, normalized.Target, snap)

	dns := diagnosisCheck(evaluation.Diagnosis, "dns")
	evaluation.Checks = append(evaluation.Checks, endpointDNSContractCheck(normalized, dns))

	tcp := diagnosisCheck(evaluation.Diagnosis, "tcp")
	evaluation.Checks = append(evaluation.Checks, endpointTCPContractCheck(tcp))

	if normalized.TLS != EndpointTLSIgnore {
		evaluation.Checks = append(evaluation.Checks, endpointTLSContractCheck(normalized.TLS, evaluation.Diagnosis, tcp))
	}
	if normalized.Service != "" {
		evaluation.Checks = append(evaluation.Checks, endpointServiceContractCheck(normalized.Service, snap))
	}
	if normalized.Container != "" {
		evaluation.Checks = append(evaluation.Checks, endpointContainerContractCheck(normalized.Container, snap))
	}

	evaluation.Status, evaluation.FirstMismatch, evaluation.Conclusion = endpointContractOutcome(evaluation.Checks)
	return evaluation, nil
}

func diagnosisCheck(d Diagnosis, name string) Check {
	for _, check := range d.Checks {
		if check.Name == name {
			return check
		}
	}
	return Check{Name: name, Status: "unknown", Evidence: "no " + name + " evidence was produced"}
}

func endpointDNSContractCheck(contract EndpointContract, check Check) EndpointContractCheck {
	if check.Status != "pass" {
		return EndpointContractCheck{
			Name:        "dns",
			Expectation: endpointDNSExpectation(contract.ExpectedAddresses),
			Status:      contractStatus(check.Status),
			Observed:    boundedEvidence(check.Evidence, 1024),
		}
	}

	observed := normalizedIPList(strings.Split(check.Evidence, ","))
	if len(contract.ExpectedAddresses) == 0 {
		if len(observed) == 0 {
			return EndpointContractCheck{
				Name:        "dns",
				Expectation: "target resolves to at least one IP address",
				Status:      "unknown",
				Observed:    "resolver reported success but no IP address could be parsed",
			}
		}
		return EndpointContractCheck{
			Name:        "dns",
			Expectation: "target resolves to at least one IP address",
			Status:      "pass",
			Observed:    strings.Join(observed, ", "),
		}
	}

	status := "pass"
	if !sameStrings(observed, contract.ExpectedAddresses) {
		status = "fail"
	}
	return EndpointContractCheck{
		Name:        "dns",
		Expectation: endpointDNSExpectation(contract.ExpectedAddresses),
		Status:      status,
		Observed:    valueOrFallback(strings.Join(observed, ", "), "no parseable IP answers"),
	}
}

func endpointDNSExpectation(addresses []string) string {
	if len(addresses) == 0 {
		return "target resolves to at least one IP address"
	}
	return "DNS answers exactly match: " + strings.Join(addresses, ", ")
}

func endpointTCPContractCheck(check Check) EndpointContractCheck {
	status := contractStatus(check.Status)
	observed := boundedEvidence(check.Evidence, 1024)
	if check.Status == "unknown" && strings.Contains(check.Evidence, "no tcp evidence") {
		observed = "TCP was not reached because an earlier diagnosis step stopped the path"
	}
	return EndpointContractCheck{
		Name:        "tcp",
		Expectation: "TCP connection succeeds",
		Status:      status,
		Observed:    observed,
	}
}

func endpointTLSContractCheck(mode string, diagnosis Diagnosis, tcp Check) EndpointContractCheck {
	expectation := map[string]string{
		EndpointTLSPresent:   "TLS handshake succeeds",
		EndpointTLSVerified:  "TLS handshake, hostname verification, and trust verification succeed",
		EndpointTLSForbidden: "TLS handshake does not succeed",
	}[mode]

	if tcp.Status != "pass" {
		return EndpointContractCheck{
			Name:        "tls",
			Expectation: expectation,
			Status:      "unknown",
			Observed:    "TCP did not succeed, so the TLS expectation could not be evaluated",
		}
	}
	if diagnosis.TLS == nil || !diagnosis.TLS.ProbeConnected {
		return EndpointContractCheck{
			Name:        "tls",
			Expectation: expectation,
			Status:      "unknown",
			Observed:    "no TLS probe evidence was produced",
		}
	}

	tls := diagnosis.TLS
	switch mode {
	case EndpointTLSPresent:
		if tls.HandshakeStatus == "pass" {
			return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "pass", Observed: tlsObservation(tls)}
		}
		return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "fail", Observed: tlsObservation(tls)}
	case EndpointTLSVerified:
		switch {
		case tls.HandshakeStatus != "pass":
			return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "fail", Observed: tlsObservation(tls)}
		case tls.HostnameStatus == "fail" || tls.TrustStatus == "fail":
			return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "fail", Observed: tlsObservation(tls)}
		case tls.HostnameStatus == "pass" && tls.TrustStatus == "pass":
			return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "pass", Observed: tlsObservation(tls)}
		default:
			return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "unknown", Observed: tlsObservation(tls)}
		}
	case EndpointTLSForbidden:
		if tls.HandshakeStatus == "pass" {
			return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "fail", Observed: tlsObservation(tls)}
		}
		if tls.HandshakeStatus == "fail" {
			return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "pass", Observed: tlsObservation(tls)}
		}
		return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "unknown", Observed: tlsObservation(tls)}
	default:
		return EndpointContractCheck{Name: "tls", Expectation: expectation, Status: "unknown", Observed: "unsupported TLS expectation"}
	}
}

func tlsObservation(tls *TLSEvidence) string {
	if tls == nil {
		return "no TLS evidence"
	}
	parts := []string{"handshake=" + valueOrFallback(tls.HandshakeStatus, "unknown")}
	if tls.HostnameStatus != "" {
		parts = append(parts, "hostname="+tls.HostnameStatus)
	}
	if tls.TrustStatus != "" {
		parts = append(parts, "trust="+tls.TrustStatus)
	}
	if tls.Protocol != "" {
		parts = append(parts, "protocol="+tls.Protocol)
	}
	if tls.HandshakeError != "" {
		parts = append(parts, "error="+boundedEvidence(tls.HandshakeError, 384))
	}
	return strings.Join(parts, "; ")
}

func endpointServiceContractCheck(unit string, snap Snapshot) EndpointContractCheck {
	expectation := unit + " is active"
	if snap.Mode == dockerDeploymentMode {
		return EndpointContractCheck{
			Name:        "service",
			Expectation: expectation,
			Status:      "unknown",
			Observed:    "native systemd service inventory is unavailable in Docker mode",
		}
	}
	if len(snap.Services) == 0 {
		return EndpointContractCheck{
			Name:        "service",
			Expectation: expectation,
			Status:      "unknown",
			Observed:    "snapshot contains no systemd service inventory",
		}
	}
	service, ok := snapshotService(snap, unit)
	if !ok {
		return EndpointContractCheck{
			Name:        "service",
			Expectation: expectation,
			Status:      "fail",
			Observed:    "service is not present in the current snapshot",
		}
	}
	status := "fail"
	if service.Active == "active" {
		status = "pass"
	}
	return EndpointContractCheck{
		Name:        "service",
		Expectation: expectation,
		Status:      status,
		Observed:    valueOrFallback(strings.TrimSpace(service.Active+"/"+service.Sub), "service state unavailable"),
	}
}

func endpointContainerContractCheck(name string, snap Snapshot) EndpointContractCheck {
	expectation := "container " + name + " is running"
	if len(snap.Containers) == 0 {
		return EndpointContractCheck{
			Name:        "container",
			Expectation: expectation,
			Status:      "unknown",
			Observed:    "snapshot contains no Docker inventory; Docker may be absent or inaccessible",
		}
	}
	for _, container := range snap.Containers {
		if strings.TrimPrefix(container.Name, "/") != name {
			continue
		}
		state := stableContainerState(container.Status)
		status := "fail"
		if state == "running" || strings.HasPrefix(state, "running ") {
			status = "pass"
		}
		return EndpointContractCheck{
			Name:        "container",
			Expectation: expectation,
			Status:      status,
			Observed:    valueOrFallback(state, "container state unavailable"),
		}
	}
	return EndpointContractCheck{
		Name:        "container",
		Expectation: expectation,
		Status:      "fail",
		Observed:    "container is not present in the current snapshot",
	}
}

func endpointContractOutcome(checks []EndpointContractCheck) (status, first, conclusion string) {
	failed := 0
	unknown := 0
	for _, check := range checks {
		switch check.Status {
		case "fail":
			failed++
			if first == "" {
				first = check.Name
			}
		case "unknown":
			unknown++
			if first == "" {
				first = check.Name
			}
		}
	}
	switch {
	case failed > 0:
		return "fail", first, fmt.Sprintf("%d expectation(s) failed; first mismatch: %s", failed, first)
	case unknown > 0:
		return "unknown", first, fmt.Sprintf("no expectation is proven false, but %d expectation(s) remain unknown; first unresolved check: %s", unknown, first)
	default:
		return "pass", "", "all endpoint expectations are satisfied"
	}
}

func contractStatus(value string) string {
	switch value {
	case "pass", "fail", "unknown":
		return value
	default:
		return "unknown"
	}
}

func normalizedIPList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		ip := net.ParseIP(strings.TrimSpace(value))
		if ip == nil {
			continue
		}
		normalized := ip.String()
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func valueOrFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
