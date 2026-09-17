package core

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	EvidenceBundleSchemaVersion = 1
	EvidenceRedactionPolicy     = "redacted-v1"
	EvidenceEventLimit          = 200
	EvidenceActionAuditLimit    = 100
	EvidenceMaxPayloadBytes     = 2 * 1024 * 1024
	EvidenceMaxJSONBytes        = 1 * 1024 * 1024
	evidenceFreeformLimit       = 8 * 1024
)

var (
	evidenceIPv4Pattern = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	evidenceMACPattern  = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{2}:){5}[0-9a-f]{2}\b`)
	evidenceUUIDPattern = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\b`)
	evidenceEmailPattern = regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`)
	evidenceServicePattern = regexp.MustCompile(`\b[A-Za-z0-9][A-Za-z0-9_.@:-]*\.service\b`)
	evidenceURLPattern = regexp.MustCompile(`(?i)https?://[^\s"'<>]+`)
	evidenceAbsPathPattern = regexp.MustCompile(`(?:^|[\s(=,:])(/[A-Za-z0-9._~@%+\-,/:]+)`)
	evidencePrivateKeyBlockPattern = regexp.MustCompile(`(?is)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	evidencePrivateKeyHeaderPattern = regexp.MustCompile(`(?i)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----`)
	evidenceSecretKVPattern = regexp.MustCompile(`(?i)\b(password|passwd|passphrase|token|access_token|refresh_token|id_token|api_key|apikey|secret|client_secret|authorization|cookie|set-cookie|session|sessionid|private_key|ssh_key)\b\s*[:=]\s*(?:"[^"]*"|'[^']*'|[^\s,;]+)`)
	evidenceAuthPattern = regexp.MustCompile(`(?i)\b(Bearer|Basic)\s+[A-Za-z0-9._~+/=\-]+`)
)

var evidenceLiteralConfigPaths = map[string]bool{
	"/etc/hosts":              true,
	"/etc/fstab":              true,
	"/etc/ssh/sshd_config":    true,
	"/etc/docker/daemon.json": true,
	"/etc/nftables.conf":      true,
}

type EvidenceFileDigest struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

type EvidenceManifest struct {
	SchemaVersion       int               `json:"schema_version"`
	HostSleuthVersion   string            `json:"hostsleuth_version"`
	GeneratedAt         time.Time         `json:"generated_at"`
	RedactionPolicy     string            `json:"redaction_policy"`
	IncludedClasses     []string          `json:"included_classes"`
	OmittedClasses      []string          `json:"omitted_classes"`
	RedactionCounts     map[string]int    `json:"redaction_counts"`
	RecordLimits        map[string]int    `json:"record_limits"`
	UnavailableEvidence []string          `json:"unavailable_evidence,omitempty"`
	Files               []EvidenceFileDigest `json:"files"`
	Warning             string            `json:"warning"`
}

type EvidencePreview struct {
	SchemaVersion       int            `json:"schema_version"`
	GeneratedAt         time.Time      `json:"generated_at"`
	RedactionPolicy     string         `json:"redaction_policy"`
	IncludedClasses     []string       `json:"included_classes"`
	OmittedClasses      []string       `json:"omitted_classes"`
	RecordCounts        map[string]int `json:"record_counts"`
	RedactionCounts     map[string]int `json:"redaction_counts"`
	Files               []string       `json:"files"`
	EstimatedBytes      int            `json:"estimated_uncompressed_bytes"`
	UnavailableEvidence []string       `json:"unavailable_evidence,omitempty"`
	Warnings            []string       `json:"warnings"`
}

type EvidenceExportResult struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

type evidenceSource struct {
	Snapshot Snapshot
	Events   []Event
	Audits   []ActionAudit
}

type evidenceBundle struct {
	Preview EvidencePreview
	Files   map[string][]byte
}

type evidenceRedactor struct {
	aliases map[string]map[string]string
	counts  map[string]int
}

func DefaultEvidenceFilename(now time.Time) string {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return "hostsleuth-evidence-" + now.UTC().Format("20060102T150405Z") + ".zip"
}

