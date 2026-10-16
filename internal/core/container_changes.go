package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Container lifecycle changes that `docker ps` status alone cannot show:
//
//   - an image update (the same name now runs a different image: a new tag,
//     or a new build behind the same tag such as :latest);
//   - a recreation with the same image (for example `docker compose up`);
//   - a restart inside one snapshot interval: the status reads "Up" before
//     and after, but Docker's start time moved.
//
// Every comparison needs the value on both sides, so upgrading from a
// version that did not record runtime evidence never replays history.
const (
	containerUpdatedPrefix   = "container updated: "
	containerRecreatedPrefix = "container recreated: "
	containerRestartedPrefix = "container restarted: "
)

func diffContainerRuntime(at, oldCapture time.Time, oldList, newList []ContainerInfo) []Event {
	oldByName := map[string]ContainerInfo{}
	for _, c := range oldList {
		oldByName[c.Name] = c
	}
	names := make([]string, 0, len(newList))
	newByName := map[string]ContainerInfo{}
	for _, c := range newList {
		if _, ok := oldByName[c.Name]; ok {
			names = append(names, c.Name)
			newByName[c.Name] = c
		}
	}
	sort.Strings(names)

	var out []Event
	for _, name := range names {
		o, n := oldByName[name], newByName[name]
		if o.ID == "" || n.ID == "" {
			continue
		}
		if !sameContainerID(o.ID, n.ID) {
			switch {
			case o.ImageID != "" && n.ImageID != "" && o.ImageID != n.ImageID:
				out = append(out, Event{At: at, Category: "container", Severity: "info", Summary: containerUpdatedPrefix + name + ": " + imageChangeText(o, n)})
			case strings.TrimSpace(o.Image) != strings.TrimSpace(n.Image) && o.Image != "" && n.Image != "":
				out = append(out, Event{At: at, Category: "container", Severity: "info", Summary: containerUpdatedPrefix + name + ": " + o.Image + " \u2192 " + n.Image})
			case o.ImageID != "" && n.ImageID != "":
				out = append(out, Event{At: at, Category: "container", Severity: "info", Summary: containerRecreatedPrefix + name + " (same image)"})
			}
			continue
		}
		running := strings.HasPrefix(strings.ToLower(strings.TrimSpace(n.Status)), "up")
		exits := exitsSince(n.RecentExits, oldCapture)
		switch {
		case running && len(exits) > 0:
			// Docker saw each exit; the container is running again.
			for _, exit := range exits {
				out = append(out, containerExitEvent(at, name, exit))
			}
		case running && o.StartedAt != "" && n.StartedAt != "" && o.StartedAt != n.StartedAt:
			// Started again, but Docker's event log did not cover the exit.
			out = append(out, Event{At: at, Category: "container", Severity: "warning", Summary: containerRestartedPrefix + name + " (how the previous run ended was not reported)"})
		}
	}
	return out
}

// Docker shortens IDs differently in different places; compare by prefix.
func sameContainerID(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if len(a) > len(b) {
		a, b = b, a
	}
	return a != "" && strings.HasPrefix(b, a)
}

func imageChangeText(o, n ContainerInfo) string {
	ref := n.Image
	if o.Image != n.Image && o.Image != "" {
		ref = o.Image + " \u2192 " + n.Image
	}
	switch {
	case o.ImageVersion != "" && n.ImageVersion != "" && o.ImageVersion != n.ImageVersion:
		return fmt.Sprintf("%s, version %s \u2192 %s", ref, o.ImageVersion, n.ImageVersion)
	case n.ImageVersion != "" && o.ImageVersion == "":
		return fmt.Sprintf("%s, now version %s", ref, n.ImageVersion)
	}
	// No version label: say what changed without printing image digests,
	// which are fingerprints the Evidence Bundle would have to redact.
	return ref + ", new image build (no version label)"
}

func exitsSince(exits []ContainerExit, after time.Time) []ContainerExit {
	if after.IsZero() {
		return nil
	}
	var out []ContainerExit
	for _, exit := range exits {
		if exit.At.After(after) {
			out = append(out, exit)
		}
	}
	return out
}

// containerExitEvent describes one exit Docker reported for a container
// that is running again: a restart, by its restart policy, autoheal, a
// person or a deploy tool. Exit codes follow the shell convention:
// 128+N means signal N (143 = SIGTERM, 137 = SIGKILL).
func containerExitEvent(at time.Time, name string, exit ContainerExit) Event {
	severity := "warning"
	how := ""
	switch {
	case exit.OOMKilled:
		how = "killed for running out of memory"
	case exit.ExitCode == 0:
		how = "stopped cleanly, exit code 0"
		severity = "info"
	case exit.ExitCode == 143:
		how = "stopped by SIGTERM, exit code 143"
		severity = "info"
	case exit.ExitCode == 137:
		how = "killed with SIGKILL, exit code 137"
	default:
		how = fmt.Sprintf("crashed, exit code %d", exit.ExitCode)
	}
	return Event{At: at, Category: "container", Severity: severity, Summary: containerRestartedPrefix + name + " (" + how + ")"}
}
