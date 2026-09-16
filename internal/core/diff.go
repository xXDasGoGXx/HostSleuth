package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func DiffSnapshots(oldSnap, newSnap Snapshot) []Event {
	at := newSnap.CapturedAt
	if at.IsZero() {
		at = time.Now().UTC()
	}
	var events []Event
	if oldSnap.Host.Kernel != "" && newSnap.Host.Kernel != oldSnap.Host.Kernel {
		events = append(events, Event{At: at, Category: "system", Severity: "info", Summary: fmt.Sprintf("kernel changed: %s -> %s", oldSnap.Host.Kernel, newSnap.Host.Kernel)})
	}
	events = append(events, diffNamedStates(at, "service", serviceMap(oldSnap.Services), serviceMap(newSnap.Services))...)
	events = append(events, diffNamedStates(at, "container", containerMap(oldSnap.Containers), containerMap(newSnap.Containers))...)
	events = append(events, diffSet(at, "listener", listenerSet(oldSnap.Listeners), listenerSet(newSnap.Listeners))...)
	if oldSnap.SchemaVersion >= snapshotSchemaVersion {
		events = append(events, diffPackageChanges(oldSnap.PackageChanges, newSnap.PackageChanges)...)
	}
	return events
}

func diffNamedStates(at time.Time, category string, oldMap, newMap map[string]string) []Event {
	keys := make([]string, 0, len(oldMap)+len(newMap))
	seen := map[string]bool{}
	for k := range oldMap {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	for k := range newMap {
		if !seen[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	sort.Strings(keys)
	var out []Event
	for _, k := range keys {
		o, okOld := oldMap[k]
		n, okNew := newMap[k]
		switch {
		case !okOld && okNew:
			out = append(out, Event{At: at, Category: category, Severity: "info", Summary: fmt.Sprintf("%s appeared: %s (%s)", category, k, n)})
		case okOld && !okNew:
			out = append(out, Event{At: at, Category: category, Severity: "warning", Summary: fmt.Sprintf("%s disappeared: %s (was %s)", category, k, o)})
		case o != n:
			sev := "info"
			if strings.Contains(n, "failed") || strings.Contains(n, "exited") || strings.Contains(n, "inactive") || strings.Contains(n, "unhealthy") || strings.Contains(n, "dead") || strings.Contains(n, "restarting") || strings.Contains(n, "paused") {
				sev = "warning"
			}
			out = append(out, Event{At: at, Category: category, Severity: sev, Summary: fmt.Sprintf("%s changed: %s: %s -> %s", category, k, o, n)})
		}
	}
	return out
}

func diffSet(at time.Time, category string, oldSet, newSet map[string]bool) []Event {
	var out []Event
	for item := range oldSet {
		if !newSet[item] {
			out = append(out, Event{At: at, Category: category, Severity: "warning", Summary: fmt.Sprintf("%s disappeared: %s", category, item)})
		}
	}
	for item := range newSet {
		if !oldSet[item] {
			out = append(out, Event{At: at, Category: category, Severity: "info", Summary: fmt.Sprintf("%s appeared: %s", category, item)})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Summary < out[j].Summary })
	return out
}

func diffPackageChanges(oldChanges, newChanges []PackageChange) []Event {
	seen := make(map[string]bool, len(oldChanges))
	for _, change := range oldChanges {
		seen[packageChangeKey(change)] = true
	}
	var out []Event
	for _, change := range newChanges {
		if seen[packageChangeKey(change)] {
			continue
		}
		at := change.At
		if at.IsZero() {
			at = time.Now().UTC()
		}
		out = append(out, Event{At: at, Category: "package", Severity: "info", Summary: packageChangeSummary(change)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].At.Equal(out[j].At) {
			return out[i].Summary < out[j].Summary
		}
		return out[i].At.Before(out[j].At)
	})
	return out
}

func serviceMap(in []ServiceInfo) map[string]string {
	m := map[string]string{}
	for _, s := range in {
		m[s.Name] = strings.TrimSpace(s.Active + "/" + s.Sub)
	}
	return m
}

func containerMap(in []ContainerInfo) map[string]string {
	m := map[string]string{}
	for _, c := range in {
		m[c.Name] = stableContainerState(c.Status)
	}
	return m
}

func stableContainerState(status string) string {
	s := strings.TrimSpace(status)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "up "):
		if strings.Contains(lower, "(paused)") {
			return "paused"
		}
		if suffix := containerStatusSuffix(s); suffix != "" {
			return "running " + suffix
		}
		return "running"
	case strings.HasPrefix(lower, "exited"):
		return containerStatusPrefixThroughParen(lower, "exited")
	case strings.HasPrefix(lower, "restarting"):
		return containerStatusPrefixThroughParen(lower, "restarting")
	case strings.HasPrefix(lower, "created"):
		return "created"
	case strings.HasPrefix(lower, "dead"):
		return "dead"
	case strings.HasPrefix(lower, "paused"):
		return "paused"
	case strings.HasPrefix(lower, "removal in progress"), strings.HasPrefix(lower, "removing"):
		return "removing"
	default:
		return lower
	}
}

func containerStatusSuffix(status string) string {
	start := strings.LastIndex(status, " (")
	if start < 0 || !strings.HasSuffix(status, ")") {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(status[start+1:]))
}

func containerStatusPrefixThroughParen(status, fallback string) string {
	if end := strings.Index(status, ")"); end >= 0 {
		return strings.TrimSpace(status[:end+1])
	}
	return fallback
}

func listenerSet(in []Listener) map[string]bool {
	m := map[string]bool{}
	for _, l := range in {
		m[l.Protocol+" "+l.Address] = true
	}
	return m
}