func BuildEvidencePreview(store Store, hostSleuthVersion string, now time.Time) (EvidencePreview, error) {
	bundle, err := buildEvidenceBundle(store, hostSleuthVersion, now)
	if err != nil {
		return EvidencePreview{}, err
	}
	return bundle.Preview, nil
}

func ExportEvidence(store Store, hostSleuthVersion, outputPath string, now time.Time) (EvidenceExportResult, error) {
	bundle, err := buildEvidenceBundle(store, hostSleuthVersion, now)
	if err != nil {
		return EvidenceExportResult{}, err
	}
	if strings.TrimSpace(outputPath) == "" {
		outputPath = DefaultEvidenceFilename(now)
	}
	outputPath = filepath.Clean(outputPath)
	if _, err := os.Stat(outputPath); err == nil {
		return EvidenceExportResult{}, fmt.Errorf("evidence output already exists: %s", outputPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return EvidenceExportResult{}, err
	}

	dir := filepath.Dir(outputPath)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return EvidenceExportResult{}, err
	}
	tmp, err := os.CreateTemp(dir, ".hostsleuth-evidence-*.tmp")
	if err != nil {
		return EvidenceExportResult{}, err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}
	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}

	zw := zip.NewWriter(tmp)
	names := make([]string, 0, len(bundle.Files))
	for name := range bundle.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(0o600)
		writer, err := zw.CreateHeader(header)
		if err != nil {
			_ = zw.Close()
			cleanup()
			return EvidenceExportResult{}, err
		}
		if _, err := writer.Write(bundle.Files[name]); err != nil {
			_ = zw.Close()
			cleanup()
			return EvidenceExportResult{}, err
		}
	}
	if err := zw.Close(); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := os.Rename(tmpName, outputPath); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := os.Chmod(outputPath, 0o600); err != nil {
		_ = os.Remove(outputPath)
		return EvidenceExportResult{}, err
	}

	archive, err := os.ReadFile(outputPath)
	if err != nil {
		return EvidenceExportResult{}, err
	}
	sum := sha256.Sum256(archive)
	return EvidenceExportResult{Path: outputPath, SHA256: hex.EncodeToString(sum[:]), Bytes: int64(len(archive))}, nil
}

