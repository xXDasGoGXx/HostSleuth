package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

const dockerDeploymentMode = "docker"

func deploymentMode() string {
	return strings.ToLower(strings.TrimSpace(os.Getenv("HOSTSLEUTH_MODE")))
}

func osReleasePath() string {
	if path := strings.TrimSpace(os.Getenv("HOSTSLEUTH_OS_RELEASE_PATH")); path != "" {
		return path
	}
	return "/etc/os-release"
}

func Collect(ctx context.Context) Snapshot {
	hostname, _ := os.Hostname()
	mode := deploymentMode()
	bootID, bootStartedAt := collectBootIdentity()
	s := Snapshot{
		SchemaVersion: snapshotSchemaVersion,
		CapturedAt:    time.Now().UTC(),
		Mode:          mode,
		Host: HostInfo{
			Hostname:      hostname,
			OS:            readOSRelease(),
			Kernel:        strings.TrimSpace(run(ctx, "uname", "-r")),
			Architecture:  runtime.GOARCH,
			CPUCount:      runtime.NumCPU(),
			MemoryTotal:   memoryTotal(),
			Uptime:        strings.TrimSpace(run(ctx, "uptime", "-p")),
			BootID:        bootID,
			BootStartedAt: bootStartedAt,
		},
	}
	s.Interfaces = collectInterfaces()
	if mode != dockerDeploymentMode {
		s.Filesystems = collectFilesystems()
	}
	s.Routes = lines(run(ctx, "ip", "route", "show"))
	s.Listeners = collectListeners(ctx)
	if mode != dockerDeploymentMode {
		s.Services = collectServices(ctx)
		s.ConfigFingerprints = collectConfigFingerprints()
	}
	s.Containers = collectContainers(ctx)
	s.PackageChanges = collectPackageChanges()
	return s
}

func readOSRelease() string {
	b, err := os.ReadFile(osReleasePath())
	if err != nil {
		return runtime.GOOS
	}
	var pretty string
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			pretty = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			break
		}
	}
	if pretty == "" {
		return runtime.GOOS
	}
	return pretty
}

func memoryTotal() string {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "MemTotal:") {
			return strings.TrimSpace(strings.TrimPrefix(s.Text(), "MemTotal:"))
		}
	}
	return ""
}

func collectInterfaces() []InterfaceInfo {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	out := make([]InterfaceInfo, 0, len(ifs))
	for _, iface := range ifs {
		addrs, _ := iface.Addrs()
		entry := InterfaceInfo{Name: iface.Name, State: "down"}
		if iface.Flags&net.FlagUp != 0 {
			entry.State = "up"
		}
		for _, a := range addrs {
			entry.Addresses = append(entry.Addresses, a.String())
		}
		sort.Strings(entry.Addresses)
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func collectFilesystems() []FilesystemInfo {
	b, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []FilesystemInfo
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) != 2 {
			continue
		}
		left := strings.Fields(parts[0])
		right := strings.Fields(parts[1])
		if len(left) < 5 || len(right) < 2 {
			continue
		}
		mountPoint := unescapeMountField(left[4])
		if seen[mountPoint] {
			continue
		}
		seen[mountPoint] = true
		entry := FilesystemInfo{MountPoint: mountPoint, FilesystemType: right[0], Source: unescapeMountField(right[1])}
		var st syscall.Statfs_t
		if syscall.Statfs(mountPoint, &st) == nil {
			entry.TotalBytes = st.Blocks * uint64(st.Bsize)
			entry.AvailableBytes = st.Bavail * uint64(st.Bsize)
		}
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MountPoint < out[j].MountPoint })
	return out
}

func unescapeMountField(s string) string {
	r := strings.NewReplacer("\\040", " ", "\\011", "\t", "\\012", "\n", "\\134", "\\")
	return r.Replace(s)
}

