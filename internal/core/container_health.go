package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Why a container is unhealthy.
//
// Docker keeps the health check's command and its last few results (exit
// code and whatever the check printed). HostSleuth reads the command and the
// newest result through the field-limited inspect, masks anything that looks
// like a secret before keeping it, and explains the failure with fixed
// rules: what the output means, what the check requires when it prints
// nothing, what changed just before it started failing, and whether it has
// done the same before. Every suggestion says how sure it is.

const (
	healthCheckLimit  = 400
	healthOutputLimit = 300
	healthTriggerSpan = 10 * time.Minute
)

// HealthCheckResult is the newest health-check run Docker reports.
type HealthCheckResult struct {
	At       string `json:"at,omitempty"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output,omitempty"`
}

// HealthWhy explains a failing health check.
type HealthWhy struct {
	Summary    string   `json:"summary"`
	Check      string   `json:"check,omitempty"`
	Needs      []string `json:"needs,omitempty"`
	LastAt     int64    `json:"last_at,omitempty"`
	LastExit   *int     `json:"last_exit,omitempty"`
	LastOutput string   `json:"last_output,omitempty"`
	Clues      []string `json:"clues"`
}

var (
	healthURLUserRe    = regexp.MustCompile(`(?i)(\b[a-z][a-z0-9+.-]*://)[^/\s@"']+@`)
	healthFlagSecretRe = regexp.MustCompile(`(?i)(--?(?:password|passwd|pass|pwd|token|api[-_]?key|secret|auth)(?:=|\s+))(?:"[^"]*"|'[^']*'|\S+)`)
	healthUserFlagRe   = regexp.MustCompile(`((?:^|\s)(?:-u|--user)(?:=|\s+))(?:"[^"]*"|'[^']*'|\S+)`)
	healthHeaderRe     = regexp.MustCompile(`(?i)((?:-H|--header)(?:=|\s+)["']?(?:authorization|proxy-authorization|x-api-key|api-key|x-auth-token|cookie)\s*:\s*)[^"'\n]*`)
	healthShortPassRe  = regexp.MustCompile(`((?:^|\s)-p)[^\s\-"']\S*`)
	healthSpaceRe      = regexp.MustCompile(`\s+`)

	healthTimeoutRe  = regexp.MustCompile(`(?i)exceeded timeout`)
	healthRefusedRe  = regexp.MustCompile(`(?i)connection refused|failed to connect|couldn'?t connect|could not connect|econnrefused|no route to host`)
	healthDNSRe      = regexp.MustCompile(`(?i)could not resolve|name or service not known|temporary failure in name resolution|no such host|nxdomain`)
	healthSlowRe     = regexp.MustCompile(`(?i)timed out|timeout`)
	healthHTTPRe     = regexp.MustCompile(`(?i)(?:HTTP/[\d.]+\s+|returned error:\s*|status(?: code)?\s*[:=]?\s*)([45]\d\d)\b`)
	healthMissingRe  = regexp.MustCompile(`(?i)executable file not found|command not found|: not found|no such file or directory`)
	healthDeniedRe   = regexp.MustCompile(`(?i)permission denied`)
	healthTestFileRe = regexp.MustCompile(`^(?:test|\[)\s+(!\s+)?-([efds])\s+("[^"]+"|\S+)(?:\s+\])?$`)
	healthTestSameRe = regexp.MustCompile(`^(?:test|\[)\s+"\$\(cat\s+([^)\s]+)\)"\s+=\s+"\$\(cat\s+([^)\s]+)\)"(?:\s+\])?$`)
)

// maskHealthText removes secrets from a health-check command or output
// before HostSleuth keeps it, and makes it one bounded line.
func maskHealthText(value string, limit int) string {
	out := strings.TrimSpace(value)
	if out == "" {
		return ""
	}
	out = evidencePrivateKeyBlockPattern.ReplaceAllString(out, "[REDACTED]")
	out = evidencePrivateKeyRemainderPattern.ReplaceAllString(out, "[REDACTED]")
	out = evidencePrivateKeyHeaderPattern.ReplaceAllString(out, "[REDACTED]")
	out = healthURLUserRe.ReplaceAllString(out, "${1}[REDACTED]@")
	out = healthHeaderRe.ReplaceAllString(out, "${1}[REDACTED]")
	out = healthFlagSecretRe.ReplaceAllString(out, "${1}[REDACTED]")
	out = healthUserFlagRe.ReplaceAllString(out, "${1}[REDACTED]")
	out = healthShortPassRe.ReplaceAllString(out, "${1}[REDACTED]")
	out = evidenceSecretKVPattern.ReplaceAllString(out, "$1=[REDACTED]")
	out = evidenceAuthPattern.ReplaceAllString(out, "$1 [REDACTED]")
	out = strings.TrimSpace(healthSpaceRe.ReplaceAllString(out, " "))
	return clipText(out, limit)
}

