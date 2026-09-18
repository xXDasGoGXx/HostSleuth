package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withActionMocks(t *testing.T, runtime func(context.Context, string) (boundedCommandResult, error), restart func(context.Context, string) (boundedCommandResult, error)) {
	t.Helper()
	oldRuntime := actionServiceRuntimeLookup
	oldRestart := actionServiceRestart
	oldPath := actionSystemctlPath
	actionServiceRuntimeLookup = runtime
	actionServiceRestart = restart
	actionSystemctlPath = func() (string, error) { return "/usr/bin/systemctl", nil }
	t.Cleanup(func() {
		actionServiceRuntimeLookup = oldRuntime
		actionServiceRestart = oldRestart
		actionSystemctlPath = oldPath
	})
}

func enabledActionManager(t *testing.T, dir string) *ActionManager {
	t.Helper()
	policy, err := NewActionPolicy(true, []string{"demo"})
	if err != nil {
		t.Fatal(err)
	}
	return &ActionManager{Policy: policy, Store: Store{Dir: dir}}
}

func TestActionPolicyRejectsUnsafeServiceNames(t *testing.T) {
	bad := []string{
		"-demo.service",
		"demo.service;reboot",
		"demo.service --no-block",
		"demo.service/../../other",
		"demo.service\nother.service",
	}
	for _, value := range bad {
		if _, err := NewActionPolicy(true, []string{value}); err == nil {
			t.Fatalf("unsafe service %q was accepted", value)
		}
	}
}

func TestTrustedExecutableIgnoresPATHAndRequiresAbsoluteCandidate(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "systemctl")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if got, err := trustedExecutable("systemctl"); err == nil {
		t.Fatalf("relative PATH command was trusted: %q", got)
	}
	got, err := trustedExecutable(fake)
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("trusted executable = %q, want %q", got, fake)
	}
}

func TestActionPreviewDisabledByDefaultDoesNotProbeSystemd(t *testing.T) {
	calls := 0
	withActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			calls++
			return boundedCommandResult{}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			t.Fatal("restart must not execute")
			return boundedCommandResult{}, nil
		},
	)
	policy, err := NewActionPolicy(false, []string{"demo"})
	if err != nil {
		t.Fatal(err)
	}
	manager := &ActionManager{Policy: policy, Store: Store{Dir: t.TempDir()}}
	preview := manager.Preview(context.Background(), actionServiceRestartID, "demo", "native")
	if preview.Available || preview.Allowed {
		t.Fatalf("disabled action unexpectedly available: %+v", preview)
	}
	if calls != 0 {
		t.Fatalf("disabled action probed systemd %d time(s)", calls)
	}
}

func TestActionPreviewRequiresAllowlistAndNativeMode(t *testing.T) {
	withActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=active\nSubState=running\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			t.Fatal("restart must not execute during preview")
			return boundedCommandResult{}, nil
		},
	)
	manager := enabledActionManager(t, t.TempDir())
	if preview := manager.Preview(context.Background(), actionServiceRestartID, "other", "native"); preview.Available || preview.Allowed {
		t.Fatalf("non-allowlisted target unexpectedly available: %+v", preview)
	}
	if preview := manager.Preview(context.Background(), actionServiceRestartID, "demo", dockerDeploymentMode); preview.Available {
		t.Fatalf("Docker action unexpectedly available: %+v", preview)
	}
}

func TestActionRunRequiresExactConfirmationAndAuditsDenial(t *testing.T) {
	restarts := 0
	withActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=failed\nSubState=failed\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			restarts++
			return boundedCommandResult{}, nil
		},
	)
	manager := enabledActionManager(t, t.TempDir())
	result, err := manager.Run(context.Background(), actionServiceRestartID, "demo", "yes", "native")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "denied" || restarts != 0 {
		t.Fatalf("confirmation boundary failed: result=%+v restarts=%d", result, restarts)
	}
	audits, err := manager.Store.ReadActionAudits(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(audits) != 1 || audits[0].Status != "denied" {
		t.Fatalf("expected one denied audit record, got %+v", audits)
	}
}

