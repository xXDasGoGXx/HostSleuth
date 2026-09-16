# HostSleuth — TO-DO

This file stays intentionally short. Completed milestone history belongs in `docs/history/DEVELOPMENT-HISTORY.md`. The ordered forward plan is in `docs/ROADMAP.md`.

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
- [x] Publish stable `v0.2.0` with native amd64/arm64 assets, checksums, and public multi-platform Docker tags.
- [x] M5 — bounded native configuration fingerprinting.
- [x] M5 schema-3 upgrade baseline preserves M4 package events.
- [x] M5 CI and isolated real OMV Debian acceptance.

Stable `v0.2.0` remains the current published release. M5 is merged on `main` but is not yet in a public release. The live OMV/Arcane deployment and `OMV-Docker-Rebuild` remain intentionally pinned to known-good `mjmalleo/hostsleuth:0.1.0`.

## Active — 1. M6 Certificate Story / TLS Detective

One bounded read-only milestone.

- [ ] Inspect current Diagnose/API/evidence structures before coding.
- [ ] Add bounded TLS handshake evidence for `host:port`.
- [ ] Record subject, SANs, issuer, serial, validity, days remaining, and SHA-256 certificate fingerprint.
- [ ] Report hostname match and bounded trust/chain evidence truthfully.
- [ ] Correlate local listener/process/container evidence when the target is local.
- [ ] On native Linux, detect Certbot when present and show bounded read-only certificate/renewal/timer evidence.
- [ ] Compare local certificate fingerprint with the certificate actually being served when both are available.
- [ ] Detect stale-served-certificate situations only when deterministic evidence supports it.
- [ ] Integrate with the accepted Diagnose/Host Story UI; do not build a generic certificate-management dashboard.
- [ ] Keep M6 read-only: no renewal, reload, install, ACME account management, or arbitrary command execution.
- [ ] Add focused TLS/certificate tests.
- [ ] Run normal CI.
- [ ] Perform one bounded real-host acceptance pass.
- [ ] Merge and close M6 if no concrete defect appears.

## Then — exact approved order

### 2. Publish stable v0.3.0

- [ ] Include M5 + M6.
- [ ] Re-check exact accepted `main` SHA and release workflow.
- [ ] Stop for explicit owner publication approval.
- [ ] Publish native amd64/arm64 binaries + `SHA256SUMS`.
- [ ] Publish `mjmalleo/hostsleuth:0.3.0` and move `latest`.
- [ ] Verify anonymous registry/platform manifests and one bounded consumer acceptance.
- [ ] Do not change live OMV merely to chase version numbers.

### 3. M7 — Service Story

Correlate systemd/journal, process/listener ownership, port collisions, containers, nearby package/config changes, TLS, and listener history to answer why a service will not start or an endpoint disappeared. No service controls.

### 4. M8 — Incident Lens

Bounded time-window context around a diagnosis/event/time. Nearby package/config/service/container/listener/TLS evidence is context, not proven causation.

### 5. M9 — HostSleuth Workbench

Small troubleshooting tools only: SHA-256/SHA-512, expected checksum verification, file fingerprint comparison, path metadata/hash, DNS, HTTP HEAD/redirect inspection, PEM inspection, and local-file-vs-served-certificate comparison. No web shell.

### 6. M10 — Reboot Story

Boot/shutdown evidence plus what failed to return after reboot. Never invent reboot cause.

### 7. M11 — Optional Safe Actions

Requires explicit security/design review before crossing the read-only boundary. First candidate: Certbot dry-run, explicit renewal, and tightly bounded associated service reload. Disabled by default, previewed, confirmed, audited, no arbitrary command field.

### 8. Later — Redacted Evidence Bundle

Only after redaction rules and threat-model work are mature enough for safe sharing/export.

## Demand-gated

Only pursue if real use proves the need:

- Authentication for direct non-loopback Web UI exposure.
- Reverse-proxy awareness beyond evidence needed for the approved roadmap.
- `.deb` packaging and signed artifacts.
- Privilege separation if future collectors genuinely justify it.
- SQLite if JSON/JSONL becomes a demonstrated limitation.
- Broader baseline-vs-incident comparison beyond M8 if real use proves necessary.
- README screenshot refresh when naturally useful; no cosmetic workstream.

## Not planned unless HostSleuth changes direction

- Multi-host controller/agent architecture.
- General dependency-graph platform.
- Pluggable diagnosis-rule ecosystem.
- AI explanation layer.
- Automatic remediation.
- Generic network-device/SNMP monitoring platform.
- Arbitrary web terminal/command execution.
- Generic server-control panel.

## Rule for new ideas

An idea must strengthen one of HostSleuth's two core jobs and fit the approved order before it becomes active. Interesting is not enough; it must improve a common troubleshooting workflow without making installation, operation, security, or the UI meaningfully worse.
