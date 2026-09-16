package core

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	rebootContextWindow          = 15 * time.Minute
	rebootContextLimit           = 100
	previousBootJournalByteLimit = 16 * 1024
)

type BootJournalEvidence struct {
	Status     string `json:"status"`
	Evidence   string `json:"evidence,omitempty"`
	Assessment string `json:"assessment"`
}

type RecoveryIssue struct {
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	CurrentState string `json:"current_state"`
	Evidence     string `json:"evidence"`
}

type RebootStory struct {
	CapturedAt      time.Time           `json:"captured_at"`
	Mode            string              `json:"mode,omitempty"`
	BootID          string              `json:"boot_id,omitempty"`
	BootStartedAt   time.Time           `json:"boot_started_at,omitempty"`
	WindowStart     time.Time           `json:"window_start,omitempty"`
	WindowEnd       time.Time           `json:"window_end,omitempty"`
	PreviousBoot    BootJournalEvidence `json:"previous_boot"`
	CurrentBoot     BootJournalEvidence `json:"current_boot"`
	FailedServices  []ServiceInfo       `json:"failed_services,omitempty"`
	RecoveryIssues  []RecoveryIssue     `json:"recovery_issues,omitempty"`
	ProblemEvents   []Event             `json:"problem_events,omitempty"`
	ContextEvents   []Event             `json:"context_events,omitempty"`
	RelatedEvents   []Event             `json:"related_events,omitempty"`
	Checks          []Check             `json:"checks"`
	Conclusion      string              `json:"conclusion"`
	Confidence      string              `json:"confidence"`
	CauseAssessment string              `json:"cause_assessment"`
	ContextNote     string              `json:"context_note"`
}

var previousBootJournalLookup = func(ctx context.Context) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("journalctl", "/usr/bin/journalctl", "/bin/journalctl")
	return runBoundedCommand(lookupCtx, previousBootJournalByteLimit, command, "-b", "-1", "-n", "48", "--no-pager", "--quiet", "-o", "short-iso")
}

var currentBootJournalLookup = func(ctx context.Context) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("journalctl", "/usr/bin/journalctl", "/bin/journalctl")
	return runBoundedCommand(lookupCtx, previousBootJournalByteLimit, command, "-b", "0", "-n", "48", "--no-pager", "--quiet", "-o", "short-iso")
}

func BuildRebootStory(ctx context.Context, snap Snapshot, events []Event) RebootStory {
	bootID := strings.TrimSpace(snap.Host.BootID)
	bootStartedAt := snap.Host.BootStartedAt
	if bootID == "" || bootStartedAt.IsZero() {
		liveID, liveStartedAt := collectBootIdentity()
		if bootID == "" {
			bootID = liveID
		}
		if bootStartedAt.IsZero() {
			bootStartedAt = liveStartedAt
		}
	}

	story := RebootStory{
		CapturedAt:      time.Now().UTC(),
		Mode:            snap.Mode,
		BootID:          bootID,
		BootStartedAt:   bootStartedAt,
		CauseAssessment: "no reboot cause is claimed from the available evidence",
		ContextNote:     "Boot-time proximity and retained changes are context, not proof of reboot cause. Recovery issues are shown only when retained post-boot evidence and the current snapshot agree that a problem remains.",
	}
	if !bootStartedAt.IsZero() {
		story.WindowStart = bootStartedAt.Add(-rebootContextWindow)
		story.WindowEnd = bootStartedAt.Add(rebootContextWindow)
		story.RelatedEvents = EventsAround(events, bootStartedAt, rebootContextWindow, rebootContextLimit)
		story.ProblemEvents = postBootProblemEvents(story.RelatedEvents, bootStartedAt)
		story.ContextEvents = rebootContextEvents(story.RelatedEvents)
		story.RecoveryIssues = unresolvedRecoveryIssues(snap, story.ProblemEvents)
	}
	story.FailedServices = currentFailedServices(snap.Services)
	story.PreviousBoot = collectPreviousBootEvidence(ctx)
	story.CurrentBoot = collectCurrentBootEvidence(ctx)
	story.Checks = rebootStoryChecks(story)
	story.Conclusion, story.Confidence = rebootStoryOutcome(story)
	return story
}

