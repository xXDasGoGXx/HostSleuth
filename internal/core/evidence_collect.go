package core

import (
	"net"
	"strings"
)

func collectAddressAlias(add func(string, string), value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if ip, _, err := net.ParseCIDR(value); err == nil {
		collectIPAlias(add, ip.String())
		return
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		host = strings.Trim(host, "[]")
		if strings.Contains(host, "%") {
			parts := strings.SplitN(host, "%", 2)
			collectIPAlias(add, parts[0])
			if parts[1] != "lo" {
				add("interface", parts[1])
			}
			return
		}
		collectIPAlias(add, host)
		return
	}
	collectIPAlias(add, strings.Trim(value, "[]"))
}

func collectIPAlias(add func(string, string), value string) {
	ip := net.ParseIP(value)
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
		return
	}
	if ip.To4() != nil {
		add("ipv4", value)
	} else {
		add("ipv6", value)
	}
}

func collectPathAlias(add func(string, string), value string) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "/") && !evidenceLiteralConfigPaths[value] {
		add("path", value)
	}
}

func collectImageAlias(add func(string, string), value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if at := strings.LastIndex(value, "@"); at >= 0 {
		value = value[:at]
	} else if colon := strings.LastIndex(value, ":"); colon > strings.LastIndex(value, "/") {
		value = value[:colon]
	}
	add("image", value)
}

func collectGenericAliases(add func(string, string), value string) {
	for _, token := range strings.Fields(value) {
		candidate := strings.Trim(token, `(){}<>"\' ,;`)
		if strings.Contains(candidate, ":") || strings.Contains(candidate, "/") {
			collectAddressAlias(add, candidate)
		}
	}
	for _, raw := range evidenceIPv4Pattern.FindAllString(value, -1) {
		collectIPAlias(add, raw)
	}
	for _, raw := range evidenceMACPattern.FindAllString(value, -1) {
		add("mac", raw)
	}
	for _, raw := range evidenceUUIDPattern.FindAllString(value, -1) {
		add("id", raw)
	}
	for _, raw := range evidenceEmailPattern.FindAllString(value, -1) {
		add("email", raw)
	}
	for _, raw := range evidenceDomainPattern.FindAllString(value, -1) {
		lower := strings.ToLower(raw)
		if strings.HasSuffix(lower, ".service") || strings.HasSuffix(lower, ".invalid") {
			continue
		}
		add("host", raw)
	}
	for _, raw := range evidenceServicePattern.FindAllString(value, -1) {
		add("service", raw)
	}
	for _, match := range evidenceAbsPathPattern.FindAllStringSubmatch(value, -1) {
		if len(match) > 1 {
			collectPathAlias(add, strings.TrimRight(match[1], ".);]"))
		}
	}
}

func collectEventAliases(add func(string, string), event Event) {
	summary := strings.TrimSpace(event.Summary)
	switch event.Category {
	case "service", "container":
		prefixes := []string{event.Category + " appeared: ", event.Category + " disappeared: ", event.Category + " changed: "}
		for _, prefix := range prefixes {
			if strings.HasPrefix(summary, prefix) {
				value := strings.TrimPrefix(summary, prefix)
				if idx := strings.IndexAny(value, "(:"); idx >= 0 {
					value = strings.TrimSpace(value[:idx])
				}
				if event.Category == "service" {
					add("service", value)
				} else {
					add("container", value)
				}
				break
			}
		}
	case "configuration":
		if idx := strings.Index(summary, ": "); idx >= 0 {
			collectPathAlias(add, strings.TrimSpace(summary[idx+2:]))
		}
	}
}

func splitList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if value := strings.TrimSpace(field); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func cloneStringIntMap(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
