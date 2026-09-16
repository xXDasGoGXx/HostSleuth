package core

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	maxPackageChanges  = 200
	maxPackageLogBytes = int64(8 << 20)
)

func packageLogRoot() string {
	if root := strings.TrimSpace(os.Getenv("HOSTSLEUTH_PACKAGE_LOG_ROOT")); root != "" {
		return root
	}
	if deploymentMode() == dockerDeploymentMode {
		return "/host/var/log"
	}
	return "/var/log"
}

func collectPackageChanges() []PackageChange {
	root := packageLogRoot()
	dpkg := parsePackageLogFiles([]string{
		filepath.Join(root, "dpkg.log.2.gz"),
		filepath.Join(root, "dpkg.log.1.gz"),
		filepath.Join(root, "dpkg.log.1"),
		filepath.Join(root, "dpkg.log"),
	}, parseDPKGLog)
	if len(dpkg) > 0 {
		return newestPackageChanges(dpkg, maxPackageChanges)
	}

	apt := parsePackageLogFiles([]string{
		filepath.Join(root, "apt", "history.log.2.gz"),
		filepath.Join(root, "apt", "history.log.1.gz"),
		filepath.Join(root, "apt", "history.log.1"),
		filepath.Join(root, "apt", "history.log"),
	}, parseAPTHistory)
	return newestPackageChanges(apt, maxPackageChanges)
}

type packageLogParser func(string) []PackageChange

func parsePackageLogFiles(paths []string, parser packageLogParser) []PackageChange {
	var out []PackageChange
	seen := map[string]bool{}
	for _, path := range paths {
		data, err := readPackageLog(path)
		if err != nil {
			continue
		}
		for _, change := range parser(string(data)) {
			key := packageChangeKey(change)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, change)
		}
	}
	sortPackageChanges(out)
	return out
}

func readPackageLog(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.HasSuffix(path, ".gz") {
		zr, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		return io.ReadAll(io.LimitReader(zr, maxPackageLogBytes))
	}

	if info, err := f.Stat(); err == nil && info.Size() > maxPackageLogBytes {
		if _, err := f.Seek(-maxPackageLogBytes, io.SeekEnd); err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(f, maxPackageLogBytes))
		if err != nil {
			return nil, err
		}
		if newline := strings.IndexByte(string(data), '\n'); newline >= 0 {
			data = data[newline+1:]
		}
		return data, nil
	}
	return io.ReadAll(io.LimitReader(f, maxPackageLogBytes))
}

func parseDPKGLog(text string) []PackageChange {
	var out []PackageChange
	scanner := newPackageScanner(text)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}
		at, ok := parsePackageTimestamp(fields[0] + " " + fields[1])
		if !ok {
			continue
		}
		name, arch := splitPackageNameArch(fields[3])
		switch fields[2] {
		case "install":
			if len(fields) < 6 {
				continue
			}
			out = append(out, PackageChange{At: at, Action: "install", Name: name, Architecture: arch, FromVersion: cleanPackageVersion(fields[4]), ToVersion: cleanPackageVersion(fields[5])})
		case "upgrade":
			if len(fields) < 6 {
				continue
			}
			out = append(out, PackageChange{At: at, Action: "update", Name: name, Architecture: arch, FromVersion: cleanPackageVersion(fields[4]), ToVersion: cleanPackageVersion(fields[5])})
		case "remove", "purge":
			out = append(out, PackageChange{At: at, Action: "remove", Name: name, Architecture: arch, FromVersion: cleanPackageVersion(fields[4])})
		}
	}
	sortPackageChanges(out)
	return out
}

func parseAPTHistory(text string) []PackageChange {
	var out []PackageChange
	var at time.Time
	scanner := newPackageScanner(text)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Start-Date:") {
			value := strings.Join(strings.Fields(strings.TrimSpace(strings.TrimPrefix(line, "Start-Date:"))), " ")
			parsed, ok := parsePackageTimestamp(value)
			if ok {
				at = parsed
			} else {
				at = time.Time{}
			}
			continue
		}
		if at.IsZero() {
			continue
		}

		action := ""
		payload := ""
		switch {
		case strings.HasPrefix(line, "Install:"):
			action = "install"
			payload = strings.TrimSpace(strings.TrimPrefix(line, "Install:"))
		case strings.HasPrefix(line, "Upgrade:"):
			action = "update"
			payload = strings.TrimSpace(strings.TrimPrefix(line, "Upgrade:"))
		case strings.HasPrefix(line, "Remove:"):
			action = "remove"
			payload = strings.TrimSpace(strings.TrimPrefix(line, "Remove:"))
		case strings.HasPrefix(line, "Purge:"):
			action = "remove"
			payload = strings.TrimSpace(strings.TrimPrefix(line, "Purge:"))
		default:
			continue
		}

		for _, item := range splitAPTPackageItems(payload) {
			change, ok := parseAPTPackageItem(at, action, item)
			if ok {
				out = append(out, change)
			}
		}
	}
	sortPackageChanges(out)
	return out
}

