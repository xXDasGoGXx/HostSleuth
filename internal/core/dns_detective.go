package core

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	dnsDetectiveResolverLimit = 4
	dnsDetectiveLookupTimeout = 5 * time.Second
	dnsResolvConfBytesLimit   = 64 * 1024
)

var dnsResolvConfPath = "/etc/resolv.conf"

type DNSRuntimeConfig struct {
	Path        string   `json:"path"`
	Nameservers []string `json:"nameservers,omitempty"`
	Search      []string `json:"search,omitempty"`
	Options     []string `json:"options,omitempty"`
	Note        string   `json:"note"`
}

type DNSAddressEvidence struct {
	Address string `json:"address"`
	Family  string `json:"family"`
	Scope   string `json:"scope"`
}

type DNSResolverEvidence struct {
	Label      string               `json:"label"`
	Server     string               `json:"server"`
	Status     string               `json:"status"`
	DurationMS int64                `json:"duration_ms"`
	Addresses  []DNSAddressEvidence `json:"addresses,omitempty"`
	CNAME      string               `json:"cname,omitempty"`
	PTR        []string             `json:"ptr,omitempty"`
	Error      string               `json:"error,omitempty"`
	CNAMEError string               `json:"cname_error,omitempty"`
}

type DNSDifference struct {
	Resolver string `json:"resolver"`
	Kind     string `json:"kind"`
	Baseline string `json:"baseline"`
	Observed string `json:"observed"`
}

type DNSDetectiveResult struct {
	Name          string                `json:"name"`
	StartedAt     time.Time             `json:"started_at"`
	Status        string                `json:"status"`
	Conclusion    string                `json:"conclusion"`
	SplitViewHint bool                  `json:"split_view_hint,omitempty"`
	Runtime       DNSRuntimeConfig      `json:"runtime"`
	Resolvers     []DNSResolverEvidence `json:"resolvers"`
	Differences   []DNSDifference       `json:"differences,omitempty"`
}

type dnsResolverSpec struct {
	Label  string
	Server string
	System bool
}

var dnsDetectiveLookup = lookupDNSResolver

func InspectDNSDetective(ctx context.Context, name string, resolverInputs []string) (*DNSDetectiveResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("DNS name or IP is required")
	}
	if len(name) > 253 && net.ParseIP(name) == nil {
		return nil, errors.New("DNS name is too long")
	}

	custom, err := normalizeDNSResolverSpecs(resolverInputs)
	if err != nil {
		return nil, err
	}
	specs := make([]dnsResolverSpec, 0, 1+len(custom))
	specs = append(specs, dnsResolverSpec{Label: "system", Server: "system resolver", System: true})
	specs = append(specs, custom...)

	result := &DNSDetectiveResult{
		Name:      name,
		StartedAt: time.Now().UTC(),
		Runtime:   readDNSRuntimeConfig(),
		Resolvers: make([]DNSResolverEvidence, len(specs)),
	}

	type indexedResult struct {
		index    int
		evidence DNSResolverEvidence
	}
	ch := make(chan indexedResult, len(specs))
	for i, spec := range specs {
		go func(index int, resolver dnsResolverSpec) {
			ch <- indexedResult{index: index, evidence: dnsDetectiveLookup(ctx, resolver, name)}
		}(i, spec)
	}
	for range specs {
		item := <-ch
		result.Resolvers[item.index] = item.evidence
	}

	result.Status, result.Conclusion, result.Differences, result.SplitViewHint = compareDNSResolverEvidence(result.Resolvers)
	return result, nil
}

func normalizeDNSResolverSpecs(inputs []string) ([]dnsResolverSpec, error) {
	if len(inputs) > dnsDetectiveResolverLimit {
		return nil, fmt.Errorf("at most %d custom DNS resolvers are allowed", dnsDetectiveResolverLimit)
	}

	out := make([]dnsResolverSpec, 0, len(inputs))
	seenServer := map[string]struct{}{}
	seenLabel := map[string]struct{}{}
	for i, raw := range inputs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		label := ""
		address := raw
		if left, right, ok := strings.Cut(raw, "="); ok {
			label = strings.TrimSpace(left)
			address = strings.TrimSpace(right)
			if label == "" || address == "" {
				return nil, fmt.Errorf("resolver %q must use LABEL=IP", raw)
			}
		}
		if label == "" {
			label = fmt.Sprintf("resolver-%d", i+1)
		}
		if err := validateDNSResolverLabel(label); err != nil {
			return nil, err
		}

		server, err := normalizeDNSResolverAddress(address)
		if err != nil {
			return nil, fmt.Errorf("resolver %q: %w", label, err)
		}
		key := strings.ToLower(server)
		if _, ok := seenServer[key]; ok {
			return nil, fmt.Errorf("resolver %q duplicates %s", label, server)
		}
		if _, ok := seenLabel[strings.ToLower(label)]; ok {
			return nil, fmt.Errorf("resolver label %q is duplicated", label)
		}
		seenServer[key] = struct{}{}
		seenLabel[strings.ToLower(label)] = struct{}{}
		out = append(out, dnsResolverSpec{Label: label, Server: server})
	}
	return out, nil
}

