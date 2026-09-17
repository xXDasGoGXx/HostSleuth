package core

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvidenceBundleRedactsSensitiveEvidenceAndPreservesUsefulShape(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	now := time.Date(2026, 9, 17, 21, 30, 0, 0, time.UTC)
	containerID := strings.Repeat("a", 64)
	configFingerprint := strings.Repeat("b", 64)
	privateKey := "-----BEGIN PRIVATE KEY-----\nMIIE-SUPER-SECRET-MATERIAL\n-----END PRIVATE KEY-----"

	snapshot := Snapshot{
		SchemaVersion: snapshotSchemaVersion,
		CapturedAt:    now.Add(-time.Minute),
		Mode:          dockerDeploymentMode,
		Host: HostInfo{
			Hostname:      "openmediavault.internal.example",
			OS:            "Debian GNU/Linux 13",
			Kernel:        "7.0.14-test",
			Architecture:  "amd64",
			CPUCount:      32,
			MemoryTotal:   "32777756 kB",
			Uptime:        "up 2 days",
			BootID:        "123e4567-e89b-42d3-a456-426614174000",
			BootStartedAt: now.Add(-48 * time.Hour),
		},
		Interfaces: []InterfaceInfo{
			{Name: "lo", Addresses: []string{"127.0.0.1/8", "::1/128"}, State: "up"},
			{Name: "enp1s0", Addresses: []string{"192.168.2.181/24"}, State: "up"},
		},
		Filesystems: []FilesystemInfo{
			{MountPoint: "/srv/private/data", FilesystemType: "ext4", Source: "/dev/disk/by-uuid/123e4567-e89b-42d3-a456-426614174001", TotalBytes: 1000, AvailableBytes: 500},
		},
		Routes: []string{"default via 192.168.2.1 dev enp1s0", "192.168.2.0/24 dev enp1s0 proto kernel src 192.168.2.181"},
		Listeners: []Listener{
			{Protocol: "tcp", Address: "192.168.2.181:8787", Process: "users:((\"secret-api\",pid=1234,fd=7))"},
			{Protocol: "tcp", Address: "127.0.0.1:22", Process: "sshd"},
		},
		Services: []ServiceInfo{{Name: "secret-api.service", Load: "loaded", Active: "active", Sub: "running"}},
		Containers: []ContainerInfo{{
			ID:       containerID,
			Name:     "secret-container",
			Image:    "registry.internal.example/team/secret-container:1.2.3",
			Status:   "Up 5 minutes",
			Ports:    "192.168.2.181:8443->8443/tcp",
			Networks: "private-net,bridge",
		}},
		PackageChanges: []PackageChange{{At: now.Add(-time.Hour), Action: "upgrade", Name: "openssl", Architecture: "amd64", FromVersion: "1.0", ToVersion: "1.1"}},
		ConfigFingerprints: []ConfigFingerprint{
			{Path: "/etc/hosts", State: "present", Fingerprint: configFingerprint, SizeBytes: 100},
			{Path: "/srv/private/custom.conf", State: "present", Fingerprint: strings.Repeat("c", 64), SizeBytes: 200},
		},
	}
	if err := store.SaveSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}

	events := []Event{
		{At: now.Add(-30 * time.Minute), Category: "service", Severity: "warning", Summary: "service disappeared: secret-api.service (was active/running)"},
		{At: now.Add(-20 * time.Minute), Category: "container", Severity: "info", Summary: "container appeared: secret-container (running)"},
		{At: now.Add(-10 * time.Minute), Category: "listener", Severity: "info", Summary: "listener appeared: tcp 192.168.2.181:8787"},
		{At: now.Add(-5 * time.Minute), Category: "configuration", Severity: "info", Summary: "configuration changed: /srv/private/custom.conf"},
		{At: now.Add(-time.Minute), Category: "system", Severity: "warning", Summary: "password=hunter2 token=abc123token Authorization: Bearer xyz987 https://alice:pw@private.example/path?q=secret " + privateKey},
	}
	if err := store.AppendEvents(events); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendActionAudit(ActionAudit{
		At:       now.Add(-2 * time.Minute),
		Phase:    "completed",
		ActionID: actionServiceRestartID,
		Target:   "secret-api.service",
		Status:   "success",
		Summary:  "allowlisted service restarted",
		Before:   &ServiceRuntimeEvidence{ActiveState: "active", MainPID: 111, ControlGroup: "/system.slice/secret-api.service"},
		After:    &ServiceRuntimeEvidence{ActiveState: "active", MainPID: 222, ControlGroup: "/system.slice/secret-api.service"},
	}); err != nil {
		t.Fatal(err)
	}

	preview, err := BuildEvidencePreview(store, "v-test", now)
	if err != nil {
		t.Fatal(err)
	}
	if preview.RecordCounts["events"] != len(events) || preview.RecordCounts["action_audits"] != 1 {
		t.Fatalf("unexpected record counts: %#v", preview.RecordCounts)
	}
	if preview.RedactionCounts["host"] == 0 || preview.RedactionCounts["ipv4"] == 0 || preview.RedactionCounts["service"] == 0 {
		t.Fatalf("expected identifier redactions, got %#v", preview.RedactionCounts)
	}
	for _, entry := range preview.Files {
		if strings.HasSuffix(entry, ".zip") {
			t.Fatalf("preview must not write or describe an archive payload as a file entry: %q", entry)
		}
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, "*.zip")); len(matches) != 0 {
		t.Fatalf("preview wrote an archive: %#v", matches)
	}

	output := filepath.Join(dir, "bundle.zip")
	result, err := ExportEvidence(store, "v-test", output, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != output || result.SHA256 == "" || result.Bytes <= 0 {
		t.Fatalf("unexpected export result: %#v", result)
	}
	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("archive mode = %o, want 600", info.Mode().Perm())
	}

	zr, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	contents := map[string][]byte{}
	for _, file := range zr.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := ioReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		contents[file.Name] = data
	}
	for _, required := range []string{"manifest.json", "summary.txt", "snapshot.json", "events.json", "action-audit.json", "checksums.txt"} {
		if _, ok := contents[required]; !ok {
			t.Fatalf("missing %s from bundle", required)
		}
	}

	all := string(evidenceFileBytes(contents))
	for _, forbidden := range []string{
		"openmediavault.internal.example", "192.168.2.181", "192.168.2.1", "enp1s0",
		"secret-api.service", "secret-container", "private-net", "registry.internal.example",
		"/srv/private/data", "/srv/private/custom.conf", configFingerprint,
		"hunter2", "abc123token", "xyz987", "alice:pw", "q=secret", "MIIE-SUPER-SECRET-MATERIAL",
	} {
		if strings.Contains(all, forbidden) {
			t.Fatalf("bundle leaked %q", forbidden)
		}
	}
	for _, expected := range []string{"127.0.0.1", "::1", "openssl", "1.1", "host-", "ipv4-", "service-", "container-", "config-fingerprint-"} {
		if !strings.Contains(all, expected) {
			t.Fatalf("bundle missing useful/redacted evidence %q", expected)
		}
	}
	if !strings.Contains(all, "/etc/hosts") {
		t.Fatal("fixed product config path should remain literal")
	}

	var manifest EvidenceManifest
	if err := json.Unmarshal(contents["manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.RedactionPolicy != EvidenceRedactionPolicy || len(manifest.Files) == 0 {
		t.Fatalf("unexpected manifest: %#v", manifest)
	}
	for _, file := range manifest.Files {
		data, ok := contents[file.Name]
		if !ok {
			t.Fatalf("manifest references missing file %s", file.Name)
		}
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != file.SHA256 {
			t.Fatalf("checksum mismatch for %s: %s != %s", file.Name, got, file.SHA256)
		}
	}
}

