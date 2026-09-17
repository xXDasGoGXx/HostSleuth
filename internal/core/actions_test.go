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