func collectPreviousBootEvidence(ctx context.Context) BootJournalEvidence {
	result, err := previousBootJournalLookup(ctx)
	journal := sanitizeJournal(result.Output)
	baseAssessment := "shutdown classification is unknown; HostSleuth does not infer orderly or abnormal shutdown from missing or bounded journal evidence"
	if errors.Is(err, exec.ErrNotFound) {
		return BootJournalEvidence{Status: "unknown", Assessment: baseAssessment, Evidence: "journalctl command is unavailable"}
	}
	if err != nil && journal == "" {
		return BootJournalEvidence{Status: "unknown", Assessment: baseAssessment, Evidence: "previous-boot journal is unavailable: " + boundedEvidence(err.Error(), 384)}
	}
	if journal == "" {
		return BootJournalEvidence{Status: "unknown", Assessment: baseAssessment, Evidence: "no readable previous-boot journal lines were returned"}
	}
	evidence := boundedEvidence(journal, 4096)
	if result.Truncated {
		evidence += " [previous-boot journal capture truncated at 16 KiB]"
	}
	lower := strings.ToLower(journal)
	switch {
	case strings.Contains(lower, "kernel panic - not syncing"):
		return BootJournalEvidence{Status: "abnormal", Assessment: "previous boot contains direct kernel panic evidence; this supports an abnormal termination classification", Evidence: evidence}
	case strings.Contains(lower, "systemd-shutdown") &&
		(strings.Contains(lower, "reached target power-off") || strings.Contains(lower, "reached target reboot") || strings.Contains(lower, "reached target shutdown")):
		return BootJournalEvidence{Status: "orderly", Assessment: "previous boot contains direct systemd shutdown target evidence supporting an orderly shutdown/reboot sequence", Evidence: evidence}
	default:
		return BootJournalEvidence{Status: "available", Assessment: baseAssessment, Evidence: evidence}
	}
}

func collectCurrentBootEvidence(ctx context.Context) BootJournalEvidence {
	result, err := currentBootJournalLookup(ctx)
	journal := sanitizeJournal(result.Output)
	assessment := "current-boot journal is bounded context only and is not used to invent a reboot cause"
	if errors.Is(err, exec.ErrNotFound) {
		return BootJournalEvidence{Status: "unknown", Assessment: assessment, Evidence: "journalctl command is unavailable"}
	}
	if err != nil && journal == "" {
		return BootJournalEvidence{Status: "unknown", Assessment: assessment, Evidence: "current-boot journal is unavailable: " + boundedEvidence(err.Error(), 384)}
	}
	if journal == "" {
		return BootJournalEvidence{Status: "unknown", Assessment: assessment, Evidence: "no readable current-boot journal lines were returned"}
	}
	evidence := boundedEvidence(journal, 4096)
	if result.Truncated {
		evidence += " [current-boot journal capture truncated at 16 KiB]"
	}
	return BootJournalEvidence{Status: "available", Assessment: assessment, Evidence: evidence}
}

