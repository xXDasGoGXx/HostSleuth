package core

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

func trustedExecutable(candidates ...string) (string, error) {
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || !strings.HasPrefix(candidate, "/") {
			continue
		}
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() || info.Mode().Perm()&0o111 == 0 {
			continue
		}
		return candidate, nil
	}
	return "", errors.New("trusted executable is unavailable")
}

func trustedActionSystemctl() (string, error) {
	return trustedExecutable("/usr/bin/systemctl", "/bin/systemctl")
}

var actionSystemctlPath = trustedActionSystemctl

var actionServiceRuntimeLookup = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	command, err := actionSystemctlPath()
	if err != nil {
		return boundedCommandResult{}, err
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return runBoundedCommand(lookupCtx, actionCommandOutputLimit, command, "show", unit, "--no-pager",
		"--property=LoadState,ActiveState,SubState,UnitFileState,Result,MainPID,ControlGroup,ExecMainCode,ExecMainStatus")
}

var actionServiceReloadCapabilityLookup = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	command, err := actionSystemctlPath()
	if err != nil {
		return boundedCommandResult{}, err
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return runBoundedCommand(lookupCtx, actionCommandOutputLimit, command, "show", unit, "--no-pager", "--property=CanReload", "--value")
}

func collectActionServiceCanReload(ctx context.Context, unit string) (bool, string, error) {
	result, err := actionServiceReloadCapabilityLookup(ctx, unit)
	value := strings.ToLower(strings.TrimSpace(result.Output))
	if err != nil {
		return false, value, err
	}
	return value == "yes", value, nil
}

func collectActionServiceRuntime(ctx context.Context, unit string) (*ServiceRuntimeEvidence, error) {
	result, err := actionServiceRuntimeLookup(ctx, unit)
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
