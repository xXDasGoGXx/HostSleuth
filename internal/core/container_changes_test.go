package core

import (
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestApplyDockerInspectParsesFieldLimitedOutput(t *testing.T) {
	cs := []ContainerInfo{{ID: "0123456789ab", Name: "media"}, {ID: "fedcba987654", Name: "vpn"}}
	out := strings.Join([]string{
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\tsha256:aaaa1111\t2\tfalse\t0\t2026-10-15T02:40:08.1Z\t2026-10-15T02:39:59.9Z\t\t1.42.1",
		"fedcba987654fedcba987654fedcba987654fedcba987654fedcba987654fedc\tsha256:bbbb2222\t0\ttrue\t137\t2026-10-15T01:00:00Z\t0001-01-01T00:00:00Z\t3\t<no value>",
		"garbage line",
	}, "\n")
	applyDockerInspect(cs, out)
	m, v := cs[0], cs[1]
	if m.ImageID != "sha256:aaaa1111" || m.RestartCount != 2 || m.OOMKilled || m.StartedAt != "2026-10-15T02:40:08.1Z" || m.ImageVersion != "1.42.1" || m.HealthFailingStreak != 0 {
		t.Fatalf("media parsed wrong: %+v", m)
	}
	if !v.OOMKilled || v.ExitCode != 137 || v.FinishedAt != "" || v.HealthFailingStreak != 3 || v.ImageVersion != "" {
		t.Fatalf("vpn parsed wrong: %+v", v)
	}
}

func TestDockerInspectFormatNeverSelectsConfigSecrets(t *testing.T) {
	for _, forbidden := range []string{".Env", ".Cmd", ".Mounts", ".HostConfig", ".Args", ".Entrypoint"} {
		if strings.Contains(dockerInspectFormat, forbidden) || strings.Contains(dockerExitFormat, forbidden) {
			t.Errorf("Docker format selects %s", forbidden)
		}
	}
	if strings.Contains(dockerInspectFormat, "{{json") || strings.Contains(dockerExitFormat, "{{json") {
		t.Error("Docker formats must select fields, not whole objects")
	}
}

func TestApplyDockerExitsMarksOOMAndOrdersExits(t *testing.T) {
	base := time.Date(2026, 10, 15, 2, 39, 0, 0, time.UTC)
	ns := func(d time.Duration) string { return itoa64(base.Add(d).UnixNano()) }
	cs := []ContainerInfo{{Name: "vpn"}, {Name: "media"}}
	out := strings.Join([]string{
		ns(30*time.Second) + "\tdie\tvpn\t1",
		ns(5*time.Second) + "\toom\tvpn\t",
		ns(6*time.Second) + "\tdie\tvpn\t137",
		ns(40*time.Second) + "\tdie\tmedia\t143",
		ns(41*time.Second) + "\tdie\tunknown\t1",
	}, "\n")
	applyDockerExits(cs, out)
	vpn := cs[0].RecentExits
	if len(vpn) != 2 || !vpn[0].OOMKilled || vpn[0].ExitCode != 137 || vpn[1].OOMKilled || vpn[1].ExitCode != 1 {
		t.Fatalf("vpn exits: %+v", vpn)
	}
	if len(cs[1].RecentExits) != 1 || cs[1].RecentExits[0].ExitCode != 143 {
		t.Fatalf("media exits: %+v", cs[1].RecentExits)
	}
}

func itoa64(v int64) string {
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	var b []byte
	for v > 0 {
		b = append([]byte{digits[v%10]}, b...)
		v /= 10
	}
	return string(b)
}

func runtimeSnap(at time.Time, cs ...ContainerInfo) Snapshot {
	return Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: at, Containers: cs}
}