func TestEvidenceRedactorSecretPatternsFailClosed(t *testing.T) {
	r := newEvidenceRedactor(evidenceSource{})
	input := "Authorization: Bearer abcdef123456 password=hunter2 cookie=sessionvalue api_key=topsecret"
	out := r.freeform(input)
	for _, forbidden := range []string{"abcdef123456", "hunter2", "sessionvalue", "topsecret"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("secret survived redaction: %q in %q", forbidden, out)
		}
	}
}

func TestEvidenceAliasesAreDeterministicForSortedObservedValues(t *testing.T) {
	a := evidenceSource{Snapshot: Snapshot{Host: HostInfo{Hostname: "z.example"}, Services: []ServiceInfo{{Name: "z.service"}, {Name: "a.service"}}}}
	b := evidenceSource{Snapshot: Snapshot{Host: HostInfo{Hostname: "z.example"}, Services: []ServiceInfo{{Name: "a.service"}, {Name: "z.service"}}}}
	ra := newEvidenceRedactor(a)
	rb := newEvidenceRedactor(b)
	if ra.register("service", "a.service") != rb.register("service", "a.service") || ra.register("service", "z.service") != rb.register("service", "z.service") {
		t.Fatalf("service aliases depend on source traversal order: %#v vs %#v", ra.aliases["service"], rb.aliases["service"])
	}
}

