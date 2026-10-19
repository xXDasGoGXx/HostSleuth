package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Container Story answers "why does this container keep restarting, and
// what changed?" from evidence HostSleuth already holds: the current
// field-limited Docker inspect, and retained timeline events (restarts with
// Docker's exit reason, health changes, image updates, state changes).
// Listener changes at the same moment are shown as context, never as cause.

const (
	containerStoryWindow   = 7 * 24 * time.Hour
	containerNearbyWindow  = 2 * time.Minute
	containerTimelineLimit = 60
	containerNearbyLimit   = 20
)

type ContainerSummary struct {
	Name          string     `json:"name"`
	Image         string     `json:"image"`
	ImageVersion  string     `json:"image_version,omitempty"`
	State         string     `json:"state"`
	Health        string     `json:"health,omitempty"`
	Ports         string     `json:"ports,omitempty"`
	RestartCount  int        `json:"restart_count"`
	Restarts24h   int        `json:"restarts_24h"`
	Restarts7d    int        `json:"restarts_7d"`
	Crashes7d     int        `json:"crashes_7d"`
	OOMKills7d    int        `json:"oom_kills_7d"`
	Unhealthy7d   int        `json:"unhealthy_7d"`
	LastRestartAt *time.Time `json:"last_restart_at,omitempty"`
	LastUpdate    string     `json:"last_update,omitempty"`
	LastUpdateAt  *time.Time `json:"last_update_at,omitempty"`
	Attention     string     `json:"attention,omitempty"`
}

type ContainerList struct {
	CapturedAt time.Time          `json:"captured_at"`
	Mode       string             `json:"mode,omitempty"`
	Window     string             `json:"window"`
	Containers []ContainerSummary `json:"containers"`
	Note       string             `json:"note,omitempty"`
}

type ContainerStory struct {
	CapturedAt time.Time        `json:"captured_at"`
	Mode       string           `json:"mode,omitempty"`
	Name       string           `json:"name"`
	Found      bool             `json:"found"`
	Status     string           `json:"status"`
	Conclusion string           `json:"conclusion"`
	Summary    ContainerSummary `json:"summary"`
	Current    *ContainerInfo   `json:"current,omitempty"`
	Timeline   []Event          `json:"timeline"`
	Nearby     []Event          `json:"nearby_listener_changes"`
	Notes      []string         `json:"notes"`
	// Why is set when the container is unhealthy now.
	Why *HealthWhy `json:"why,omitempty"`
}

// containerEventName returns the container an event is about and the kind
// of change, for every container summary shape HostSleuth records.
func containerEventName(e Event) (name, kind string) {
	if e.Category != "container" {
		return "", ""
	}
	s := e.Summary
	prefixes := []struct{ prefix, kind, stop string }{
		{containerRestartedPrefix, "restart", " ("},
		{containerUpdatedPrefix, "update", ": "},
		{containerRecreatedPrefix, "recreate", " ("},
		{"container changed: ", "state", ": "},
		{"container appeared: ", "appeared", " ("},
		{"container disappeared: ", "disappeared", " (was "},
	}
	for _, p := range prefixes {
		if !strings.HasPrefix(s, p.prefix) {
			continue
		}
		rest := strings.TrimPrefix(s, p.prefix)
		if idx := strings.Index(rest, p.stop); idx >= 0 {
			rest = rest[:idx]
		}
		return strings.TrimSpace(rest), p.kind
	}
	return "", ""
}

func containerHealthFromStatus(status string) string {
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "(unhealthy)"):
		return "unhealthy"
	case strings.Contains(lower, "(health: starting)"):
		return "starting"
	case strings.Contains(lower, "(healthy)"):
		return "healthy"
	}
	return ""
}

func summarizeContainer(c ContainerInfo, events []Event, now time.Time) ContainerSummary {
	sum := ContainerSummary{
		Name:         c.Name,
		Image:        c.Image,
		ImageVersion: c.ImageVersion,
		State:        stableContainerState(c.Status),
		Health:       containerHealthFromStatus(c.Status),
		Ports:        c.Ports,
		RestartCount: c.RestartCount,
	}
	for _, e := range events {
		name, kind := containerEventName(e)
		if name != c.Name || now.Sub(e.At) > containerStoryWindow {
			continue
		}
		switch kind {
		case "restart":
			sum.Restarts7d++
			if now.Sub(e.At) <= 24*time.Hour {
				sum.Restarts24h++
			}
			if e.Severity == "warning" && !strings.Contains(e.Summary, "was not reported") {
				sum.Crashes7d++
			}
			if strings.Contains(e.Summary, "out of memory") {
				sum.OOMKills7d++
			}
			if sum.LastRestartAt == nil || e.At.After(*sum.LastRestartAt) {
				at := e.At
				sum.LastRestartAt = &at
			}
		case "update", "recreate":
			if sum.LastUpdateAt == nil || e.At.After(*sum.LastUpdateAt) {
				at := e.At
				sum.LastUpdateAt = &at
				sum.LastUpdate = e.Summary
			}
		case "state":
			if strings.HasSuffix(e.Summary, "-> running (unhealthy)") {
				sum.Unhealthy7d++
			}
		}
	}
	switch {
	case sum.Health == "unhealthy":
		sum.Attention = "unhealthy"
	case strings.HasPrefix(sum.State, "restarting"):
		sum.Attention = "restarting"
	case strings.HasPrefix(sum.State, "exited") || sum.State == "dead":
		sum.Attention = "stopped"
	case sum.Crashes7d >= 3:
		sum.Attention = "crash-looping"
	case sum.Crashes7d > 0 || sum.Unhealthy7d > 0:
		sum.Attention = "unstable"
	}
	return sum
}

