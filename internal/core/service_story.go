package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	serviceRuntimeOutputLimit = 8 * 1024
	serviceJournalOutputLimit = 16 * 1024
	serviceContextLimit       = 12
)

type ServiceRuntimeEvidence struct {
	LoadState      string `json:"load_state,omitempty"`
	ActiveState    string `json:"active_state,omitempty"`
	SubState       string `json:"sub_state,omitempty"`
	UnitFileState  string `json:"unit_file_state,omitempty"`
	Result         string `json:"result,omitempty"`
	MainPID        int    `json:"main_pid,omitempty"`
	PIDs           []int  `json:"pids,omitempty"`
	ControlGroup   string `json:"control_group,omitempty"`
	ExecMainCode   string `json:"exec_main_code,omitempty"`
	ExecMainStatus string `json:"exec_main_status,omitempty"`
}

type ServiceStory struct {
	Service       string                  `json:"service"`
	Target        string                  `json:"target,omitempty"`
	StartedAt     time.Time               `json:"started_at"`
	State         *ServiceInfo            `json:"state,omitempty"`
	Runtime       *ServiceRuntimeEvidence `json:"runtime,omitempty"`
	Listeners     []Listener              `json:"listeners,omitempty"`
	Containers    []ContainerInfo         `json:"containers,omitempty"`
	Checks        []Check                 `json:"checks"`
	RelatedEvents []Event                 `json:"related_events,omitempty"`
	Endpoint      *Diagnosis              `json:"endpoint,omitempty"`
	Conclusion    string                  `json:"conclusion"`
	Confidence    string                  `json:"confidence"`
}

var listenerPIDPattern = regexp.MustCompile(`pid=(\d+)`)

var serviceRuntimeLookup = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("systemctl", "/usr/bin/systemctl", "/bin/systemctl")
	return runBoundedCommand(lookupCtx, serviceRuntimeOutputLimit, command, "show", unit, "--no-pager",
		"--property=LoadState,ActiveState,SubState,UnitFileState,Result,MainPID,ControlGroup,ExecMainCode,ExecMainStatus")
}

var serviceJournalLookup = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("journalctl", "/usr/bin/journalctl", "/bin/journalctl")
	return runBoundedCommand(lookupCtx, serviceJournalOutputLimit, command, "--no-pager", "--quiet", "-b", "-u", unit, "-n", "20", "-o", "short-iso")
}

var serviceCgroupPIDs = func(controlGroup string) ([]int, error) {
	controlGroup = strings.TrimSpace(controlGroup)
	if controlGroup == "" || controlGroup == "/" {
		return nil, errors.New("service control group is unavailable")
	}
	clean := filepath.Clean("/" + strings.TrimPrefix(controlGroup, "/"))
	path := filepath.Join("/sys/fs/cgroup", clean, "cgroup.procs")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) > 64*1024 {
		b = b[:64*1024]
	}
	var out []int
	for _, field := range strings.Fields(string(b)) {
		pid, err := strconv.Atoi(field)
		if err == nil && pid > 0 {
			out = append(out, pid)
		}
	}
	return out, nil
}

