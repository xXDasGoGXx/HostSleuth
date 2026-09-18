package core

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	actionSchemaVersion      = 1
	actionServiceRestartID   = "service.restart"
	actionServiceReloadID    = "service.reload"
	actionCommandOutputLimit = 8 * 1024
	actionExecutionTimeout   = 10 * time.Second
)

var safeActionServiceUnitPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.@:-]*\.service$`)

type ActionPolicy struct {
	Enabled                bool            `json:"enabled"`
	AllowedRestartServices map[string]bool `json:"allowed_restart_services,omitempty"`
	AllowedReloadServices  map[string]bool `json:"allowed_reload_services,omitempty"`
}

type ActionCapability struct {
	ActionID       string   `json:"action_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Enabled        bool     `json:"enabled"`
	Available      bool     `json:"available"`
	Reason         string   `json:"reason,omitempty"`
	AllowedTargets []string `json:"allowed_targets,omitempty"`
}

type ActionPreview struct {
	SchemaVersion int                     `json:"schema_version"`
	ActionID      string                  `json:"action_id"`
	Target        string                  `json:"target"`
	Enabled       bool                    `json:"enabled"`
	Allowed       bool                    `json:"allowed"`
	Available     bool                    `json:"available"`
	Effect        string                  `json:"effect,omitempty"`
	Command       []string                `json:"command,omitempty"`
	Confirmation  string                  `json:"confirmation,omitempty"`
	Before        *ServiceRuntimeEvidence `json:"before,omitempty"`
	Checks        []Check                 `json:"checks"`
}

type ActionResult struct {
	SchemaVersion int                     `json:"schema_version"`
	ActionID      string                  `json:"action_id"`
	Target        string                  `json:"target"`
	StartedAt     time.Time               `json:"started_at"`
	FinishedAt    time.Time               `json:"finished_at"`
	Status        string                  `json:"status"`
	Summary       string                  `json:"summary"`
	Before        *ServiceRuntimeEvidence `json:"before,omitempty"`
	After         *ServiceRuntimeEvidence `json:"after,omitempty"`
	CommandOutput string                  `json:"command_output,omitempty"`
}

type ActionAudit struct {
	SchemaVersion int                     `json:"schema_version"`
	At            time.Time               `json:"at"`
	Phase         string                  `json:"phase"`
	ActionID      string                  `json:"action_id"`
	Target        string                  `json:"target"`
	Status        string                  `json:"status"`
	Summary       string                  `json:"summary"`
	Before        *ServiceRuntimeEvidence `json:"before,omitempty"`
	After         *ServiceRuntimeEvidence `json:"after,omitempty"`
}

type ActionManager struct {
	Policy ActionPolicy
	Store  Store
	mu     sync.Mutex
}

var actionServiceRestart = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	command, err := actionSystemctlPath()
	if err != nil {
		return boundedCommandResult{}, err
	}
	actionCtx, cancel := context.WithTimeout(ctx, actionExecutionTimeout)
	defer cancel()
	return runBoundedCommand(actionCtx, actionCommandOutputLimit, command, "restart", unit)
}

var actionServiceReload = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	command, err := actionSystemctlPath()
	if err != nil {
		return boundedCommandResult{}, err
	}
	actionCtx, cancel := context.WithTimeout(ctx, actionExecutionTimeout)
	defer cancel()
	return runBoundedCommand(actionCtx, actionCommandOutputLimit, command, "reload", unit)
}

func NewActionPolicy(enabled bool, allowedRestartServices []string) (ActionPolicy, error) {
	return NewActionPolicyWithReload(enabled, allowedRestartServices, nil)
}

func NewActionPolicyWithReload(enabled bool, allowedRestartServices, allowedReloadServices []string) (ActionPolicy, error) {
	policy := ActionPolicy{
		Enabled:                enabled,
		AllowedRestartServices: map[string]bool{},
		AllowedReloadServices:  map[string]bool{},
	}
	for _, raw := range allowedRestartServices {
		unit := normalizeServiceUnit(raw)
		if !validActionServiceUnit(unit) {
			return ActionPolicy{}, fmt.Errorf("invalid restart-allowlisted systemd service %q", raw)
		}
		policy.AllowedRestartServices[unit] = true
	}
	for _, raw := range allowedReloadServices {
		unit := normalizeServiceUnit(raw)
		if !validActionServiceUnit(unit) {
			return ActionPolicy{}, fmt.Errorf("invalid reload-allowlisted systemd service %q", raw)
		}
		policy.AllowedReloadServices[unit] = true
	}
	return policy, nil
}