func validateDNSResolverLabel(label string) error {
	if len(label) > 40 {
		return errors.New("resolver label must be 40 characters or fewer")
	}
	for _, r := range label {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return fmt.Errorf("resolver label %q contains unsupported characters", label)
	}
	return nil
}

func normalizeDNSResolverAddress(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("DNS resolver IP is required")
	}

	if ip := net.ParseIP(strings.Trim(value, "[]")); ip != nil {
		return net.JoinHostPort(ip.String(), "53"), nil
	}

	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return "", errors.New("DNS resolver must be an IP address, optionally with :53")
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return "", errors.New("DNS resolver must be an IP address")
	}
	if port != "53" {
		return "", errors.New("custom DNS resolvers are restricted to port 53")
	}
	return net.JoinHostPort(ip.String(), "53"), nil
}

func lookupDNSResolver(ctx context.Context, spec dnsResolverSpec, name string) DNSResolverEvidence {
	started := time.Now()
	evidence := DNSResolverEvidence{Label: spec.Label, Server: spec.Server, Status: "fail"}
	lookupCtx, cancel := context.WithTimeout(ctx, dnsDetectiveLookupTimeout)
	defer cancel()

	resolver := net.DefaultResolver
	if !spec.System {
		server := spec.Server
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
				dialer := &net.Dialer{Timeout: 3 * time.Second}
				return dialer.DialContext(ctx, network, server)
			},
		}
	}

	if ip := net.ParseIP(strings.Trim(name, "[]")); ip != nil {
		names, err := resolver.LookupAddr(lookupCtx, ip.String())
		evidence.DurationMS = time.Since(started).Milliseconds()
		if err != nil {
			evidence.Error = boundedEvidence(err.Error(), 512)
			return evidence
		}
		for _, value := range names {
			value = normalizeDNSName(value)
			if value != "" {
				evidence.PTR = append(evidence.PTR, value)
			}
		}
		sort.Strings(evidence.PTR)
		evidence.PTR = uniqueStrings(evidence.PTR)
		evidence.Status = "pass"
		return evidence
	}

	ips, err := resolver.LookupIP(lookupCtx, "ip", name)
	if err != nil {
		evidence.DurationMS = time.Since(started).Milliseconds()
		evidence.Error = boundedEvidence(err.Error(), 512)
		return evidence
	}
	for _, ip := range ips {
		evidence.Addresses = append(evidence.Addresses, dnsAddressEvidence(ip))
	}
	sort.Slice(evidence.Addresses, func(i, j int) bool {
		return evidence.Addresses[i].Address < evidence.Addresses[j].Address
	})
	evidence.Addresses = uniqueDNSAddresses(evidence.Addresses)

	cname, cnameErr := resolver.LookupCNAME(lookupCtx, name)
	if cnameErr != nil {
		evidence.CNAMEError = boundedEvidence(cnameErr.Error(), 384)
	} else if normalized := normalizeDNSName(cname); normalized != "" && normalized != normalizeDNSName(name) {
		evidence.CNAME = normalized
	}
	evidence.DurationMS = time.Since(started).Milliseconds()
	evidence.Status = "pass"
	return evidence
}

func dnsAddressEvidence(ip net.IP) DNSAddressEvidence {
	family := "IPv6"
	if ip.To4() != nil {
		family = "IPv4"
	}
	return DNSAddressEvidence{
		Address: ip.String(),
		Family:  family,
		Scope:   dnsAddressScope(ip),
	}
}

func dnsAddressScope(ip net.IP) string {
	switch {
	case ip == nil:
		return "other"
	case ip.IsLoopback():
		return "loopback"
	case ip.IsPrivate():
		return "private"
	case ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast():
		return "link-local"
	case ip.IsUnspecified():
		return "unspecified"
	case ip.IsMulticast():
		return "multicast"
	case ip.IsGlobalUnicast():
		return "global"
	default:
		return "other"
	}
}

func uniqueDNSAddresses(values []DNSAddressEvidence) []DNSAddressEvidence {
	if len(values) == 0 {
		return values
	}
	out := values[:0]
	for i, value := range values {
		if i == 0 || value.Address != values[i-1].Address {
			out = append(out, value)
		}
	}
	return out
}