func ServiceStoryFor(ctx context.Context, service, target string, snap Snapshot, events []Event) ServiceStory {
	unit := normalizeServiceUnit(service)
	story := ServiceStory{Service: unit, Target: strings.TrimSpace(target), StartedAt: time.Now().UTC(), Confidence: "medium"}
	if unit == "" {
		story.Checks = append(story.Checks, Check{Name: "service-unit", Status: "fail", Evidence: "service unit is required"})
		story.Conclusion = "service unit is required"
		story.Confidence = "high"
		return story
	}
	if snap.Mode == dockerDeploymentMode {
		story.Checks = append(story.Checks, Check{Name: "service-runtime", Status: "unknown", Evidence: "systemd service evidence is unavailable in Docker deployment mode"})
		story.Conclusion = "service story is unavailable in Docker deployment mode"
		story.Confidence = "high"
		return story
	}

	if state, ok := snapshotService(snap, unit); ok {
		copy := state
		story.State = &copy
	}

	runtime, runtimeErr := collectServiceRuntime(ctx, unit)
	story.Runtime = runtime
	story.Checks = append(story.Checks, serviceRuntimeCheck(story.State, runtime, runtimeErr))
	story.Checks = append(story.Checks, serviceJournalCheck(ctx, unit))

	pids := servicePIDSet(runtime)
	story.Listeners = listenersOwnedByPIDs(snap.Listeners, pids)
	if story.Target == "" {
		story.Checks = append(story.Checks, serviceListenerInventoryCheck(story.Listeners, len(pids) > 0))
		story.RelatedEvents = serviceContextEvents(events, unit, "")
		story.Conclusion, story.Confidence = serviceStoryOutcome(story, "", false, false)
		return story
	}

	_, port, err := net.SplitHostPort(story.Target)
	if err != nil {
		story.Checks = append(story.Checks, Check{Name: "service-target", Status: "fail", Evidence: "expected endpoint must be host:port"})
		story.RelatedEvents = serviceContextEvents(events, unit, "")
		story.Conclusion = "invalid expected endpoint"
		story.Confidence = "high"
		return story
	}
	if _, err := strconv.Atoi(port); err != nil {
		story.Checks = append(story.Checks, Check{Name: "service-target", Status: "fail", Evidence: "expected endpoint port must be numeric"})
		story.Conclusion = "invalid expected endpoint"
		story.Confidence = "high"
		return story
	}

	portListeners := listenersForTCPPort(snap.Listeners, port)
	ownedOnPort := listenersOwnedByPIDs(portListeners, pids)
	occupiedByOther := len(pids) > 0 && len(portListeners) > 0 && len(ownedOnPort) == 0
	missingExpected := len(portListeners) == 0
	story.Checks = append(story.Checks, expectedPortOwnershipCheck(port, portListeners, ownedOnPort, len(pids) > 0))
	story.Containers = containersForHostPort(snap.Containers, port)
	if len(story.Containers) > 0 {
		story.Checks = append(story.Checks, Check{Name: "service-container", Status: "unknown", Evidence: boundedEvidence("container runtime also exposes TCP/"+port+": "+formatServiceContainers(story.Containers), 1024)})
	}

	diagnosis := Diagnose(ctx, story.Target, snap)
	story.Endpoint = &diagnosis
	story.Checks = append(story.Checks, endpointSummaryCheck(diagnosis))
	story.RelatedEvents = serviceContextEvents(events, unit, port)
	story.Conclusion, story.Confidence = serviceStoryOutcome(story, port, occupiedByOther, missingExpected)
	return story
}

func normalizeServiceUnit(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.Contains(value, ".") {
		value += ".service"
	}
	return value
}

func snapshotService(snap Snapshot, unit string) (ServiceInfo, bool) {
	for _, service := range snap.Services {
		if service.Name == unit {
			return service, true
		}
	}
	return ServiceInfo{}, false
}

