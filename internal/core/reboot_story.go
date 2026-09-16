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

type PreviousBootEvidence struct {
	Status     string `json:"status"`
	Evidence   string `json:"evidence,omitempty"`
	Assessment string `json:"assessment"`
}

type RebootStory struct {
	CapturedAt       time.Time             `json:"captured_at"`
	Mode             string                `json:"mode,omitempty"`
	BootID           string                `json:"boot_id,omitempty"`
	BootStartedAt    time.Time             `json:"boot_started_at,omitempty"`
	WindowStart      time.Time             `json:"window_start,omitempty"`
	WindowEnd        time.Time             `json:"window_end,omitempty"`
	PreviousBoot     PreviousBootEvidence  `json:"previous_boot"`
	FailedServices   []ServiceInfo         `json:"failed_services,omitempty"`
	ProblemEvents    []Event               `json:"problem_events,omitempty"`
	RelatedEvents    []Event               `json:"related_events,omitempty"`
	Checks           []Check               `json:"checks"`
	Conclusion       string                `json:"conclusion"`
	Confidence       string                `json:"confidence"`
	CauseAssessment  string                `json:"cause_assessment"`
	ContextNote      string                `json:"context_note"`
}

var previousBootJournalLookup = func(ctx context.Context) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("journalctl", "/usr/bin/journalctl", "/bin/journalctl")
	return runBoundedCommand(lookupCtx, previousBootJournalByteLimit, command, "-b", "-1", "-n", "24", "--no-pager", "--quiet", "-o", "short-iso")
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
		ContextNote:     "Boot-time proximity and retained change events are context, not proof of reboot cause. Missing retained evidence does not prove that every service recovered.",
	}
	if !bootStartedAt.IsZero() {
		story.WindowStart = bootStartedAt.Add(-rebootContextWindow)
		story.WindowEnd = bootStartedAt.Add(rebootContextWindow)
		story.RelatedEvents = EventsAround(events, bootStartedAt, rebootContextWindow, rebootContextLimit)
		story.ProblemEvents = postBootProblemEvents(story.RelatedEvents, bootStartedAt)
	}
	story.FailedServices = currentFailedServices(snap.Services)
	story.PreviousBoot = collectPreviousBootEvidence(ctx)
	story.Checks = rebootStoryChecks(story)
	story.Conclusion, story.Confidence = rebootStoryOutcome(story)
	return story
}

func collectPreviousBootEvidence(ctx context.Context) PreviousBootEvidence {
	result, err := previousBootJournalLookup(ctx)
	journal := sanitizeJournal(result.Output)
	assessment := "shutdown classification is unknown; HostSleuth does not infer orderly or abnormal shutdown from missing or bounded journal evidence"
	if errors.Is(err, exec.ErrNotFound) {
		return PreviousBootEvidence{Status: "unknown", Assessment: assessment, Evidence: "journalctl command is unavailable"}
	}
	if err != nil && journal == "" {
		return PreviousBootEvidence{Status: "unknown", Assessment: assessment, Evidence: "previous-boot journal is unavailable: " + boundedEvidence(err.Error(), 384)}
	}
	if journal == "" {
		return PreviousBootEvidence{Status: "unknown", Assessment: assessment, Evidence: "no readable previous-boot journal lines were returned"}
	}
	evidence := boundedEvidence(journal, 4096)
	if result.Truncated {
		evidence += " [previous-boot journal capture truncated at 16 KiB]"
	}
	return PreviousBootEvidence{Status: "available", Assessment: assessment, Evidence: evidence}
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
	if story.PreviousBoot.Status == "available" {
		previousStatus = "pass"
	}
	checks = append(checks, Check{Name: "previous-boot-journal", Status: previousStatus, Evidence: boundedEvidence(story.PreviousBoot.Evidence, 1536)})

	if story.Mode == dockerDeploymentMode {
		checks = append(checks, Check{Name: "post-boot-services", Status: "unknown", Evidence: "native systemd service inventory is unavailable in Docker deployment mode"})
	} else if len(story.FailedServices) > 0 {
		checks = append(checks, Check{Name: "post-boot-services", Status: "fail", Evidence: fmt.Sprintf("%d current systemd service(s) are failed", len(story.FailedServices))})
	} else {
		checks = append(checks, Check{Name: "post-boot-services", Status: "pass", Evidence: "no currently failed systemd services are present in the snapshot"})
	}

	if story.BootStartedAt.IsZero() {
		checks = append(checks, Check{Name: "post-boot-changes", Status: "unknown", Evidence: "boot start time is unavailable, so retained changes cannot be anchored to this boot"})
	} else if len(story.ProblemEvents) > 0 {
		checks = append(checks, Check{Name: "post-boot-changes", Status: "fail", Evidence: fmt.Sprintf("%d warning service/listener/container event(s) were retained after boot inside the bounded window", len(story.ProblemEvents))})
	} else {
		checks = append(checks, Check{Name: "post-boot-changes", Status: "pass", Evidence: "no warning service/listener/container events were retained after boot inside the bounded window"})
	}
	return checks
}

func rebootStoryOutcome(story RebootStory) (string, string) {
	if story.BootStartedAt.IsZero() {
		return "current boot start time is unavailable", "medium"
	}
	prefix := "current boot started " + story.BootStartedAt.Format(time.RFC3339)
	if len(story.FailedServices) > 0 || len(story.ProblemEvents) > 0 {
		return prefix + "; post-boot service/listener/container problems are visible in current or retained evidence", "high"
	}
	return prefix + "; no retained post-boot service/listener/container problem evidence was found", "medium"
}