func TestContainerUpdateEvents(t *testing.T) {
	t0 := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	old := ContainerInfo{ID: "aaaaaaaaaaaa", Name: "media", Image: "example/media:latest", Status: "Up 2 days", ImageID: "sha256:old", ImageVersion: "1.41.0", StartedAt: "2026-10-13T00:00:00Z"}

	cases := []struct {
		name string
		next ContainerInfo
		want string
	}{
		{"same tag, new version", ContainerInfo{ID: "bbbbbbbbbbbb", Name: "media", Image: "example/media:latest", Status: "Up 5 seconds", ImageID: "sha256:new", ImageVersion: "1.42.1", StartedAt: "2026-10-15T03:00:30Z"},
			"container updated: media: example/media:latest, version 1.41.0 → 1.42.1"},
		{"new tag", ContainerInfo{ID: "bbbbbbbbbbbb", Name: "media", Image: "example/media:1.42", Status: "Up 5 seconds", ImageID: "sha256:new", ImageVersion: "1.42.1", StartedAt: "2026-10-15T03:00:30Z"},
			"container updated: media: example/media:latest → example/media:1.42, version 1.41.0 → 1.42.1"},
		{"no version label", ContainerInfo{ID: "bbbbbbbbbbbb", Name: "media", Image: "example/media:latest", Status: "Up 5 seconds", ImageID: "sha256:new", StartedAt: "2026-10-15T03:00:30Z"},
			"container updated: media: example/media:latest, new image build (no version label)"},
		{"recreated, same image", ContainerInfo{ID: "bbbbbbbbbbbb", Name: "media", Image: "example/media:latest", Status: "Up 5 seconds", ImageID: "sha256:old", ImageVersion: "1.41.0", StartedAt: "2026-10-15T03:00:30Z"},
			"container recreated: media (same image)"},
	}
	for _, tc := range cases {
		events := DiffSnapshots(runtimeSnap(t0, old), runtimeSnap(t1, tc.next))
		found := false
		for _, e := range events {
			if e.Summary == tc.want {
				found = true
				if e.Severity != "info" || e.Category != "container" {
					t.Errorf("%s: %+v", tc.name, e)
				}
			}
			if strings.Contains(e.Summary, "sha256") {
				t.Errorf("%s: summary prints an image digest: %s", tc.name, e.Summary)
			}
		}
		if !found {
			t.Errorf("%s: want %q, got %+v", tc.name, tc.want, events)
		}
	}
}

