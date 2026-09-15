package core

import (
	"net"
	"sort"
	"strconv"
	"strings"
)

const dockerEvidenceCandidateLimit = 4

type dockerPortRelation string

const (
	dockerPortUnavailable        dockerPortRelation = "unavailable"
	dockerPortPublishedTarget    dockerPortRelation = "published-target"
	dockerPortPublishedOtherBind dockerPortRelation = "published-other-bind"
	dockerPortInternalOnly       dockerPortRelation = "internal-only"
	dockerPortNoMatchingExposure dockerPortRelation = "no-match"
)

type dockerPortAssessment struct {
	Check    Check
	Relation dockerPortRelation
}

type dockerPortBinding struct {
	Container     string
	State         string
	Networks      string
	HostIP        string
	HostPort      string
	ContainerPort string
	Protocol      string
	Published     bool
	Raw           string
}

func dockerPortCheck(snap Snapshot, port string, resolvedIPs []string) Check {
	return assessDockerPort(snap, port, resolvedIPs).Check
}

func assessDockerPort(snap Snapshot, port string, resolvedIPs []string) dockerPortAssessment {
	if len(snap.Containers) == 0 {
		return dockerPortAssessment{
			Relation: dockerPortUnavailable,
			Check:    Check{Name: "docker-port", Status: "unknown", Evidence: "snapshot contains no Docker inventory; Docker may be absent or inaccessible"},
		}
	}

	bindings := dockerPortBindings(snap.Containers)
	published := make([]dockerPortBinding, 0)
	internalOnly := make([]dockerPortBinding, 0)
	for _, binding := range bindings {
		if binding.Protocol != "tcp" {
			continue
		}
		if binding.Published && binding.HostPort == port {
			published = append(published, binding)
		}
		if !binding.Published && binding.ContainerPort == port {
			internalOnly = append(internalOnly, binding)
		}
	}

	if len(published) > 0 {
		direct := make([]dockerPortBinding, 0)
		for _, binding := range published {
			if dockerBindingMatchesTarget(binding.HostIP, resolvedIPs) {
				direct = append(direct, binding)
			}
		}
		if len(direct) > 0 {
			return dockerPortAssessment{
				Relation: dockerPortPublishedTarget,
				Check: Check{
					Name:     "docker-port",
					Status:   "fail",
					Evidence: boundedEvidence("Docker publishes the requested host port but TCP still failed: "+formatDockerBindings(direct), 1024),
				},
			}
		}
		return dockerPortAssessment{
			Relation: dockerPortPublishedOtherBind,
			Check: Check{
				Name:     "docker-port",
				Status:   "fail",
				Evidence: boundedEvidence("Docker publishes TCP/"+port+" only on different host address(es): "+formatDockerBindings(published), 1024),
			},
		}
	}

	if len(internalOnly) > 0 {
		return dockerPortAssessment{
			Relation: dockerPortInternalOnly,
			Check: Check{
				Name:     "docker-port",
				Status:   "unknown",
				Evidence: boundedEvidence("Docker container port TCP/"+port+" is exposed internally but not published on the host: "+formatDockerBindings(internalOnly), 1024),
			},
		}
	}

	return dockerPortAssessment{
		Relation: dockerPortNoMatchingExposure,
		Check: Check{
			Name:     "docker-port",
			Status:   "pass",
			Evidence: "Docker inventory contains no TCP publication or internal exposure matching port " + port,
		},
	}
}

func dockerPortBindings(containers []ContainerInfo) []dockerPortBinding {
	bindings := make([]dockerPortBinding, 0)
	for _, container := range containers {
		state := stableContainerState(container.Status)
		for _, raw := range strings.Split(container.Ports, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			binding, ok := parseDockerPortBinding(raw)
			if !ok {
				continue
			}
			binding.Container = container.Name
			binding.State = state
			binding.Networks = strings.TrimSpace(container.Networks)
			bindings = append(bindings, binding)
		}
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].Container != bindings[j].Container {
			return bindings[i].Container < bindings[j].Container
		}
		if bindings[i].HostPort != bindings[j].HostPort {
			return bindings[i].HostPort < bindings[j].HostPort
		}
		if bindings[i].ContainerPort != bindings[j].ContainerPort {
			return bindings[i].ContainerPort < bindings[j].ContainerPort
		}
		return bindings[i].HostIP < bindings[j].HostIP
	})
	return bindings
}

func parseDockerPortBinding(raw string) (dockerPortBinding, bool) {
	binding := dockerPortBinding{Raw: raw}
	if left, right, ok := strings.Cut(raw, "->"); ok {
		containerPort, protocol, ok := splitDockerPortProtocol(right)
		if !ok {
			return dockerPortBinding{}, false
		}
		host, hostPort, err := net.SplitHostPort(strings.TrimSpace(left))
		if err != nil {
			return dockerPortBinding{}, false
		}
		binding.HostIP = host
		binding.HostPort = hostPort
		binding.ContainerPort = containerPort
		binding.Protocol = protocol
		binding.Published = true
		return binding, true
	}

	containerPort, protocol, ok := splitDockerPortProtocol(raw)
	if !ok {
		return dockerPortBinding{}, false
	}
	binding.ContainerPort = containerPort
	binding.Protocol = protocol
	return binding, true
}

func splitDockerPortProtocol(value string) (string, string, bool) {
	value = strings.TrimSpace(value)
	port, protocol, ok := strings.Cut(value, "/")
	if !ok {
		return "", "", false
	}
	port = strings.TrimSpace(port)
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	if port == "" || protocol == "" {
		return "", "", false
	}
	return port, protocol, true
}

func dockerBindingMatchesTarget(hostIP string, resolvedIPs []string) bool {
	bindingIP := net.ParseIP(hostIP)
	if bindingIP == nil {
		return false
	}
	for _, value := range resolvedIPs {
		targetIP := net.ParseIP(value)
		if targetIP == nil {
			continue
		}
		if bindingIP.IsUnspecified() {
			if (bindingIP.To4() != nil) == (targetIP.To4() != nil) {
				return true
			}
			continue
		}
		if bindingIP.Equal(targetIP) {
			return true
		}
	}
	return false
}

func formatDockerBindings(bindings []dockerPortBinding) string {
	limit := len(bindings)
	if limit > dockerEvidenceCandidateLimit {
		limit = dockerEvidenceCandidateLimit
	}
	parts := make([]string, 0, limit+1)
	for _, binding := range bindings[:limit] {
		state := binding.State
		if state == "" {
			state = "state unknown"
		}
		context := state
		if binding.Networks != "" {
			context += "; networks=" + binding.Networks
		}
		if binding.Published {
			host := binding.HostIP
			if strings.Contains(host, ":") {
				host = "[" + host + "]"
			}
			parts = append(parts, binding.Container+" ("+context+") "+host+":"+binding.HostPort+"->"+binding.ContainerPort+"/"+binding.Protocol)
		} else {
			parts = append(parts, binding.Container+" ("+context+") "+binding.ContainerPort+"/"+binding.Protocol+" internal-only")
		}
	}
	if len(bindings) > limit {
		parts = append(parts, "+"+strconv.Itoa(len(bindings)-limit)+" more")
	}
	return strings.Join(parts, " | ")
}