func buildEvidenceBundle(store Store, hostSleuthVersion string, now time.Time) (evidenceBundle, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	source, unavailable, err := loadEvidenceSource(store)
	if err != nil {
		return evidenceBundle{}, err
	}
	redactor := newEvidenceRedactor(source)
	redactedSnapshot := redactor.snapshot(source.Snapshot)
	redactedEvents := make([]Event, 0, len(source.Events))
	for _, event := range source.Events {
		redactedEvents = append(redactedEvents, redactor.event(event))
	}
	redactedAudits := make([]ActionAudit, 0, len(source.Audits))
	for _, audit := range source.Audits {
		redactedAudits = append(redactedAudits, redactor.audit(audit))
	}

	files := map[string][]byte{}
	if files["snapshot.json"], err = marshalEvidenceJSON(redactedSnapshot); err != nil {
		return evidenceBundle{}, fmt.Errorf("snapshot payload: %w", err)
	}
	if files["events.json"], err = marshalEvidenceJSON(redactedEvents); err != nil {
		return evidenceBundle{}, fmt.Errorf("events payload: %w", err)
	}
	if len(redactedAudits) > 0 {
		if files["action-audit.json"], err = marshalEvidenceJSON(redactedAudits); err != nil {
			return evidenceBundle{}, fmt.Errorf("action audit payload: %w", err)
		}
	}
	files["summary.txt"] = []byte(evidenceSummary(redactedSnapshot, len(redactedEvents), len(redactedAudits)))

	included := []string{"snapshot", "events"}
	if len(redactedAudits) > 0 {
		included = append(included, "action-audit")
	}
	omitted := []string{
		"raw-journal", "raw-action-command-output", "action-preview-command-and-confirmation",
		"arbitrary-file-contents", "arbitrary-configuration-contents", "workbench-arbitrary-inputs",
	}
	manifest := EvidenceManifest{
		SchemaVersion:       EvidenceBundleSchemaVersion,
		HostSleuthVersion:   hostSleuthVersion,
		GeneratedAt:         now,
		RedactionPolicy:     EvidenceRedactionPolicy,
		IncludedClasses:     included,
		OmittedClasses:      omitted,
		RedactionCounts:     cloneStringIntMap(redactor.counts),
		RecordLimits:        map[string]int{"events": EvidenceEventLimit, "action_audits": EvidenceActionAuditLimit, "max_uncompressed_bytes": EvidenceMaxPayloadBytes, "max_json_bytes": EvidenceMaxJSONBytes},
		UnavailableEvidence: unavailable,
		Warning:             "Redaction reduces disclosure risk but cannot guarantee anonymity; review the preview and bundle before sharing.",
	}
	payloadNames := make([]string, 0, len(files))
	for name := range files {
		payloadNames = append(payloadNames, name)
	}
	sort.Strings(payloadNames)
	for _, name := range payloadNames {
		sum := sha256.Sum256(files[name])
		manifest.Files = append(manifest.Files, EvidenceFileDigest{Name: name, SHA256: hex.EncodeToString(sum[:]), Bytes: len(files[name])})
	}
	manifestBytes, err := marshalEvidenceJSON(manifest)
	if err != nil {
		return evidenceBundle{}, fmt.Errorf("manifest payload: %w", err)
	}
	files["manifest.json"] = manifestBytes
	files["checksums.txt"] = evidenceChecksums(files)

	total := 0
	fileNames := make([]string, 0, len(files))
	for name, data := range files {
		total += len(data)
		fileNames = append(fileNames, name)
	}
	if total > EvidenceMaxPayloadBytes {
		return evidenceBundle{}, fmt.Errorf("evidence payload exceeds %d byte limit", EvidenceMaxPayloadBytes)
	}
	sort.Strings(fileNames)
	preview := EvidencePreview{
		SchemaVersion:       EvidenceBundleSchemaVersion,
		GeneratedAt:         now,
		RedactionPolicy:     EvidenceRedactionPolicy,
		IncludedClasses:     included,
		OmittedClasses:      omitted,
		RecordCounts:        map[string]int{"events": len(redactedEvents), "action_audits": len(redactedAudits)},
		RedactionCounts:     cloneStringIntMap(redactor.counts),
		Files:               fileNames,
		EstimatedBytes:      total,
		UnavailableEvidence: unavailable,
		Warnings:            []string{manifest.Warning},
	}
	return evidenceBundle{Preview: preview, Files: files}, nil
}

func loadEvidenceSource(store Store) (evidenceSource, []string, error) {
	snapshot, err := store.LoadSnapshot()
	if err != nil {
		return evidenceSource{}, nil, fmt.Errorf("load snapshot: %w", err)
	}
	events, err := store.ReadEvents(EvidenceEventLimit)
	if err != nil {
		return evidenceSource{}, nil, fmt.Errorf("load events: %w", err)
	}
	audits, err := store.ReadActionAudits(EvidenceActionAuditLimit)
	if err != nil {
		return evidenceSource{}, nil, fmt.Errorf("load action audits: %w", err)
	}
	unavailable := []string{}
	if len(events) == 0 {
		unavailable = append(unavailable, "no retained events were available")
	}
	if len(audits) == 0 {
		unavailable = append(unavailable, "no Safe Action audit records were available")
	}
	return evidenceSource{Snapshot: snapshot, Events: events, Audits: audits}, unavailable, nil
}

func marshalEvidenceJSON(value any) ([]byte, error) {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	if len(b) > EvidenceMaxJSONBytes {
		return nil, fmt.Errorf("JSON payload exceeds %d byte limit", EvidenceMaxJSONBytes)
	}
	return b, nil
}

func evidenceSummary(snapshot Snapshot, events, audits int) string {
	return fmt.Sprintf("HostSleuth redacted evidence bundle\nMode: %s\nCaptured snapshot: %s\nEvents included: %d\nAction audits included: %d\nPolicy: %s\nReview manifest.json before sharing.\n",
		snapshot.Mode, snapshot.CapturedAt.UTC().Format(time.RFC3339), events, audits, EvidenceRedactionPolicy)
}

