package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDiffSnapshotsDetectsServiceAndListenerChanges(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Services:   []ServiceInfo{{Name: "demo.service", Active: "active", Sub: "running"}},
		Listeners:  []Listener{{Protocol: "tcp", Address: "0.0.0.0:8080"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Services:   []ServiceInfo{{Name: "demo.service", Active: "failed", Sub: "failed"}},
	}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(events))
	}
}

func TestDiffSnapshotsIgnoresContainerUptimeOnlyChanges(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 5 minutes (healthy)"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 6 minutes (healthy)"}},
	}
	if events := DiffSnapshots(oldSnap, newSnap); len(events) != 0 {
		t.Fatalf("expected uptime-only container change to be ignored, got %#v", events)
	}
}

func TestDiffSnapshotsIgnoresContainerRestartTimerChurn(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Restarting (1) 5 seconds ago"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Restarting (1) 15 seconds ago"}},
	}
	if events := DiffSnapshots(oldSnap, newSnap); len(events) != 0 {
		t.Fatalf("expected restart-timer-only container change to be ignored, got %#v", events)
	}
}

func TestDiffSnapshotsDetectsContainerHealthChange(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 5 minutes (healthy)"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 6 minutes (unhealthy)"}},
	}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) != 1 {
		t.Fatalf("expected one health-change event, got %#v", events)
	}
	if events[0].Severity != "warning" {
		t.Fatalf("expected warning severity, got %q", events[0].Severity)
	}
	if events[0].Summary != "container changed: demo: running (healthy) -> running (unhealthy)" {
		t.Fatalf("unexpected summary: %q", events[0].Summary)
	}
}

func TestDiffSnapshotsDetectsContainerExitWithoutAgeNoise(t *testing.T) {
	oldSnap := Snapshot{
		CapturedAt: time.Now().Add(-time.Minute),
		Containers: []ContainerInfo{{Name: "demo", Status: "Up 5 minutes"}},
	}
	newSnap := Snapshot{
		CapturedAt: time.Now(),
		Containers: []ContainerInfo{{Name: "demo", Status: "Exited (1) 3 seconds ago"}},
	}
	events := DiffSnapshots(oldSnap, newSnap)
	if len(events) != 1 {
		t.Fatalf("expected one exit event, got %#v", events)
	}
	if events[0].Severity != "warning" || !strings.Contains(events[0].Summary, "running -> exited (1)") {
		t.Fatalf("unexpected exit event: %#v", events[0])
	}
}

func TestDiagnoseRejectsInvalidTarget(t *testing.T) {
	d := Diagnose(context.Background(), "not-a-host-port", Snapshot{})
	if d.Conclusion != "invalid target" {
		t.Fatalf("unexpected conclusion: %q", d.Conclusion)
	}
	if d.Confidence != "high" {
		t.Fatalf("unexpected confidence: %q", d.Confidence)
	}
}

func TestStorePersistsSchemaAndReturnsEmptyEvents(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	if err := store.SaveSnapshot(Snapshot{Host: HostInfo{Hostname: "test-host"}}); err != nil {
		t.Fatal(err)
	}
	got, err := store.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != 1 {
		t.Fatalf("expected snapshot schema 1, got %d", got.SchemaVersion)
	}
	events, err := store.ReadEvents(10)
	if err != nil {
		t.Fatal(err)
	}
	if events == nil || len(events) != 0 {
		t.Fatalf("expected empty non-nil events slice, got %#v", events)
	}
}

func TestStoreAddsEventSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	if err := store.AppendEvents([]Event{{Summary: "test"}}); err != nil {
		t.Fatal(err)
	}
	events, err := store.ReadEvents(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].SchemaVersion != 1 {
		t.Fatalf("unexpected events: %#v", events)
	}
}