// healthCheckCommand turns Docker's Healthcheck.Test (JSON array) into the
// command as written, masked.
func healthCheckCommand(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var test []string
	if err := json.Unmarshal([]byte(raw), &test); err != nil || len(test) == 0 {
		return ""
	}
	switch test[0] {
	case "NONE":
		return ""
	case "CMD", "CMD-SHELL":
		test = test[1:]
	}
	return maskHealthText(strings.Join(test, " "), healthCheckLimit)
}

// healthLastResult reads the newest of Docker's retained results, written by
// the inspect template as "exit|end|json-output" records separated by RS.
func healthLastResult(raw string) *HealthCheckResult {
	var last string
	for _, rec := range strings.Split(raw, "\x1e") {
		if strings.TrimSpace(rec) != "" {
			last = rec
		}
	}
	parts := strings.SplitN(strings.TrimSpace(last), "|", 3)
	if len(parts) != 3 {
		return nil
	}
	code := 0
	if _, err := fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &code); err != nil {
		return nil
	}
	var output string
	_ = json.Unmarshal([]byte(parts[2]), &output)
	return &HealthCheckResult{At: dockerTime(parts[1]), ExitCode: code, Output: maskHealthText(output, healthOutputLimit)}
}

// healthProblemFileRe matches file names programs write to report trouble.
var healthProblemFileRe = regexp.MustCompile(`(?i)fail|error|unclassified|alert|problem|broken|corrupt|stale|lock|warn`)

// healthNeeds reads a check made only of file tests joined by && and says
// in words what must be true for it to pass. ok is false for anything else.
// suspects are the conditions that fail when a program reports a problem
// by writing a file ("FAILED must not exist", "ERRORS.txt must be empty"),
// phrased as what is probably true now.
func healthNeeds(check string) (needs []string, compares [2]string, suspects []string, ok bool) {
	if strings.TrimSpace(check) == "" || strings.Contains(check, "||") {
		return nil, compares, nil, false
	}
	for _, part := range strings.Split(check, "&&") {
		part = strings.TrimSpace(part)
		if m := healthTestSameRe.FindStringSubmatch(part); m != nil {
			needs = append(needs, fmt.Sprintf("%s and %s hold the same value", m[1], m[2]))
			compares = [2]string{m[1], m[2]}
			continue
		}
		m := healthTestFileRe.FindStringSubmatch(part)
		if m == nil {
			return nil, compares, nil, false
		}
		not, flag, path := m[1] != "", m[2], strings.Trim(m[3], `"`)
		if not && healthProblemFileRe.MatchString(path[strings.LastIndex(path, "/")+1:]) {
			if flag == "s" {
				suspects = append(suspects, path+" is not empty")
			} else {
				suspects = append(suspects, path+" exists")
			}
		}
		switch {
		case flag == "s" && not:
			needs = append(needs, path+" is empty or missing")
		case flag == "s":
			needs = append(needs, path+" exists and is not empty")
		case flag == "d" && not:
			needs = append(needs, "folder "+path+" does not exist")
		case flag == "d":
			needs = append(needs, "folder "+path+" exists")
		case not:
			needs = append(needs, path+" does not exist")
		default:
			needs = append(needs, path+" exists")
		}
	}
	return needs, compares, suspects, len(needs) > 0
}

// healthOutputReason reads what the check printed.
func healthOutputReason(out string, exit int) string {
	switch {
	case healthTimeoutRe.MatchString(out):
		return "The check ran out of time: the program did not answer before Docker's time limit. That usually means it is overloaded or stuck."
	case healthRefusedRe.MatchString(out):
		return "Nothing answered where the check connects, so the program inside is not listening, has not finished starting, or listens on a different port."
	case healthDNSRe.MatchString(out):
		return "The check could not look up a name, so this is a name (DNS) problem inside the container."
	case exit == 127 || healthMissingRe.MatchString(out):
		return "The check's own command or a file it needs is missing inside the container. That often follows an image update that removed the tool the check uses."
	case exit == 126 || healthDeniedRe.MatchString(out):
		return "The check's command was not allowed to run (permission denied)."
	}
	if m := healthHTTPRe.FindStringSubmatch(out); m != nil {
		return fmt.Sprintf("The program answered with an error (HTTP %s), so it is running but reports a problem of its own.", m[1])
	}
	if healthSlowRe.MatchString(out) {
		return "The check's request timed out: the program accepted it but did not answer in time."
	}
	return ""
}

