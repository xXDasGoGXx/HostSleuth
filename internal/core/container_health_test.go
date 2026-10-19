package core

import (
	"strings"
	"testing"
	"time"
)

func TestMaskHealthTextRemovesSecrets(t *testing.T) {
	cases := map[string]string{
		"curl -f http://admin:pa55@localhost:8080/ping":                 "curl -f http://[REDACTED]@localhost:8080/ping",
		"mysqladmin ping -h localhost -uroot -pS3cret":                  "mysqladmin ping -h localhost -uroot -p[REDACTED]",
		"redis-cli --pass hunter2 ping":                                 "redis-cli --pass [REDACTED] ping",
		"app check --api-key=abc123 --verbose":                          "app check --api-key=[REDACTED] --verbose",
		`curl -H "Authorization: Bearer eyJhbGciOi" http://localhost/`:  `curl -H "Authorization=[REDACTED] http://localhost/`,
		"wget -qO- http://localhost/health?token=abcdef":                "wget -qO- http://localhost/health?token=[REDACTED]",
		"curl -u user:pw http://x/":                                     "curl -u [REDACTED] http://x/",
		"test -f /staging/READY && test ! -s /staging/UNCLASSIFIED.txt": "test -f /staging/READY && test ! -s /staging/UNCLASSIFIED.txt",
		"pg_isready -U postgres -d app\n\n  || exit 1":                  "pg_isready -U postgres -d app || exit 1",
	}
	for in, want := range cases {
		if got := maskHealthText(in, healthCheckLimit); got != want {
			t.Errorf("maskHealthText(%q)\n got %q\nwant %q", in, got, want)
		}
	}
	if got := maskHealthText(strings.Repeat("x", 1000), 50); len([]rune(got)) > 50 {
		t.Fatalf("not bounded: %d", len(got))
	}
}

func TestHealthCheckCommandAndLastResult(t *testing.T) {
	if got := healthCheckCommand(`["CMD","curl","-f","http://localhost:9000/"]`); got != "curl -f http://localhost:9000/" {
		t.Fatalf("CMD: %q", got)
	}
	if got := healthCheckCommand(`["NONE"]`); got != "" {
		t.Fatalf("NONE: %q", got)
	}
	if got := healthCheckCommand(`not json`); got != "" {
		t.Fatalf("garbage: %q", got)
	}
	last := healthLastResult("0|2026-10-16T10:00:00Z|\"ok\"\x1e1|2026-10-16T10:05:00Z|\"\"\x1e")
	if last == nil || last.ExitCode != 1 || last.At != "2026-10-16T10:05:00Z" || last.Output != "" {
		t.Fatalf("newest result: %+v", last)
	}
	if healthLastResult("") != nil {
		t.Fatal("no results must give nil")
	}
}

func TestReplayExplainsTheFailingBackup(t *testing.T) {
	snap, events := loadReplayWeek(t)
	r := BuildReplay(snap, events)
	var why *HealthWhy
	for _, p := range r.Parts {
		if p.ID == "backup-snapshot" {
			why = p.Why
		} else if p.Why != nil {
			t.Fatalf("%s is healthy but has an explanation", p.ID)
		}
	}
	if why == nil {
		t.Fatal("unhealthy container has no explanation")
	}
	if !strings.HasPrefix(why.Summary, "Most likely /work/staging/SNAPSHOT_FAILED exists or /work/staging/UNCLASSIFIED.txt is not empty.") {
		t.Fatalf("summary: %q", why.Summary)
	}
	if strings.Join(why.Needs, "; ") != "/work/staging/READY exists; /work/staging/SNAPSHOT_FAILED does not exist; /work/staging/UNCLASSIFIED.txt is empty or missing" {
		t.Fatalf("needs: %q", why.Needs)
	}
	clues := strings.Join(why.Clues, "\n")
	for _, want := range []string{"It has not failed its health check before this week.", "Docker's last 210 checks in a row failed."} {
		if !strings.Contains(clues, want) {
			t.Errorf("missing clue %q in\n%s", want, clues)
		}
	}
	if why.LastExit == nil || *why.LastExit != 1 || why.LastAt == 0 {
		t.Fatalf("last result: %+v", why)
	}
}

// The uploader's check passes when it has uploaded the newest snapshot.
func uploaderCase(now time.Time) (ContainerInfo, map[string]string) {
	c := ContainerInfo{
		Name: "uploader", Status: "Up 2 days (unhealthy)", ComposeProject: "backup", HealthFailingStreak: 2,
		HealthCheck: `test -f /staging/READY && test -f /cache/last-ready && test "$(cat /staging/READY)" = "$(cat /cache/last-ready)"`,
		HealthLast:  &HealthCheckResult{At: now.Add(-time.Minute).Format(time.RFC3339Nano), ExitCode: 1},
	}
	return c, map[string]string{"uploader": "backup", "snapshotter": "backup", "web": "site"}
}