func TestActionRunRestartsExactAllowlistedUnitAndVerifiesPostcondition(t *testing.T) {
	runtimeCalls := 0
	restartedUnit := ""
	withActionMocks(t,
		func(_ context.Context, unit string) (boundedCommandResult, error) {
			if unit != "demo.service" {
				t.Fatalf("unexpected runtime unit %q", unit)
			}
			runtimeCalls++
			if runtimeCalls == 1 {
				return boundedCommandResult{Output: "LoadState=loaded\nActiveState=failed\nSubState=failed\nResult=exit-code\n"}, nil
			}
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=active\nSubState=running\nResult=success\n"}, nil
		},
		func(_ context.Context, unit string) (boundedCommandResult, error) {
			restartedUnit = unit
			return boundedCommandResult{Output: ""}, nil
		},
	)
	manager := enabledActionManager(t, t.TempDir())
	result, err := manager.Run(context.Background(), actionServiceRestartID, "demo", "RESTART demo.service", "native")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "success" {
		t.Fatalf("expected success, got %+v", result)
	}
	if restartedUnit != "demo.service" {
		t.Fatalf("restart target = %q", restartedUnit)
	}
	if result.After == nil || result.After.ActiveState != "active" {
		t.Fatalf("postcondition was not captured: %+v", result.After)
	}
	audits, err := manager.Store.ReadActionAudits(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(audits) != 2 || audits[0].Phase != "requested" || audits[1].Phase != "completed" || audits[1].Status != "success" {
		t.Fatalf("unexpected audit trail: %+v", audits)
	}
}

func TestActionRunFailsClosedWhenPostconditionIsNotActive(t *testing.T) {
	withActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=failed\nSubState=failed\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{}, nil
		},
	)
	manager := enabledActionManager(t, t.TempDir())
	result, err := manager.Run(context.Background(), actionServiceRestartID, "demo", "RESTART demo.service", "native")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "failed" || !strings.Contains(result.Summary, "ActiveState=active") {
		t.Fatalf("expected verified postcondition failure, got %+v", result)
	}
}

func TestActionRunRefusesExecutionWhenAuditLogUnavailable(t *testing.T) {
	restarts := 0
	withActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=failed\nSubState=failed\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			restarts++
			return boundedCommandResult{}, nil
		},
	)
	manager := enabledActionManager(t, "")
	result, err := manager.Run(context.Background(), actionServiceRestartID, "demo", "RESTART demo.service", "native")
	if err == nil || !strings.Contains(err.Error(), "audit") {
		t.Fatalf("expected audit failure, result=%+v err=%v", result, err)
	}
	if restarts != 0 {
		t.Fatalf("restart executed even though audit preflight failed")
	}
}

func TestActionCommandFailureIsRecordedAsFailure(t *testing.T) {
	runtimeCalls := 0
	withActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			runtimeCalls++
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=failed\nSubState=failed\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "permission denied"}, errors.New("exit status 1")
		},
	)
	manager := enabledActionManager(t, t.TempDir())
	result, err := manager.Run(context.Background(), actionServiceRestartID, "demo", "RESTART demo.service", "native")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "failed" || !strings.Contains(result.Summary, "restart command failed") {
		t.Fatalf("expected command failure, got %+v", result)
	}
	if runtimeCalls != 2 {
		t.Fatalf("expected before and after evidence, got %d runtime calls", runtimeCalls)
	}
}

func withReloadActionMocks(t *testing.T, runtime func(context.Context, string) (boundedCommandResult, error), canReload func(context.Context, string) (boundedCommandResult, error), reload func(context.Context, string) (boundedCommandResult, error)) {
	t.Helper()
	oldRuntime := actionServiceRuntimeLookup
	oldCanReload := actionServiceReloadCapabilityLookup
	oldReload := actionServiceReload
	oldRestart := actionServiceRestart
	oldPath := actionSystemctlPath
	actionServiceRuntimeLookup = runtime
	actionServiceReloadCapabilityLookup = canReload
	actionServiceReload = reload
	actionServiceRestart = func(context.Context, string) (boundedCommandResult, error) {
		t.Fatal("restart must never be used as reload fallback")
		return boundedCommandResult{}, nil
	}
	actionSystemctlPath = func() (string, error) { return "/usr/bin/systemctl", nil }
	t.Cleanup(func() {
		actionServiceRuntimeLookup = oldRuntime
		actionServiceReloadCapabilityLookup = oldCanReload
		actionServiceReload = oldReload
		actionServiceRestart = oldRestart
		actionSystemctlPath = oldPath
	})
}

