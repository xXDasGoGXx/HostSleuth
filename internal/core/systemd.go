package core

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	systemdFailureCandidateLimit = 3
	systemdJournalOutputLimit    = 8 * 1024
)

var (
	journalSensitiveAssignment = regexp.MustCompile(`(?i)\b(password|passwd|pwd|secret|token|api[_-]?key|authorization|cookie)(\s*[:=]\s*)(\S+)`)
	journalBearerToken         = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+`)
)

var journalUnitLookup = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("journalctl", "/usr/bin/journalctl", "/bin/journalctl")
	return runBoundedCommand(lookupCtx, systemdJournalOutputLimit, command, "--no-pager", "--quiet", "-b", "-u", unit, "-n", "6", "-o", "cat")
}

func systemdFailureCheck(ctx context.Context, snap Snapshot) Check {
	if snap.Mode == dockerDeploymentMode {
		return Check{
			Name:     "systemd-failures",
			Status:   "unknown",
			Evidence: "systemd runtime evidence is unavailable in Docker deployment mode",
		}
	}

	failed := failedServices(snap)
	if len(failed) == 0 {
		return Check{Name: "systemd-failures", Status: "pass", Evidence: "snapshot contains no failed systemd services"}
	}

	limit := len(failed)
	if limit > systemdFailureCandidateLimit {
		limit = systemdFailureCandidateLimit
	}
	parts := make([]string, 0, limit+1)
	for _, service := range failed[:limit] {
		detail := service.Name + " " + service.Active + "/" + service.Sub
		result, err := journalUnitLookup(ctx, service.Name)
		journal := sanitizeJournal(result.Output)
		if err != nil {
			switch {
			case errors.Is(err, exec.ErrNotFound):
				detail += "; journal unavailable: journalctl command is unavailable"
			case journal != "":
				detail += "; journal read unavailable: " + boundedEvidence(journal, 384)
			default:
				detail += "; journal read unavailable: " + boundedEvidence(err.Error(), 256)
			}
		} else if journal == "" {
			detail += "; recent journal: no current-boot lines"
		} else {
			detail += "; recent journal: " + boundedEvidence(journal, 384)
		}
		if result.Truncated {
			detail += " [journal capture truncated at 8 KiB]"
		}
		parts = append(parts, detail)
	}
	if len(failed) > limit {
		parts = append(parts, "+"+strconv.Itoa(len(failed)-limit)+" additional failed service(s) omitted")
	}

	return Check{
		Name:     "systemd-failures",
		Status:   "unknown",
		Evidence: boundedEvidence("failed systemd service candidates (not proven related to port): "+strings.Join(parts, " | "), 2048),
	}
}

func failedServices(snap Snapshot) []ServiceInfo {
	failed := make([]ServiceInfo, 0)
	for _, service := range snap.Services {
		if strings.EqualFold(service.Active, "failed") || strings.EqualFold(service.Sub, "failed") {
			failed = append(failed, service)
		}
	}
	sort.Slice(failed, func(i, j int) bool { return failed[i].Name < failed[j].Name })
	return failed
}

func sanitizeJournal(value string) string {
	value = journalBearerToken.ReplaceAllString(value, "Bearer [REDACTED]")
	value = journalSensitiveAssignment.ReplaceAllString(value, "$1$2[REDACTED]")
	return strings.TrimSpace(value)
}
