package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadOSReleaseUsesConfiguredPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "os-release")
	if err := os.WriteFile(path, []byte("PRETTY_NAME=\"Host Linux\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOSTSLEUTH_OS_RELEASE_PATH", path)

	if got := readOSRelease(); got != "Host Linux" {
		t.Fatalf("expected host OS release, got %q", got)
	}
}

func TestSystemdFailureCheckDockerModeIsUnavailable(t *testing.T) {
	oldJournal := journalUnitLookup
	journalUnitLookup = func(context.Context, string) (boundedCommandResult, error) {
		t.Fatal("Docker mode must not attempt to read the host journal")
		return boundedCommandResult{}, nil
	}
	defer func() { journalUnitLookup = oldJournal }()

	check := systemdFailureCheck(context.Background(), Snapshot{Mode: dockerDeploymentMode})
	if check.Status != "unknown" {
		t.Fatalf("expected unknown status, got %#v", check)
	}
	if !strings.Contains(check.Evidence, "unavailable in Docker deployment mode") {
		t.Fatalf("unexpected evidence: %q", check.Evidence)
	}
}

func TestStorePreservesDockerSnapshotMode(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	if err := store.SaveSnapshot(Snapshot{Mode: dockerDeploymentMode}); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != dockerDeploymentMode {
		t.Fatalf("expected mode %q, got %q", dockerDeploymentMode, got.Mode)
	}
}
