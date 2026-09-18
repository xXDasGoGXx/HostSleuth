package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	permissionCommandOutputLimit = 12 * 1024
	permissionDockerMountLimit   = 16
	permissionContainerLimit     = 64
)

type PermissionIdentity struct {
	PID               int      `json:"pid"`
	UID               uint32   `json:"uid"`
	GID               uint32   `json:"gid"`
	SupplementaryGIDs []uint32 `json:"supplementary_gids,omitempty"`
	ConfiguredUser    string   `json:"configured_user,omitempty"`
	ConfiguredGroup   string   `json:"configured_group,omitempty"`
	WorkingDirectory  string   `json:"working_directory,omitempty"`
	Executable        string   `json:"executable,omitempty"`
	EffectiveCaps     string   `json:"effective_caps,omitempty"`
}

type PermissionNode struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	UID       uint32 `json:"uid"`
	GID       uint32 `json:"gid"`
	Mode      string `json:"mode"`
	Relation  string `json:"relation"`
	IsTarget  bool   `json:"is_target,omitempty"`
	StatError string `json:"stat_error,omitempty"`
}

type PermissionDecision struct {
	Path      string `json:"path"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Relation  string `json:"relation,omitempty"`
	Required  string `json:"required,omitempty"`
	Evidence  string `json:"evidence"`
}

type DockerBindMountEvidence struct {
	Container   string `json:"container"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	ReadWrite   bool   `json:"read_write"`
	MatchSide   string `json:"match_side"`
}

type PermissionStory struct {
	Service          string                    `json:"service"`
	RequestedPath    string                    `json:"requested_path"`
	ResolvedPath     string                    `json:"resolved_path,omitempty"`
	StartedAt        time.Time                 `json:"started_at"`
	Status           string                    `json:"status"`
	Conclusion       string                    `json:"conclusion"`
	FirstProblem     string                    `json:"first_problem,omitempty"`
	Identity         *PermissionIdentity       `json:"identity,omitempty"`
	Chain            []PermissionNode          `json:"chain,omitempty"`
	Decisions        []PermissionDecision      `json:"decisions,omitempty"`
	DockerBindMounts []DockerBindMountEvidence `json:"docker_bind_mounts,omitempty"`
	ScopeNotes       []string                  `json:"scope_notes,omitempty"`
}

var permissionSystemdLookup = func(ctx context.Context, unit string) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("systemctl", "/usr/bin/systemctl", "/bin/systemctl")
	return runBoundedCommand(lookupCtx, permissionCommandOutputLimit, command, "show", unit, "--no-pager",
		"--property=LoadState,ActiveState,MainPID,User,Group,WorkingDirectory")
}

var permissionReadFile = os.ReadFile
var permissionReadlink = os.Readlink
var permissionStat = os.Stat
var permissionDockerInspect = inspectRelevantDockerBindMounts

func BuildPermissionStory(ctx context.Context, service, requestedPath string, snap Snapshot) PermissionStory {
	unit := normalizeServiceUnit(service)
	story := PermissionStory{
		Service:       unit,
		RequestedPath: strings.TrimSpace(requestedPath),
		StartedAt:     time.Now().UTC(),
		ScopeNotes: []string{
			"Access reasoning uses Linux ownership/mode metadata and the observed effective process identity.",
			"POSIX ACLs, Linux Security Modules, namespaces, and application-level checks can still change the real outcome.",
			"No file contents are read and no recursive filesystem crawl is performed.",
		},
	}
	if unit == "" {
		return permissionStoryInputFailure(story, "service is required")
	}
	if story.RequestedPath == "" {
		return permissionStoryInputFailure(story, "path is required")
	}
	if !filepath.IsAbs(story.RequestedPath) {
		return permissionStoryInputFailure(story, "path must be absolute")
	}
	if snap.Mode == dockerDeploymentMode {
		story.Status = "unknown"
		story.Conclusion = "native service identity is unavailable in Docker deployment mode"
		story.FirstProblem = "service-identity"
		return story
	}

	props, err := collectPermissionServiceProperties(ctx, unit)
	if err != nil {
		story.Status = "fail"
		story.Conclusion = "could not read native service identity"
		story.FirstProblem = "service-identity"
		return story
	}
	if props["LoadState"] == "not-found" || strings.TrimSpace(props["MainPID"]) == "" || props["MainPID"] == "0" {
		story.Status = "fail"
		story.Conclusion = "service does not have a running main process"
		story.FirstProblem = "service-identity"
		return story
	}

	pid, err := strconv.Atoi(props["MainPID"])
	if err != nil || pid <= 0 {
		story.Status = "fail"
		story.Conclusion = "service returned an invalid main process identity"
		story.FirstProblem = "service-identity"
		return story
	}
	identity, err := collectPermissionIdentity(pid, props)
	if err != nil {
		story.Status = "fail"
		story.Conclusion = "could not inspect the running service process"
		story.FirstProblem = "service-identity"
		return story
	}
	story.Identity = identity

	resolved, chain := permissionPathChain(story.RequestedPath, identity)
	story.ResolvedPath = resolved
	story.Chain = chain
	story.Decisions = permissionDecisions(chain, identity)
	story.DockerBindMounts = permissionDockerInspect(ctx, story.RequestedPath, resolved, snap.Containers)

	story.Status, story.FirstProblem, story.Conclusion = permissionStoryOutcome(story)
	return story
}

