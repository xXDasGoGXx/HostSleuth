package core

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestEventsAroundIncludesAllCategoriesInsideWindowOnly(t *testing.T) {
	anchor := time.Date(2026, 9, 16, 20, 0, 0, 0, time.UTC)
	events := []Event{
		{At: anchor.Add(-16 * time.Minute), Category: "package", Summary: "too early"},
		{At: anchor.Add(-15 * time.Minute), Category: "service", Summary: "service failed"},
		{At: anchor.Add(-2 * time.Minute), Category: "listener", Summary: "listener disappeared"},
		{At: anchor, Category: "configuration", Summary: "configuration changed"},
		{At: anchor.Add(4 * time.Minute), Category: "container", Summary: "container restarted"},
		{At: anchor.Add(15 * time.Minute), Category: "system", Summary: "kernel changed"},
		{At: anchor.Add(16 * time.Minute), Category: "package", Summary: "too late"},
	}
	got := EventsAround(events, anchor, 15*time.Minute, 100)
	if len(got) != 5 {
		t.Fatalf("expected 5 bounded events, got %#v", got)
	}
	if got[0].Summary != "service failed" || got[len(got)-1].Summary != "kernel changed" {
		t.Fatalf("events are not chronological: %#v", got)
	}
}

func TestEventsAroundOrderingIsDeterministicAtSameTimestamp(t *testing.T) {
	anchor := time.Date(2026, 9, 16, 20, 0, 0, 0, time.UTC)
	events := []Event{
		{At: anchor, Category: "service", Summary: "z"},
		{At: anchor, Category: "configuration", Summary: "b"},
		{At: anchor, Category: "configuration", Summary: "a"},
	}
	got := EventsAround(events, anchor, time.Minute, 100)
	if len(got) != 3 || got[0].Summary != "a" || got[1].Summary != "b" || got[2].Summary != "z" {
		t.Fatalf("unexpected deterministic order: %#v", got)
	}
}

func TestIncidentLensDoesNotClaimNearbyEventsCausedIncident(t *testing.T) {
	anchor := time.Date(2026, 9, 16, 20, 0, 0, 0, time.UTC)
	lens := BuildIncidentLens(context.Background(), anchor, "", Snapshot{}, []Event{{At: anchor, Category: "package", Summary: "package updated"}})
	if lens.Conclusion != "1 retained host change was recorded inside the incident window" {
		t.Fatalf("unexpected conclusion: %#v", lens)
	}
	if !strings.Contains(strings.ToLower(lens.ContextNote), "does not prove causation") {
		t.Fatalf("missing non-causal boundary: %#v", lens)
	}
	if lens.CurrentEndpoint != nil || lens.EndpointCaptured != nil {
		t.Fatalf("endpoint evidence and timestamp must be absent without target: %#v", lens)
	}
}

func TestIncidentLensTargetEvidenceIsCurrentAndSeparate(t *testing.T) {
	anchor := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	oldRoute := routeLookup
	oldTCP := tcpConnect
	oldTLS := tlsProbeLookup
	routeLookup = func(context.Context, string) (string, error) { return "local 127.0.0.1 dev lo src 127.0.0.1\n", nil }
	tcpConnect = func(context.Context, string) error { return nil }
	tlsProbeLookup = func(context.Context, string, string) *TLSEvidence { return nil }
	t.Cleanup(func() {
		routeLookup = oldRoute
		tcpConnect = oldTCP
		tlsProbeLookup = oldTLS
	})

	lens := BuildIncidentLens(context.Background(), anchor, "127.0.0.1:8080", Snapshot{}, nil)
	if lens.CurrentEndpoint == nil || lens.CurrentEndpoint.Conclusion != "target is reachable" {
		t.Fatalf("expected current endpoint evidence: %#v", lens)
	}
	if lens.EndpointCaptured == nil || lens.EndpointCaptured.IsZero() || !lens.EndpointCaptured.After(anchor) {
		t.Fatalf("current endpoint timestamp must remain separate from historical anchor: %#v", lens)
	}
}

func TestEventsAroundHonorsBoundedLimit(t *testing.T) {
	anchor := time.Date(2026, 9, 16, 20, 0, 0, 0, time.UTC)
	var events []Event
	for i := 0; i < 10; i++ {
		events = append(events, Event{At: anchor.Add(time.Duration(i) * time.Second), Category: "service", Summary: string(rune('a' + i))})
	}
	got := EventsAround(events, anchor, time.Minute, 3)
	if len(got) != 3 {
		t.Fatalf("expected bounded result, got %d", len(got))
	}
}
