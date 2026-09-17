package core

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
)

func (r *evidenceRedactor) freeform(value string) string {
	if strings.TrimSpace(value) == "" {
		return value
	}
	out := evidencePrivateKeyBlockPattern.ReplaceAllString(value, "[REDACTED_PRIVATE_KEY]")
	out = evidencePrivateKeyRemainderPattern.ReplaceAllString(out, "[REDACTED_PRIVATE_KEY]")
	out = evidencePrivateKeyHeaderPattern.ReplaceAllString(out, "[REDACTED_PRIVATE_KEY]")
	out = evidenceSecretKVPattern.ReplaceAllString(out, "$1=[REDACTED_SECRET]")
	out = evidenceAuthPattern.ReplaceAllString(out, "$1 [REDACTED_SECRET]")
	out = evidenceURLPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.redactURL(raw) })
	out = evidenceEmailPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("email", raw) })
	out = evidenceDomainPattern.ReplaceAllStringFunc(out, func(raw string) string {
		lower := strings.ToLower(raw)
		if strings.HasSuffix(lower, ".invalid") || strings.HasSuffix(lower, ".service") {
			return raw
		}
		return r.alias("host", raw)
	})

	out = r.replaceKnownAliases(out)
	out = r.redactIPv6Tokens(out)
	out = evidenceMACPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("mac", raw) })
	out = evidenceUUIDPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("id", raw) })
	out = evidenceServicePattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("service", raw) })
	out = evidenceIPv4Pattern.ReplaceAllStringFunc(out, func(raw string) string {
		if net.ParseIP(raw) == nil {
			return raw
		}
		return r.ip(raw)
	})
	out = evidenceAbsPathPattern.ReplaceAllStringFunc(out, func(match string) string {
		idx := strings.Index(match, "/")
		if idx < 0 {
			return match
		}
		prefix, path := match[:idx], match[idx:]
		path = strings.TrimRight(path, ".);]")
		if prefix == ":" && strings.HasPrefix(path, "//") {
			return match
		}
		if evidenceCIDRSuffixPattern.MatchString(path) {
			return match
		}
		return prefix + r.path(path)
	})
	if len(out) > evidenceFreeformLimit {
		out = out[:evidenceFreeformLimit] + " [truncated after redaction]"
	}
	return out
}

func (r *evidenceRedactor) replaceKnownAliases(value string) string {
	type replacement struct{ original, alias, category string }
	var replacements []replacement
	for category, values := range r.aliases {
		for original, alias := range values {
			if original == "" || len(original) < 3 {
				continue
			}
			replacements = append(replacements, replacement{original: original, alias: alias, category: category})
		}
	}
	sort.Slice(replacements, func(i, j int) bool { return len(replacements[i].original) > len(replacements[j].original) })
	out := value
	for _, item := range replacements {
		count := strings.Count(out, item.original)
		if count == 0 {
			continue
		}
		out = strings.ReplaceAll(out, item.original, item.alias)
		r.counts[item.category] += count
	}
	return out
}

func (r *evidenceRedactor) redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		r.counts["url"]++
		return "[REDACTED_URL]"
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	host := u.Hostname()
	port := u.Port()
	redactedHost := r.host(host)
	if port != "" {
		u.Host = net.JoinHostPort(redactedHost, port)
	} else {
		u.Host = redactedHost
	}
	path := u.Path
	switch {
	case path == "":
		path = ""
	case path == "/":
		path = "/"
	default:
		path = "/[redacted]"
	}
	r.counts["url"]++
	return u.Scheme + "://" + u.Host + path
}

func (r *evidenceRedactor) redactIPv6Tokens(value string) string {
	out := value
	for _, field := range strings.Fields(value) {
		candidate := strings.Trim(field, `(){}<>"\' ,;`)
		redacted, ok := r.redactIPv6Candidate(candidate)
		if !ok || redacted == candidate {
			continue
		}
		out = strings.Replace(out, candidate, redacted, 1)
	}
	return out
}

func (r *evidenceRedactor) redactIPv6Candidate(value string) (string, bool) {
	if value == "" {
		return value, false
	}
	if host, port, err := net.SplitHostPort(value); err == nil {
		host = strings.Trim(host, "[]")
		ip := net.ParseIP(host)
		if ip != nil && ip.To4() == nil {
			return r.ip(host) + ":" + port, true
		}
	}
	trimmed := strings.Trim(value, "[]")
	if ip, network, err := net.ParseCIDR(trimmed); err == nil && ip.To4() == nil {
		prefix, _ := network.Mask.Size()
		return r.ip(ip.String()) + fmt.Sprintf("/%d", prefix), true
	}
	if ip := net.ParseIP(trimmed); ip != nil && ip.To4() == nil {
		return r.ip(ip.String()), true
	}
	return value, false
}
