package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func DiffSnapshots(oldSnap, newSnap Snapshot) []Event {
	at := newSnap.CapturedAt
	if at.IsZero() { at = time.Now().UTC() }
	var events []Event
	if oldSnap.Host.Kernel != "" && newSnap.Host.Kernel != oldSnap.Host.Kernel {
		events = append(events, Event{At: at, Category: "system", Severity: "info", Summary: fmt.Sprintf("kernel changed: %s -> %s", oldSnap.Host.Kernel, newSnap.Host.Kernel)})
	}
	events = append(events, diffNamedStates(at, "service", serviceMap(oldSnap.Services), serviceMap(newSnap.Services))...)
	events = append(events, diffNamedStates(at, "container", containerMap(oldSnap.Containers), containerMap(newSnap.Containers))...)
	events = append(events, diffSet(at, "listener", listenerSet(oldSnap.Listeners), listenerSet(newSnap.Listeners))...)
	return events
}

func diffNamedStates(at time.Time, category string, oldMap, newMap map[string]string) []Event {
	keys := make([]string, 0, len(oldMap)+len(newMap))
	seen := map[string]bool{}
	for k := range oldMap { if !seen[k] { keys = append(keys, k); seen[k] = true } }
	for k := range newMap { if !seen[k] { keys = append(keys, k); seen[k] = true } }
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
			if strings.Contains(n, "failed") || strings.Contains(n, "exited") || strings.Contains(n, "inactive") { sev = "warning" }
			out = append(out, Event{At: at, Category: category, Severity: sev, Summary: fmt.Sprintf("%s changed: %s: %s -> %s", category, k, o, n)})
		}
	}
	return out
}

func diffSet(at time.Time, category string, oldSet, newSet map[string]bool) []Event {
	var out []Event
	for item := range oldSet {
		if !newSet[item] { out = append(out, Event{At: at, Category: category, Severity: "warning", Summary: fmt.Sprintf("%s disappeared: %s", category, item)}) }
	}
	for item := range newSet {
		if !oldSet[item] { out = append(out, Event{At: at, Category: category, Severity: "info", Summary: fmt.Sprintf("%s appeared: %s", category, item)}) }
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Summary < out[j].Summary })
	return out
}

func serviceMap(in []ServiceInfo) map[string]string {
	m := map[string]string{}
	for _, s := range in { m[s.Name] = strings.TrimSpace(s.Active + "/" + s.Sub) }
	return m
}

func containerMap(in []ContainerInfo) map[string]string {
	m := map[string]string{}
	for _, c := range in { m[c.Name] = c.Status }
	return m
}

func listenerSet(in []Listener) map[string]bool {
	m := map[string]bool{}
	for _, l := range in { m[l.Protocol+" "+l.Address] = true }
	return m
}
