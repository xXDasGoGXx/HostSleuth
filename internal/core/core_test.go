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