var healthVerb = map[string]string{
	"restart": "restarted", "clean": "restarted", "crash": "restarted after being stopped or killed",
	"oom": "was killed for running out of memory", "update": "was updated", "recreate": "was recreated",
	"recovered": "passed its health check again", "unhealthy": "started failing its own health check",
	"slow": "finally passed its health check", "appeared": "appeared", "stopped": "stopped",
}

// healthTrigger finds what changed shortly before a failure began: the
// container itself first, then another container in the same compose
// project, then the host.
func healthTrigger(name, stack string, began int64, events []ReplayEvent, stackOf map[string]string) (ReplayEvent, string, bool) {
	var own, peer, host *ReplayEvent
	for i := range events {
		e := &events[i]
		if e.T > began || began-e.T > healthTriggerSpan.Milliseconds() || (e.Container == name && e.Kind == "unhealthy") {
			continue
		}
		switch {
		case e.Container == name:
			own = e
		case e.Container != "" && stack != "" && stack != "standalone" && stackOf[e.Container] == stack:
			peer = e
		case e.Container == "" && (e.Kind == "reboot" || e.Kind == "docker" || e.Kind == "package" || e.Kind == "config"):
			host = e
		}
	}
	switch {
	case own != nil:
		return *own, "own", true
	case peer != nil:
		return *peer, "peer", true
	case host != nil:
		return *host, "host", true
	}
	return ReplayEvent{}, "", false
}

func healthGap(d int64) string {
	if d < 60_000 {
		return "less than a minute"
	}
	return replayDuration(time.Duration(d) * time.Millisecond)
}

func healthTriggerText(name string, e ReplayEvent, scope string, began int64) string {
	gap := healthGap(began - e.T)
	switch scope {
	case "own":
		return fmt.Sprintf("It started failing %s after it %s.", gap, healthVerb[e.Kind])
	case "peer":
		verb := healthVerb[e.Kind]
		if verb == "" {
			verb = "changed"
		}
		return fmt.Sprintf("It started failing %s after %s (same stack) %s.", gap, e.Container, verb)
	}
	what := map[string]string{"reboot": "the host rebooted", "docker": "Docker restarted every container", "package": "packages changed on the host", "config": "the host's configuration changed"}[e.Kind]
	return fmt.Sprintf("It started failing %s after %s.", gap, what)
}

