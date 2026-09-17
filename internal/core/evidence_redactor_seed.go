package core

import (
	"fmt"
	"sort"
	"strings"
)

func newEvidenceRedactor(source evidenceSource) *evidenceRedactor {
	r := &evidenceRedactor{aliases: map[string]map[string]string{}, counts: map[string]int{}}
	sets := map[string]map[string]bool{}
	add := func(category, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if sets[category] == nil {
			sets[category] = map[string]bool{}
		}
		sets[category][value] = true
	}

	add("host", source.Snapshot.Host.Hostname)
	add("id", source.Snapshot.Host.BootID)
	for _, iface := range source.Snapshot.Interfaces {
		if iface.Name != "lo" {
			add("interface", iface.Name)
		}
		for _, address := range iface.Addresses {
			collectAddressAlias(add, address)
		}
	}
	for _, fs := range source.Snapshot.Filesystems {
		collectPathAlias(add, fs.MountPoint)
		collectPathAlias(add, fs.Source)
		collectGenericAliases(add, fs.Source)
	}
	for _, route := range source.Snapshot.Routes {
		collectGenericAliases(add, route)
	}
	for _, listener := range source.Snapshot.Listeners {
		collectAddressAlias(add, listener.Address)
		collectGenericAliases(add, listener.Process)
	}
	for _, service := range source.Snapshot.Services {
		add("service", service.Name)
	}
	for _, container := range source.Snapshot.Containers {
		add("id", container.ID)
		add("container", container.Name)
		collectImageAlias(add, container.Image)
		for _, network := range splitList(container.Networks) {
			if network != "bridge" && network != "host" && network != "none" {
				add("network", network)
			}
		}
		collectGenericAliases(add, container.Ports)
	}
	for _, fingerprint := range source.Snapshot.ConfigFingerprints {
		if !evidenceLiteralConfigPaths[fingerprint.Path] {
			collectPathAlias(add, fingerprint.Path)
		}
		add("config-fingerprint", fingerprint.Fingerprint)
	}
	for _, event := range source.Events {
		collectEventAliases(add, event)
		collectGenericAliases(add, event.Summary)
	}
	for _, audit := range source.Audits {
		add("service", audit.Target)
		collectGenericAliases(add, audit.Summary)
		if audit.Before != nil {
			collectGenericAliases(add, audit.Before.ControlGroup)
		}
		if audit.After != nil {
			collectGenericAliases(add, audit.After.ControlGroup)
		}
	}

	categories := make([]string, 0, len(sets))
	for category := range sets {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	for _, category := range categories {
		values := make([]string, 0, len(sets[category]))
		for value := range sets[category] {
			values = append(values, value)
		}
		sort.Strings(values)
		for _, value := range values {
			r.register(category, value)
		}
	}
	return r
}

func (r *evidenceRedactor) register(category, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if r.aliases[category] == nil {
		r.aliases[category] = map[string]string{}
	}
	if alias := r.aliases[category][value]; alias != "" {
		return alias
	}
	n := len(r.aliases[category]) + 1
	var alias string
	switch category {
	case "host":
		alias = fmt.Sprintf("host-%02d.invalid", n)
	case "ipv4":
		alias = fmt.Sprintf("ipv4-%02d", n)
	case "ipv6":
		alias = fmt.Sprintf("ipv6-%02d", n)
	case "mac":
		alias = fmt.Sprintf("mac-%02d", n)
	case "path":
		alias = fmt.Sprintf("path-%02d", n)
	case "service":
		alias = fmt.Sprintf("service-%02d.service", n)
	case "container":
		alias = fmt.Sprintf("container-%02d", n)
	case "network":
		alias = fmt.Sprintf("network-%02d", n)
	case "interface":
		alias = fmt.Sprintf("interface-%02d", n)
	case "image":
		alias = fmt.Sprintf("image-%02d", n)
	case "email":
		alias = fmt.Sprintf("email-%02d.invalid", n)
	case "config-fingerprint":
		alias = fmt.Sprintf("config-fingerprint-%02d", n)
	case "cert-fingerprint":
		alias = fmt.Sprintf("certificate-fingerprint-%02d", n)
	default:
		alias = fmt.Sprintf("id-%02d", n)
	}
	r.aliases[category][value] = alias
	return alias
}

func (r *evidenceRedactor) alias(category, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	alias := r.register(category, value)
	r.counts[category]++
	return alias
}
