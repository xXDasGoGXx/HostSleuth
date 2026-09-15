package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxProxyConfigFiles = 256
	maxProxyConfigBytes = 256 * 1024
	maxProxyEvidence    = 4
)

type dockerMount struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
}

var dockerMountLookup = func(ctx context.Context, containerID string) ([]dockerMount, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(lookupCtx, "docker", "inspect", "--format", "{{json .Mounts}}", containerID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker mount inspection unavailable: %s", boundedEvidence(string(out), 384))
	}
	var mounts []dockerMount
	if err := json.Unmarshal(out, &mounts); err != nil {
		return nil, fmt.Errorf("decode docker mounts: %w", err)
	}
	return mounts, nil
}

func collectReverseProxies(ctx context.Context, containers []ContainerInfo) []ReverseProxyRoute {
	var routes []ReverseProxyRoute
	for _, container := range containers {
		if !isNginxProxyManager(container) || container.ID == "" {
			continue
		}
		mounts, err := dockerMountLookup(ctx, container.ID)
		if err != nil {
			continue
		}
		for _, mount := range mounts {
			if filepath.Clean(mount.Destination) != "/data" || !filepath.IsAbs(mount.Source) {
				continue
			}
			routes = append(routes, collectNPMRoutesFromDataDir(container, mount.Source)...)
			break
		}
	}
	sortReverseProxyRoutes(routes)
	return routes
}

func isNginxProxyManager(container ContainerInfo) bool {
	image := strings.ToLower(strings.TrimSpace(container.Image))
	name := strings.ToLower(strings.TrimSpace(container.Name))
	return strings.Contains(image, "jc21/nginx-proxy-manager") || strings.Contains(name, "nginx-proxy-manager")
}

func collectNPMRoutesFromDataDir(container ContainerInfo, dataDir string) []ReverseProxyRoute {
	proxyDir := filepath.Join(dataDir, "nginx", "proxy_host")
	info, err := os.Lstat(proxyDir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	entries, err := os.ReadDir(proxyDir)
	if err != nil {
		return nil
	}
	if len(entries) > maxProxyConfigFiles {
		entries = entries[:maxProxyConfigFiles]
	}
	var routes []ReverseProxyRoute
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".conf") {
			continue
		}
		path := filepath.Join(proxyDir, entry.Name())
		data, ok := readRegularFileBounded(path, maxProxyConfigBytes)
		if !ok {
			continue
		}
		route, ok := parseNPMProxyConfig(string(data))
		if !ok {
			continue
		}
		route.Provider = "nginx-proxy-manager"
		route.Container = container.Name
		route.RouteID = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		routes = append(routes, route)
	}
	sortReverseProxyRoutes(routes)
	return routes
}

func readRegularFileBounded(path string, limit int64) ([]byte, bool) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > limit {
		return nil, false
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, false
	}
	return data, true
}

func parseNPMProxyConfig(text string) (ReverseProxyRoute, bool) {
	route := ReverseProxyRoute{}
	hostnames := map[string]bool{}
	ports := map[string]bool{}

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "server_name "):
			value := directiveValue(line, "server_name")
			for _, hostname := range strings.Fields(value) {
				hostname = strings.TrimSpace(strings.Trim(hostname, "\"'"))
				if hostname != "" && hostname != "_" && !strings.HasPrefix(hostname, "$") {
					hostnames[hostname] = true
				}
			}
		case strings.HasPrefix(line, "listen "):
			if port := nginxListenPort(directiveValue(line, "listen")); port != "" {
				ports[port] = true
			}
		case strings.HasPrefix(line, "set $forward_scheme ") && route.BackendScheme == "":
			route.BackendScheme = cleanNginxValue(strings.TrimPrefix(line, "set $forward_scheme "))
		case strings.HasPrefix(line, "set $server ") && route.BackendHost == "":
			route.BackendHost = cleanNginxValue(strings.TrimPrefix(line, "set $server "))
		case strings.HasPrefix(line, "set $port ") && route.BackendPort == "":
			route.BackendPort = cleanNginxValue(strings.TrimPrefix(line, "set $port "))
		}
	}

	for hostname := range hostnames {
		route.Hostnames = append(route.Hostnames, hostname)
	}
	for port := range ports {
		route.ListenPorts = append(route.ListenPorts, port)
	}
	sort.Strings(route.Hostnames)
	sort.Slice(route.ListenPorts, func(i, j int) bool {
		a, _ := strconv.Atoi(route.ListenPorts[i])
		b, _ := strconv.Atoi(route.ListenPorts[j])
		return a < b
	})

	if route.BackendHost == "" || route.BackendPort == "" {
		return ReverseProxyRoute{}, false
	}
	if route.BackendScheme == "" {
		route.BackendScheme = "http"
	}
	return route, true
}