func collectServiceRuntime(ctx context.Context, unit string) (*ServiceRuntimeEvidence, error) {
	result, err := serviceRuntimeLookup(ctx, unit)
	if errors.Is(err, exec.ErrNotFound) {
		return nil, err
	}
	values := map[string]string{}
	for _, line := range strings.Split(result.Output, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	if len(values) == 0 {
		return nil, err
	}
	runtime := &ServiceRuntimeEvidence{
		LoadState:      values["LoadState"],
		ActiveState:    values["ActiveState"],
		SubState:       values["SubState"],
		UnitFileState:  values["UnitFileState"],
		Result:         values["Result"],
		ControlGroup:   values["ControlGroup"],
		ExecMainCode:   values["ExecMainCode"],
		ExecMainStatus: values["ExecMainStatus"],
	}
	runtime.MainPID, _ = strconv.Atoi(values["MainPID"])
	if pids, readErr := serviceCgroupPIDs(runtime.ControlGroup); readErr == nil {
		runtime.PIDs = pids
	}
	return runtime, err
}

func serviceRuntimeCheck(state *ServiceInfo, runtime *ServiceRuntimeEvidence, err error) Check {
	if runtime == nil {
		if errors.Is(err, exec.ErrNotFound) {
			return Check{Name: "service-runtime", Status: "unknown", Evidence: "systemctl command is unavailable"}
		}
		if err != nil {
			return Check{Name: "service-runtime", Status: "unknown", Evidence: "systemd runtime lookup unavailable: " + boundedEvidence(err.Error(), 384)}
		}
		return Check{Name: "service-runtime", Status: "unknown", Evidence: "systemd returned no runtime properties"}
	}
	active, sub := runtime.ActiveState, runtime.SubState
	if active == "" && state != nil {
		active, sub = state.Active, state.Sub
	}
	evidence := fmt.Sprintf("load=%s active=%s sub=%s enabled=%s result=%s main_pid=%d exec=%s/%s",
		fallback(runtime.LoadState, "unknown"), fallback(active, "unknown"), fallback(sub, "unknown"), fallback(runtime.UnitFileState, "unknown"), fallback(runtime.Result, "unknown"), runtime.MainPID, fallback(runtime.ExecMainCode, "unknown"), fallback(runtime.ExecMainStatus, "unknown"))
	if runtime.LoadState == "not-found" {
		return Check{Name: "service-runtime", Status: "fail", Evidence: evidence}
	}
	if strings.EqualFold(active, "failed") || strings.EqualFold(sub, "failed") {
		return Check{Name: "service-runtime", Status: "fail", Evidence: evidence}
	}
	if strings.EqualFold(active, "active") {
		return Check{Name: "service-runtime", Status: "pass", Evidence: evidence}
	}
	return Check{Name: "service-runtime", Status: "fail", Evidence: evidence}
}

func serviceJournalCheck(ctx context.Context, unit string) Check {
	result, err := serviceJournalLookup(ctx, unit)
	journal := sanitizeJournal(result.Output)
	if errors.Is(err, exec.ErrNotFound) {
		return Check{Name: "service-journal", Status: "unknown", Evidence: "journalctl command is unavailable"}
	}
	if err != nil && journal == "" {
		return Check{Name: "service-journal", Status: "unknown", Evidence: "journal read unavailable: " + boundedEvidence(err.Error(), 384)}
	}
	if journal == "" {
		return Check{Name: "service-journal", Status: "unknown", Evidence: "no current-boot journal lines were returned for the service"}
	}
	evidence := boundedEvidence(journal, 4096)
	if result.Truncated {
		evidence += " [journal capture truncated at 16 KiB]"
	}
	return Check{Name: "service-journal", Status: "unknown", Evidence: evidence}
}

func servicePIDSet(runtime *ServiceRuntimeEvidence) map[int]bool {
	out := map[int]bool{}
	if runtime == nil {
		return out
	}
	if runtime.MainPID > 0 {
		out[runtime.MainPID] = true
	}
	for _, pid := range runtime.PIDs {
		if pid > 0 {
			out[pid] = true
		}
	}
	return out
}

func listenerPIDs(listener Listener) []int {
	matches := listenerPIDPattern.FindAllStringSubmatch(listener.Process, -1)
	var out []int
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		pid, err := strconv.Atoi(match[1])
		if err == nil {
			out = append(out, pid)
		}
	}
	return out
}

func listenersOwnedByPIDs(listeners []Listener, pids map[int]bool) []Listener {
	var out []Listener
	for _, listener := range listeners {
		owned := false
		for _, pid := range listenerPIDs(listener) {
			if pids[pid] {
				owned = true
				break
			}
		}
		if owned {
			out = append(out, listener)
		}
	}
	return out
}

func listenersForTCPPort(listeners []Listener, port string) []Listener {
	var out []Listener
	for _, listener := range listeners {
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(listener.Protocol)), "tcp") {
			continue
		}
		_, candidatePort, err := net.SplitHostPort(strings.TrimSpace(listener.Address))
		if err == nil && candidatePort == port {
			out = append(out, listener)
		}
	}
	return out
}

func serviceListenerInventoryCheck(listeners []Listener, havePIDs bool) Check {
	if !havePIDs {
		return Check{Name: "service-listeners", Status: "unknown", Evidence: "service process membership is unavailable, so listener ownership cannot be proven"}
	}
	if len(listeners) == 0 {
		return Check{Name: "service-listeners", Status: "unknown", Evidence: "no current listening socket is owned by a process in the service control group; the service may not be network-facing"}
	}
	return Check{Name: "service-listeners", Status: "pass", Evidence: boundedEvidence("service-owned listeners: "+formatListeners(listeners), 1536)}
}

func expectedPortOwnershipCheck(port string, all, owned []Listener, havePIDs bool) Check {
	if len(all) == 0 {
		return Check{Name: "service-port", Status: "fail", Evidence: "no current TCP listener exists on expected port " + port}
	}
	if !havePIDs {
		return Check{Name: "service-port", Status: "unknown", Evidence: boundedEvidence("TCP/"+port+" is listening, but service process membership is unavailable: "+formatListeners(all), 1536)}
	}
	if len(owned) > 0 {
		return Check{Name: "service-port", Status: "pass", Evidence: boundedEvidence("expected port is owned by the service: "+formatListeners(owned), 1536)}
	}
	return Check{Name: "service-port", Status: "fail", Evidence: boundedEvidence("expected TCP/"+port+" is occupied by another process: "+formatListeners(all), 1536)}
}

