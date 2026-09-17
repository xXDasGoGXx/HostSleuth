package core

import (
	"regexp"
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
	evidenceIPv4Pattern             = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	evidenceMACPattern              = regexp.MustCompile(`(?i)\b(?:[0-9a-f]{2}:){5}[0-9a-f]{2}\b`)
	evidenceUUIDPattern             = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\b`)
	evidenceEmailPattern            = regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`)
	evidenceServicePattern          = regexp.MustCompile(`\b[A-Za-z0-9][A-Za-z0-9_.@:-]*\.service\b`)
	evidenceURLPattern              = regexp.MustCompile(`(?i)https?://[^\s"'<>]+`)
	evidenceAbsPathPattern          = regexp.MustCompile(`(?:^|[\s(=,:])(/[A-Za-z0-9._~@%+\-,/:]+)`)
	evidencePrivateKeyBlockPattern  = regexp.MustCompile(`(?is)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	evidencePrivateKeyHeaderPattern = regexp.MustCompile(`(?i)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----`)
	evidenceSecretKVPattern         = regexp.MustCompile(`(?i)\b(password|passwd|passphrase|token|access_token|refresh_token|id_token|api_key|apikey|secret|client_secret|authorization|cookie|set-cookie|session|sessionid|private_key|ssh_key)\b\s*[:=]\s*(?:(?:Bearer|Basic)\s+)?(?:"[^"]*"|'[^']*'|[^\s,;]+)`)
	evidenceAuthPattern             = regexp.MustCompile(`(?i)\b(Bearer|Basic)\s+[A-Za-z0-9._~+/=\-]+`)
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
	SchemaVersion       int                  `json:"schema_version"`
	HostSleuthVersion   string               `json:"hostsleuth_version"`
	GeneratedAt         time.Time            `json:"generated_at"`
	RedactionPolicy     string               `json:"redaction_policy"`
	IncludedClasses     []string             `json:"included_classes"`
	OmittedClasses      []string             `json:"omitted_classes"`
	RedactionCounts     map[string]int       `json:"redaction_counts"`
	RecordLimits        map[string]int       `json:"record_limits"`
	UnavailableEvidence []string             `json:"unavailable_evidence,omitempty"`
	Files               []EvidenceFileDigest `json:"files"`
	Warning             string               `json:"warning"`
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
