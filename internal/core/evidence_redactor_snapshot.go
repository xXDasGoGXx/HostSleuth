package core

import (
	"fmt"
	"net"
	"strings"
)

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