func evidenceChecksums(files map[string][]byte) []byte {
	names := make([]string, 0, len(files))
	for name := range files {
		if name != "checksums.txt" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		fmt.Fprintf(&b, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	return []byte(b.String())
}

func newEvidenceRedactor(source evidenceSource) *evidenceRedactor {
	r := &evidenceRedactor{aliases: map[string]map[string]string{}, counts: map[string]int{}}
	sets := map[string]map[string]bool{}
	add := func(category, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if sets[category] == nil {
			sets[category] = map[string]bool{}
		}
		sets[category][value] = true
	}

	add("host", source.Snapshot.Host.Hostname)
	add("id", source.Snapshot.Host.BootID)
	for _, iface := range source.Snapshot.Interfaces {
		if iface.Name != "lo" {
			add("interface", iface.Name)
		}
		for _, address := range iface.Addresses {
			collectAddressAlias(add, address)
		}
	}
	for _, fs := range source.Snapshot.Filesystems {
		collectPathAlias(add, fs.MountPoint)
		collectPathAlias(add, fs.Source)
		collectGenericAliases(add, fs.Source)
	}
	for _, route := range source.Snapshot.Routes {
		collectGenericAliases(add, route)
	}
	for _, listener := range source.Snapshot.Listeners {
		collectAddressAlias(add, listener.Address)
		collectGenericAliases(add, listener.Process)
	}
	for _, service := range source.Snapshot.Services {
		add("service", service.Name)
	}
	for _, container := range source.Snapshot.Containers {
		add("id", container.ID)
		add("container", container.Name)
		collectImageAlias(add, container.Image)
		for _, network := range splitList(container.Networks) {
			if network != "bridge" && network != "host" && network != "none" {
				add("network", network)
			}
		}
		collectGenericAliases(add, container.Ports)
	}
	for _, fingerprint := range source.Snapshot.ConfigFingerprints {
		if !evidenceLiteralConfigPaths[fingerprint.Path] {
			collectPathAlias(add, fingerprint.Path)
		}
		add("config-fingerprint", fingerprint.Fingerprint)
	}
	for _, event := range source.Events {
		collectEventAliases(add, event)
		collectGenericAliases(add, event.Summary)
	}
	for _, audit := range source.Audits {
		add("service", audit.Target)
		collectGenericAliases(add, audit.Summary)
		if audit.Before != nil {
			collectGenericAliases(add, audit.Before.ControlGroup)
		}
		if audit.After != nil {
			collectGenericAliases(add, audit.After.ControlGroup)
		}
	}

	categories := make([]string, 0, len(sets))
	for category := range sets {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	for _, category := range categories {
		values := make([]string, 0, len(sets[category]))
		for value := range sets[category] {
			values = append(values, value)
		}
		sort.Strings(values)
		for _, value := range values {
			r.register(category, value)
		}
	}
	return r
}

func (r *evidenceRedactor) register(category, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if r.aliases[category] == nil {
		r.aliases[category] = map[string]string{}
	}
	if alias := r.aliases[category][value]; alias != "" {
		return alias
	}
	n := len(r.aliases[category]) + 1
	var alias string
	switch category {
	case "host":
		alias = fmt.Sprintf("host-%02d.invalid", n)
	case "ipv4":
		alias = fmt.Sprintf("ipv4-%02d", n)
	case "ipv6":
		alias = fmt.Sprintf("ipv6-%02d", n)
	case "mac":
		alias = fmt.Sprintf("mac-%02d", n)
	case "path":
		alias = fmt.Sprintf("path-%02d", n)
	case "service":
		alias = fmt.Sprintf("service-%02d.service", n)
	case "container":
		alias = fmt.Sprintf("container-%02d", n)
	case "network":
		alias = fmt.Sprintf("network-%02d", n)
	case "interface":
		alias = fmt.Sprintf("interface-%02d", n)
	case "image":
		alias = fmt.Sprintf("image-%02d", n)
	case "email":
		alias = fmt.Sprintf("email-%02d.invalid", n)
	case "config-fingerprint":
		alias = fmt.Sprintf("config-fingerprint-%02d", n)
	case "cert-fingerprint":
		alias = fmt.Sprintf("certificate-fingerprint-%02d", n)
	default:
		alias = fmt.Sprintf("id-%02d", n)
	}
	r.aliases[category][value] = alias
	return alias
}

func (r *evidenceRedactor) alias(category, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	alias := r.register(category, value)
	r.counts[category]++
	return alias
}

func (r *evidenceRedactor) snapshot(in Snapshot) Snapshot {
	out := in
	out.Host.Hostname = r.alias("host", in.Host.Hostname)
	out.Host.BootID = r.alias("id", in.Host.BootID)
	out.Interfaces = make([]InterfaceInfo, len(in.Interfaces))
	for i, iface := range in.Interfaces {
		out.Interfaces[i] = iface
		if iface.Name != "lo" {
			out.Interfaces[i].Name = r.alias("interface", iface.Name)
		}
		out.Interfaces[i].Addresses = make([]string, len(iface.Addresses))
		for j, address := range iface.Addresses {
			out.Interfaces[i].Addresses[j] = r.address(address)
		}
	}
	out.Filesystems = make([]FilesystemInfo, len(in.Filesystems))
	for i, fs := range in.Filesystems {
		out.Filesystems[i] = fs
		out.Filesystems[i].MountPoint = r.path(fs.MountPoint)
		out.Filesystems[i].Source = r.freeform(fs.Source)
	}
	out.Routes = make([]string, len(in.Routes))
	for i, route := range in.Routes {
		out.Routes[i] = r.freeform(route)
	}
	out.Listeners = make([]Listener, len(in.Listeners))
	for i, listener := range in.Listeners {
		out.Listeners[i] = listener
		out.Listeners[i].Address = r.address(listener.Address)
		out.Listeners[i].Process = r.freeform(listener.Process)
	}
	out.Services = make([]ServiceInfo, len(in.Services))
	for i, service := range in.Services {
		out.Services[i] = service
		out.Services[i].Name = r.alias("service", service.Name)
	}
	out.Containers = make([]ContainerInfo, len(in.Containers))
	for i, container := range in.Containers {
		out.Containers[i] = container
		out.Containers[i].ID = r.alias("id", container.ID)
		out.Containers[i].Name = r.alias("container", container.Name)
		out.Containers[i].Image = r.image(container.Image)
		out.Containers[i].Status = r.freeform(container.Status)
		out.Containers[i].Ports = r.freeform(container.Ports)
		networks := splitList(container.Networks)
		for j, network := range networks {
			if network != "bridge" && network != "host" && network != "none" {
				networks[j] = r.alias("network", network)
			}
		}
		out.Containers[i].Networks = strings.Join(networks, ",")
	}
	out.ConfigFingerprints = make([]ConfigFingerprint, len(in.ConfigFingerprints))
	for i, fingerprint := range in.ConfigFingerprints {
		out.ConfigFingerprints[i] = fingerprint
		out.ConfigFingerprints[i].Path = r.path(fingerprint.Path)
		out.ConfigFingerprints[i].Fingerprint = r.alias("config-fingerprint", fingerprint.Fingerprint)
	}
	return out
}

func (r *evidenceRedactor) event(in Event) Event {
	out := in
	out.Summary = r.freeform(in.Summary)
	return out
}

func (r *evidenceRedactor) audit(in ActionAudit) ActionAudit {
	out := in
	out.Target = r.alias("service", in.Target)
	out.Summary = r.freeform(in.Summary)
	out.Before = r.runtime(in.Before)
	out.After = r.runtime(in.After)
	return out
}

func (r *evidenceRedactor) runtime(in *ServiceRuntimeEvidence) *ServiceRuntimeEvidence {
	if in == nil {
		return nil
	}
	out := *in
	out.ControlGroup = r.freeform(in.ControlGroup)
	return &out
}

func (r *evidenceRedactor) path(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || evidenceLiteralConfigPaths[value] {
		return value
	}
	if strings.HasPrefix(value, "/") {
		return r.alias("path", value)
	}
	return r.freeform(value)
}

func (r *evidenceRedactor) address(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if ip, network, err := net.ParseCIDR(value); err == nil {
		_ = ip
		prefix, _ := network.Mask.Size()
		return r.ip(strings.Split(value, "/")[0]) + fmt.Sprintf("/%d", prefix)
	}
	if host, port, err := net.SplitHostPort(value); err == nil {
		return net.JoinHostPort(r.host(host), port)
	}
	return r.host(value)
}

func (r *evidenceRedactor) host(value string) string {
	value = strings.Trim(strings.TrimSpace(value), "[]")
	if value == "" || value == "*" || value == "localhost" || value == "0.0.0.0" || value == "::" || value == "127.0.0.1" || value == "::1" {
		return value
	}
	if strings.Contains(value, "%") {
		parts := strings.SplitN(value, "%", 2)
		return r.ip(parts[0]) + "%" + r.interfaceName(parts[1])
	}
	if net.ParseIP(value) != nil {
		return r.ip(value)
	}
	return r.alias("host", value)
}

func (r *evidenceRedactor) ip(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return r.alias("id", value)
	}
	if ip.IsLoopback() || ip.IsUnspecified() {
		return value
	}
	if ip.To4() != nil {
		return r.alias("ipv4", value)
	}
	return r.alias("ipv6", value)
}

func (r *evidenceRedactor) interfaceName(value string) string {
	if value == "" || value == "lo" {
		return value
	}
	return r.alias("interface", value)
}

func (r *evidenceRedactor) image(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	base := value
	suffix := ""
	if at := strings.LastIndex(base, "@"); at >= 0 {
		suffix = base[at:]
		base = base[:at]
	} else if colon := strings.LastIndex(base, ":"); colon > strings.LastIndex(base, "/") {
		suffix = base[colon:]
		base = base[:colon]
	}
	return r.alias("image", base) + suffix
}

func (r *evidenceRedactor) freeform(value string) string {
	if strings.TrimSpace(value) == "" {
		return value
	}
	out := evidencePrivateKeyBlockPattern.ReplaceAllString(value, "[REDACTED_PRIVATE_KEY]")
	out = evidencePrivateKeyHeaderPattern.ReplaceAllString(out, "[REDACTED_PRIVATE_KEY]")
	out = evidenceSecretKVPattern.ReplaceAllString(out, "$1=[REDACTED_SECRET]")
	out = evidenceAuthPattern.ReplaceAllString(out, "$1 [REDACTED_SECRET]")
	out = evidenceURLPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.redactURL(raw) })
	out = evidenceEmailPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("email", raw) })

	out = r.replaceKnownAliases(out)
	out = evidenceMACPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("mac", raw) })
	out = evidenceUUIDPattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("id", raw) })
	out = evidenceServicePattern.ReplaceAllStringFunc(out, func(raw string) string { return r.alias("service", raw) })
	out = evidenceIPv4Pattern.ReplaceAllStringFunc(out, func(raw string) string {
		if net.ParseIP(raw) == nil {
			return raw
		}
		return r.ip(raw)
	})
	out = evidenceAbsPathPattern.ReplaceAllStringFunc(out, func(match string) string {
		idx := strings.Index(match, "/")
		if idx < 0 {
			return match
		}
		prefix, path := match[:idx], match[idx:]
		path = strings.TrimRight(path, ".);]")
		return prefix + r.path(path)
	})
	if len(out) > evidenceFreeformLimit {
		out = out[:evidenceFreeformLimit] + " [truncated after redaction]"
	}
	return out
}

