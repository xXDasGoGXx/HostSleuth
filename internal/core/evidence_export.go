package core

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func DefaultEvidenceFilename(now time.Time) string {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return "hostsleuth-evidence-" + now.UTC().Format("20060102T150405Z") + ".zip"
}

func BuildEvidencePreview(store Store, hostSleuthVersion string, now time.Time) (EvidencePreview, error) {
	bundle, err := buildEvidenceBundle(store, hostSleuthVersion, now)
	if err != nil {
		return EvidencePreview{}, err
	}
	return bundle.Preview, nil
}

func ExportEvidence(store Store, hostSleuthVersion, outputPath string, now time.Time) (EvidenceExportResult, error) {
	bundle, err := buildEvidenceBundle(store, hostSleuthVersion, now)
	if err != nil {
		return EvidenceExportResult{}, err
	}
	if strings.TrimSpace(outputPath) == "" {
		outputPath = DefaultEvidenceFilename(now)
	}
	outputPath = filepath.Clean(outputPath)
	if _, err := os.Stat(outputPath); err == nil {
		return EvidenceExportResult{}, fmt.Errorf("evidence output already exists: %s", outputPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return EvidenceExportResult{}, err
	}

	dir := filepath.Dir(outputPath)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return EvidenceExportResult{}, err
	}
	tmp, err := os.CreateTemp(dir, ".hostsleuth-evidence-*.tmp")
	if err != nil {
		return EvidenceExportResult{}, err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}
	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}

	zw := zip.NewWriter(tmp)
	names := make([]string, 0, len(bundle.Files))
	for name := range bundle.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(0o600)
		writer, err := zw.CreateHeader(header)
		if err != nil {
			_ = zw.Close()
			cleanup()
			return EvidenceExportResult{}, err
		}
		if _, err := writer.Write(bundle.Files[name]); err != nil {
			_ = zw.Close()
			cleanup()
			return EvidenceExportResult{}, err
		}
	}
	if err := zw.Close(); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := os.Rename(tmpName, outputPath); err != nil {
		cleanup()
		return EvidenceExportResult{}, err
	}
	if err := os.Chmod(outputPath, 0o600); err != nil {
		_ = os.Remove(outputPath)
		return EvidenceExportResult{}, err
	}

	archive, err := os.ReadFile(outputPath)
	if err != nil {
		return EvidenceExportResult{}, err
	}
	sum := sha256.Sum256(archive)
	return EvidenceExportResult{Path: outputPath, SHA256: hex.EncodeToString(sum[:]), Bytes: int64(len(archive))}, nil
}
