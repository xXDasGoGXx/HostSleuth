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
- [x] M7 — read-only Service Story with systemd/journal/listener/endpoint/change correlation.
- [x] M7 real-host regression: hidden listener PID remains unknown and cannot become a false collision claim.
- [x] M8 — read-only Incident Lens.
- [x] M8 explicit event/time anchor with bounded +/- 15 minute retained-event window.
- [x] M8 package/configuration/service/container/listener and other retained event categories shown without causal claims.
- [x] M8 optional current endpoint Diagnose/TLS evidence kept explicitly separate from historical event context.
- [x] M8 Changes-integrated Web UI, CLI `incident`, and `/api/incident-lens`.
- [x] M8 focused tests and isolated real OMV CLI/API/UI acceptance.

Stable `v0.3.0` remains the current published release. M7 and M8 are newer source capabilities and are not claimed to be included in v0.3.0. The live OMV/Arcane deployment and `OMV-Docker-Rebuild` remain intentionally pinned to known-good `mjmalleo/hostsleuth:0.1.0`.

## Next — 5. M9 HostSleuth Workbench

M9 is the next approved milestone. Keep it a deliberately small set of read-only troubleshooting tools, not a miscellaneous utilities page or browser shell.

Initial approved candidate set:

- [ ] SHA-256 / SHA-512 calculation for a selected local file without storing contents.
- [ ] Expected-checksum verification.
- [ ] Compare two files by fingerprint.
- [ ] Path owner/group/permissions/mtime/size/hash inspection.
- [ ] DNS inspection for common records with deterministic evidence.
- [ ] HTTP HEAD / redirect-chain inspection.
- [ ] PEM certificate inspection.
- [ ] Local certificate file fingerprint vs certificate actually served by an endpoint.
- [ ] Use consumer/product research to refine the strongest workflows without silently adding unrelated tools.
- [ ] No arbitrary command box, file editor, generic server controls, or hidden shell hooks.
- [ ] Focused tests, normal CI, and bounded real-host acceptance.

## Consumer/product research — separate backlog, not active scope

Research what users currently assemble from monitoring tools, admin consoles, certificate clients, TLS scanners, scripts, and one-off websites. Preserve differentiated opportunities without changing the ordered roadmap.

Promising later themes include:

- protocol-aware TLS/STARTTLS inspection for mail and other non-HTTPS services;
- certificate source -> destination -> actually-served fingerprint verification;
- certificate rollout consistency across several local consumers/endpoints;
- expected-endpoint contracts tying service/listener/protocol/certificate expectations to retained changes;
- tightly bounded certificate deployment recipes with preview, audit, reload, and postcondition verification.

Any write/action feature remains reserved for M11 security/design review or another explicitly approved later milestone.

## Then — exact approved order

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
