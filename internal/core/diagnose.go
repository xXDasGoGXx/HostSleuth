package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var routeLookup = func(ctx context.Context, destination string) (string, error) {
	cmd := exec.CommandContext(ctx, "ip", "route", "get", destination)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

var tcpConnect = func(ctx context.Context, address string) error {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func Diagnose(ctx context.Context, target string, snap Snapshot) Diagnosis {
	d := Diagnosis{Target: target, StartedAt: time.Now().UTC(), Confidence: "medium"}
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		d.Checks = append(d.Checks, Check{Name: "target", Status: "fail", Evidence: "target must be host:port"})
		d.Conclusion = "invalid target"
		d.Confidence = "high"
		return d
	}
	if _, err := strconv.Atoi(port); err != nil {
		d.Checks = append(d.Checks, Check{Name: "port", Status: "fail", Evidence: "port must be numeric"})
		d.Conclusion = "invalid port"
		d.Confidence = "high"
		return d
	}

	ips, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		d.Checks = append(d.Checks, Check{Name: "dns", Status: "fail", Evidence: err.Error()})
		d.Conclusion = "name resolution failed"
		d.Confidence = "high"
		return d
	}
	d.Checks = append(d.Checks, Check{Name: "dns", Status: "pass", Evidence: strings.Join(ips, ", ")})
	local := resolvedTargetIsLocal(ips)

	route := Check{Name: "route", Status: "unknown", Evidence: "no resolved IP was available for route lookup"}
	if destinationIP := firstResolvedIP(ips); destinationIP != "" {
		route = routeCheck(ctx, destinationIP)
	}
	d.Checks = append(d.Checks, route)

	if err := tcpConnect(ctx, net.JoinHostPort(host, port)); err == nil {
		d.Checks = append(d.Checks, Check{Name: "tcp", Status: "pass", Evidence: fmt.Sprintf("TCP/%s accepted a connection", port)})
		if local {
			d.Checks = append(d.Checks, reachableLocalEvidenceChecks(snap, port, ips)...)
		}

		d.TLS = tlsProbeLookup(ctx, net.JoinHostPort(host, port), host)
		if d.TLS != nil && d.TLS.ProbeConnected {
			d.Checks = append(d.Checks, tlsCheck(d.TLS, port))
		}
		if d.TLS != nil && d.TLS.HandshakeStatus == "pass" {
			d.Checks = append(d.Checks, tlsCertificateCheck(d.TLS))
			d.Checks = append(d.Checks, tlsHostnameCheck(d.TLS))
			d.Checks = append(d.Checks, tlsTrustCheck(d.TLS))
		}

		if local && snap.Mode != dockerDeploymentMode && d.TLS != nil && d.TLS.HandshakeStatus == "pass" {
			var served *CertificateEvidence
			if d.TLS != nil {
				served = d.TLS.Certificate
			}
			d.Certbot = collectCertbotEvidence(ctx, host, served)
			d.Checks = append(d.Checks, certbotCheck(d.Certbot))
			if served != nil {
				d.Checks = append(d.Checks, certificateComparisonCheck(d.Certbot))
			}
		}

		d.Conclusion, d.Confidence = reachableOutcome(d, port)
		return d
	} else {
		d.Checks = append(d.Checks, Check{Name: "tcp", Status: "fail", Evidence: err.Error()})
	}

	// Evidence precedence is deliberate:
	// 1. successful TCP is definitive for transport reachability;
	// 2. for local failures, exact listener/Docker binding evidence outranks
	//    firewall and failed-unit candidates;
	// 3. for remote failures, a kernel no-route result outranks firewall evidence;
	// 4. unavailable optional evidence remains neutral and cannot lower a
	//    conclusion supported by stronger evidence.
	if local {
		docker := assessDockerPort(snap, port, ips)
		if listener, ok := LocalListenerForTarget(snap, port, ips); ok {
			d.Checks = append(d.Checks, Check{Name: "local-listener", Status: "pass", Evidence: listener.Protocol + " " + listener.Address + " " + listener.Process})
			d.Checks = append(d.Checks, docker.Check)
			appendLocalCertbotWithoutServed(ctx, &d, snap, host, port)
			d.Checks = append(d.Checks, firewallCheck(ctx, port))
			d.Conclusion = "snapshot shows a listener on the requested local address but the current TCP connection failed; inspect firewall, network namespace, or snapshot freshness"
			d.Confidence = "medium"
			return d
		}

		d.Checks = append(d.Checks, Check{Name: "local-listener", Status: "fail", Evidence: "no local listener found for TCP/" + port + " on the requested address"})
		d.Checks = append(d.Checks, docker.Check)
		appendLocalCertbotWithoutServed(ctx, &d, snap, host, port)
		d.Checks = append(d.Checks, firewallCheck(ctx, port))
		d.Checks = append(d.Checks, systemdFailureCheck(ctx, snap))
		d.Conclusion, d.Confidence = localNoListenerOutcome(port, docker.Relation)
		return d
	}

	if route.Status == "fail" {
		d.Conclusion = "kernel route lookup reports the destination unreachable; inspect routing before the remote service"
		d.Confidence = "high"
		return d
	}

	d.Checks = append(d.Checks, firewallCheck(ctx, port))
	d.Conclusion = "remote TCP connection failed; inspect routing, firewall policy, and the destination service"
	return d
}