func directiveValue(line, directive string) string {
	value := strings.TrimSpace(strings.TrimPrefix(line, directive))
	if idx := strings.IndexByte(value, ';'); idx >= 0 {
		value = value[:idx]
	}
	return strings.TrimSpace(value)
}

func cleanNginxValue(value string) string {
	if idx := strings.IndexByte(value, ';'); idx >= 0 {
		value = value[:idx]
	}
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(value), "\"'"))
}

func nginxListenPort(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	endpoint := strings.TrimSpace(fields[0])
	if n, err := strconv.Atoi(endpoint); err == nil && n > 0 && n <= 65535 {
		return strconv.Itoa(n)
	}
	if host, port, err := net.SplitHostPort(endpoint); err == nil {
		_ = host
		if n, err := strconv.Atoi(port); err == nil && n > 0 && n <= 65535 {
			return strconv.Itoa(n)
		}
	}
	return ""
}

func sortReverseProxyRoutes(routes []ReverseProxyRoute) {
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Provider != routes[j].Provider {
			return routes[i].Provider < routes[j].Provider
		}
		if routes[i].Container != routes[j].Container {
			return routes[i].Container < routes[j].Container
		}
		return routes[i].RouteID < routes[j].RouteID
	})
}

func reverseProxyBackendCheck(snap Snapshot, targetHost, port string, resolvedIPs []string) (Check, bool) {
	var matches []ReverseProxyRoute
	for _, route := range snap.ReverseProxies {
		if route.BackendPort != port || !proxyBackendMatchesTarget(route.BackendHost, targetHost, resolvedIPs) {
			continue
		}
		matches = append(matches, route)
	}
	if len(matches) == 0 {
		return Check{}, false
	}
	sortReverseProxyRoutes(matches)
	return Check{
		Name:     "reverse-proxy",
		Status:   "unknown",
		Evidence: boundedEvidence("known reverse-proxy frontend(s) depend on this backend (candidate dependency evidence): "+formatReverseProxyRoutes(matches), 1536),
	}, true
}

func proxyBackendMatchesTarget(backendHost, targetHost string, resolvedIPs []string) bool {
	backendHost = strings.Trim(strings.TrimSpace(backendHost), "[]")
	if backendHost == "" {
		return false
	}
	if strings.EqualFold(backendHost, strings.Trim(strings.TrimSpace(targetHost), "[]")) {
		return true
	}
	backendIP := net.ParseIP(backendHost)
	if backendIP == nil {
		return false
	}
	for _, value := range resolvedIPs {
		if targetIP := net.ParseIP(value); targetIP != nil && backendIP.Equal(targetIP) {
			return true
		}
	}
	return false
}

func formatReverseProxyRoutes(routes []ReverseProxyRoute) string {
	limit := len(routes)
	if limit > maxProxyEvidence {
		limit = maxProxyEvidence
	}
	parts := make([]string, 0, limit+1)
	for _, route := range routes[:limit] {
		frontend := strings.Join(route.Hostnames, ",")
		if frontend == "" {
			frontend = "route " + route.RouteID
		}
		listens := strings.Join(route.ListenPorts, ",")
		if listens != "" {
			frontend += " listen=" + listens
		}
		provider := route.Provider
		if route.Container != "" {
			provider += "/" + route.Container
		}
		parts = append(parts, provider+" ["+frontend+"] -> "+route.BackendScheme+"://"+net.JoinHostPort(route.BackendHost, route.BackendPort))
	}
	if len(routes) > limit {
		parts = append(parts, "+"+strconv.Itoa(len(routes)-limit)+" more")
	}
	return strings.Join(parts, " | ")
}