func normalizeDNSName(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func compareDNSResolverEvidence(results []DNSResolverEvidence) (status, conclusion string, differences []DNSDifference, splitHint bool) {
	if len(results) == 0 {
		return "partial", "no resolver evidence was produced", nil, false
	}

	successful := make([]DNSResolverEvidence, 0, len(results))
	failures := 0
	for _, result := range results {
		if result.Status == "pass" {
			successful = append(successful, result)
		} else {
			failures++
		}
	}
	if len(results) == 1 {
		if len(successful) == 1 {
			return "single", "system resolver answered; add another resolver to compare DNS views", nil, false
		}
		return "partial", "the system resolver did not return a usable answer", nil, false
	}
	if len(successful) < 2 {
		return "partial", fmt.Sprintf("%d of %d resolver views answered successfully", len(successful), len(results)), nil, false
	}

	baseline := successful[0]
	baselineSig := dnsResolverSignature(baseline)
	for _, current := range successful[1:] {
		currentSig := dnsResolverSignature(current)
		if currentSig == baselineSig {
			continue
		}
		differences = append(differences, DNSDifference{
			Resolver: current.Label,
			Kind:     dnsDifferenceKind(baseline, current),
			Baseline: dnsResolverDisplayValue(baseline),
			Observed: dnsResolverDisplayValue(current),
		})
	}
	if len(differences) > 0 {
		splitHint = dnsPrivateGlobalSplit(successful)
		conclusion = fmt.Sprintf("%d resolver view(s) disagree with the baseline resolver", len(differences))
		if splitHint {
			conclusion += "; private/local and global answers differ, which is consistent with split-view DNS or resolver-specific overrides"
		}
		return "diverge", conclusion, differences, splitHint
	}
	if failures > 0 {
		return "partial", fmt.Sprintf("successful resolver views agree, but %d resolver lookup(s) failed", failures), nil, false
	}
	return "agree", "all compared resolver views agree", nil, false
}

func dnsResolverSignature(value DNSResolverEvidence) string {
	if len(value.PTR) > 0 {
		ptr := append([]string(nil), value.PTR...)
		sort.Strings(ptr)
		return "ptr=" + strings.Join(ptr, ",")
	}
	addresses := make([]string, 0, len(value.Addresses))
	for _, address := range value.Addresses {
		addresses = append(addresses, address.Address)
	}
	sort.Strings(addresses)
	return "addr=" + strings.Join(addresses, ",") + "|cname=" + normalizeDNSName(value.CNAME)
}

func dnsResolverDisplayValue(value DNSResolverEvidence) string {
	if len(value.PTR) > 0 {
		return "PTR " + strings.Join(value.PTR, ", ")
	}
	addresses := make([]string, 0, len(value.Addresses))
	for _, address := range value.Addresses {
		addresses = append(addresses, address.Address)
	}
	text := strings.Join(addresses, ", ")
	if value.CNAME != "" {
		if text != "" {
			text += "; "
		}
		text += "CNAME " + value.CNAME
	}
	if text == "" {
		return "no answer"
	}
	return text
}

func dnsDifferenceKind(left, right DNSResolverEvidence) string {
	if normalizeDNSName(left.CNAME) != normalizeDNSName(right.CNAME) {
		return "cname"
	}
	return "address-set"
}

func dnsPrivateGlobalSplit(results []DNSResolverEvidence) bool {
	hasLocal := false
	hasGlobal := false
	for _, result := range results {
		if result.Status != "pass" {
			continue
		}
		for _, address := range result.Addresses {
			switch address.Scope {
			case "private", "link-local", "loopback":
				hasLocal = true
			case "global":
				hasGlobal = true
			}
		}
	}
	return hasLocal && hasGlobal
}

func readDNSRuntimeConfig() DNSRuntimeConfig {
	result := DNSRuntimeConfig{
		Path: dnsResolvConfPath,
		Note: "resolver configuration visible to the HostSleuth process; in Docker mode this may be the container runtime resolver configuration rather than the host-native file",
	}
	f, err := os.Open(dnsResolvConfPath)
	if err != nil {
		return result
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), dnsResolvConfBytesLimit)
	readBytes := 0
	for scanner.Scan() {
		line := scanner.Text()
		readBytes += len(line) + 1
		if readBytes > dnsResolvConfBytesLimit {
			break
		}
		if before, _, ok := strings.Cut(line, "#"); ok {
			line = before
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "nameserver":
			if ip := net.ParseIP(fields[1]); ip != nil {
				result.Nameservers = append(result.Nameservers, ip.String())
			}
		case "search":
			for _, value := range fields[1:] {
				if normalized := normalizeDNSName(value); normalized != "" {
					result.Search = append(result.Search, normalized)
				}
			}
		case "domain":
			if normalized := normalizeDNSName(fields[1]); normalized != "" {
				result.Search = append(result.Search, normalized)
			}
		case "options":
			for _, option := range fields[1:] {
				if dnsOptionAllowed(option) {
					result.Options = append(result.Options, option)
				}
			}
		}
	}
	sort.Strings(result.Nameservers)
	sort.Strings(result.Search)
	sort.Strings(result.Options)
	result.Nameservers = uniqueStrings(result.Nameservers)
	result.Search = uniqueStrings(result.Search)
	result.Options = uniqueStrings(result.Options)
	return result
}

func dnsOptionAllowed(value string) bool {
	key := value
	if before, _, ok := strings.Cut(value, ":"); ok {
		key = before
	}
	switch key {
	case "ndots", "timeout", "attempts", "rotate", "single-request", "single-request-reopen", "use-vc", "trust-ad":
		return true
	default:
		return false
	}
}