func (r *evidenceRedactor) replaceKnownAliases(value string) string {
	type replacement struct{ original, alias, category string }
	var replacements []replacement
	for category, values := range r.aliases {
		for original, alias := range values {
			if original == "" || len(original) < 3 {
				continue
			}
			replacements = append(replacements, replacement{original: original, alias: alias, category: category})
		}
	}
	sort.Slice(replacements, func(i, j int) bool { return len(replacements[i].original) > len(replacements[j].original) })
	out := value
	for _, item := range replacements {
		count := strings.Count(out, item.original)
		if count == 0 {
			continue
		}
		out = strings.ReplaceAll(out, item.original, item.alias)
		r.counts[item.category] += count
	}
	return out
}

func (r *evidenceRedactor) redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		r.counts["url"]++
		return "[REDACTED_URL]"
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	host := u.Hostname()
	port := u.Port()
	redactedHost := r.host(host)
	if port != "" {
		u.Host = net.JoinHostPort(redactedHost, port)
	} else {
		u.Host = redactedHost
	}
	if u.Path != "" && u.Path != "/" {
		u.Path = "/[redacted]"
		u.RawPath = ""
	}
	r.counts["url"]++
	return u.String()
}

func collectAddressAlias(add func(string, string), value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if ip, _, err := net.ParseCIDR(value); err == nil {
		collectIPAlias(add, ip.String())
		return
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		host = strings.Trim(host, "[]")
		if strings.Contains(host, "%") {
			parts := strings.SplitN(host, "%", 2)
			collectIPAlias(add, parts[0])
			if parts[1] != "lo" {
				add("interface", parts[1])
			}
			return
		}
		collectIPAlias(add, host)
		return
	}
	collectIPAlias(add, strings.Trim(value, "[]"))
}

