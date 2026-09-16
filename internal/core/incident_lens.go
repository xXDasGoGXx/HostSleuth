package core

import (
	"context"
	"sort"
	"strings"
	"time"
)

const (
	incidentDefaultWindow = 15 * time.Minute
	incidentEventLimit    = 100
)

type IncidentLens struct {
	AnchorAt         time.Time  `json:"anchor_at"`
	WindowStart      time.Time  `json:"window_start"`
	WindowEnd        time.Time  `json:"window_end"`
	WindowMinutes    int        `json:"window_minutes"`
	Target           string     `json:"target,omitempty"`
	Events           []Event    `json:"events"`
	CurrentEndpoint  *Diagnosis `json:"current_endpoint,omitempty"`
	EndpointCaptured time.Time  `json:"endpoint_captured_at,omitempty"`
	Conclusion       string     `json:"conclusion"`
	ContextNote      string     `json:"context_note"`
}

// BuildIncidentLens returns retained events close to an explicit anchor. Nearby
// events are context only: temporal proximity is never promoted to causation.
// When target is supplied, endpoint evidence is captured now and is explicitly
// separate from the historical event window.
func BuildIncidentLens(ctx context.Context, anchor time.Time, target string, snap Snapshot, events []Event) IncidentLens {
	anchor = anchor.UTC()
	lens := IncidentLens{
		AnchorAt:      anchor,
		WindowStart:   anchor.Add(-incidentDefaultWindow),
		WindowEnd:     anchor.Add(incidentDefaultWindow),
		WindowMinutes: int(incidentDefaultWindow / time.Minute),
		Target:        strings.TrimSpace(target),
		Events:        EventsAround(events, anchor, incidentDefaultWindow, incidentEventLimit),
		ContextNote:   "Events are shown because they occurred within the bounded incident window; temporal proximity does not prove causation.",
	}

	if lens.Target != "" {
		diagnosis := Diagnose(ctx, lens.Target, snap)
		lens.CurrentEndpoint = &diagnosis
		lens.EndpointCaptured = diagnosis.StartedAt
	}

	switch len(lens.Events) {
	case 0:
		lens.Conclusion = "no retained host changes were recorded inside the incident window"
	case 1:
		lens.Conclusion = "1 retained host change was recorded inside the incident window"
	default:
		lens.Conclusion = "multiple retained host changes were recorded inside the incident window"
	}
	return lens
}

// EventsAround is the shared bounded event-window primitive for incident
// correlation. Results are chronological and stable for equal timestamps.
func EventsAround(events []Event, anchor time.Time, window time.Duration, limit int) []Event {
	if anchor.IsZero() || window < 0 || limit <= 0 {
		return []Event{}
	}
	start := anchor.Add(-window)
	end := anchor.Add(window)
	out := make([]Event, 0)
	for _, event := range events {
		if event.At.IsZero() || event.At.Before(start) || event.At.After(end) {
			continue
		}
		out = append(out, event)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].At.Equal(out[j].At) {
			if out[i].Category == out[j].Category {
				return out[i].Summary < out[j].Summary
			}
			return out[i].Category < out[j].Category
		}
		return out[i].At.Before(out[j].At)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
