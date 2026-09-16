package core

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxConfigFileBytes = int64(4 << 20)

var defaultConfigFingerprintPaths = []string{
	"/etc/hosts",
	"/etc/fstab",
	"/etc/ssh/sshd_config",
	"/etc/docker/daemon.json",
	"/etc/nftables.conf",
}

func configFingerprintRoot() string {
	if root := strings.TrimSpace(os.Getenv("HOSTSLEUTH_CONFIG_ROOT")); root != "" {
		return root
	}
	return "/"
}

func collectConfigFingerprints() []ConfigFingerprint {
	root := configFingerprintRoot()
	out := make([]ConfigFingerprint, 0, len(defaultConfigFingerprintPaths))
	for _, canonicalPath := range defaultConfigFingerprintPaths {
		out = append(out, fingerprintConfigFile(root, canonicalPath))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func fingerprintConfigFile(root, canonicalPath string) ConfigFingerprint {
	entry := ConfigFingerprint{Path: canonicalPath}
	actualPath := canonicalPath
	if root != "/" {
		actualPath = filepath.Join(root, strings.TrimPrefix(canonicalPath, "/"))
	}

	f, err := os.Open(actualPath)
	if err != nil {
		if os.IsNotExist(err) {
			entry.State = "missing"
		} else {
			entry.State = "unreadable"
		}
		return entry
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxConfigFileBytes {
		entry.State = "unreadable"
		return entry
	}

	h := sha256.New()
	if _, err := io.Copy(h, io.LimitReader(f, maxConfigFileBytes+1)); err != nil {
		entry.State = "unreadable"
		return entry
	}

	entry.State = "present"
	entry.Fingerprint = hex.EncodeToString(h.Sum(nil))
	entry.SizeBytes = info.Size()
	return entry
}
