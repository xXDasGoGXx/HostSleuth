# HostSleuth — TO-DO

This file stays intentionally short. Completed milestone detail belongs in `docs/history/`. The ordered forward plan is in `docs/ROADMAP.md`.

## Completed

- [x] M0 — repository foundation.
- [x] M1 — deployable single-host MVP.
- [x] M2 — deeper deterministic diagnosis.
- [x] M3 — Product Experience and supported Docker acceptance.
- [x] M3.4 — public container distribution and stable v0.1.0 publication.
- [x] M4 — bounded native Debian/Ubuntu package-change timeline.
- [x] Publish stable v0.2.0.
- [x] M5 — bounded native configuration fingerprinting.
- [x] M6 — bounded read-only Certificate Story / TLS Detective.
- [x] Publish stable v0.3.0 from accepted source `6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1` with native amd64/arm64 assets, checksums, and public multi-platform Docker tags.
- [x] M7 — read-only Service Story.
- [x] M7 systemd runtime/result + bounded sanitized journal evidence.
- [x] M7 cgroup/main-PID to listener correlation and deterministic port-collision evidence.
- [x] M7 endpoint/TLS reuse, container-port context, and bounded nearby retained changes.
- [x] M7 Diagnose-integrated Web UI, CLI, and API.
- [x] M7 full tests/vet/build/JS checks plus isolated real OMV and Web/API acceptance.
- [x] M7 regression: hidden listener PID remains unknown and cannot become a false collision claim.

Stable `v0.3.0` remains the current published release. M7 is newer source capability and is not claimed to be included in v0.3.0. The live OMV/Arcane deployment and `OMV-Docker-Rebuild` remain intentionally pinned to known-good `mjmalleo/hostsleuth:0.1.0`.

## Active — 4. M8 Incident Lens

Keep M8 bounded and evidence-first:

- [ ] Allow a diagnosis/event/time to anchor a bounded incident window, initially +/- 15 minutes.
- [ ] Show nearby package events.
- [ ] Show nearby configuration-fingerprint events.
- [ ] Show nearby systemd service changes.
- [ ] Show nearby container changes.
- [ ] Show nearby listener changes.
- [ ] Include TLS/certificate context where retained/current evidence can be connected honestly.
- [ ] Include boot/reboot context only when already available; do not pull M10 forward.
- [ ] Label temporal proximity as context, never proof of causation.
- [ ] Reuse the existing event model; do not create a time-series monitoring database.
- [ ] Integrate with existing Host Story/Diagnose patterns; do not create a generic monitoring dashboard.
- [ ] Add focused tests, normal CI, and one bounded real-host acceptance.

## Consumer/product research — separate backlog, not active scope

Research what users currently assemble from monitoring tools, admin consoles, certificate clients, TLS scanners, scripts, and one-off websites. Record differentiated opportunities without changing the ordered roadmap or smuggling them into M8.

Promising themes to evaluate for later milestones include:

- protocol-aware TLS/STARTTLS inspection for mail and other non-HTTPS services;
- certificate source/destination/served fingerprint verification;
- tightly bounded certificate deployment recipes with preview/postcondition verification;
- rollout consistency checks that can prove different endpoints are serving different certificates;
- expected-endpoint contracts tying a service, listener, protocol/TLS expectation, local certificate evidence, and retained changes into one troubleshooting story.

Any write/action feature remains reserved for M11 security/design review.

## Then — exact approved order

### 5. M9 — HostSleuth Workbench

Small troubleshooting tools only: SHA-256/SHA-512, expected checksum verification, file fingerprint comparison, path metadata/hash, DNS, HTTP HEAD/redirect inspection, PEM inspection, and local-file-vs-served-certificate comparison. No web shell.

### 6. M10 — Reboot Story

Boot/shutdown evidence plus what failed to return after reboot. Never invent reboot cause.

### 7. M11 — Optional Safe Actions

Requires explicit security/design review before crossing the read-only boundary. Actions must be disabled by default, previewed, confirmed, audited, and expose no arbitrary command field.

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