func collectIPAlias(add func(string, string), value string) {
	ip := net.ParseIP(value)
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
		return
	}
	if ip.To4() != nil {
		add("ipv4", value)
	} else {
		add("ipv6", value)
	}
}

func collectPathAlias(add func(string, string), value string) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "/") && !evidenceLiteralConfigPaths[value] {
		add("path", value)
	}
}

func collectImageAlias(add func(string, string), value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if at := strings.LastIndex(value, "@"); at >= 0 {
		value = value[:at]
	} else if colon := strings.LastIndex(value, ":"); colon > strings.LastIndex(value, "/") {
		value = value[:colon]
	}
	add("image", value)
}

func collectGenericAliases(add func(string, string), value string) {
	for _, raw := range evidenceIPv4Pattern.FindAllString(value, -1) {
		collectIPAlias(add, raw)
	}
	for _, raw := range evidenceMACPattern.FindAllString(value, -1) {
		add("mac", raw)
	}
	for _, raw := range evidenceUUIDPattern.FindAllString(value, -1) {
		add("id", raw)
	}
	for _, raw := range evidenceEmailPattern.FindAllString(value, -1) {
		add("email", raw)
	}
	for _, raw := range evidenceServicePattern.FindAllString(value, -1) {
		add("service", raw)
	}
	for _, match := range evidenceAbsPathPattern.FindAllStringSubmatch(value, -1) {
		if len(match) > 1 {
			collectPathAlias(add, strings.TrimRight(match[1], ".);]"))
		}
	}
}

func collectEventAliases(add func(string, string), event Event) {
	summary := strings.TrimSpace(event.Summary)
	switch event.Category {
	case "service", "container":
		prefixes := []string{event.Category + " appeared: ", event.Category + " disappeared: ", event.Category + " changed: "}
		for _, prefix := range prefixes {
			if strings.HasPrefix(summary, prefix) {
				value := strings.TrimPrefix(summary, prefix)
				if idx := strings.IndexAny(value, "(:"); idx >= 0 {
					value = strings.TrimSpace(value[:idx])
				}
				if event.Category == "service" {
					add("service", value)
				} else {
					add("container", value)
				}
				break
			}
		}
	case "configuration":
		if idx := strings.Index(summary, ": "); idx >= 0 {
			collectPathAlias(add, strings.TrimSpace(summary[idx+2:]))
		}
	}
}

func splitList(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if value := strings.TrimSpace(field); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func cloneStringIntMap(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func evidenceFileBytes(files map[string][]byte) []byte {
	var b bytes.Buffer
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		b.WriteString(name)
		b.WriteByte('\n')
		b.Write(files[name])
		b.WriteByte('\n')
	}
	return b.Bytes()
}
