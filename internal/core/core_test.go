package core

import (
	"context"
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