func validActionServiceUnit(unit string) bool {
	return safeActionServiceUnitPattern.MatchString(strings.TrimSpace(unit))
}

func actionAllowedTargets(values map[string]bool) []string {
	targets := make([]string, 0, len(values))
	for target := range values {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return targets
}

func actionCapability(enabled bool, mode, actionID, title, description, emptyReason string, targets []string) ActionCapability {
	capability := ActionCapability{
		ActionID:       actionID,
		Title:          title,
		Description:    description,
		Enabled:        enabled,
		Available:      enabled && strings.TrimSpace(mode) != dockerDeploymentMode && len(targets) > 0,
		AllowedTargets: targets,
	}
	switch {
	case !enabled:
		capability.Reason = "actions are disabled; start HostSleuth with --enable-actions to opt in"
	case strings.TrimSpace(mode) == dockerDeploymentMode:
		capability.Reason = "systemd service actions are unavailable in Docker deployment mode"
	case len(targets) == 0:
		capability.Reason = emptyReason
	}
	return capability
}

func (m *ActionManager) Capabilities(mode string) []ActionCapability {
	restartTargets := actionAllowedTargets(m.Policy.AllowedRestartServices)
	reloadTargets := actionAllowedTargets(m.Policy.AllowedReloadServices)
	return []ActionCapability{
		actionCapability(
			m.Policy.Enabled,
			mode,
			actionServiceRestartID,
			"Restart allowlisted service",
			"Restart one explicitly allowlisted systemd service and verify that it returns active.",
			"no restart services are allowlisted",
			restartTargets,
		),
		actionCapability(
			m.Policy.Enabled,
			mode,
			actionServiceReloadID,
			"Reload allowlisted service",
			"Reload one explicitly allowlisted, active, reload-capable systemd service without restart fallback.",
			"no reload services are allowlisted",
			reloadTargets,
		),
	}
}

func (m *ActionManager) Preview(ctx context.Context, actionID, target, mode string) ActionPreview {
	preview := ActionPreview{
		SchemaVersion: actionSchemaVersion,
		ActionID:      strings.TrimSpace(actionID),
		Target:        normalizeServiceUnit(target),
		Enabled:       m.Policy.Enabled,
	}
	if preview.ActionID != actionServiceRestartID && preview.ActionID != actionServiceReloadID {
		preview.Checks = append(preview.Checks, Check{Name: "action-id", Status: "fail", Evidence: "unknown action ID"})
		return preview
	}
	if !m.Policy.Enabled {
		preview.Checks = append(preview.Checks, Check{Name: "action-policy", Status: "fail", Evidence: "actions are disabled by default; explicit enablement is required"})
		return preview
	}
	if strings.TrimSpace(mode) == dockerDeploymentMode {
		preview.Checks = append(preview.Checks, Check{Name: "action-mode", Status: "fail", Evidence: "systemd service actions are unavailable in Docker deployment mode"})
		return preview
	}
	if !validActionServiceUnit(preview.Target) {
		preview.Checks = append(preview.Checks, Check{Name: "action-target", Status: "fail", Evidence: "target is not a valid conservative systemd .service unit name"})
		return preview
	}

	var (
		allowed          bool
		policyEvidence   string
		commandVerb      string
		confirmationVerb string
		effect           string
	)
	switch preview.ActionID {
	case actionServiceRestartID:
		allowed = m.Policy.AllowedRestartServices[preview.Target]
		policyEvidence = "target is explicitly allowlisted for service.restart"
		commandVerb = "restart"
		confirmationVerb = "RESTART"
		effect = "Restart only the allowlisted systemd unit " + preview.Target + " and verify ActiveState=active afterward."
	case actionServiceReloadID:
		allowed = m.Policy.AllowedReloadServices[preview.Target]
		policyEvidence = "target is explicitly allowlisted for service.reload"
		commandVerb = "reload"
		confirmationVerb = "RELOAD"
		effect = "Reload only the allowlisted systemd unit " + preview.Target + " and verify ActiveState=active afterward; never fall back to restart."
	}
	if !allowed {
		preview.Checks = append(preview.Checks, Check{Name: "action-policy", Status: "fail", Evidence: "target is not explicitly allowlisted for " + preview.ActionID})
		return preview
	}

	command, commandErr := actionSystemctlPath()
	if commandErr != nil {
		preview.Checks = append(preview.Checks, Check{Name: "action-command", Status: "fail", Evidence: boundedEvidence(commandErr.Error(), 384)})
		return preview
	}
	preview.Allowed = true
	preview.Checks = append(preview.Checks, Check{Name: "action-policy", Status: "pass", Evidence: policyEvidence})

	before, err := collectActionServiceRuntime(ctx, preview.Target)
	preview.Before = before
	if before == nil {
		evidence := "systemd runtime evidence is unavailable"
		if errors.Is(err, exec.ErrNotFound) {
			evidence = "systemctl is unavailable"
		} else if err != nil {
			evidence += ": " + boundedEvidence(err.Error(), 384)
		}
		preview.Checks = append(preview.Checks, Check{Name: "action-precondition", Status: "fail", Evidence: evidence})
		return preview
	}
	if before.LoadState == "not-found" {
		preview.Checks = append(preview.Checks, Check{Name: "action-precondition", Status: "fail", Evidence: "allowlisted systemd unit is not installed"})
		return preview
	}
	if preview.ActionID == actionServiceReloadID && before.ActiveState != "active" {
		preview.Checks = append(preview.Checks, Check{
			Name:     "action-precondition",
			Status:   "fail",
			Evidence: "service.reload requires the allowlisted unit to be active before reload",
		})
		return preview
	}
	preview.Checks = append(preview.Checks, Check{
		Name:   "action-precondition",
		Status: "pass",
		Evidence: fmt.Sprintf("load=%s active=%s sub=%s result=%s",
			fallback(before.LoadState, "unknown"), fallback(before.ActiveState, "unknown"), fallback(before.SubState, "unknown"), fallback(before.Result, "unknown")),
	})

	if preview.ActionID == actionServiceReloadID {
		canReload, raw, reloadErr := collectActionServiceCanReload(ctx, preview.Target)
		if reloadErr != nil {
			preview.Checks = append(preview.Checks, Check{
				Name:     "reload-capability",
				Status:   "fail",
				Evidence: "systemd reload capability evidence is unavailable: " + boundedEvidence(reloadErr.Error(), 384),
			})
			return preview
		}
		if !canReload {
			preview.Checks = append(preview.Checks, Check{
				Name:     "reload-capability",
				Status:   "fail",
				Evidence: "systemd does not report CanReload=yes (observed " + fallback(raw, "unknown") + ")",
			})
			return preview
		}
		preview.Checks = append(preview.Checks, Check{Name: "reload-capability", Status: "pass", Evidence: "systemd reports CanReload=yes"})
	}

	preview.Command = []string{command, commandVerb, preview.Target}
	preview.Effect = effect
	preview.Confirmation = confirmationVerb + " " + preview.Target
	preview.Available = true
	return preview
}

func (m *ActionManager) Run(ctx context.Context, actionID, target, confirmation, mode string) (ActionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	started := time.Now().UTC()
	preview := m.Preview(ctx, actionID, target, mode)
	result := ActionResult{
		SchemaVersion: actionSchemaVersion,
		ActionID:      preview.ActionID,
		Target:        preview.Target,
		StartedAt:     started,
		Before:        preview.Before,
	}
	finish := func(status, summary string) {
		result.Status = status
		result.Summary = summary
		result.FinishedAt = time.Now().UTC()
	}

	if !preview.Available {
		finish("denied", actionPreviewFailureSummary(preview))
		if err := m.Store.AppendActionAudit(actionAuditFromResult("denied", result)); err != nil {
			return result, err
		}
		return result, nil
	}
	if confirmation != preview.Confirmation {
		finish("denied", "confirmation did not exactly match the preview value")
		if err := m.Store.AppendActionAudit(actionAuditFromResult("denied", result)); err != nil {
			return result, err
		}
		return result, nil
	}

	requested := ActionAudit{
		SchemaVersion: actionSchemaVersion,
		At:            time.Now().UTC(),
		Phase:         "requested",
		ActionID:      result.ActionID,
		Target:        result.Target,
		Status:        "requested",
		Summary:       preview.Effect,
		Before:        preview.Before,
	}
	if err := m.Store.AppendActionAudit(requested); err != nil {
		finish("blocked", "action was not executed because the audit log is unavailable")
		return result, fmt.Errorf("refusing action without durable audit log: %w", err)
	}

	var commandResult boundedCommandResult
	var commandErr error
	actionName := ""
	switch result.ActionID {
	case actionServiceRestartID:
		actionName = "restart"
		commandResult, commandErr = actionServiceRestart(ctx, result.Target)
	case actionServiceReloadID:
		actionName = "reload"
		commandResult, commandErr = actionServiceReload(ctx, result.Target)
	default:
		commandErr = errors.New("unknown action ID after successful preview")
	}
	result.CommandOutput = boundedEvidence(strings.TrimSpace(commandResult.Output), actionCommandOutputLimit)
	after, afterErr := collectActionServiceRuntime(ctx, result.Target)
	result.After = after

	switch {
	case commandErr != nil:
		finish("failed", "systemd "+actionName+" command failed: "+boundedEvidence(commandErr.Error(), 384))
	case afterErr != nil && after == nil:
		finish("failed", actionName+" returned but postcondition evidence is unavailable: "+boundedEvidence(afterErr.Error(), 384))
	case after == nil:
		finish("failed", actionName+" returned but postcondition evidence is unavailable")
	case after.ActiveState != "active":
		finish("failed", actionName+" did not satisfy the required postcondition ActiveState=active")
	case result.ActionID == actionServiceReloadID:
		finish("success", "allowlisted service reload command succeeded and postcondition ActiveState=active was verified")
	default:
		finish("success", "allowlisted service restarted and postcondition ActiveState=active was verified")
	}

	if err := m.Store.AppendActionAudit(actionAuditFromResult("completed", result)); err != nil {
		return result, fmt.Errorf("action completed but final audit write failed: %w", err)
	}
	return result, nil
}

func actionPreviewFailureSummary(preview ActionPreview) string {
	for _, check := range preview.Checks {
		if check.Status == "fail" && check.Evidence != "" {
			return check.Evidence
		}
	}
	return "action is unavailable"
}

func actionAuditFromResult(phase string, result ActionResult) ActionAudit {
	return ActionAudit{
		SchemaVersion: actionSchemaVersion,
		At:            time.Now().UTC(),
		Phase:         phase,
		ActionID:      result.ActionID,
		Target:        result.Target,
		Status:        result.Status,
		Summary:       result.Summary,
		Before:        result.Before,
		After:         result.After,
	}
}

func (s Store) ActionsPath() string { return filepath.Join(s.Dir, "actions.jsonl") }

func (s Store) AppendActionAudit(audit ActionAudit) error {
	if err := s.ensure(); err != nil {
		return err
	}
	if audit.SchemaVersion == 0 {
		audit.SchemaVersion = actionSchemaVersion
	}
	if audit.At.IsZero() {
		audit.At = time.Now().UTC()
	}
	f, err := os.OpenFile(s.ActionsPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(audit)
}

func (s Store) ReadActionAudits(limit int) ([]ActionAudit, error) {
	f, err := os.Open(s.ActionsPath())
	if errors.Is(err, os.ErrNotExist) {
		return []ActionAudit{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var audits []ActionAudit
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var audit ActionAudit
		if json.Unmarshal(scanner.Bytes(), &audit) == nil {
			audits = append(audits, audit)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limit > 0 && len(audits) > limit {
		audits = audits[len(audits)-limit:]
	}
	return audits, nil
}