func reachableOutcome(d Diagnosis, port string) (string, string) {
	if d.Certbot != nil && d.Certbot.StaleServedCert {
		return "the local certificate on disk is newer/different than the certificate this endpoint is serving", "high"
	}
	if d.TLS == nil || d.TLS.HandshakeStatus != "pass" {
		if d.TLS != nil && d.TLS.ProbeConnected && d.TLS.HandshakeStatus == "fail" && likelyTLSPort(port) {
			return "TCP connection succeeds, but TLS handshake failed", "high"
		}
		return "target is reachable", "high"
	}
	if problem := certificateValidityProblem(d.TLS.Certificate, time.Now().UTC()); problem != "" {
		return "TCP and TLS are reachable, but " + problem, "high"
	}
	if d.TLS.HostnameStatus == "fail" {
		return "TCP and TLS are reachable, but the served certificate does not match the requested host", "high"
	}
	if d.TLS.TrustStatus == "fail" {
		return "TCP and TLS are reachable, but the served certificate chain is not trusted by this host", "high"
	}
	return "target is reachable", "high"
}

func appendLocalCertbotWithoutServed(ctx context.Context, d *Diagnosis, snap Snapshot, host, port string) {
	if d == nil || snap.Mode == dockerDeploymentMode || !likelyTLSPort(port) {
		return
	}
	d.Certbot = collectCertbotEvidence(ctx, host, nil)
	d.Checks = append(d.Checks, certbotCheck(d.Certbot))
}

func reachableLocalEvidenceChecks(snap Snapshot, port string, resolvedIPs []string) []Check {
	checks := make([]Check, 0, 2)
	if listener, ok := LocalListenerForTarget(snap, port, resolvedIPs); ok {
		checks = append(checks, Check{Name: "local-listener", Status: "pass", Evidence: listener.Protocol + " " + listener.Address + " " + listener.Process})
	}

	if len(snap.Containers) == 0 {
		return checks
	}
	bindings := dockerPortBindings(snap.Containers)
	direct := make([]dockerPortBinding, 0)
	other := make([]dockerPortBinding, 0)
	internalOnly := make([]dockerPortBinding, 0)
	for _, binding := range bindings {
		if binding.Protocol != "tcp" {
			continue
		}
		if binding.Published && binding.HostPort == port {
			if dockerBindingMatchesTarget(binding.HostIP, resolvedIPs) {
				direct = append(direct, binding)
			} else {
				other = append(other, binding)
			}
		}
		if !binding.Published && binding.ContainerPort == port {
			internalOnly = append(internalOnly, binding)
		}
	}
	switch {
	case len(direct) > 0:
		checks = append(checks, Check{Name: "docker-port", Status: "pass", Evidence: boundedEvidence("Docker publishes the requested local endpoint: "+formatDockerBindings(direct), 1024)})
	case len(other) > 0:
		checks = append(checks, Check{Name: "docker-port", Status: "unknown", Evidence: boundedEvidence("Docker publishes TCP/"+port+" only on different host address(es): "+formatDockerBindings(other), 1024)})
	case len(internalOnly) > 0:
		checks = append(checks, Check{Name: "docker-port", Status: "unknown", Evidence: boundedEvidence("Docker container port TCP/"+port+" is exposed internally but not published on the host: "+formatDockerBindings(internalOnly), 1024)})
	}
	return checks
}

func resolvedTargetIsLocal(ips []string) bool {
	for _, ip := range ips {
		parsed := net.ParseIP(ip)
		if parsed != nil && (parsed.IsLoopback() || isLocalIP(parsed)) {
			return true
		}
	}
	return false
}

func localNoListenerOutcome(port string, relation dockerPortRelation) (string, string) {
	switch relation {
	case dockerPortPublishedOtherBind:
		return "no process is listening on the requested local address; Docker publishes TCP/" + port + " only on a different host address", "high"
	case dockerPortInternalOnly:
		return "no process is listening on the requested local address; Docker exposes TCP/" + port + " inside a container but does not publish it on the host", "high"
	case dockerPortPublishedTarget:
		return "snapshot says Docker publishes TCP/" + port + " on the requested local address, but current TCP/listener evidence disagrees; inspect Docker proxy/network state or snapshot freshness", "medium"
	default:
		return "no process appears to be listening on the requested local address and port", "high"
	}
}

func firstResolvedIP(ips []string) string {
	for _, ip := range ips {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

func routeCheck(ctx context.Context, destination string) Check {
	out, err := routeLookup(ctx, destination)
	evidence := boundedEvidence(out, 512)
	if err == nil {
		if evidence == "" {
			return Check{Name: "route", Status: "unknown", Evidence: "kernel route lookup returned no evidence"}
		}
		return Check{Name: "route", Status: "pass", Evidence: evidence}
	}

	if evidence != "" {
		lower := strings.ToLower(evidence)
		if strings.Contains(lower, "network is unreachable") || strings.Contains(lower, "no route") || strings.Contains(lower, "unreachable") {
			return Check{Name: "route", Status: "fail", Evidence: evidence}
		}
		return Check{Name: "route", Status: "unknown", Evidence: evidence}
	}
	if errors.Is(err, exec.ErrNotFound) {
		return Check{Name: "route", Status: "unknown", Evidence: "ip command is unavailable"}
	}
	return Check{Name: "route", Status: "unknown", Evidence: "route lookup unavailable: " + boundedEvidence(err.Error(), 384)}
}

func boundedEvidence(value string, max int) string {
	value = strings.TrimSpace(value)
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func isLocalIP(target net.IP) bool {
	ifs, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifs {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.Equal(target) {
				return true
			}
		}
	}
	return false
}
