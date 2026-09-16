package core

import "time"

const (
	packageHistorySchemaVersion    = 2
	configFingerprintSchemaVersion = 3
	snapshotSchemaVersion          = rebootStorySchemaVersion
)

type Snapshot struct {
	SchemaVersion      int                 `json:"schema_version"`
	CapturedAt         time.Time           `json:"captured_at"`
	Mode               string              `json:"mode,omitempty"`
	Host               HostInfo            `json:"host"`
	Interfaces         []InterfaceInfo     `json:"interfaces,omitempty"`
	Filesystems        []FilesystemInfo    `json:"filesystems,omitempty"`
	Routes             []string            `json:"routes,omitempty"`
	Listeners          []Listener          `json:"listeners,omitempty"`
	Services           []ServiceInfo       `json:"services,omitempty"`
	Containers         []ContainerInfo     `json:"containers,omitempty"`
	PackageChanges     []PackageChange     `json:"package_changes,omitempty"`
	ConfigFingerprints []ConfigFingerprint `json:"config_fingerprints,omitempty"`
}

type HostInfo struct {
	Hostname      string    `json:"hostname"`
	OS            string    `json:"os"`
	Kernel        string    `json:"kernel"`
	Architecture  string    `json:"architecture"`
	CPUCount      int       `json:"cpu_count"`
	MemoryTotal   string    `json:"memory_total,omitempty"`
	Uptime        string    `json:"uptime,omitempty"`
	BootID        string    `json:"boot_id,omitempty"`
	BootStartedAt time.Time `json:"boot_started_at,omitempty"`
}

type InterfaceInfo struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses,omitempty"`
	State     string   `json:"state,omitempty"`
}

type FilesystemInfo struct {
	MountPoint     string `json:"mount_point"`
	FilesystemType string `json:"filesystem_type"`
	Source         string `json:"source,omitempty"`
	TotalBytes     uint64 `json:"total_bytes,omitempty"`
	AvailableBytes uint64 `json:"available_bytes,omitempty"`
}

type Listener struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Process  string `json:"process,omitempty"`
}

type ServiceInfo struct {
	Name   string `json:"name"`
	Load   string `json:"load,omitempty"`
	Active string `json:"active,omitempty"`
	Sub    string `json:"sub,omitempty"`
}

type ContainerInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	Status   string `json:"status"`
	Ports    string `json:"ports,omitempty"`
	Networks string `json:"networks,omitempty"`
}

type PackageChange struct {
	At           time.Time `json:"at"`
	Action       string    `json:"action"`
	Name         string    `json:"name"`
	Architecture string    `json:"architecture,omitempty"`
	FromVersion  string    `json:"from_version,omitempty"`
	ToVersion    string    `json:"to_version,omitempty"`
}

type ConfigFingerprint struct {
	Path        string `json:"path"`
	State       string `json:"state"`
	Fingerprint string `json:"fingerprint,omitempty"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
}

type Event struct {
	SchemaVersion int       `json:"schema_version"`
	At            time.Time `json:"at"`
	Category      string    `json:"category"`
	Severity      string    `json:"severity"`
	Summary       string    `json:"summary"`
}

type Check struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type CertificateEvidence struct {
	Subject           string    `json:"subject"`
	Issuer            string    `json:"issuer"`
	Serial            string    `json:"serial"`
	SANs              []string  `json:"sans,omitempty"`
	ValidFrom         time.Time `json:"valid_from"`
	ValidUntil        time.Time `json:"valid_until"`
	DaysRemaining     int       `json:"days_remaining"`
	SHA256Fingerprint string    `json:"sha256_fingerprint"`
}

type TLSEvidence struct {
	ProbeConnected   bool                 `json:"probe_connected"`
	HandshakeStatus  string               `json:"handshake_status"`
	HandshakeError   string               `json:"handshake_error,omitempty"`
	ServerName       string               `json:"server_name,omitempty"`
	Protocol         string               `json:"protocol,omitempty"`
	CipherSuite      string               `json:"cipher_suite,omitempty"`
	HostnameStatus   string               `json:"hostname_status,omitempty"`
	HostnameEvidence string               `json:"hostname_evidence,omitempty"`
	TrustStatus      string               `json:"trust_status,omitempty"`
	TrustEvidence    string               `json:"trust_evidence,omitempty"`
	ChainSubjects    []string             `json:"chain_subjects,omitempty"`
	Certificate      *CertificateEvidence `json:"certificate,omitempty"`
}

type CertbotLineageEvidence struct {
	Name             string               `json:"name"`
	RenewalConfig    string               `json:"renewal_config"`
	CertificatePath  string               `json:"certificate_path,omitempty"`
	FullchainPath    string               `json:"fullchain_path,omitempty"`
	RenewalState     string               `json:"renewal_state,omitempty"`
	Certificate      *CertificateEvidence `json:"certificate,omitempty"`
	HostnameMatch    bool                 `json:"hostname_match"`
	ServedComparison string               `json:"served_comparison,omitempty"`
}

type CertbotEvidence struct {
	Status          string                   `json:"status"`
	Executable      string                   `json:"executable,omitempty"`
	TimerState      string                   `json:"timer_state,omitempty"`
	ServiceState    string                   `json:"service_state,omitempty"`
	Lineages        []CertbotLineageEvidence `json:"lineages,omitempty"`
	Comparison      string                   `json:"comparison,omitempty"`
	StaleServedCert bool                     `json:"stale_served_certificate,omitempty"`
}

type Diagnosis struct {
	Target     string           `json:"target"`
	StartedAt  time.Time        `json:"started_at"`
	Checks     []Check          `json:"checks"`
	Conclusion string           `json:"conclusion"`
	Confidence string           `json:"confidence"`
	TLS        *TLSEvidence     `json:"tls,omitempty"`
	Certbot    *CertbotEvidence `json:"certbot,omitempty"`
}
