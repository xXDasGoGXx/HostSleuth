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
- [x] M8 — read-only Incident Lens with event/time/diagnosis anchors and bounded +/- 15 minute context.
- [x] M9 — read-only HostSleuth Workbench.
- [x] M9 file SHA-256/SHA-512 plus expected-checksum verification and path metadata.
- [x] M9 file-to-file fingerprint comparison.
- [x] M9 common DNS record inspection through the host system resolver.
- [x] M9 direct HTTP HEAD and bounded redirect-chain inspection.
- [x] M9 public PEM certificate inspection and exact local-file-vs-served certificate comparison.
- [x] M9 dedicated CLI/API/Web UI with no arbitrary command box, file editor, custom HTTP headers/credentials, or hidden shell hooks.
- [x] M9 Workbench Web/API operations restricted to loopback clients; local CLI and SSH-tunnel workflows remain available.
- [x] M9 focused tests, full validation, isolated real OMV CLI acceptance, and isolated API/UI smoke.
- [x] CI now syntax-checks every UI fragment plus the exact concatenated served JavaScript and includes a Docker Workbench smoke path.

Stable `v0.3.0` remains the current published release. M7, M8, and M9 are newer source capabilities and are not claimed to be included in v0.3.0. The live OMV/Arcane deployment and `OMV-Docker-Rebuild` remain intentionally pinned to known-good `mjmalleo/hostsleuth:0.1.0`.

## Active — 6. M10 Reboot Story

M10 is the next approved milestone. Keep it evidence-first and read-only.

- [ ] Capture current boot time and bounded previous boot/shutdown evidence where available.
- [ ] Distinguish orderly vs abnormal shutdown only when direct evidence supports it.
- [ ] Correlate package/kernel/configuration changes near reboot without claiming cause from timing alone.
- [ ] Surface services failed after boot.
- [ ] Surface listeners that existed before but did not return only when retained evidence supports that statement.
- [ ] Include container state changes relevant to boot recovery.
- [ ] Attach current service/listener/certificate context for missing endpoints only where evidence connects honestly.
- [ ] Do not invent a reboot cause.
- [ ] Reuse Incident Lens/event primitives instead of creating a second history store.
- [ ] Focused tests, normal CI, and bounded real-host acceptance.

## Consumer/product research — separate backlog, not active scope

Continue studying what users currently assemble from CLI tools, admin consoles, monitoring products, DNS/HTTP/TLS websites, package tools, log viewers, scripts, and other one-off troubleshooting utilities. Preserve differentiated opportunities without changing the ordered roadmap.

The research backlog is intentionally broader than certificate management. Promising themes include:

- expected-endpoint contracts tying service/listener/protocol/certificate expectations to retained changes;
- protocol-aware TLS/STARTTLS inspection for mail and other non-HTTPS services;
- certificate source -> destination -> actually-served fingerprint verification and rollout consistency;
- DNS path/delegation or resolver-comparison workflows when they answer a concrete troubleshooting question;
- HTTP/reverse-proxy evidence that explains redirect, host-header, or upstream mismatches without becoming a proxy manager;
- permissions/ownership/deployment-path evidence that explains why a service cannot consume a file it is expected to use;
- tightly bounded safe-action recipes with preview, audit, postcondition verification, and no arbitrary shell.

Any write/action feature remains reserved for M11 security/design review or another explicitly approved later milestone.

## Then — exact approved order

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