func currentFailedServices(services []ServiceInfo) []ServiceInfo {
	var out []ServiceInfo
	for _, service := range services {
		if strings.EqualFold(service.Active, "failed") || strings.EqualFold(service.Sub, "failed") {
			out = append(out, service)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	if len(out) > 16 {
		out = out[:16]
	}
	return out
}

func postBootProblemEvents(events []Event, bootStartedAt time.Time) []Event {
	if bootStartedAt.IsZero() {
		return nil
	}
	var out []Event
	for _, event := range events {
		if event.At.Before(bootStartedAt) || !strings.EqualFold(event.Severity, "warning") {
			continue
		}
		switch event.Category {
		case "service", "listener", "container":
			out = append(out, event)
		}
	}
	if len(out) > 32 {
		out = out[:32]
	}
	return out
}

func rebootContextEvents(events []Event) []Event {
	var out []Event
	for _, event := range events {
		switch event.Category {
		case "package", "configuration", "system":
			out = append(out, event)
		}
	}
	if len(out) > 32 {
		out = out[:32]
	}
	return out
}

func unresolvedRecoveryIssues(snap Snapshot, events []Event) []RecoveryIssue {
	currentServices := serviceMap(snap.Services)
	currentContainers := containerMap(snap.Containers)
	currentListeners := listenerSet(snap.Listeners)
	var out []RecoveryIssue
	seen := map[string]bool{}

	add := func(issue RecoveryIssue) {
		key := issue.Kind + "\x00" + issue.Name
		if issue.Name == "" || seen[key] {
			return
		}
		seen[key] = true
		out = append(out, issue)
	}

	for _, event := range events {
		switch event.Category {
		case "listener":
			const prefix = "listener disappeared: "
			if !strings.HasPrefix(event.Summary, prefix) {
				continue
			}
			name := strings.TrimSpace(strings.TrimPrefix(event.Summary, prefix))
			if !currentListeners[name] {
				add(RecoveryIssue{Kind: "listener", Name: name, CurrentState: "not present in current snapshot", Evidence: event.Summary})
			}
		case "service":
			name := namedStateEventName(event.Summary, "service")
			state, present := currentServices[name]
			if name == "" {
				continue
			}
			if !present {
				add(RecoveryIssue{Kind: "service", Name: name, CurrentState: "not present in current snapshot", Evidence: event.Summary})
			} else if problematicStableState(state) {
				add(RecoveryIssue{Kind: "service", Name: name, CurrentState: state, Evidence: event.Summary})
			}
		case "container":
			name := namedStateEventName(event.Summary, "container")
			state, present := currentContainers[name]
			if name == "" {
				continue
			}
			if !present {
				add(RecoveryIssue{Kind: "container", Name: name, CurrentState: "not present in current snapshot", Evidence: event.Summary})
			} else if problematicStableState(state) {
				add(RecoveryIssue{Kind: "container", Name: name, CurrentState: state, Evidence: event.Summary})
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind == out[j].Kind {
			return out[i].Name < out[j].Name
		}
		return out[i].Kind < out[j].Kind
	})
	if len(out) > 32 {
		out = out[:32]
	}
	return out
}

func namedStateEventName(summary, category string) string {
	disappeared := category + " disappeared: "
	if strings.HasPrefix(summary, disappeared) {
		rest := strings.TrimSpace(strings.TrimPrefix(summary, disappeared))
		if idx := strings.Index(rest, " (was "); idx >= 0 {
			return strings.TrimSpace(rest[:idx])
		}
		return rest
	}
	changed := category + " changed: "
	if strings.HasPrefix(summary, changed) {
		rest := strings.TrimSpace(strings.TrimPrefix(summary, changed))
		if idx := strings.Index(rest, ": "); idx >= 0 {
			return strings.TrimSpace(rest[:idx])
		}
	}
	return ""
}

func problematicStableState(state string) bool {
	lower := strings.ToLower(strings.TrimSpace(state))
	for _, marker := range []string{"failed", "inactive", "dead", "exited", "restarting", "paused", "removing", "created"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func rebootStoryChecks(story RebootStory) []Check {
	var checks []Check
	if story.BootID == "" && story.BootStartedAt.IsZero() {
		checks = append(checks, Check{Name: "boot-identity", Status: "unknown", Evidence: "kernel boot identity and start time are unavailable"})
	} else {
		evidence := fmt.Sprintf("boot_id=%s", fallback(story.BootID, "unknown"))
		if !story.BootStartedAt.IsZero() {
			evidence += " started_at=" + story.BootStartedAt.Format(time.RFC3339)
		}
		checks = append(checks, Check{Name: "boot-identity", Status: "pass", Evidence: evidence})
	}

	previousStatus := "unknown"
	if story.PreviousBoot.Status == "available" || story.PreviousBoot.Status == "orderly" || story.PreviousBoot.Status == "abnormal" {
		previousStatus = "pass"
	}
	checks = append(checks, Check{Name: "previous-boot-journal", Status: previousStatus, Evidence: boundedEvidence(story.PreviousBoot.Assessment+"; "+story.PreviousBoot.Evidence, 1536)})

	currentStatus := "unknown"
	if story.CurrentBoot.Status == "available" {
		currentStatus = "pass"
	}
	checks = append(checks, Check{Name: "current-boot-journal", Status: currentStatus, Evidence: boundedEvidence(story.CurrentBoot.Assessment+"; "+story.CurrentBoot.Evidence, 1536)})

	if story.Mode == dockerDeploymentMode {
		checks = append(checks, Check{Name: "post-boot-services", Status: "unknown", Evidence: "native systemd service inventory is unavailable in Docker deployment mode"})
	} else if len(story.FailedServices) > 0 {
		checks = append(checks, Check{Name: "post-boot-services", Status: "fail", Evidence: fmt.Sprintf("%d current systemd service(s) are failed", len(story.FailedServices))})
	} else {
		checks = append(checks, Check{Name: "post-boot-services", Status: "pass", Evidence: "no currently failed systemd services are present in the snapshot"})
	}

	if story.BootStartedAt.IsZero() {
		checks = append(checks, Check{Name: "post-boot-recovery", Status: "unknown", Evidence: "boot start time is unavailable, so retained changes cannot be anchored to this boot"})
	} else if len(story.RecoveryIssues) > 0 {
		checks = append(checks, Check{Name: "post-boot-recovery", Status: "fail", Evidence: fmt.Sprintf("%d service/listener/container recovery issue(s) are supported by both retained post-boot evidence and current snapshot state", len(story.RecoveryIssues))})
	} else if len(story.ProblemEvents) > 0 {
		checks = append(checks, Check{Name: "post-boot-recovery", Status: "pass", Evidence: "warning service/listener/container events occurred after boot, but the current snapshot does not confirm those problems remain"})
	} else {
		checks = append(checks, Check{Name: "post-boot-recovery", Status: "pass", Evidence: "no warning service/listener/container events were retained after boot inside the bounded window"})
	}
	return checks
}

func rebootStoryOutcome(story RebootStory) (string, string) {
	if story.BootStartedAt.IsZero() {
		return "current boot start time is unavailable", "medium"
	}
	prefix := "current boot started " + story.BootStartedAt.Format(time.RFC3339)
	if len(story.RecoveryIssues) > 0 || len(story.FailedServices) > 0 {
		return prefix + "; current recovery problems are supported by retained and/or current host evidence", "high"
	}
	if len(story.ProblemEvents) > 0 {
		return prefix + "; post-boot problems were retained, but the current snapshot does not confirm they remain", "medium"
	}
	return prefix + "; no retained post-boot service/listener/container problem evidence was found", "medium"
}