func TestExplainCatchingUpAfterAPeerChanged(t *testing.T) {
	now := time.Date(2026, 10, 16, 17, 0, 0, 0, time.UTC)
	c, stackOf := uploaderCase(now)
	m := func(d time.Duration) int64 { return now.Add(d).UnixMilli() }
	events := []ReplayEvent{
		{T: m(-48 * time.Hour), Kind: "crash", Container: "snapshotter"},
		{T: m(-48*time.Hour + 30*time.Second), Kind: "unhealthy", Container: "uploader"},
		{T: m(-48*time.Hour + 5*time.Minute + 30*time.Second), Kind: "recovered", Container: "uploader"},
		{T: m(-20*time.Minute - 40*time.Second), Kind: "update", Container: "web"},
		{T: m(-20 * time.Minute), Kind: "crash", Container: "snapshotter"},
		{T: m(-19 * time.Minute), Kind: "unhealthy", Container: "uploader"},
	}
	why := explainUnhealthy(c, "backup", events, stackOf)
	want := "It is probably catching up with snapshotter: its check compares /staging/READY with /cache/last-ready, and snapshotter restarted after being stopped or killed just before it started failing. It did the same once earlier this week after snapshotter changed, and cleared on its own each time."
	if why.Summary != want {
		t.Fatalf("summary\n got %q\nwant %q", why.Summary, want)
	}
	clues := strings.Join(why.Clues, "\n")
	for _, s := range []string{"It started failing about a minute after snapshotter (same stack) restarted after being stopped or killed.", "It also failed once earlier this week and cleared on its own after 5 minutes.", "Docker's last 2 checks in a row failed."} {
		if !strings.Contains(clues, s) {
			t.Errorf("missing clue %q in\n%s", s, clues)
		}
	}
	if len(why.Needs) != 3 || why.Needs[2] != "/staging/READY and /cache/last-ready hold the same value" {
		t.Fatalf("needs: %q", why.Needs)
	}
}

func TestExplainCatchingUpWithoutARecordedTrigger(t *testing.T) {
	now := time.Date(2026, 10, 16, 17, 0, 0, 0, time.UTC)
	c, stackOf := uploaderCase(now)
	m := func(d time.Duration) int64 { return now.Add(d).UnixMilli() }
	events := []ReplayEvent{
		{T: m(-72 * time.Hour), Kind: "unhealthy", Container: "uploader"},
		{T: m(-72*time.Hour + 5*time.Minute), Kind: "recovered", Container: "uploader"},
		{T: m(-24 * time.Hour), Kind: "unhealthy", Container: "uploader"},
		{T: m(-24*time.Hour + 10*time.Hour), Kind: "recovered", Container: "uploader"},
		{T: m(-10 * time.Minute), Kind: "unhealthy", Container: "uploader"},
	}
	why := explainUnhealthy(c, "backup", events, stackOf)
	if !strings.HasPrefix(why.Summary, "It is probably waiting to catch up: its check passes only when /staging/READY and /cache/last-ready hold the same value, and each time this happened earlier this week it cleared on its own after 5 minutes to 10 hours.") {
		t.Fatalf("summary: %q", why.Summary)
	}
	if !strings.Contains(strings.Join(why.Clues, "\n"), "Nothing else HostSleuth records changed in the ten minutes before it started failing.") {
		t.Fatalf("clues: %q", why.Clues)
	}
}

func TestExplainReadsWhatTheCheckPrinted(t *testing.T) {
	cases := []struct {
		exit         int
		output, want string
	}{
		{1, "curl: (7) Failed to connect to localhost port 8080 after 0 ms: Connection refused", "Nothing answered where the check connects"},
		{-1, "Health check exceeded timeout (10s)", "The check ran out of time"},
		{22, "curl: (22) The requested URL returned error: 503", "HTTP 503"},
		{127, "/bin/sh: 1: curl: not found", "The check's own command or a file it needs is missing"},
		{1, "wget: bad address 'db'", ""},
		{1, "getaddrinfo ENOTFOUND db: Temporary failure in name resolution", "name (DNS) problem"},
	}
	for _, tc := range cases {
		c := ContainerInfo{Name: "app", Status: "Up 1 hour (unhealthy)", HealthCheck: "curl -f http://localhost:8080/health", HealthLast: &HealthCheckResult{ExitCode: tc.exit, Output: tc.output}}
		why := explainUnhealthy(c, "app", nil, nil)
		if tc.want == "" {
			if why.Summary != "The check printed: “wget: bad address 'db'”." {
				t.Errorf("unknown output: %q", why.Summary)
			}
			continue
		}
		if !strings.Contains(why.Summary, tc.want) {
			t.Errorf("%q → %q, want %q", tc.output, why.Summary, tc.want)
		}
	}
}

func TestExplainWithoutCheckDetails(t *testing.T) {
	why := explainUnhealthy(ContainerInfo{Name: "old", Status: "Up 3 days (unhealthy)"}, "standalone", nil, nil)
	if !strings.HasPrefix(why.Summary, "Docker did not report the check or its results") || why.Clues == nil {
		t.Fatalf("%+v", why)
	}
}

func TestEvidenceBundleRedactsHealthCheckFields(t *testing.T) {
	snap := Snapshot{
		Host: HostInfo{Hostname: "nas-basement"},
		Containers: []ContainerInfo{{
			ID: "abc123", Name: "photos", Status: "Up 1 hour (unhealthy)",
			HealthCheck: "curl -f http://nas-basement:2342/api/v1/status",
			HealthLast:  &HealthCheckResult{ExitCode: 7, Output: "curl: (7) Failed to connect to nas-basement port 2342 from photos"},
		}},
	}
	r := newEvidenceRedactor(evidenceSource{Snapshot: snap})
	out := r.snapshot(snap)
	c := out.Containers[0]
	for _, s := range []string{c.HealthCheck, c.HealthLast.Output} {
		if strings.Contains(s, "nas-basement") || strings.Contains(s, "photos") {
			t.Fatalf("not redacted: %q", s)
		}
	}
	if snap.Containers[0].HealthLast.Output == c.HealthLast.Output {
		t.Fatal("redaction changed the caller's snapshot or did nothing")
	}
}