var containerAttentionRank = map[string]int{"unhealthy": 0, "restarting": 1, "crash-looping": 2, "stopped": 3, "unstable": 4, "": 5}

// BuildContainerList summarizes every container, most troubled first.
func BuildContainerList(snap Snapshot, events []Event) ContainerList {
	now := snap.CapturedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	list := ContainerList{CapturedAt: snap.CapturedAt, Mode: snap.Mode, Window: "7 days", Containers: []ContainerSummary{}}
	for _, c := range snap.Containers {
		list.Containers = append(list.Containers, summarizeContainer(c, events, now))
	}
	sort.SliceStable(list.Containers, func(i, j int) bool {
		a, b := list.Containers[i], list.Containers[j]
		if containerAttentionRank[a.Attention] != containerAttentionRank[b.Attention] {
			return containerAttentionRank[a.Attention] < containerAttentionRank[b.Attention]
		}
		if a.Restarts7d != b.Restarts7d {
			return a.Restarts7d > b.Restarts7d
		}
		return a.Name < b.Name
	})
	if len(snap.Containers) == 0 {
		list.Note = "The current snapshot has no Docker inventory; Docker may be absent or not readable by HostSleuth."
	}
	return list
}

// BuildContainerStory explains one container. Events should be the
// detailed (uncollapsed) timeline so listener changes can be matched.
func BuildContainerStory(snap Snapshot, events []Event, name string) ContainerStory {
	name = strings.TrimSpace(name)
	now := snap.CapturedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	story := ContainerStory{CapturedAt: snap.CapturedAt, Mode: snap.Mode, Name: name, Timeline: []Event{}, Nearby: []Event{}, Notes: []string{}}
	var current *ContainerInfo
	for i := range snap.Containers {
		if snap.Containers[i].Name == name {
			c := snap.Containers[i]
			current = &c
		}
	}

	var anchors []time.Time
	for _, e := range events {
		evName, kind := containerEventName(e)
		if evName != name || now.Sub(e.At) > containerStoryWindow {
			continue
		}
		story.Timeline = append(story.Timeline, e)
		if containerChangeAnchors(kind, e.Summary) {
			anchors = append(anchors, e.At)
		}
	}
	var ports map[string]bool
	if current != nil {
		ports = publishedHostPorts(*current)
	}
	if len(story.Timeline) > containerTimelineLimit {
		story.Timeline = story.Timeline[len(story.Timeline)-containerTimelineLimit:]
		story.Notes = append(story.Notes, fmt.Sprintf("Showing the newest %d of this container's changes.", containerTimelineLimit))
	}
	seen := map[int]bool{}
	for i, e := range events {
		if e.Category != "listener" || seen[i] {
			continue
		}
		if len(ports) > 0 && !ports[listenerEventPort(e)] {
			continue
		}
		for _, a := range anchors {
			d := e.At.Sub(a)
			if d < 0 {
				d = -d
			}
			if d <= containerNearbyWindow {
				seen[i] = true
				story.Nearby = append(story.Nearby, e)
				break
			}
		}
	}
	if len(story.Nearby) > containerNearbyLimit {
		story.Nearby = story.Nearby[len(story.Nearby)-containerNearbyLimit:]
	}

	if current == nil {
		story.Status = "unknown"
		story.Conclusion = fmt.Sprintf("No container named %q is in the current snapshot.", name)
		if len(story.Timeline) > 0 {
			story.Conclusion += " Its retained history is shown below."
		}
		return story
	}
	story.Found = true
	story.Current = current
	if containerHealthFromStatus(current.Status) == "unhealthy" {
		for _, p := range BuildReplay(snap, events).Parts {
			if p.ID == name {
				story.Why = p.Why
			}
		}
	}
	story.Summary = summarizeContainer(*current, events, now)
	story.Status, story.Conclusion = containerConclusion(story.Summary, *current)
	if current.StartedAt == "" && current.ImageID == "" {
		story.Notes = append(story.Notes, "Docker runtime details (start time, exits, image build) were not reported in this snapshot.")
	}
	switch {
	case len(story.Nearby) > 0 && len(ports) > 0:
		story.Notes = append(story.Notes, "Listener changes are shown for host ports this container publishes, within two minutes of its restarts, updates or failures.")
	case len(story.Nearby) > 0:
		story.Notes = append(story.Notes, "This container publishes no host ports Docker reports (host networking, for example), so listener changes within two minutes of its restarts, updates or failures are shown as context, not proof that they belong to it.")
	}
	return story
}

