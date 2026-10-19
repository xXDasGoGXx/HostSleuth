package core

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Replay turns the snapshot and the retained timeline into a drawing of the
// host's containers and a narrated week that the Web UI can rewind.
//
// Every narrated event comes from a recorded event by a fixed rule. Where a
// rule infers something (several containers restarting together "fits a
// Docker restart"), the text says so. Outage windows are short, visual
// approximations anchored at the recorded time; the recorded time itself is
// always shown.

const (
	replayWindow    = 7 * 24 * time.Hour
	replayMaxEvents = 400
	replayBlip      = 10 * time.Second
)

type Replay struct {
	Now      int64                         `json:"now"`
	Start    int64                         `json:"start"`
	Mode     string                        `json:"mode,omitempty"`
	Host     ReplayHost                    `json:"host"`
	Stacks   []ReplayStack                 `json:"stacks"`
	Parts    []ReplayPart                  `json:"parts"`
	Trans    map[string][]ReplayTransition `json:"trans"`
	Events   []ReplayEvent                 `json:"events"`
	Versions map[string]ReplayVersion      `json:"versions"`
	Notes    []string                      `json:"notes"`
}

type ReplayHost struct {
	Name   string `json:"name"`
	OS     string `json:"os"`
	Kernel string `json:"kernel"`
}

type ReplayStack struct {
	ID    string   `json:"id"`
	Parts []string `json:"parts"`
}

type ReplayFlag struct {
	Net   string `json:"net"`
	Label string `json:"label"`
}

type ReplayPart struct {
	ID            string       `json:"id"`
	Stack         string       `json:"stack"`
	Base          string       `json:"base"`
	Version       string       `json:"ver"`
	Status        string       `json:"status"`
	Streak        int          `json:"streak,omitempty"`
	HealthChecked bool         `json:"hc"`
	Flags         []ReplayFlag `json:"flags"`
	// Why is set for a container that is unhealthy in the snapshot.
	Why *HealthWhy `json:"why,omitempty"`
}

// ReplayTransition is [time, from, to] on the wire, matching the UI.
type ReplayTransition [3]any

type ReplayEvent struct {
	N         int     `json:"n"`
	T         int64   `json:"t"`
	Kind      string  `json:"k"`
	Container string  `json:"c,omitempty"`
	Text      string  `json:"text"`
	Short     string  `json:"short"`
	Fault     bool    `json:"fault"`
	Down      []int64 `json:"down,omitempty"`
}

type ReplayVersion struct {
	T    int64  `json:"t"`
	From string `json:"from"`
	To   string `json:"to"`
}

var (
	replayChangedRe = regexp.MustCompile(`^container changed: (.+?): (.+) -> (.+)$`)
	replayPortRe    = regexp.MustCompile(`^\s*(\[[^\]]*\]|[0-9.]+):(\d+)->`)
	replayVersionRe = regexp.MustCompile(`version (\S+) \x{2192} (\S+)$`)
	replayRefRe     = regexp.MustCompile(`:(\S+) \x{2192} \S+:(\S+)$`)
	replayExitRe    = regexp.MustCompile(`\((\d+)\)`)
)

func ms(t time.Time) int64 { return t.UnixMilli() }

