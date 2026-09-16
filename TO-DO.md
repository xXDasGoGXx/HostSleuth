# HostSleuth — TO-DO

This file stays intentionally short. Completed milestone history belongs in `docs/history/DEVELOPMENT-HISTORY.md`.

## Completed

- [x] M0 — repository foundation.
- [x] M1 — deployable single-host MVP.
- [x] M2 — deeper deterministic diagnosis.
- [x] M3.1 — modern responsive Web UI.
- [x] M3.2 — usability and installation clarity.
- [x] M3.3 — supported Docker Compose deployment with amd64/arm64 CI validation.
- [x] M3 acceptance — native and Docker runtime validation.
- [x] M3.4 — public container distribution and stable `v0.1.0` publication.
- [x] M4 — bounded native Debian/Ubuntu package-change timeline.
- [x] Parse recent `dpkg` install/update/remove history with `apt` history fallback.
- [x] Feed package changes into the existing Changes timeline as `category=package` events.
- [x] Preserve package timestamps, architecture, and old/new versions.
- [x] Bound package-log reads and retained history.
- [x] Baseline existing package history across the schema-1 -> schema-2 upgrade instead of replaying it as new events.
- [x] Keep unsupported/unavailable package logs quiet and collection read-only.
- [x] Keep default Docker host mounts unchanged; package history remains native-mode evidence by default.
- [x] Focused apt/dpkg parser, event, fallback, and schema-upgrade tests.
- [x] Final M4 CI passed: tests/vet/build, JS syntax, Compose, amd64, arm64, Docker smoke.
- [x] Real OMV Debian acceptance passed with actual dpkg history plus one copied-log synthetic event; real package log remained unchanged.

M4 is closed after PR #24 merge. Do not create a newer release/tag/image without explicit owner approval.

## Next capability — deliberately unselected

Do **not** start another capability automatically. Choose one deliberately when development resumes.

Candidates that fit HostSleuth's two core jobs:

- Configuration fingerprinting that records change evidence without storing configuration secrets by default.
- Bounded TLS/certificate diagnosis for common host:port failures: handshake, hostname, expiry, and trust evidence.
- General redaction rules and threat-model documentation before exporting or sharing richer diagnostic data.

## Demand-gated

Only pursue these if real use proves the need:

- Authentication for direct non-loopback Web UI exposure. Loopback + SSH tunneling remains the simple default.
- Reverse-proxy awareness for common proxies. Do not build a proxy-management layer.
- `.deb` packaging and signed artifacts if installation becomes a real adoption problem.
- Privilege separation if future collectors require enough elevated access to justify the complexity.
- SQLite only if JSON/JSONL becomes a demonstrated operational limitation.
- Baseline-vs-incident comparison only if the event timeline proves insufficient in real incidents.
- Export/import support bundle only after redaction rules are strong enough to make sharing safe.
- Refresh the README screenshot when a useful real public-image capture is naturally available; do not create a cosmetic workstream for it.

## Not planned unless HostSleuth changes direction

- Multi-host controller/agent architecture.
- General dependency-graph platform.
- Pluggable diagnosis-rule ecosystem.
- AI explanation layer.
- Automatic or guided repair/remediation workflow.
- Generic network-device/SNMP monitoring platform.

## Rule for new ideas

An idea goes here first. It becomes active only when it clearly improves a common HostSleuth workflow without making installation, operation, or the UI meaningfully harder.