func enabledReloadActionManager(t *testing.T, dir string) *ActionManager {
	t.Helper()
	policy, err := NewActionPolicyWithReload(true, nil, []string{"demo"})
	if err != nil {
		t.Fatal(err)
	}
	return &ActionManager{Policy: policy, Store: Store{Dir: dir}}
}

func TestActionPolicyKeepsRestartAndReloadAllowlistsIndependent(t *testing.T) {
	policy, err := NewActionPolicyWithReload(true, []string{"restart-only"}, []string{"reload-only"})
	if err != nil {
		t.Fatal(err)
	}
	if !policy.AllowedRestartServices["restart-only.service"] || policy.AllowedRestartServices["reload-only.service"] {
		t.Fatalf("unexpected restart allowlist: %+v", policy.AllowedRestartServices)
	}
	if !policy.AllowedReloadServices["reload-only.service"] || policy.AllowedReloadServices["restart-only.service"] {
		t.Fatalf("unexpected reload allowlist: %+v", policy.AllowedReloadServices)
	}
	if _, err := NewActionPolicyWithReload(true, nil, []string{"demo.service;restart"}); err == nil {
		t.Fatal("unsafe reload allowlist target was accepted")
	}
}

func TestActionCapabilitiesExposeReloadSeparately(t *testing.T) {
	policy, err := NewActionPolicyWithReload(true, []string{"restart-only"}, []string{"reload-only"})
	if err != nil {
		t.Fatal(err)
	}
	caps := (&ActionManager{Policy: policy}).Capabilities("native")
	if len(caps) != 2 {
		t.Fatalf("capabilities=%+v", caps)
	}
	if caps[0].ActionID != actionServiceRestartID || len(caps[0].AllowedTargets) != 1 || caps[0].AllowedTargets[0] != "restart-only.service" {
		t.Fatalf("unexpected restart capability: %+v", caps[0])
	}
	if caps[1].ActionID != actionServiceReloadID || len(caps[1].AllowedTargets) != 1 || caps[1].AllowedTargets[0] != "reload-only.service" {
		t.Fatalf("unexpected reload capability: %+v", caps[1])
	}
	for _, cap := range (&ActionManager{Policy: policy}).Capabilities(dockerDeploymentMode) {
		if cap.Available {
			t.Fatalf("Docker action unexpectedly available: %+v", cap)
		}
	}
}

func TestReloadDoesNotInheritRestartAllowlist(t *testing.T) {
	policy, err := NewActionPolicyWithReload(true, []string{"demo"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	manager := &ActionManager{Policy: policy, Store: Store{Dir: t.TempDir()}}
	preview := manager.Preview(context.Background(), actionServiceReloadID, "demo", "native")
	if preview.Available || preview.Allowed {
		t.Fatalf("reload inherited restart allowlist: %+v", preview)
	}
	if len(preview.Checks) == 0 || !strings.Contains(preview.Checks[0].Evidence, "service.reload") {
		t.Fatalf("reload denial did not identify the independent allowlist: %+v", preview.Checks)
	}
}

func TestReloadPreviewRequiresActiveReloadCapableUnit(t *testing.T) {
	withReloadActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=active\nSubState=running\nResult=success\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "yes\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			t.Fatal("reload must not execute during preview")
			return boundedCommandResult{}, nil
		},
	)
	preview := enabledReloadActionManager(t, t.TempDir()).Preview(context.Background(), actionServiceReloadID, "demo", "native")
	if !preview.Available || !preview.Allowed {
		t.Fatalf("reload preview unavailable: %+v", preview)
	}
	if preview.Confirmation != "RELOAD demo.service" {
		t.Fatalf("confirmation=%q", preview.Confirmation)
	}
	if got := strings.Join(preview.Command, " "); got != "/usr/bin/systemctl reload demo.service" {
		t.Fatalf("command=%q", got)
	}
	if !strings.Contains(preview.Effect, "never fall back to restart") {
		t.Fatalf("effect does not preserve no-fallback boundary: %q", preview.Effect)
	}
}