func TestContainerRestartEventsUseDockerExitEvidence(t *testing.T) {
	t0 := time.Date(2026, 10, 15, 2, 38, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	old := ContainerInfo{ID: "cccccccccccc", Name: "vpn", Image: "example/vpn:5", Status: "Up 3 hours (healthy)", ImageID: "sha256:x", StartedAt: "2026-10-14T23:00:00Z",
		RecentExits: []ContainerExit{{At: t0.Add(-10 * time.Minute), ExitCode: 1}}} // already reported
	next := old
	next.Status = "Up 20 seconds (health: starting)"
	next.StartedAt = "2026-10-15T02:38:40Z"
	next.RecentExits = []ContainerExit{
		{At: t0.Add(-10 * time.Minute), ExitCode: 1},
		{At: t0.Add(10 * time.Second), ExitCode: 137, OOMKilled: true},
		{At: t0.Add(30 * time.Second), ExitCode: 143},
		{At: t0.Add(35 * time.Second), ExitCode: 2},
	}
	var got []string
	sev := map[string]string{}
	for _, e := range DiffSnapshots(runtimeSnap(t0, old), runtimeSnap(t1, next)) {
		if strings.HasPrefix(e.Summary, containerRestartedPrefix) {
			got = append(got, e.Summary)
			sev[e.Summary] = e.Severity
		}
	}
	want := []string{
		"container restarted: vpn (killed for running out of memory)",
		"container restarted: vpn (stopped by SIGTERM, exit code 143)",
		"container restarted: vpn (crashed, exit code 2)",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("restart events:\n%v\nwant\n%v", got, want)
	}
	if sev[want[0]] != "warning" || sev[want[1]] != "info" || sev[want[2]] != "warning" {
		t.Fatalf("severities: %v", sev)
	}
}

func TestContainerRestartWithoutExitEvidenceAndBaselines(t *testing.T) {
	t0 := time.Date(2026, 10, 15, 2, 38, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	old := ContainerInfo{ID: "dddddddddddd", Name: "app", Image: "example/app:1", Status: "Up 2 hours", StartedAt: "2026-10-15T00:00:00Z"}
	next := old
	next.StartedAt = "2026-10-15T02:38:30Z"
	events := DiffSnapshots(runtimeSnap(t0, old), runtimeSnap(t1, next))
	if len(events) != 1 || events[0].Summary != "container restarted: app (how the previous run ended was not reported)" {
		t.Fatalf("fallback restart: %+v", events)
	}

	// Upgrading from a snapshot without runtime evidence must not replay.
	legacy := ContainerInfo{ID: "dddddddddddd", Name: "app", Image: "example/app:1", Status: "Up 2 hours"}
	upgraded := next
	upgraded.RecentExits = []ContainerExit{{At: t0.Add(10 * time.Second), ExitCode: 1}}
	if events := DiffSnapshots(Snapshot{SchemaVersion: snapshotSchemaVersion, Containers: []ContainerInfo{legacy}}, runtimeSnap(t1, upgraded)); len(events) != 0 {
		t.Fatalf("baseline replayed history: %+v", events)
	}

	// A container that stays down is already reported by its status change;
	// its exit is not reported a second time as a restart.
	down := old
	down.Status = "Exited (1) 10 seconds ago"
	down.RecentExits = []ContainerExit{{At: t0.Add(20 * time.Second), ExitCode: 1}}
	for _, e := range DiffSnapshots(runtimeSnap(t0, old), runtimeSnap(t1, down)) {
		if strings.HasPrefix(e.Summary, containerRestartedPrefix) {
			t.Fatalf("stopped container reported as restarted: %+v", e)
		}
	}
}

// Docker executes inspect templates on the decoded JSON map with missing
// keys as errors. Containers without a healthcheck have no State.Health.
func TestDockerInspectFormatToleratesMissingOptionalKeys(t *testing.T) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(dockerInspectFormat)
	if err != nil {
		t.Fatal(err)
	}
	noHealth := map[string]any{
		"Id": "abc", "Image": "sha256:x", "RestartCount": float64(2),
		"State":  map[string]any{"OOMKilled": false, "ExitCode": float64(0), "StartedAt": "2026-10-15T00:00:00Z", "FinishedAt": "0001-01-01T00:00:00Z"},
		"Config": map[string]any{"Labels": nil},
	}
	withHealth := map[string]any{
		"Id": "def", "Image": "sha256:y", "RestartCount": float64(0),
		"State":  map[string]any{"OOMKilled": true, "ExitCode": float64(137), "StartedAt": "s", "FinishedAt": "f", "Health": map[string]any{"Status": "unhealthy", "FailingStreak": float64(4)}},
		"Config": map[string]any{"Labels": map[string]any{"org.opencontainers.image.version": "1.2.3", "com.docker.compose.project": "media"}},
	}
	for name, data := range map[string]map[string]any{"no health": noHealth, "health": withHealth} {
		var b strings.Builder
		if err := tmpl.Execute(&b, data); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		cs := []ContainerInfo{{ID: data["Id"].(string)}}
		applyDockerInspect(cs, b.String())
		if cs[0].ImageID == "" || cs[0].StartedAt == "" {
			t.Fatalf("%s: output %q did not parse: %+v", name, b.String(), cs[0])
		}
		if name == "health" && (cs[0].HealthFailingStreak != 4 || cs[0].ImageVersion != "1.2.3" || !cs[0].OOMKilled || cs[0].ExitCode != 137 || cs[0].ComposeProject != "media") {
			t.Fatalf("health fields: %+v", cs[0])
		}
		if name == "no health" && (cs[0].ImageVersion != "" || cs[0].RestartCount != 2 || cs[0].ComposeProject != "") {
			t.Fatalf("no-health fields: %+v", cs[0])
		}
	}
}