func containerConclusion(s ContainerSummary, c ContainerInfo) (string, string) {
	restarts := func() string {
		if s.Restarts7d == 0 {
			return ""
		}
		parts := []string{}
		if s.Crashes7d > 0 {
			parts = append(parts, plural(s.Crashes7d, "crash", "crashes"))
		}
		if s.OOMKills7d > 0 {
			parts = append(parts, plural(s.OOMKills7d, "out-of-memory kill", "out-of-memory kills"))
		}
		detail := ""
		if len(parts) > 0 {
			detail = " (" + strings.Join(parts, ", ") + ")"
		}
		return fmt.Sprintf(" It restarted %s in 7 days%s.", times(s.Restarts7d), detail)
	}()
	switch s.Attention {
	case "unhealthy":
		streak := ""
		if c.HealthFailingStreak > 0 {
			streak = fmt.Sprintf(", failing %s in a row", plural(c.HealthFailingStreak, "health check", "health checks"))
		}
		return "fail", fmt.Sprintf("%s is running but Docker reports it unhealthy%s.%s", s.Name, streak, restarts)
	case "restarting":
		return "fail", fmt.Sprintf("%s is restarting right now.%s", s.Name, restarts)
	case "stopped":
		return "fail", fmt.Sprintf("%s is not running (%s).%s", s.Name, s.State, restarts)
	case "crash-looping", "unstable":
		msg := fmt.Sprintf("%s is running now.%s", s.Name, restarts)
		if s.Unhealthy7d > 0 {
			msg += fmt.Sprintf(" Docker reported it unhealthy %s.", times(s.Unhealthy7d))
		}
		return "warning", msg
	}
	update := ""
	if s.LastUpdate != "" {
		if idx := strings.Index(s.LastUpdate, s.Name+": "); strings.HasPrefix(s.LastUpdate, containerUpdatedPrefix) && idx >= 0 {
			update = " Its image was updated in the last 7 days (" + s.LastUpdate[idx+len(s.Name)+2:] + ")."
		} else if strings.HasPrefix(s.LastUpdate, containerRecreatedPrefix) {
			update = " It was recreated from the same image in the last 7 days."
		}
	}
	if s.Restarts7d > 0 {
		return "pass", fmt.Sprintf("%s is running and healthy.%s None were crashes.%s", s.Name, restarts, update)
	}
	return "pass", fmt.Sprintf("%s is running with no restarts, crashes or health problems in retained history.%s", s.Name, update)
}

// containerChangeAnchors marks the changes worth matching listener
// changes against: restarts, updates, recreations and changes into a
// problem state. Recovering ("health: starting -> healthy") is not one.
func containerChangeAnchors(kind, summary string) bool {
	switch kind {
	case "restart", "update", "recreate", "disappeared":
		return true
	case "state":
		arrow := strings.LastIndex(summary, "-> ")
		if arrow < 0 {
			return false
		}
		to := summary[arrow+3:]
		return strings.Contains(to, "unhealthy") || strings.HasPrefix(to, "exited") || strings.HasPrefix(to, "restarting") || strings.HasPrefix(to, "dead")
	}
	return false
}

func publishedHostPorts(c ContainerInfo) map[string]bool {
	ports := map[string]bool{}
	for _, b := range dockerPortBindings([]ContainerInfo{c}) {
		if b.Published && b.HostPort != "" {
			ports[b.HostPort] = true
		}
	}
	return ports
}

func listenerEventPort(e Event) string {
	key, ok := listenerEventKey(e, listenerAppearedPrefix)
	if !ok {
		key, ok = listenerEventKey(e, listenerDisappearedPrefix)
	}
	if !ok {
		return ""
	}
	if idx := strings.LastIndex(key, ":"); idx >= 0 {
		return key[idx+1:]
	}
	return ""
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

func times(n int) string {
	switch n {
	case 1:
		return "once"
	case 2:
		return "twice"
	}
	return fmt.Sprintf("%d times", n)
}
