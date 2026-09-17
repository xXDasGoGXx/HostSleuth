package core

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func withEndpointContractStubs(t *testing.T, tcpErr error, tls *TLSEvidence) {
	t.Helper()
	oldTCP := tcpConnect
	oldTLS := tlsProbeLookup
	oldRoute := routeLookup
	tcpConnect = func(context.Context, string) error { return tcpErr }
	tlsProbeLookup = func(context.Context, string, string) *TLSEvidence { return tls }
	routeLookup = func(context.Context, string) (string, error) { return "local 127.0.0.1 dev lo src 127.0.0.1", nil }
	t.Cleanup(func() {
		tcpConnect = oldTCP
		tlsProbeLookup = oldTLS
		routeLookup = oldRoute
	})
}

func TestNormalizeEndpointContractRejectsUnsafeInputs(t *testing.T) {
	tests := []EndpointContract{
		{Target: "missing-port"},
		{Target: "example.com:0"},
		{Target: "example.com:443", TLS: "magic"},
		{Target: "example.com:443", ExpectedAddresses: []string{"not-an-ip"}},
	}
	for _, contract := range tests {
		if _, err := NormalizeEndpointContract(contract); err == nil {
			t.Fatalf("expected validation error for %#v", contract)
		}
	}
}

func TestEvaluateEndpointContractPassesWhenRequestedEvidenceMatches(t *testing.T) {
	withEndpointContractStubs(t, nil, &TLSEvidence{
		ProbeConnected:  true,
		HandshakeStatus: "pass",
		HostnameStatus:  "pass",
		TrustStatus:     "pass",
		Protocol:        "TLS 1.3",
	})
	snap := Snapshot{
		Mode: dockerDeploymentMode,
		Containers: []ContainerInfo{{
			Name:   "web",
			Status: "Up 2 hours (healthy)",
		}},
	}
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Name:              "web",
		Target:            "127.0.0.1:443",
		ExpectedAddresses: []string{"127.0.0.1"},
		TLS:               EndpointTLSVerified,
		Container:         "web",
	}, snap)
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "pass" || evaluation.FirstMismatch != "" {
		t.Fatalf("unexpected evaluation: %#v", evaluation)
	}
	if evaluation.Conclusion != "all endpoint expectations are satisfied" {
		t.Fatalf("unexpected conclusion: %s", evaluation.Conclusion)
	}
}

func TestEvaluateEndpointContractReportsFirstDNSMismatchBeforeReachability(t *testing.T) {
	withEndpointContractStubs(t, nil, nil)
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Target:            "127.0.0.1:8080",
		ExpectedAddresses: []string{"127.0.0.2"},
	}, Snapshot{Mode: dockerDeploymentMode})
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "fail" || evaluation.FirstMismatch != "dns" {
		t.Fatalf("expected DNS to be first mismatch: %#v", evaluation)
	}
	if !strings.Contains(evaluation.Checks[0].Expectation, "127.0.0.2") {
		t.Fatalf("unexpected DNS expectation: %#v", evaluation.Checks[0])
	}
}

func TestEvaluateEndpointContractTCPFailureLeavesTLSUnknown(t *testing.T) {
	withEndpointContractStubs(t, errors.New("connection refused"), nil)
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Target: "127.0.0.1:8443",
		TLS:    EndpointTLSPresent,
	}, Snapshot{Mode: dockerDeploymentMode})
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "fail" || evaluation.FirstMismatch != "tcp" {
		t.Fatalf("unexpected result: %#v", evaluation)
	}
	var tlsCheck EndpointContractCheck
	for _, check := range evaluation.Checks {
		if check.Name == "tls" {
			tlsCheck = check
		}
	}
	if tlsCheck.Status != "unknown" {
		t.Fatalf("expected TLS to remain unknown after TCP failure: %#v", tlsCheck)
	}
}

func TestEvaluateEndpointContractVerifiedTLSRejectsHostnameOrTrustFailure(t *testing.T) {
	withEndpointContractStubs(t, nil, &TLSEvidence{
		ProbeConnected:  true,
		HandshakeStatus: "pass",
		HostnameStatus:  "fail",
		TrustStatus:     "pass",
	})
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Target: "127.0.0.1:443",
		TLS:    EndpointTLSVerified,
	}, Snapshot{Mode: dockerDeploymentMode})
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "fail" || evaluation.FirstMismatch != "tls" {
		t.Fatalf("unexpected result: %#v", evaluation)
	}
}

func TestEvaluateEndpointContractForbiddenTLSPassesOnPlaintextEndpoint(t *testing.T) {
	withEndpointContractStubs(t, nil, &TLSEvidence{
		ProbeConnected:  true,
		HandshakeStatus: "fail",
		HandshakeError:  "first record does not look like a TLS handshake",
		HostnameStatus:  "unknown",
		TrustStatus:     "unknown",
	})
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Target: "127.0.0.1:8080",
		TLS:    EndpointTLSForbidden,
	}, Snapshot{Mode: dockerDeploymentMode})
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "pass" {
		t.Fatalf("unexpected result: %#v", evaluation)
	}
}

func TestEvaluateEndpointContractServiceUnknownInDockerMode(t *testing.T) {
	withEndpointContractStubs(t, nil, nil)
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Target:  "127.0.0.1:8080",
		Service: "nginx",
	}, Snapshot{Mode: dockerDeploymentMode})
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "unknown" || evaluation.FirstMismatch != "service" {
		t.Fatalf("unexpected result: %#v", evaluation)
	}
}

func TestEvaluateEndpointContractServiceAndContainerState(t *testing.T) {
	withEndpointContractStubs(t, nil, nil)
	snap := Snapshot{
		Services:   []ServiceInfo{{Name: "nginx.service", Active: "failed", Sub: "failed"}},
		Containers: []ContainerInfo{{Name: "web", Status: "Exited (1) 10 seconds ago"}},
	}
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Target:    "127.0.0.1:8080",
		Service:   "nginx",
		Container: "web",
	}, snap)
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "fail" || evaluation.FirstMismatch != "service" {
		t.Fatalf("unexpected result: %#v", evaluation)
	}
	if len(evaluation.Checks) != 4 {
		t.Fatalf("expected dns, tcp, service, container checks, got %d", len(evaluation.Checks))
	}
}


func TestEndpointContractFailurePrecedesEarlierUnknownForFirstMismatch(t *testing.T) {
	withEndpointContractStubs(t, nil, nil)
	evaluation, err := EvaluateEndpointContract(context.Background(), EndpointContract{
		Target:    "127.0.0.1:8080",
		TLS:       EndpointTLSPresent,
		Container: "missing",
	}, Snapshot{Mode: dockerDeploymentMode, Containers: []ContainerInfo{{Name: "other", Status: "Up 1 minute"}}})
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Status != "fail" || evaluation.FirstMismatch != "container" {
		t.Fatalf("expected first proven failure to win over earlier unknown: %#v", evaluation)
	}
}
