package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Listener blips.
//
// When a host reboots, or Docker or a service restarts, HostSleuth's next
// snapshot can catch listeners before they are back: every one of them is
// recorded as "disappeared", and the following snapshot records them all as
// "appeared". On a busy host that is dozens of lines that cancel out, and
// they crowd the reboot itself out of the recent timeline.
//
// CollapseListenerBlips pairs each disappearance with the same socket's
// reappearance within listenerBlipWindow and replaces every group that
// disappeared in the same snapshot with one summary event. A listener that
// did not come back keeps its own "disappeared" warning. This is a view
// over recorded events: the timeline file is unchanged.

const listenerBlipWindow = 15 * time.Minute

// rebootProximity is how long before a disappearance a recorded reboot
// still counts as its context. The first post-boot snapshot records the
// reboot and the disappearances together, so this only needs to cover a
// slow first capture.
const rebootProximity = 10 * time.Minute

const listenerBlipSamples = 3

const (
	listenerAppearedPrefix    = "listener appeared: "
	listenerDisappearedPrefix = "listener disappeared: "
	rebootSummaryPrefix       = "host reboot detected"
)

// CollapseListenerBlips expects events oldest first and preserves order.
func CollapseListenerBlips(events []Event) []Event {
	type blip struct {
		key string
		gap time.Duration
	}
	appeared := map[string][]int{}
	for i, e := range events {
		if key, ok := listenerEventKey(e, listenerAppearedPrefix); ok {
			appeared[key] = append(appeared[key], i)
		}
	}

	drop := make([]bool, len(events))
	groups := map[time.Time][]blip{}
	groupStart := map[time.Time]int{}
	for i, e := range events {
		key, ok := listenerEventKey(e, listenerDisappearedPrefix)
		if !ok {
			continue
		}
		for _, j := range appeared[key] {
			if j <= i || drop[j] {
				continue
			}
			gap := events[j].At.Sub(e.At)
			if gap <= 0 {
				continue
			}
			if gap > listenerBlipWindow {
				break
			}
			drop[i], drop[j] = true, true
			at := e.At.UTC()
			if _, seen := groupStart[at]; !seen {
				groupStart[at] = i
			}
			groups[at] = append(groups[at], blip{key: key, gap: gap})
			break
		}
	}
	if len(groups) == 0 {
		return events
	}

	summaries := map[int]Event{}
	for at, members := range groups {
		keys := make([]string, 0, len(members))
		var longest time.Duration
		for _, m := range members {
			keys = append(keys, m.key)
			if m.gap > longest {
				longest = m.gap
			}
		}
		afterReboot := rebootBefore(events, at)
		severity := "warning"
		if afterReboot {
			severity = "info"
		}
		summaries[groupStart[at]] = Event{
			SchemaVersion: 1,
			At:            events[groupStart[at]].At,
			Category:      "listener",
			Severity:      severity,
			Summary:       listenerBlipSummary(keys, longest, afterReboot),
		}
	}

	out := make([]Event, 0, len(events))
	for i, e := range events {
		if summary, ok := summaries[i]; ok {
			out = append(out, summary)
			continue
		}
		if !drop[i] {
			out = append(out, e)
		}
	}
	return out
}

func listenerEventKey(e Event, prefix string) (string, bool) {
	if e.Category != "listener" || !strings.HasPrefix(e.Summary, prefix) {
		return "", false
	}
	key := strings.TrimSpace(strings.TrimPrefix(e.Summary, prefix))
	return key, key != ""
}

func rebootBefore(events []Event, at time.Time) bool {
	for _, e := range events {
		if e.Category != "system" || !strings.HasPrefix(e.Summary, rebootSummaryPrefix) {
			continue
		}
		if !e.At.After(at) && at.Sub(e.At) <= rebootProximity {
			return true
		}
	}
	return false
}

func listenerBlipSummary(keys []string, longest time.Duration, afterReboot bool) string {
	sort.Slice(keys, func(i, j int) bool {
		ti, tj := strings.HasPrefix(keys[i], "tcp "), strings.HasPrefix(keys[j], "tcp ")
		if ti != tj {
			return ti
		}
		return keys[i] < keys[j]
	})
	samples := keys
	if len(samples) > listenerBlipSamples {
		samples = samples[:listenerBlipSamples]
	}
	list := strings.Join(samples, ", ")
	if extra := len(keys) - len(samples); extra > 0 {
		list += fmt.Sprintf(" and %d more", extra)
	}
	within := blipDuration(longest)
	context := ""
	if afterReboot {
		context = " after the host reboot"
	}
	if len(keys) == 1 {
		return fmt.Sprintf("listener went away%s and was back within %s: %s", context, within, list)
	}
	return fmt.Sprintf("%d listeners went away%s and were back within %s: %s", len(keys), context, within, list)
}

// blipDuration shows short gaps in whole seconds and longer ones in whole
// minutes, rounding up so the stated bound is never shorter than observed.
func blipDuration(d time.Duration) string {
	if d < 2*time.Minute {
		seconds := int((d + time.Second - 1) / time.Second)
		if seconds < 1 {
			seconds = 1
		}
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dm", int((d+time.Minute-1)/time.Minute))
}
