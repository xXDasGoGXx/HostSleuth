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
- [x] Publish stable `v0.2.0` from M4-complete `main`.
- [x] Publish native linux/amd64 and linux/arm64 v0.2.0 assets plus `SHA256SUMS`.
- [x] Publish public multi-platform `mjmalleo/hostsleuth:0.2.0` and move `latest` to the same image.
- [x] Verify anonymous Docker Registry access, linux/amd64 + linux/arm64 manifests, and matching `0.2.0` / `latest` index digest.
- [x] M5 — bounded native configuration fingerprinting.
- [x] Fingerprint only the explicit high-value default set rather than recursively crawling `/etc`.
- [x] Store path/state/SHA-256/size only; never store configuration contents in snapshots or events.
- [x] Feed appeared/disappeared/content-changed configuration evidence into the existing Changes timeline.
- [x] Baseline configuration fingerprints across schema 2 -> 3 while preserving M4 package events.
- [x] Keep missing/unreadable files truthful and unreadable transitions quiet.
- [x] Final M5 CI passed tests/vet/build, JS syntax, Compose, amd64/arm64 builds, and Docker runtime smoke.
- [x] Real OMV Debian acceptance confirmed the five explicit candidates without privilege escalation or live deployment replacement.

Stable `v0.2.0` remains the current published release. M5 is merged in source on `main` and is not yet claimed to be in a published release. The live OMV/Arcane deployment and `OMV-Docker-Rebuild` remain intentionally pinned to known-good `mjmalleo/hostsleuth:0.1.0`.

## Next capability — deliberately unselected

Do **not** start another capability automatically. Choose one deliberately when development resumes.

Candidates that fit HostSleuth's two core jobs:

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
