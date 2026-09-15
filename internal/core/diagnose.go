package core

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

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

	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err == nil {
		_ = conn.Close()
		d.Checks = append(d.Checks, Check{Name: "tcp", Status: "pass", Evidence: fmt.Sprintf("TCP/%s accepted a connection", port)})
		d.Conclusion = "target is reachable"
		d.Confidence = "high"
		return d
	}
	d.Checks = append(d.Checks, Check{Name: "tcp", Status: "fail", Evidence: err.Error()})

	local := false
	for _, ip := range ips {
		parsed := net.ParseIP(ip)
		if parsed != nil && (parsed.IsLoopback() || isLocalIP(parsed)) { local = true; break }
	}
	if local {
		if l, ok := LocalListenerForPort(snap, port); ok {
			d.Checks = append(d.Checks, Check{Name: "local-listener", Status: "pass", Evidence: l.Protocol + " " + l.Address + " " + l.Process})
			d.Conclusion = "service is listening locally but the TCP connection failed; inspect bind address, firewall, or network namespace"
			d.Confidence = "medium"
		} else {
			d.Checks = append(d.Checks, Check{Name: "local-listener", Status: "fail", Evidence: "no local listener found for TCP/" + port})
			d.Conclusion = "no process appears to be listening on the requested local port"
			d.Confidence = "high"
		}
		return d
	}

	d.Conclusion = "remote TCP connection failed; inspect routing, firewall policy, and the destination service"
	return d
}

func isLocalIP(target net.IP) bool {
	ifs, err := net.Interfaces()
	if err != nil { return false }
	for _, iface := range ifs {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet: ip = v.IP
			case *net.IPAddr: ip = v.IP
			}
			if ip != nil && ip.Equal(target) { return true }
		}
	}
	return false
}