// BuildReplay builds the Replay for the snapshot's moment. Events must be
// oldest first and should include detail (ReadDetailedEvents).
func BuildReplay(snap Snapshot, events []Event) Replay {
	now := snap.CapturedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	start := now.Add(-replayWindow)
	r := Replay{
		Now: ms(now), Start: ms(start), Mode: snap.Mode,
		Host:     ReplayHost{Name: snap.Host.Hostname, OS: snap.Host.OS, Kernel: snap.Host.Kernel},
		Stacks:   []ReplayStack{},
		Parts:    []ReplayPart{},
		Trans:    map[string][]ReplayTransition{},
		Events:   []ReplayEvent{},
		Versions: map[string]ReplayVersion{},
		Notes:    []string{},
	}
	if len(snap.Containers) == 0 {
		r.Notes = append(r.Notes, "This snapshot has no Docker containers, so there is nothing to draw yet. Replay draws containers and the host events around them.")
	}

	known := map[string]ContainerInfo{}
	for _, c := range snap.Containers {
		known[c.Name] = c
	}

	// Health transitions per container, inside the window.
	type trans struct {
		at       time.Time
		from, to string
	}
	byName := map[string][]trans{}
	var window []Event
	for _, e := range events {
		if e.At.Before(start) || e.At.After(now.Add(time.Minute)) {
			continue
		}
		window = append(window, e)
		if m := replayChangedRe.FindStringSubmatch(e.Summary); m != nil && e.Category == "container" {
			byName[m[1]] = append(byName[m[1]], trans{e.At, m[2], m[3]})
		}
	}
	for name, tl := range byName {
		if _, ok := known[name]; !ok {
			continue
		}
		for _, t := range tl {
			r.Trans[name] = append(r.Trans[name], ReplayTransition{ms(t.at), t.from, t.to})
		}
	}

	// Parts and stacks.
	stacks := map[string][]string{}
	for _, c := range snap.Containers {
		stack := strings.TrimSpace(c.ComposeProject)
		if stack == "" {
			stack = "standalone"
		}
		base, ver := replayImage(c)
		hc := len(byName[c.Name]) > 0 || strings.Contains(c.Status, "healthy)") || strings.Contains(c.Status, "health:")
		r.Parts = append(r.Parts, ReplayPart{
			ID: c.Name, Stack: stack, Base: base, Version: ver, Status: c.Status,
			Streak: c.HealthFailingStreak, HealthChecked: hc, Flags: replayFlags(c),
		})
		stacks[stack] = append(stacks[stack], c.Name)
	}
	sort.SliceStable(r.Parts, func(i, j int) bool {
		if r.Parts[i].Stack != r.Parts[j].Stack {
			return r.Parts[i].Stack < r.Parts[j].Stack
		}
		return r.Parts[i].ID < r.Parts[j].ID
	})
	for id, parts := range stacks {
		sort.Strings(parts)
		r.Stacks = append(r.Stacks, ReplayStack{ID: id, Parts: parts})
	}
	sort.Slice(r.Stacks, func(i, j int) bool {
		if len(r.Stacks[i].Parts) != len(r.Stacks[j].Parts) {
			return len(r.Stacks[i].Parts) > len(r.Stacks[j].Parts)
		}
		return r.Stacks[i].ID < r.Stacks[j].ID
	})

	add := func(at time.Time, kind, container, text, short string, fault bool, down []int64) {
		r.Events = append(r.Events, ReplayEvent{T: ms(at), Kind: kind, Container: container, Text: text, Short: short, Fault: fault, Down: down})
	}
	blip := func(at time.Time) []int64 { return []int64{ms(at), ms(at.Add(replayBlip))} }

	// Host events: reboots (with any kernel change), packages, configuration.
	var reboots []time.Time
	kernelAt := map[int64]string{}
	for _, e := range window {
		if e.Category == "system" && strings.HasPrefix(e.Summary, "kernel changed: ") {
			kernelAt[e.At.Unix()/60] = strings.Replace(strings.TrimPrefix(e.Summary, "kernel changed: "), " -> ", " → ", 1)
		}
	}
	packages := map[int64][]Event{}
	for _, e := range window {
		switch {
		case e.Category == "system" && strings.HasPrefix(e.Summary, "host reboot detected"):
			boot := e.At
			if idx := strings.LastIndex(e.Summary, "started "); idx >= 0 {
				if t, err := time.Parse(time.RFC3339, strings.TrimSpace(e.Summary[idx+8:])); err == nil && !t.After(e.At) {
					boot = t
				}
			}
			reboots = append(reboots, boot, e.At)
			text := "The host rebooted. Every container stopped and came back."
			if k, ok := kernelAt[e.At.Unix()/60]; ok {
				text = fmt.Sprintf("The host rebooted (kernel %s). Every container stopped and came back.", k)
			}
			end := e.At
			if end.Sub(boot) < replayBlip {
				end = boot.Add(replayBlip)
			}
			add(boot, "reboot", "", text, "host  reboot", true, []int64{ms(boot), ms(end)})
		case e.Category == "package":
			packages[e.At.Unix()/60] = append(packages[e.At.Unix()/60], e)
		case e.Category == "configuration":
			add(e.At, "config", "", cap1(e.Summary)+" on the host.", "host  "+strings.TrimPrefix(strings.TrimPrefix(e.Summary, "configuration "), "changed: "), false, nil)
		}
	}
	for _, group := range packages {
		names := make([]string, 0, len(group))
		for _, e := range group {
			names = append(names, e.Summary)
		}
		text := names[0] + "."
		short := "host  " + clipText(group[0].Summary, 26)
		if len(group) > 1 {
			text = fmt.Sprintf("%d package changes, including %s.", len(group), strings.Join(names[:min(3, len(names))], "; "))
			short = fmt.Sprintf("host  %d package changes", len(group))
		}
		add(group[0].At, "package", "", cap1(text), short, false, nil)
	}
	nearReboot := func(t time.Time) bool {
		for _, b := range reboots {
			if d := t.Sub(b); d > -2*time.Minute && d < 2*time.Minute {
				return true
			}
		}
		return false
	}

	// Several health-checked containers restarting in the same minute.
	healthChecked := 0
	for _, p := range r.Parts {
		if p.HealthChecked {
			healthChecked++
		}
	}
	massMinutes := map[int64]bool{}
	minuteSet := map[int64]map[string]time.Time{}
	for name, tl := range byName {
		for _, t := range tl {
			if t.to == "running (health: starting)" && !strings.HasPrefix(t.from, "restarting") {
				m := t.at.Unix() / 60
				if minuteSet[m] == nil {
					minuteSet[m] = map[string]time.Time{}
				}
				minuteSet[m][name] = t.at
			}
		}
	}
	for m, set := range minuteSet {
		if len(set) < 4 || float64(len(set)) < 0.6*float64(healthChecked) {
			continue
		}
		massMinutes[m] = true
		first := time.Time{}
		for _, at := range set {
			if first.IsZero() || at.Before(first) {
				first = at
			}
		}
		if nearReboot(first) {
			continue
		}
		who := "All"
		if len(set) < healthChecked {
			who = fmt.Sprintf("%d of the %d", len(set), healthChecked)
		}
		add(first, "docker", "", fmt.Sprintf("%s containers with a health check restarted at the same moment. That pattern fits a Docker restart.", replayCount(who, len(set))), fmt.Sprintf("all  %d containers restarted", len(set)), false, blip(first))
	}

	// Per-container health rules.
	autoheal := false
	for _, c := range snap.Containers {
		if strings.Contains(strings.ToLower(c.Name+" "+c.Image), "autoheal") {
			autoheal = true
		}
	}
	for name, tl := range byName {
		if _, ok := known[name]; !ok {
			continue
		}
		var startSince time.Time
		for i, t := range tl {
			grouped := massMinutes[t.at.Unix()/60] || nearReboot(t.at)
			switch {
			case t.to == "running (health: starting)" && !strings.HasPrefix(t.from, "restarting"):
				startSince = t.at
				if !grouped {
					add(t.at, "restart", name, name+" restarted.", name+"  restarted", false, blip(t.at))
				}
			case t.to == "running (unhealthy)":
				text := name + " started failing its health check."
				down := []int64(nil)
				if started, err := time.Parse(time.RFC3339Nano, known[name].StartedAt); err == nil {
					if gap := started.Sub(t.at); gap > -time.Minute && gap < 2*time.Minute {
						secs := int(gap.Round(time.Second) / time.Second)
						if secs < 1 {
							secs = 1
						}
						text += fmt.Sprintf(" Docker shows it started again %d seconds later.", secs)
						if autoheal {
							text += " An autoheal container runs on this host; it restarts unhealthy containers."
						}
						down = []int64{ms(t.at.Add(3 * time.Second)), ms(t.at.Add(13 * time.Second))}
					}
				}
				recovered := false
				for _, later := range tl[i+1:] {
					if later.from == "running (unhealthy)" {
						recovered = true
						break
					}
				}
				if !recovered && down == nil {
					text += " It has not passed since."
				}
				add(t.at, "unhealthy", name, text, name+"  unhealthy", true, down)
			case t.from == "running (unhealthy)" && t.to == "running (healthy)":
				began := t.at
				for j := i; j >= 0; j-- {
					if tl[j].to == "running (unhealthy)" {
						began = tl[j].at
						break
					}
				}
				add(t.at, "recovered", name, fmt.Sprintf("%s passed its health check again after %s.", name, replayDuration(t.at.Sub(began))), name+"  healthy again", false, nil)
			case strings.HasPrefix(t.to, "restarting"):
				code := "?"
				if m := replayExitRe.FindStringSubmatch(t.to); m != nil {
					code = m[1]
				}
				add(t.at, "crash", name, fmt.Sprintf("%s crashed with exit code %s and Docker restarted it.", name, code), fmt.Sprintf("%s  crash, exit %s", name, code), true, []int64{ms(t.at), ms(t.at.Add(time.Minute))})
			case strings.HasPrefix(t.to, "exited") || strings.HasPrefix(t.to, "dead"):
				add(t.at, "stopped", name, fmt.Sprintf("%s stopped (%s).", name, t.to), name+"  stopped", true, nil)
			case t.from == "running (health: starting)" && t.to == "running (healthy)" && !startSince.IsZero() && t.at.Sub(startSince) > 30*time.Minute:
				d := replayDuration(t.at.Sub(startSince))
				add(t.at, "slow", name, fmt.Sprintf("%s passed its health check for the first time since it restarted %s earlier.", name, d), fmt.Sprintf("%s  healthy after %s", name, d), false, nil)
				startSince = time.Time{}
			case t.to == "running (healthy)":
				startSince = time.Time{}
			}
		}
	}

	// Lifecycle events recorded by HostSleuth 1.3+.
	for _, e := range window {
		if e.Category != "container" {
			continue
		}
		s := e.Summary
		switch {
		case strings.HasPrefix(s, containerUpdatedPrefix):
			rest := strings.TrimPrefix(s, containerUpdatedPrefix)
			name, detail, _ := strings.Cut(rest, ": ")
			from, to := "", ""
			if m := replayVersionRe.FindStringSubmatch(detail); m != nil {
				from, to = m[1], m[2]
			} else if m := replayRefRe.FindStringSubmatch(detail); m != nil {
				from, to = m[1], m[2]
			}
			text, short := name+" updated: "+detail+".", name+"  updated"
			if from != "" {
				text = fmt.Sprintf("%s updated %s → %s.", name, from, to)
				short = fmt.Sprintf("%s  %s→%s", name, clipText(from, 12), clipText(to, 12))
				if _, ok := known[name]; ok {
					r.Versions[name] = ReplayVersion{T: ms(e.At), From: from, To: to}
				}
			}
			add(e.At, "update", name, text, short, false, blip(e.At))
		case strings.HasPrefix(s, containerRecreatedPrefix):
			name, _, _ := strings.Cut(strings.TrimPrefix(s, containerRecreatedPrefix), " (")
			add(e.At, "recreate", name, name+" was recreated from the same image.", name+"  recreated", false, blip(e.At))
		case strings.HasPrefix(s, containerRestartedPrefix):
			name, how, _ := strings.Cut(strings.TrimPrefix(s, containerRestartedPrefix), " (")
			how = strings.TrimSuffix(how, ")")
			kind, fault := "clean", false
			switch {
			case strings.Contains(how, "out of memory"):
				kind, fault = "oom", true
			case strings.Contains(how, "crashed") || strings.Contains(how, "SIGKILL"):
				kind, fault = "crash", true
			case strings.Contains(how, "not reported"):
				kind = "restart"
			}
			short := name + "  restart"
			if m := regexp.MustCompile(`exit code (\d+)`).FindStringSubmatch(how); m != nil {
				short += ", exit " + m[1]
			} else if kind == "oom" {
				short = name + "  out of memory"
			}
			add(e.At, kind, name, fmt.Sprintf("%s restarted (%s).", name, how), short, fault, blip(e.At))
		case strings.HasPrefix(s, "container appeared: "):
			name, _, _ := strings.Cut(strings.TrimPrefix(s, "container appeared: "), " (")
			add(e.At, "appeared", name, name+" appeared on the host.", name+"  appeared", false, nil)
		case strings.HasPrefix(s, "container disappeared: "):
			name, _, _ := strings.Cut(strings.TrimPrefix(s, "container disappeared: "), " (")
			add(e.At, "disappeared", name, name+" disappeared from the host.", name+"  disappeared", true, nil)
		}
	}

	sort.SliceStable(r.Events, func(i, j int) bool { return r.Events[i].T < r.Events[j].T })
	if len(r.Events) > replayMaxEvents {
		r.Events = r.Events[len(r.Events)-replayMaxEvents:]
		r.Notes = append(r.Notes, fmt.Sprintf("Showing the newest %d changes of the week.", replayMaxEvents))
	}
	for i := range r.Events {
		r.Events[i].N = i + 1
	}

	// Why each currently unhealthy container is failing.
	stackOf := map[string]string{}
	for _, p := range r.Parts {
		stackOf[p.ID] = p.Stack
	}
	for i := range r.Parts {
		p := &r.Parts[i]
		if c, ok := known[p.ID]; ok && containerHealthFromStatus(c.Status) == "unhealthy" {
			p.Why = explainUnhealthy(c, p.Stack, r.Events, stackOf)
		}
	}
	return r
}