func TestEvidenceJSONLimitIsEnforced(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	snapshot := Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: time.Now().UTC(), Host: HostInfo{OS: strings.Repeat("x", EvidenceMaxJSONBytes+1024)}}
	if err := store.SaveSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildEvidencePreview(store, "v-test", time.Now().UTC()); err == nil || !strings.Contains(err.Error(), "JSON payload exceeds") {
		t.Fatalf("expected JSON size-limit error, got %v", err)
	}
}

func ioReadAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var out []byte
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				return out, nil
			}
			return out, err
		}
	}
}

func TestEvidenceFreeformRedactsPlainDomainIPv6AndTruncatedPrivateKey(t *testing.T) {
	source := evidenceSource{Events: []Event{{
		Category: "system",
		Summary:  "resolver controlplane.private.example reached [2001:db8:abcd::42]:443 then -----BEGIN OPENSSH PRIVATE KEY----- TRUNCATED-SECRET-MATERIAL",
	}}}
	r := newEvidenceRedactor(source)
	out := r.freeform(source.Events[0].Summary)
	for _, forbidden := range []string{"controlplane.private.example", "2001:db8:abcd::42", "TRUNCATED-SECRET-MATERIAL"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("sensitive free-form value survived redaction: %q in %q", forbidden, out)
		}
	}
	for _, expected := range []string{"host-", "ipv6-", "[REDACTED_PRIVATE_KEY]"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in redacted output %q", expected, out)
		}
	}
}

func TestEvidenceRecordLimitsAreEnforced(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	now := time.Date(2026, 9, 17, 22, 0, 0, 0, time.UTC)
	if err := store.SaveSnapshot(Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: now, Host: HostInfo{Hostname: "limit-test.example"}}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < EvidenceEventLimit+25; i++ {
		if err := store.AppendEvents([]Event{{SchemaVersion: 1, At: now.Add(time.Duration(i) * time.Second), Category: "system", Severity: "info", Summary: "bounded event"}}); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < EvidenceActionAuditLimit+25; i++ {
		if err := store.AppendActionAudit(ActionAudit{SchemaVersion: actionSchemaVersion, At: now.Add(time.Duration(i) * time.Second), Phase: "completed", ActionID: actionServiceRestartID, Target: "bounded.service", Status: "success", Summary: "bounded audit"}); err != nil {
			t.Fatal(err)
		}
	}
	preview, err := BuildEvidencePreview(store, "v-test", now)
	if err != nil {
		t.Fatal(err)
	}
	if preview.RecordCounts["events"] != EvidenceEventLimit {
		t.Fatalf("events = %d, want %d", preview.RecordCounts["events"], EvidenceEventLimit)
	}
	if preview.RecordCounts["action_audits"] != EvidenceActionAuditLimit {
		t.Fatalf("action audits = %d, want %d", preview.RecordCounts["action_audits"], EvidenceActionAuditLimit)
	}
}

func TestEvidenceExportFailureLeavesNoOutput(t *testing.T) {
	dir := t.TempDir()
	store := Store{Dir: dir}
	now := time.Now().UTC()
	if err := store.SaveSnapshot(Snapshot{SchemaVersion: snapshotSchemaVersion, CapturedAt: now, Host: HostInfo{OS: strings.Repeat("x", EvidenceMaxJSONBytes+1024)}}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "must-not-exist.zip")
	if _, err := ExportEvidence(store, "v-test", out, now); err == nil {
		t.Fatal("expected export to fail on oversized JSON")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("failed export left output behind: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(dir, ".hostsleuth-evidence-*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("failed export left temporary files: %#v", matches)
	}
}

func TestEvidencePolicyIsModeIndependent(t *testing.T) {
	base := Snapshot{Host: HostInfo{Hostname: "mode-private.example"}, Interfaces: []InterfaceInfo{{Name: "eth0", Addresses: []string{"10.20.30.40/24"}}}}
	one := base
	one.Mode = "mode-a"
	two := base
	two.Mode = "mode-b"
	rOne := newEvidenceRedactor(evidenceSource{Snapshot: one})
	rTwo := newEvidenceRedactor(evidenceSource{Snapshot: two})
	a := rOne.snapshot(one)
	b := rTwo.snapshot(two)
	if a.Host.Hostname != b.Host.Hostname || a.Interfaces[0].Addresses[0] != b.Interfaces[0].Addresses[0] {
		t.Fatalf("redaction differs by deployment mode: one=%#v two=%#v", a, b)
	}
}
