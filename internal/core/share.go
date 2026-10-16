package core

import "strings"

// ShareTextLimit bounds text submitted for redaction.
const ShareTextLimit = 64 * 1024

// RedactForSharing applies the Evidence Bundle's redaction to free text a
// person wants to post publicly (a forum, an issue tracker, a chat). It is
// seeded with this host's own names and addresses, so identifiers that the
// generic patterns would miss, such as container and service names, are
// replaced with the same stable aliases throughout the text.
//
// Redaction lowers disclosure risk; it cannot guarantee anonymity. The
// person still reviews the result before posting.
func RedactForSharing(snap Snapshot, events []Event, text string) string {
	r := newEvidenceRedactor(evidenceSource{Snapshot: snap, Events: events})
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = r.freeform(line)
	}
	return strings.Join(lines, "\n")
}