// replayFlags reduces Docker's port list to bind classes; addresses are
// never sent, only all / LAN (one specific address) / local.
func replayFlags(c ContainerInfo) []ReplayFlag {
	nets := splitList(c.Networks)
	for _, n := range nets {
		if n == "host" {
			return []ReplayFlag{{Net: "host", Label: "host network"}}
		}
	}
	if len(nets) == 1 && nets[0] == "none" {
		return []ReplayFlag{{Net: "none", Label: "no network"}}
	}
	groups := map[string]map[int]bool{}
	for _, part := range strings.Split(c.Ports, ",") {
		m := replayPortRe.FindStringSubmatch(part)
		if m == nil {
			continue
		}
		port, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		bind := strings.Trim(m[1], "[]")
		class := "lan"
		switch {
		case bind == "0.0.0.0" || bind == "::":
			class = "all"
		case strings.HasPrefix(bind, "127.") || bind == "::1":
			class = "local"
		}
		if groups[class] == nil {
			groups[class] = map[int]bool{}
		}
		groups[class][port] = true
	}
	label := map[string]string{"all": "all", "lan": "LAN", "local": "local"}
	out := []ReplayFlag{}
	for _, class := range []string{"all", "lan", "local"} {
		if len(groups[class]) == 0 {
			continue
		}
		ports := make([]int, 0, len(groups[class]))
		for p := range groups[class] {
			ports = append(ports, p)
		}
		sort.Ints(ports)
		parts := []string{label[class]}
		for _, p := range ports {
			parts = append(parts, ":"+strconv.Itoa(p))
		}
		out = append(out, ReplayFlag{Net: class, Label: strings.Join(parts, " ")})
	}
	return out
}

func replayImage(c ContainerInfo) (string, string) {
	img := c.Image
	if at := strings.Index(img, "@"); at >= 0 {
		img = img[:at]
	}
	if slash := strings.LastIndex(img, "/"); slash >= 0 {
		img = img[slash+1:]
	}
	base, tag, _ := strings.Cut(img, ":")
	ver := strings.TrimSpace(c.ImageVersion)
	if ver == "" {
		ver = tag
	}
	return base, ver
}

func replayDuration(d time.Duration) string {
	m := int(d.Round(time.Minute) / time.Minute)
	switch {
	case m <= 1:
		return "about a minute"
	case m < 90:
		return fmt.Sprintf("%d minutes", m)
	}
	return fmt.Sprintf("%d hours", int(d.Round(time.Hour)/time.Hour))
}

func replayCount(who string, n int) string {
	if who == "All" {
		return fmt.Sprintf("All %d", n)
	}
	return who
}

func cap1(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func clipText(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