func newPackageScanner(text string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 64*1024), int(maxPackageLogBytes))
	return scanner
}

func splitAPTPackageItems(payload string) []string {
	var out []string
	start := 0
	depth := 0
	for i, r := range payload {
		switch r {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				item := strings.TrimSpace(payload[start:i])
				if item != "" {
					out = append(out, item)
				}
				start = i + 1
			}
		}
	}
	if item := strings.TrimSpace(payload[start:]); item != "" {
		out = append(out, item)
	}
	return out
}

func parseAPTPackageItem(at time.Time, action, item string) (PackageChange, bool) {
	open := strings.Index(item, " (")
	close := strings.LastIndex(item, ")")
	if open <= 0 || close <= open+2 {
		return PackageChange{}, false
	}
	name, arch := splitPackageNameArch(strings.TrimSpace(item[:open]))
	if name == "" {
		return PackageChange{}, false
	}
	parts := strings.Split(item[open+2:close], ", ")
	change := PackageChange{At: at, Action: action, Name: name, Architecture: arch}
	switch action {
	case "install":
		if len(parts) < 1 {
			return PackageChange{}, false
		}
		change.ToVersion = cleanPackageVersion(parts[0])
	case "update":
		if len(parts) < 2 {
			return PackageChange{}, false
		}
		change.FromVersion = cleanPackageVersion(parts[0])
		change.ToVersion = cleanPackageVersion(parts[1])
	case "remove":
		if len(parts) < 1 {
			return PackageChange{}, false
		}
		change.FromVersion = cleanPackageVersion(parts[0])
	default:
		return PackageChange{}, false
	}
	return change, true
}

func parsePackageTimestamp(value string) (time.Time, bool) {
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", strings.Join(strings.Fields(value), " "), time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func splitPackageNameArch(value string) (string, string) {
	value = strings.TrimSpace(value)
	if i := strings.LastIndexByte(value, ':'); i > 0 && i < len(value)-1 {
		return value[:i], value[i+1:]
	}
	return value, ""
}

func cleanPackageVersion(value string) string {
	value = strings.TrimSpace(value)
	if value == "<none>" {
		return ""
	}
	return value
}

func newestPackageChanges(changes []PackageChange, limit int) []PackageChange {
	if len(changes) == 0 {
		return nil
	}
	sortPackageChanges(changes)
	if limit > 0 && len(changes) > limit {
		changes = changes[len(changes)-limit:]
	}
	return changes
}

func sortPackageChanges(changes []PackageChange) {
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].At.Equal(changes[j].At) {
			return packageChangeKeyWithoutTime(changes[i]) < packageChangeKeyWithoutTime(changes[j])
		}
		return changes[i].At.Before(changes[j].At)
	})
}

func packageChangeKey(change PackageChange) string {
	return change.At.UTC().Format(time.RFC3339Nano) + "|" + packageChangeKeyWithoutTime(change)
}

func packageChangeKeyWithoutTime(change PackageChange) string {
	return strings.Join([]string{change.Action, change.Name, change.Architecture, change.FromVersion, change.ToVersion}, "|")
}

func packageChangeSummary(change PackageChange) string {
	name := change.Name
	if change.Architecture != "" {
		name += ":" + change.Architecture
	}
	switch change.Action {
	case "install":
		if change.ToVersion != "" {
			return fmt.Sprintf("package installed: %s %s", name, change.ToVersion)
		}
		return "package installed: " + name
	case "update":
		if change.FromVersion != "" && change.ToVersion != "" {
			return fmt.Sprintf("package updated: %s %s -> %s", name, change.FromVersion, change.ToVersion)
		}
		return "package updated: " + name
	case "remove":
		if change.FromVersion != "" {
			return fmt.Sprintf("package removed: %s %s", name, change.FromVersion)
		}
		return "package removed: " + name
	default:
		return "package changed: " + name
	}
}