func permissionStoryInputFailure(story PermissionStory, msg string) PermissionStory {
	story.Status = "fail"
	story.Conclusion = msg
	story.FirstProblem = "input"
	return story
}

func collectPermissionServiceProperties(ctx context.Context, unit string) (map[string]string, error) {
	result, err := permissionSystemdLookup(ctx, unit)
	if errors.Is(err, exec.ErrNotFound) {
		return nil, err
	}
	values := map[string]string{}
	for _, line := range strings.Split(result.Output, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok {
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	if len(values) == 0 {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("systemd returned no service properties")
	}
	return values, err
}

func collectPermissionIdentity(pid int, props map[string]string) (*PermissionIdentity, error) {
	statusBytes, err := permissionReadFile(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if err != nil {
		return nil, err
	}
	fields := map[string]string{}
	for _, line := range strings.Split(string(statusBytes), "\n") {
		key, value, ok := strings.Cut(line, ":")
		if ok {
			fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	uid, ok := parseEffectiveID(fields["Uid"])
	if !ok {
		return nil, errors.New("process effective UID is unavailable")
	}
	gid, ok := parseEffectiveID(fields["Gid"])
	if !ok {
		return nil, errors.New("process effective GID is unavailable")
	}
	identity := &PermissionIdentity{
		PID:               pid,
		UID:               uid,
		GID:               gid,
		SupplementaryGIDs: parseIDList(fields["Groups"]),
		ConfiguredUser:    props["User"],
		ConfiguredGroup:   props["Group"],
		EffectiveCaps:     fields["CapEff"],
	}
	if value, readErr := permissionReadlink(filepath.Join("/proc", strconv.Itoa(pid), "cwd")); readErr == nil {
		identity.WorkingDirectory = value
	} else {
		identity.WorkingDirectory = props["WorkingDirectory"]
	}
	if value, readErr := permissionReadlink(filepath.Join("/proc", strconv.Itoa(pid), "exe")); readErr == nil {
		identity.Executable = value
	}
	return identity, nil
}

func parseEffectiveID(value string) (uint32, bool) {
	fields := strings.Fields(value)
	if len(fields) < 2 {
		return 0, false
	}
	n, err := strconv.ParseUint(fields[1], 10, 32)
	return uint32(n), err == nil
}

func parseIDList(value string) []uint32 {
	var out []uint32
	for _, field := range strings.Fields(value) {
		n, err := strconv.ParseUint(field, 10, 32)
		if err == nil {
			out = append(out, uint32(n))
		}
	}
	return out
}

func permissionPathChain(requested string, identity *PermissionIdentity) (string, []PermissionNode) {
	clean := filepath.Clean(requested)
	resolved := clean
	if value, err := filepath.EvalSymlinks(clean); err == nil {
		resolved = value
	}
	parts := pathPrefixes(resolved)
	nodes := make([]PermissionNode, 0, len(parts))
	for i, path := range parts {
		node := PermissionNode{Path: path, IsTarget: i == len(parts)-1}
		info, err := permissionStat(path)
		if err != nil {
			node.Kind = "missing"
			node.StatError = boundedEvidence(err.Error(), 384)
			nodes = append(nodes, node)
			break
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			node.Kind = permissionKind(info.Mode())
			node.Mode = fmt.Sprintf("%04o", info.Mode().Perm())
			node.StatError = "ownership metadata unavailable"
			nodes = append(nodes, node)
			continue
		}
		node.Kind = permissionKind(info.Mode())
		node.UID = stat.Uid
		node.GID = stat.Gid
		node.Mode = fmt.Sprintf("%04o", info.Mode().Perm())
		node.Relation = permissionRelation(identity, stat.Uid, stat.Gid)
		nodes = append(nodes, node)
	}
	return resolved, nodes
}

func pathPrefixes(path string) []string {
	clean := filepath.Clean(path)
	if clean == "/" {
		return []string{"/"}
	}
	parts := strings.Split(strings.TrimPrefix(clean, "/"), string(filepath.Separator))
	out := []string{"/"}
	current := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		current = filepath.Join(current, part)
		out = append(out, "/"+current)
	}
	return out
}

func permissionKind(mode os.FileMode) string {
	switch {
	case mode.IsDir():
		return "directory"
	case mode.IsRegular():
		return "file"
	case mode&os.ModeSocket != 0:
		return "socket"
	case mode&os.ModeSymlink != 0:
		return "symlink"
	case mode&os.ModeNamedPipe != 0:
		return "fifo"
	case mode&os.ModeDevice != 0:
		return "device"
	default:
		return "other"
	}
}

func permissionRelation(identity *PermissionIdentity, uid, gid uint32) string {
	if identity == nil {
		return "unknown"
	}
	if identity.UID == uid {
		return "owner"
	}
	if identity.GID == gid || containsID(identity.SupplementaryGIDs, gid) {
		return "group"
	}
	return "other"
}

func containsID(values []uint32, want uint32) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func permissionDecisions(chain []PermissionNode, identity *PermissionIdentity) []PermissionDecision {
	if len(chain) == 0 {
		return nil
	}
	var out []PermissionDecision
	for i, node := range chain {
		if node.StatError != "" {
			out = append(out, PermissionDecision{
				Path: node.Path, Operation: "stat", Status: "fail",
				Evidence: "metadata unavailable: " + node.StatError,
			})
			break
		}
		if i < len(chain)-1 {
			out = append(out, permissionDecision(node, identity, "traverse", 1))
			continue
		}
		switch node.Kind {
		case "directory":
			out = append(out,
				permissionDecision(node, identity, "read", 4),
				permissionDecision(node, identity, "write", 2),
				permissionDecision(node, identity, "traverse", 1),
			)
		case "socket":
			out = append(out, permissionDecision(node, identity, "connect", 2))
		default:
			out = append(out,
				permissionDecision(node, identity, "read", 4),
				permissionDecision(node, identity, "write", 2),
				permissionDecision(node, identity, "execute", 1),
			)
		}
	}
	return out
}

func permissionDecision(node PermissionNode, identity *PermissionIdentity, operation string, bit uint32) PermissionDecision {
	relation := permissionRelation(identity, node.UID, node.GID)
	modeValue, err := strconv.ParseUint(node.Mode, 8, 32)
	if err != nil {
		return PermissionDecision{Path: node.Path, Operation: operation, Status: "unknown", Relation: relation, Evidence: "mode metadata is unavailable"}
	}
	shift := uint32(0)
	switch relation {
	case "owner":
		shift = 6
	case "group":
		shift = 3
	case "other":
		shift = 0
	default:
		return PermissionDecision{Path: node.Path, Operation: operation, Status: "unknown", Relation: relation, Evidence: "identity relation is unavailable"}
	}
	allowed := ((uint32(modeValue) >> shift) & bit) != 0
	status := "fail"
	if allowed {
		status = "pass"
	}
	required := map[uint32]string{4: "r", 2: "w", 1: "x"}[bit]
	evidence := fmt.Sprintf("%s class requires %s; mode=%s uid=%d gid=%d", relation, required, node.Mode, node.UID, node.GID)

	if !allowed && identity != nil && identity.UID == 0 {
		capOverride, capsKnown := permissionHasCapability(identity.EffectiveCaps, 1)
		capReadSearch, readKnown := permissionHasCapability(identity.EffectiveCaps, 2)
		if !capsKnown || !readKnown {
			status = "unknown"
			evidence += "; effective UID is 0 but capability evidence is unavailable"
		} else {
			bypass := capOverride
			if operation == "read" || operation == "traverse" {
				bypass = bypass || capReadSearch
			}
			if operation == "execute" && node.Kind != "directory" && uint32(modeValue)&0111 == 0 {
				bypass = false
			}
			if bypass {
				status = "pass"
				evidence += "; effective capabilities bypass this mode-bit denial"
			}
		}
	}
	return PermissionDecision{Path: node.Path, Operation: operation, Status: status, Relation: relation, Required: required, Evidence: evidence}
}

func permissionHasCapability(raw string, bit uint) (bool, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, false
	}
	value, err := strconv.ParseUint(raw, 16, 64)
	if err != nil {
		return false, false
	}
	return value&(uint64(1)<<bit) != 0, true
}

type dockerInspectMount struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	RW          bool   `json:"RW"`
}

func inspectRelevantDockerBindMounts(ctx context.Context, requested, resolved string, containers []ContainerInfo) []DockerBindMountEvidence {
	if len(containers) == 0 {
		return nil
	}
	command := resolveCommand("docker", "/usr/bin/docker", "/usr/local/bin/docker")
	if _, err := exec.LookPath(command); err != nil && !filepath.IsAbs(command) {
		return nil
	}
	limit := len(containers)
	if limit > permissionContainerLimit {
		limit = permissionContainerLimit
	}
	var out []DockerBindMountEvidence
	for _, container := range containers[:limit] {
		name := strings.TrimSpace(container.Name)
		if name == "" {
			continue
		}
		lookupCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		result, err := runBoundedCommand(lookupCtx, permissionCommandOutputLimit, command, "inspect", "--format", "{{json .Mounts}}", name)
		cancel()
		if err != nil || strings.TrimSpace(result.Output) == "" {
			continue
		}
		var mounts []dockerInspectMount
		if json.Unmarshal([]byte(strings.TrimSpace(result.Output)), &mounts) != nil {
			continue
		}
		for _, mount := range mounts {
			if mount.Type != "bind" {
				continue
			}
			side := relevantMountSide(requested, resolved, mount.Source, mount.Destination)
			if side == "" {
				continue
			}
			out = append(out, DockerBindMountEvidence{
				Container: name, Source: mount.Source, Destination: mount.Destination,
				ReadWrite: mount.RW, MatchSide: side,
			})
			if len(out) >= permissionDockerMountLimit {
				return out
			}
		}
	}
	return out
}

func relevantMountSide(requested, resolved, source, destination string) string {
	for _, candidate := range []string{requested, resolved} {
		if pathWithin(candidate, source) {
			return "host-source"
		}
		if pathWithin(candidate, destination) {
			return "container-destination"
		}
	}
	return ""
}

func pathWithin(path, root string) bool {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(root) == "" || !filepath.IsAbs(path) || !filepath.IsAbs(root) {
		return false
	}
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	if path == root {
		return true
	}
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func permissionStoryOutcome(story PermissionStory) (status, firstProblem, conclusion string) {
	if story.Identity == nil {
		return "unknown", "service-identity", "service identity could not be established"
	}
	for _, decision := range story.Decisions {
		if decision.Status == "fail" {
			id := decision.Operation + ":" + decision.Path
			if decision.Operation == "stat" {
				return "fail", id, "the requested path or one of its parents is unavailable"
			}
			if decision.Operation == "traverse" && decision.Path != story.ResolvedPath {
				return "fail", id, "the service identity cannot traverse the full parent-directory chain"
			}
			return "fail", id, "the service identity is denied at the target by the observed mode bits"
		}
	}
	for _, decision := range story.Decisions {
		if decision.Status == "unknown" {
			return "unknown", decision.Operation + ":" + decision.Path, "mode-bit evidence is incomplete"
		}
	}
	return "pass", "", "the observed mode-bit chain permits the tested access classes"
}
