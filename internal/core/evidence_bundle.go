package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func buildEvidenceBundle(store Store, hostSleuthVersion string, now time.Time) (evidenceBundle, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	source, unavailable, err := loadEvidenceSource(store)
	if err != nil {
		return evidenceBundle{}, err
	}
	redactor := newEvidenceRedactor(source)
	redactedSnapshot := redactor.snapshot(source.Snapshot)
	redactedEvents := make([]Event, 0, len(source.Events))
	for _, event := range source.Events {
		redactedEvents = append(redactedEvents, redactor.event(event))
	}
	redactedAudits := make([]ActionAudit, 0, len(source.Audits))
	for _, audit := range source.Audits {
		redactedAudits = append(redactedAudits, redactor.audit(audit))
	}

	files := map[string][]byte{}
	if files["snapshot.json"], err = marshalEvidenceJSON(redactedSnapshot); err != nil {
		return evidenceBundle{}, fmt.Errorf("snapshot payload: %w", err)
	}
	if files["events.json"], err = marshalEvidenceJSON(redactedEvents); err != nil {
		return evidenceBundle{}, fmt.Errorf("events payload: %w", err)
	}
	if len(redactedAudits) > 0 {
		if files["action-audit.json"], err = marshalEvidenceJSON(redactedAudits); err != nil {
			return evidenceBundle{}, fmt.Errorf("action audit payload: %w", err)
		}
	}
	files["summary.txt"] = []byte(evidenceSummary(redactedSnapshot, len(redactedEvents), len(redactedAudits)))

	included := []string{"snapshot", "events"}
	if len(redactedAudits) > 0 {
		included = append(included, "action-audit")
	}
	omitted := []string{
		"raw-journal", "raw-action-command-output", "action-preview-command-and-confirmation",
		"arbitrary-file-contents", "arbitrary-configuration-contents", "workbench-arbitrary-inputs",
	}
	manifest := EvidenceManifest{
		SchemaVersion:       EvidenceBundleSchemaVersion,
		HostSleuthVersion:   hostSleuthVersion,
		GeneratedAt:         now,
		RedactionPolicy:     EvidenceRedactionPolicy,
		IncludedClasses:     included,
		OmittedClasses:      omitted,
		RedactionCounts:     cloneStringIntMap(redactor.counts),
		RecordLimits:        map[string]int{"events": EvidenceEventLimit, "action_audits": EvidenceActionAuditLimit, "max_uncompressed_bytes": EvidenceMaxPayloadBytes, "max_json_bytes": EvidenceMaxJSONBytes},
		UnavailableEvidence: unavailable,
		Warning:             "Redaction reduces disclosure risk but cannot guarantee anonymity; review the preview and bundle before sharing.",
	}
	payloadNames := make([]string, 0, len(files))
	for name := range files {
		payloadNames = append(payloadNames, name)
	}
	sort.Strings(payloadNames)
	for _, name := range payloadNames {
		sum := sha256.Sum256(files[name])
		manifest.Files = append(manifest.Files, EvidenceFileDigest{Name: name, SHA256: hex.EncodeToString(sum[:]), Bytes: len(files[name])})
	}
	manifestBytes, err := marshalEvidenceJSON(manifest)
	if err != nil {
		return evidenceBundle{}, fmt.Errorf("manifest payload: %w", err)
	}
	files["manifest.json"] = manifestBytes
	files["checksums.txt"] = evidenceChecksums(files)

	total := 0
	fileNames := make([]string, 0, len(files))
	for name, data := range files {
		total += len(data)
		fileNames = append(fileNames, name)
	}
	if total > EvidenceMaxPayloadBytes {
		return evidenceBundle{}, fmt.Errorf("evidence payload exceeds %d byte limit", EvidenceMaxPayloadBytes)
	}
	sort.Strings(fileNames)
	preview := EvidencePreview{
		SchemaVersion:       EvidenceBundleSchemaVersion,
		GeneratedAt:         now,
		RedactionPolicy:     EvidenceRedactionPolicy,
		IncludedClasses:     included,
		OmittedClasses:      omitted,
		RecordCounts:        map[string]int{"events": len(redactedEvents), "action_audits": len(redactedAudits)},
		RedactionCounts:     cloneStringIntMap(redactor.counts),
		Files:               fileNames,
		EstimatedBytes:      total,
		UnavailableEvidence: unavailable,
		Warnings:            []string{manifest.Warning},
	}
	return evidenceBundle{Preview: preview, Files: files}, nil
}

func loadEvidenceSource(store Store) (evidenceSource, []string, error) {
	snapshot, err := store.LoadSnapshot()
	if err != nil {
		return evidenceSource{}, nil, fmt.Errorf("load snapshot: %w", err)
	}
	events, err := store.ReadEvents(EvidenceEventLimit)
	if err != nil {
		return evidenceSource{}, nil, fmt.Errorf("load events: %w", err)
	}
	audits, err := store.ReadActionAudits(EvidenceActionAuditLimit)
	if err != nil {
		return evidenceSource{}, nil, fmt.Errorf("load action audits: %w", err)
	}
	unavailable := []string{}
	if len(events) == 0 {
		unavailable = append(unavailable, "no retained events were available")
	}
	if len(audits) == 0 {
		unavailable = append(unavailable, "no Safe Action audit records were available")
	}
	return evidenceSource{Snapshot: snapshot, Events: events, Audits: audits}, unavailable, nil
}

func marshalEvidenceJSON(value any) ([]byte, error) {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	if len(b) > EvidenceMaxJSONBytes {
		return nil, fmt.Errorf("JSON payload exceeds %d byte limit", EvidenceMaxJSONBytes)
	}
	return b, nil
}

func evidenceSummary(snapshot Snapshot, events, audits int) string {
	return fmt.Sprintf("HostSleuth redacted evidence bundle\nMode: %s\nCaptured snapshot: %s\nEvents included: %d\nAction audits included: %d\nPolicy: %s\nReview manifest.json before sharing.\n",
		snapshot.Mode, snapshot.CapturedAt.UTC().Format(time.RFC3339), events, audits, EvidenceRedactionPolicy)
}

func evidenceChecksums(files map[string][]byte) []byte {
	names := make([]string, 0, len(files))
	for name := range files {
		if name != "checksums.txt" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		fmt.Fprintf(&b, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	return []byte(b.String())
}