func containersForHostPort(containers []ContainerInfo, port string) []ContainerInfo {
	seen := map[string]bool{}
	var out []ContainerInfo
	for _, binding := range dockerPortBindings(containers) {
		if binding.Protocol != "tcp" || !binding.Published || binding.HostPort != port || seen[binding.Container] {
			continue
		}
		for _, container := range containers {
			if container.Name == binding.Container {
				out = append(out, container)
				seen[binding.Container] = true
				break
			}
		}
	}
	return out
}

func endpointSummaryCheck(d Diagnosis) Check {
	status := "unknown"
	if d.Conclusion == "target is reachable" {
		status = "pass"
	} else if d.Confidence == "high" {
		status = "fail"
	}
	return Check{Name: "service-endpoint", Status: status, Evidence: d.Conclusion}
}

func serviceContextEvents(events []Event, unit, port string) []Event {
	var direct []Event
	for _, event := range events {
		summary := strings.ToLower(event.Summary)
		if strings.Contains(summary, strings.ToLower(unit)) || (port != "" && event.Category == "listener" && strings.Contains(summary, ":"+port)) {
			direct = append(direct, event)
		}
	}
	var anchor time.Time
	for _, event := range direct {
		if event.At.After(anchor) {
			anchor = event.At
		}
	}
	out := append([]Event(nil), direct...)
	if !anchor.IsZero() {
		window := 15 * time.Minute
		for _, event := range events {
			if event.Category != "package" && event.Category != "configuration" && event.Category != "container" {
				continue
			}
			delta := event.At.Sub(anchor)
			if delta < 0 {
				delta = -delta
			}
			if delta <= window && !containsEvent(out, event) {
				out = append(out, event)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	if len(out) > serviceContextLimit {
		out = out[:serviceContextLimit]
	}
	return out
}

func containsEvent(events []Event, candidate Event) bool {
	for _, event := range events {
		if event.At.Equal(candidate.At) && event.Category == candidate.Category && event.Summary == candidate.Summary {
			return true
		}
	}
	return false
}

func serviceStoryOutcome(story ServiceStory, port string, occupiedByOther, missingExpected bool) (string, string) {
	active, sub, load := "", "", ""
	if story.Runtime != nil {
		active, sub, load = story.Runtime.ActiveState, story.Runtime.SubState, story.Runtime.LoadState
	} else if story.State != nil {
		active, sub, load = story.State.Active, story.State.Sub, story.State.Load
	}
	if load == "not-found" {
		return "systemd unit was not found", "high"
	}
	if strings.EqualFold(active, "failed") || strings.EqualFold(sub, "failed") {
		if occupiedByOther {
			return "service is failed and the expected port is occupied by another process", "high"
		}
		return "service is failed", "high"
	}
	if active != "" && !strings.EqualFold(active, "active") {
		if occupiedByOther {
			return "service is not active and the expected port is occupied by another process", "high"
		}
		return "service is not active", "high"
	}
	if port != "" && occupiedByOther {
		return "service is active, but the expected port is owned by another process", "high"
	}
	if port != "" && missingExpected {
		return "service is active, but no process is listening on the expected port", "high"
	}
	if story.Endpoint != nil {
		if story.Endpoint.Conclusion == "target is reachable" {
			return "service is active and the expected endpoint is reachable", "high"
		}
		return "service is active and owns the expected local evidence, but endpoint diagnosis still reports a problem", "medium"
	}
	if strings.EqualFold(active, "active") {
		return "service is active", "high"
	}
	return "service state is inconclusive", "medium"
}

func formatListeners(listeners []Listener) string {
	parts := make([]string, 0, len(listeners))
	for _, listener := range listeners {
		parts = append(parts, strings.TrimSpace(listener.Protocol+" "+listener.Address+" "+listener.Process))
	}
	return strings.Join(parts, " | ")
}

func formatServiceContainers(containers []ContainerInfo) string {
	parts := make([]string, 0, len(containers))
	for _, container := range containers {
		parts = append(parts, container.Name+" ("+stableContainerState(container.Status)+") "+container.Ports)
	}
	return strings.Join(parts, " | ")
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return strings.TrimSpace(value)
}
