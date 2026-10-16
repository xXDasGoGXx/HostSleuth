package core

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"
)

// dockerInspectFormat selects only the runtime fields HostSleuth reports.
// It must never select .Config.Env, .Config.Cmd, .Mounts or .HostConfig:
// container configuration routinely carries passwords and tokens, and a
// field-limited template keeps them out of process memory and snapshots.
const dockerInspectFormat = `{{.Id}}{{"\t"}}{{.Image}}{{"\t"}}{{.RestartCount}}{{"\t"}}{{.State.OOMKilled}}{{"\t"}}{{.State.ExitCode}}{{"\t"}}{{.State.StartedAt}}{{"\t"}}{{.State.FinishedAt}}{{"\t"}}{{if .State.Health}}{{.State.Health.FailingStreak}}{{end}}{{"\t"}}{{with .Config.Labels}}{{index . "org.opencontainers.image.version"}}{{end}}`

// enrichContainers adds runtime evidence to containers listed by
// `docker ps`. One `docker inspect` call covers every container; if it
// fails, containers keep their `docker ps` fields only.
func enrichContainers(ctx context.Context, containers []ContainerInfo) {
	if len(containers) == 0 {
		return
	}
	args := []string{"inspect", "--format", dockerInspectFormat}
	for _, c := range containers {
		if id := strings.TrimSpace(c.ID); id != "" {
			args = append(args, id)
		}
	}
	if len(args) == 3 {
		return
	}
	applyDockerInspect(containers, run(ctx, "docker", args...))
	applyDockerExits(containers, collectDockerExitLog(ctx, time.Now()))
}

// containerExitWindow bounds how far back each snapshot asks Docker for
// container deaths. Diffing keeps only exits after the previous snapshot,
// so overlapping windows never record an exit twice. Docker itself keeps a
// bounded in-memory event log.
const containerExitWindow = 30 * time.Minute

// dockerExitFormat selects the event time, action, container name and exit
// code only; event attributes can include every container label.
const dockerExitFormat = `{{.TimeNano}}{{"\t"}}{{.Action}}{{"\t"}}{{.Actor.Attributes.name}}{{"\t"}}{{index .Actor.Attributes "exitCode"}}`

func collectDockerExitLog(ctx context.Context, now time.Time) string {
	return run(ctx, "docker", "events",
		"--since", strconv.FormatInt(now.Add(-containerExitWindow).Unix(), 10),
		"--until", strconv.FormatInt(now.Unix(), 10),
		"--filter", "type=container", "--filter", "event=die", "--filter", "event=oom",
		"--format", dockerExitFormat)
}

// applyDockerExits turns die/oom events into per-container exits. Docker
// emits "oom" just before the matching "die", so an oom within a few
// seconds before a die marks that exit as out-of-memory.
func applyDockerExits(containers []ContainerInfo, text string) {
	type raw struct {
		at       time.Time
		action   string
		name     string
		exitCode int
	}
	var events []raw
	for _, line := range strings.Split(text, "\n") {
		f := strings.Split(strings.TrimRight(line, "\r"), "\t")
		if len(f) < 4 {
			continue
		}
		nanos, err := strconv.ParseInt(strings.TrimSpace(f[0]), 10, 64)
		if err != nil || nanos <= 0 {
			continue
		}
		code, _ := strconv.Atoi(strings.TrimSpace(f[3]))
		events = append(events, raw{at: time.Unix(0, nanos).UTC(), action: strings.TrimSpace(f[1]), name: strings.TrimSpace(f[2]), exitCode: code})
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].at.Before(events[j].at) })

	byName := map[string]int{}
	for i := range containers {
		byName[containers[i].Name] = i
	}
	lastOOM := map[string]time.Time{}
	for _, e := range events {
		idx, ok := byName[e.name]
		if !ok {
			continue
		}
		switch e.action {
		case "oom":
			lastOOM[e.name] = e.at
		case "die":
			exit := ContainerExit{At: e.at, ExitCode: e.exitCode}
			if t, seen := lastOOM[e.name]; seen && !t.After(e.at) && e.at.Sub(t) <= 10*time.Second {
				exit.OOMKilled = true
				delete(lastOOM, e.name)
			}
			c := &containers[idx]
			c.RecentExits = append(c.RecentExits, exit)
			if len(c.RecentExits) > 20 {
				c.RecentExits = c.RecentExits[len(c.RecentExits)-20:]
			}
		}
	}
}

func applyDockerInspect(containers []ContainerInfo, text string) {
	for _, line := range strings.Split(text, "\n") {
		f := strings.Split(strings.TrimRight(line, "\r"), "\t")
		if len(f) < 9 || f[0] == "" {
			continue
		}
		for i := range containers {
			short := strings.TrimSpace(containers[i].ID)
			if short == "" || !strings.HasPrefix(f[0], short) {
				continue
			}
			c := &containers[i]
			c.ImageID = strings.TrimSpace(f[1])
			c.RestartCount, _ = strconv.Atoi(strings.TrimSpace(f[2]))
			c.OOMKilled = strings.TrimSpace(f[3]) == "true"
			c.ExitCode, _ = strconv.Atoi(strings.TrimSpace(f[4]))
			c.StartedAt = dockerTime(f[5])
			c.FinishedAt = dockerTime(f[6])
			c.HealthFailingStreak, _ = strconv.Atoi(strings.TrimSpace(f[7]))
			c.ImageVersion = boundedLabel(f[8])
			break
		}
	}
}

// dockerTime drops Docker's zero time ("0001-01-01T00:00:00Z"), which
// means "never".
func dockerTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "0001-01-01") {
		return ""
	}
	return value
}

func boundedLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "<no value>" {
		return ""
	}
	if len(value) > 80 {
		value = value[:80]
	}
	return value
}