func TestReloadPreviewRejectsInactiveUnitBeforeCapabilityProbe(t *testing.T) {
	capabilityCalls := 0
	withReloadActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=inactive\nSubState=dead\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			capabilityCalls++
			return boundedCommandResult{Output: "yes\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			t.Fatal("reload must not execute")
			return boundedCommandResult{}, nil
		},
	)
	preview := enabledReloadActionManager(t, t.TempDir()).Preview(context.Background(), actionServiceReloadID, "demo", "native")
	if preview.Available || !preview.Allowed {
		t.Fatalf("inactive reload preview=%+v", preview)
	}
	if capabilityCalls != 0 {
		t.Fatalf("reload capability probed despite inactive precondition: %d", capabilityCalls)
	}
}

func TestReloadPreviewFailsClosedWhenCanReloadIsNotYes(t *testing.T) {
	withReloadActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=active\nSubState=running\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "no\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			t.Fatal("reload must not execute")
			return boundedCommandResult{}, nil
		},
	)
	preview := enabledReloadActionManager(t, t.TempDir()).Preview(context.Background(), actionServiceReloadID, "demo", "native")
	if preview.Available || !preview.Allowed {
		t.Fatalf("non-reloadable unit became available: %+v", preview)
	}
	found := false
	for _, check := range preview.Checks {
		if check.Name == "reload-capability" && check.Status == "fail" && strings.Contains(check.Evidence, "CanReload=yes") {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing fail-closed reload capability evidence: %+v", preview.Checks)
	}
}

func TestReloadRunUsesExactActionAndVerifiesActivePostcondition(t *testing.T) {
	runtimeCalls := 0
	reloadedUnit := ""
	withReloadActionMocks(t,
		func(_ context.Context, unit string) (boundedCommandResult, error) {
			if unit != "demo.service" {
				t.Fatalf("unexpected runtime unit %q", unit)
			}
			runtimeCalls++
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=active\nSubState=running\nResult=success\n"}, nil
		},
		func(_ context.Context, unit string) (boundedCommandResult, error) {
			if unit != "demo.service" {
				t.Fatalf("unexpected capability unit %q", unit)
			}
			return boundedCommandResult{Output: "yes\n"}, nil
		},
		func(_ context.Context, unit string) (boundedCommandResult, error) {
			reloadedUnit = unit
			return boundedCommandResult{}, nil
		},
	)
	manager := enabledReloadActionManager(t, t.TempDir())
	result, err := manager.Run(context.Background(), actionServiceReloadID, "demo", "RELOAD demo.service", "native")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "success" || reloadedUnit != "demo.service" || runtimeCalls != 2 {
		t.Fatalf("unexpected reload result: result=%+v unit=%q runtimeCalls=%d", result, reloadedUnit, runtimeCalls)
	}
	if !strings.Contains(result.Summary, "reload command succeeded") || strings.Contains(strings.ToLower(result.Summary), "configuration applied") {
		t.Fatalf("reload result overclaims semantics: %q", result.Summary)
	}
	audits, err := manager.Store.ReadActionAudits(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(audits) != 2 || audits[0].ActionID != actionServiceReloadID || audits[0].Phase != "requested" || audits[1].Phase != "completed" || audits[1].Status != "success" {
		t.Fatalf("unexpected reload audit trail: %+v", audits)
	}
}

func TestReloadFailureNeverFallsBackToRestart(t *testing.T) {
	withReloadActionMocks(t,
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "LoadState=loaded\nActiveState=active\nSubState=running\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "yes\n"}, nil
		},
		func(context.Context, string) (boundedCommandResult, error) {
			return boundedCommandResult{Output: "reload failed"}, errors.New("exit status 1")
		},
	)
	result, err := enabledReloadActionManager(t, t.TempDir()).Run(context.Background(), actionServiceReloadID, "demo", "RELOAD demo.service", "native")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "failed" || !strings.Contains(result.Summary, "reload command failed") {
		t.Fatalf("unexpected reload failure: %+v", result)
	}
}
