package core

import "time"

type Snapshot struct {
	SchemaVersion  int                 `json:"schema_version"`
	CapturedAt     time.Time           `json:"captured_at"`
	Host           HostInfo            `json:"host"`
	Interfaces     []InterfaceInfo     `json:"interfaces,omitempty"`
	Filesystems    []FilesystemInfo    `json:"filesystems,omitempty"`
	Routes         []string            `json:"routes,omitempty"`
	Listeners      []Listener          `json:"listeners,omitempty"`
	Services       []ServiceInfo       `json:"services,omitempty"`
	Containers     []ContainerInfo     `json:"containers,omitempty"`
	ReverseProxies []ReverseProxyRoute `json:"reverse_proxies,omitempty"`
}

type HostInfo struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	CPUCount     int    `json:"cpu_count"`
	MemoryTotal  string `json:"memory_total,omitempty"`
	Uptime       string `json:"uptime,omitempty"`
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

type ReverseProxyRoute struct {
	Provider      string   `json:"provider"`
	Container     string   `json:"container,omitempty"`
	RouteID       string   `json:"route_id,omitempty"`
	Hostnames     []string `json:"hostnames,omitempty"`
	ListenPorts   []string `json:"listen_ports,omitempty"`
	BackendScheme string   `json:"backend_scheme,omitempty"`
	BackendHost   string   `json:"backend_host,omitempty"`
	BackendPort   string   `json:"backend_port,omitempty"`
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

type Diagnosis struct {
	Target     string    `json:"target"`
	StartedAt  time.Time `json:"started_at"`
	Checks     []Check   `json:"checks"`
	Conclusion string    `json:"conclusion"`
	Confidence string    `json:"confidence"`
}