// explainUnhealthy builds the explanation for a container that is unhealthy
// in the snapshot. events are the Replay's narrated events (oldest first).
func explainUnhealthy(c ContainerInfo, stack string, events []ReplayEvent, stackOf map[string]string) *HealthWhy {
	why := &HealthWhy{Check: c.HealthCheck, Clues: []string{}}
	if c.HealthLast != nil {
		code := c.HealthLast.ExitCode
		why.LastExit = &code
		why.LastOutput = c.HealthLast.Output
		if t, err := time.Parse(time.RFC3339Nano, c.HealthLast.At); err == nil {
			why.LastAt = ms(t)
		}
	}

	// Episodes: each "unhealthy" for this container and how it ended.
	type episode struct {
		began, ended int64
		trigger      string
	}
	var eps []episode
	for i, e := range events {
		if e.Container != c.Name || e.Kind != "unhealthy" {
			continue
		}
		ep := episode{began: e.T}
		for _, later := range events[i+1:] {
			if later.Container == c.Name && (later.Kind == "recovered" || later.Kind == "unhealthy") {
				if later.Kind == "recovered" {
					ep.ended = later.T
				}
				break
			}
		}
		if t, scope, ok := healthTrigger(c.Name, stack, e.T, events, stackOf); ok && scope == "peer" {
			ep.trigger = t.Container
		}
		eps = append(eps, ep)
	}

	var trigger ReplayEvent
	var scope string
	var current *episode
	if n := len(eps); n > 0 && eps[n-1].ended == 0 {
		current = &eps[n-1]
		var ok bool
		trigger, scope, ok = healthTrigger(c.Name, stack, current.began, events, stackOf)
		if ok {
			why.Clues = append(why.Clues, healthTriggerText(c.Name, trigger, scope, current.began))
		} else {
			why.Clues = append(why.Clues, "Nothing else HostSleuth records changed in the ten minutes before it started failing.")
		}
	} else {
		why.Clues = append(why.Clues, "It was already failing before the oldest change HostSleuth keeps for this week.")
	}

	// Has it done this before?
	pattern, took := "", ""
	cleared := 0
	if current != nil {
		earlier := eps[:len(eps)-1]
		same := 0
		var shortest, longest int64
		for _, ep := range earlier {
			if ep.ended > 0 {
				cleared++
				d := ep.ended - ep.began
				if shortest == 0 || d < shortest {
					shortest = d
				}
				if d > longest {
					longest = d
				}
			}
			if scope == "peer" && ep.trigger == trigger.Container {
				same++
			}
		}
		switch {
		case len(earlier) == 0:
			why.Clues = append(why.Clues, "It has not failed its health check before this week.")
		default:
			took = replayDuration(time.Duration(shortest) * time.Millisecond)
			if longest-shortest >= 60_000 {
				took += " to " + replayDuration(time.Duration(longest)*time.Millisecond)
			}
			line := fmt.Sprintf("It also failed %s earlier this week", times(len(earlier)))
			switch {
			case cleared == len(earlier) && cleared == 1:
				line += fmt.Sprintf(" and cleared on its own after %s.", took)
			case cleared == len(earlier):
				line += fmt.Sprintf(" and cleared on its own every time, after %s.", took)
			case cleared > 0:
				line += fmt.Sprintf("; %d of those cleared on their own, after %s.", cleared, took)
			default:
				line += "."
			}
			why.Clues = append(why.Clues, line)
			if same > 0 {
				lead := fmt.Sprintf("%d of those", same)
				if same == len(earlier) {
					lead = map[bool]string{true: "That one", false: "Each of those"}[same == 1]
				}
				why.Clues = append(why.Clues, fmt.Sprintf("%s also began right after %s changed.", lead, trigger.Container))
				if same == len(earlier) && cleared == len(earlier) {
					pattern = fmt.Sprintf(" It did the same %s earlier this week after %s changed, and cleared on its own each time.", times(same), trigger.Container)
				}
			}
		}
	}
	if c.HealthFailingStreak == 1 {
		why.Clues = append(why.Clues, "Docker's last check failed.")
	} else if c.HealthFailingStreak > 1 {
		why.Clues = append(why.Clues, fmt.Sprintf("Docker's last %d checks in a row failed.", c.HealthFailingStreak))
	}

	// The best single suggestion.
	exit := 0
	if why.LastExit != nil {
		exit = *why.LastExit
	}
	needs, compares, suspects, fileCheck := healthNeeds(c.HealthCheck)
	if fileCheck {
		why.Needs = needs
	}
	switch reason := healthOutputReason(why.LastOutput, exit); {
	case reason != "":
		why.Summary = reason
	case fileCheck && len(suspects) > 0:
		look := "that file"
		if len(suspects) > 1 {
			look = "those files"
		}
		why.Summary = fmt.Sprintf("Most likely %s. Its check prints nothing, so Docker records no reason, but it fails when the program reports a problem that way. Look in %s for what went wrong.", joinOr(suspects), look)
	case fileCheck && compares[0] != "" && scope == "peer":
		why.Summary = fmt.Sprintf("It is probably catching up with %s: its check compares %s with %s, and %s %s just before it started failing.", trigger.Container, compares[0], compares[1], trigger.Container, healthVerb[trigger.Kind])
	case fileCheck && compares[0] != "" && cleared > 0:
		why.Summary = fmt.Sprintf("It is probably waiting to catch up: its check passes only when %s and %s hold the same value, and each time this happened earlier this week it cleared on its own after %s.", compares[0], compares[1], took)
		pattern = ""
	case fileCheck && compares[0] != "":
		why.Summary = fmt.Sprintf("It is probably waiting to catch up: its check passes only when %s and %s hold the same value. That usually clears on its own once the program finishes its current job.", compares[0], compares[1])
	case fileCheck:
		why.Summary = "Its check only looks at files and prints nothing, so Docker records no reason. One of the conditions below is not true right now."
	case why.LastOutput != "":
		why.Summary = fmt.Sprintf("The check printed: “%s”.", why.LastOutput)
	case c.HealthCheck == "" && c.HealthLast == nil:
		why.Summary = "Docker did not report the check or its results, so HostSleuth cannot see why. The clues below are what changed around it."
	case scope != "":
		why.Summary = why.Clues[0]
	default:
		why.Summary = fmt.Sprintf("The check failed with exit code %d and printed nothing, so Docker records no reason.", exit)
	}
	why.Summary += pattern
	return why
}

// joinOr writes "a", "a or b", "a, b or c".
func joinOr(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " or " + items[len(items)-1]
}
