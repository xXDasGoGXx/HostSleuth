package core

import (
	"net/netip"
	"strconv"
	"strings"
)

// Listener churn classification for the change timeline.
//
// A Snapshot keeps every listening socket as current evidence. The change
// timeline, however, should only remember listener changes an administrator
// would act on. Three families of sockets follow interfaces or client
// activity rather than services, and on a container host they can produce
// hundreds of "appeared/disappeared" pairs per day, pushing real history
// (reboots, container health, service state) out of the retained window:
//
//  1. Link-local and multicast sockets (fe80::/10, 169.254.0.0/16,
//     ff00::/8, 224.0.0.0/4). Discovery daemons bind one per interface,
//     including every Docker veth and bridge.
//  2. Discovery-protocol binds on one specific interface address
//     (mDNS 5353, LLMNR 5355, SSDP 1900, WS-Discovery 3702/5357).
//     They follow interfaces being created and removed. The wildcard
//     binding (for example 0.0.0.0:5353) is still recorded, because the
//     discovery daemon itself starting or stopping is meaningful.
//  3. UDP sockets in the Linux default ephemeral port range
//     (32768-60999). These are unconnected client sockets that `ss -l`
//     lists alongside real UDP services. A small set of well-known fixed
//     services inside that range is still recorded.
//
// TCP listeners on arbitrary high ports are never classified as churn:
// a TCP socket in LISTEN state is always a real server.

const (
	ephemeralUDPLow  = 32768
	ephemeralUDPHigh = 60999
)

// Fixed-port UDP services that live inside the ephemeral range and whose
// appearance or disappearance is meaningful.
var fixedHighUDPServices = map[int]bool{
	41641: true, // Tailscale
	51820: true, // WireGuard
}

var discoveryPorts = map[string]map[int]bool{
	"udp": {1900: true, 3702: true, 5353: true, 5355: true},
	"tcp": {5355: true, 5357: true},
}

// IsRoutineListener reports whether a listening socket is interface or
// client churn that should not be recorded in the change timeline.
func IsRoutineListener(protocol, address string) bool {
	proto := strings.ToLower(strings.TrimSpace(protocol))
	host, zone, port, ok := splitListenerAddress(address)
	if !ok {
		return false
	}
	wildcard := host == "*" || host == "0.0.0.0" || host == "::"
	var addr netip.Addr
	if !wildcard {
		parsed, err := netip.ParseAddr(host)
		if err != nil {
			return false
		}
		addr = parsed.Unmap()
	}

	if addr.IsValid() && (addr.IsLinkLocalUnicast() || addr.IsMulticast()) {
		return true
	}
	if !wildcard && discoveryPorts[proto][port] && (zone != "" || !addr.IsLoopback()) {
		return true
	}
	if proto == "udp" && port >= ephemeralUDPLow && port <= ephemeralUDPHigh && !fixedHighUDPServices[port] {
		return true
	}
	return false
}

// splitListenerAddress parses `ss` local-address forms such as
// "*:22", "0.0.0.0:111", "[::]:22", "127.0.0.53%lo:53",
// "[fe80::1]%veth0:3702" and "239.255.255.250:1900".
func splitListenerAddress(address string) (host, zone string, port int, ok bool) {
	address = strings.TrimSpace(address)
	idx := strings.LastIndex(address, ":")
	if idx <= 0 || idx == len(address)-1 {
		return "", "", 0, false
	}
	p, err := strconv.Atoi(address[idx+1:])
	if err != nil || p < 0 || p > 65535 {
		return "", "", 0, false
	}
	host = address[:idx]
	if pct := strings.Index(host, "%"); pct >= 0 {
		zone = strings.Trim(host[pct+1:], "[]")
		host = host[:pct]
	}
	host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	if pct := strings.Index(host, "%"); pct >= 0 {
		if zone == "" {
			zone = host[pct+1:]
		}
		host = host[:pct]
	}
	if host == "" {
		return "", "", 0, false
	}
	return host, zone, p, true
}

// IsRoutineEvent reports whether a recorded event is listener churn under
// IsRoutineListener. It lets timelines recorded by earlier versions be
// read without that noise, without rewriting history the user did not ask
// to change.
func IsRoutineEvent(e Event) bool {
	if e.Category != "listener" {
		return false
	}
	rest := ""
	switch {
	case strings.HasPrefix(e.Summary, "listener appeared: "):
		rest = strings.TrimPrefix(e.Summary, "listener appeared: ")
	case strings.HasPrefix(e.Summary, "listener disappeared: "):
		rest = strings.TrimPrefix(e.Summary, "listener disappeared: ")
	default:
		return false
	}
	proto, address, found := strings.Cut(rest, " ")
	if !found {
		return false
	}
	return IsRoutineListener(proto, address)
}

// MeaningfulEvents returns events with routine listener churn removed,
// preserving order.
func MeaningfulEvents(events []Event) []Event {
	out := make([]Event, 0, len(events))
	for _, e := range events {
		if !IsRoutineEvent(e) {
			out = append(out, e)
		}
	}
	return out
}