func collectListeners(ctx context.Context) []Listener {
	text := run(ctx, "ss", "-lntupH")
	var out []Listener
	for _, line := range lines(text) {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		proto := fields[0]
		addr := fields[4]
		process := ""
		if len(fields) > 6 {
			process = strings.Join(fields[6:], " ")
		}
		out = append(out, Listener{Protocol: proto, Address: addr, Process: process})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Protocol == out[j].Protocol {
			return out[i].Address < out[j].Address
		}
		return out[i].Protocol < out[j].Protocol
	})
	return out
}

func collectServices(ctx context.Context) []ServiceInfo {
	text := run(ctx, "systemctl", "list-units", "--type=service", "--all", "--no-legend", "--no-pager", "--plain")
	var out []ServiceInfo
	for _, line := range lines(text) {
		f := strings.Fields(line)
		if len(f) < 4 {
			continue
		}
		out = append(out, ServiceInfo{Name: f[0], Load: f[1], Active: f[2], Sub: f[3]})
	}
	return out
}

func collectContainers(ctx context.Context) []ContainerInfo {
	format := "{{.ID}}\\t{{.Names}}\\t{{.Image}}\\t{{.Status}}\\t{{.Ports}}\\t{{.Networks}}"
	text := run(ctx, "docker", "ps", "-a", "--format", format)
	var out []ContainerInfo
	for _, line := range lines(text) {
		f := strings.Split(line, "\t")
		if len(f) < 4 {
			continue
		}
		c := ContainerInfo{ID: f[0], Name: f[1], Image: f[2], Status: f[3]}
		if len(f) > 4 {
			c.Ports = f[4]
		}
		if len(f) > 5 {
			c.Networks = f[5]
		}
		out = append(out, c)
	}
	return out
}

func run(ctx context.Context, name string, args ...string) string {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return ""
	}
	return stdout.String()
}

func lines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func LocalListenerForPort(s Snapshot, port string) (Listener, bool) {
	needle := ":" + port
	for _, l := range s.Listeners {
		if strings.HasSuffix(strings.TrimSpace(l.Address), needle) {
			return l, true
		}
	}
	return Listener{}, false
}

func LocalListenerForTarget(s Snapshot, port string, resolvedIPs []string) (Listener, bool) {
	for _, listener := range s.Listeners {
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(listener.Protocol)), "tcp") {
			continue
		}
		host, listenerPort, err := net.SplitHostPort(strings.TrimSpace(listener.Address))
		if err != nil || listenerPort != port {
			continue
		}
		if listenerHostMatchesTarget(host, resolvedIPs) {
			return listener, true
		}
	}
	return Listener{}, false
}

func listenerHostMatchesTarget(host string, resolvedIPs []string) bool {
	host = strings.TrimSpace(host)
	if host == "*" {
		return len(resolvedIPs) > 0
	}
	if zone := strings.LastIndexByte(host, '%'); zone >= 0 {
		host = host[:zone]
	}
	listenerIP := net.ParseIP(host)
	if listenerIP == nil {
		return false
	}
	for _, value := range resolvedIPs {
		targetIP := net.ParseIP(value)
		if targetIP == nil {
			continue
		}
		if listenerIP.IsUnspecified() {
			if (listenerIP.To4() != nil) == (targetIP.To4() != nil) {
				return true
			}
			continue
		}
		if listenerIP.Equal(targetIP) {
			return true
		}
	}
	return false
}

func SnapshotSummary(s Snapshot) string {
	if s.Mode == dockerDeploymentMode {
		return fmt.Sprintf("%s | %s | mode=docker filesystems=unavailable services=unavailable listeners=%d containers=%d", s.Host.Hostname, s.Host.OS, len(s.Listeners), len(s.Containers))
	}
	return fmt.Sprintf("%s | %s | filesystems=%d services=%d listeners=%d containers=%d", s.Host.Hostname, s.Host.OS, len(s.Filesystems), len(s.Services), len(s.Listeners), len(s.Containers))
}
