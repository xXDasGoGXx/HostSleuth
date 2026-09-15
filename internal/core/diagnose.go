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

	route := Check{Name: "route", Status: "unknown", Evidence: "no resolved IP was available for route lookup"}
	if destinationIP := firstResolvedIP(ips); destinationIP != "" {
		route = routeCheck(ctx, destinationIP)
	}
	d.Checks = append(d.Checks, route)

	if err := tcpConnect(ctx, net.JoinHostPort(host, port)); err == nil {
		d.Checks = append(d.Checks, Check{Name: "tcp", Status: "pass", Evidence: fmt.Sprintf("TCP/%s accepted a connection", port)})
		d.Conclusion = "target is reachable"
		d.Confidence = "high"
		return d
	} else {
		d.Checks = append(d.Checks, Check{Name: "tcp", Status: "fail", Evidence: err.Error()})
	}

	local := false
	for _, ip := range ips {
		parsed := net.ParseIP(ip)
		if parsed != nil && (parsed.IsLoopback() || isLocalIP(parsed)) {
			local = true
			break
		}
	}

	// Evidence precedence is deliberate:
	// 1. successful TCP is definitive and already returned above;
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
			d.Checks = append(d.Checks, firewallCheck(ctx, port))
			d.Conclusion = "snapshot shows a listener on the requested local address but the current TCP connection failed; inspect firewall, network namespace, or snapshot freshness"
			d.Confidence = "medium"
			return d
		}

		d.Checks = append(d.Checks, Check{Name: "local-listener", Status: "fail", Evidence: "no local listener found for TCP/" + port + " on the requested address"})
		d.Checks = append(d.Checks, docker.Check)
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
