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
- [x] M8 — read-only Incident Lens with event/time/diagnosis anchors and bounded +/- 15 minute context.
- [x] M9 — read-only HostSleuth Workbench for bounded file identity, DNS, HTTP, and public-certificate inspection.
- [x] M10 — read-only Reboot Story.
- [x] M10 schema v4 boot ID and exact `/proc/stat` `btime` evidence.
- [x] M10 deterministic known-boot-ID change detection with no uptime inference.
- [x] M10 schema-upgrade regression: an older snapshot without boot ID cannot become a false reboot event.
- [x] M10 CLI, JSON API, runtime Web assets, and Reboot-tab integration.
- [x] M10 bounded previous/current boot journal evidence with inaccessible/unavailable history reported as `unknown`.
- [x] M10 orderly-vs-abnormal classification only when direct evidence supports it.
- [x] M10 service/listener/container recovery evidence requiring retained post-boot evidence plus current-state agreement.
- [x] M10 package/kernel/system/configuration context retained as non-causal context.
- [x] M10 focused and full Go validation, vet, formatting, individual/exact-served JavaScript syntax, native build, Docker smoke, linux/amd64 build, and linux/arm64 build.
- [x] M10 isolated real-OMV acceptance without changing the live HostSleuth deployment.
- [x] M10 real-host regression: `No journal files were opened due to insufficient permissions.` is unavailable evidence and remains `unknown` without privilege expansion.
- [x] Post-M10 live deployment audit verified `mjmalleo/hostsleuth:0.3.0` in production.
- [x] With explicit owner approval, `OMV-Docker-Rebuild` recovery was repinned to `mjmalleo/hostsleuth:0.3.0` in PR #3 without redeploying production.

Full M10 implementation and acceptance record: `docs/history/M10-REBOOT-STORY.md`.

Stable `v0.3.0` remains the current published release. M7, M8, M9, and M10 are newer source capabilities and are not claimed to be included in v0.3.0.

Production and disaster recovery are now aligned at the HostSleuth image-tag level on `mjmalleo/hostsleuth:0.3.0`. The recovery repin changed only the disaster-recovery Git source of truth; it did not restart or redeploy the live HostSleuth container.

## Active — M11 Optional Safe Actions — IN PROGRESS

Owner direction to begin M11 was given on 2026-09-17. The explicit security/design review is captured in `docs/history/M11-OPTIONAL-SAFE-ACTIONS-WIP.md`.

Current implementation on `m11-safe-actions` / PR #37:

- [x] Actions disabled by default; explicit `--enable-actions` opt-in required.
- [x] Fixed `service.restart` action only; no generic command field or generic service controller.
- [x] Repeated `--allow-restart-service UNIT` allowlist with conservative unit-name validation.
- [x] Deterministic preview with exact target/effect/argv and exact confirmation value.
- [x] Direct argv execution of `systemctl restart UNIT`; no shell interpretation.
- [x] Durable pre-execution audit requirement plus final outcome audit.
- [x] Bounded before/after systemd evidence and required `ActiveState=active` success postcondition.
- [x] Loopback-only Action Web/API surface with JSON-only requests and explicit action header.
- [x] Docker mode remains unavailable for systemd restart and receives no writable Docker socket.
- [x] CLI, JSON API, Web UI, and focused security/regression tests implemented.
- [x] CI definition extended for Actions JavaScript and disabled-by-default Docker smoke coverage.
- [ ] Full PR CI green.
- [ ] Isolated native-host acceptance of preview, denied path, audit path, and one deliberately selected harmless service restart.
- [ ] Replace WIP history with final M11 closeout; update roadmap/handoff and mark M11 complete.

Do not add additional action families during M11 merely because the framework exists. Prove this one narrow action first.

## Consumer/product research — separate backlog, not active scope

Continue studying recurring troubleshooting workflows where HostSleuth can add correlation, verification, and boundedness without becoming a generic administration platform.

Promising themes include:

- endpoint-path and expected-state contracts;
- DNS resolver/delegation/split-view evidence;
- HTTP/reverse-proxy/upstream mismatch evidence;
- port/listener/bind ownership;
- file permissions/ownership/deployment-path problems;
- container disappearance/dependency evidence;
- protocol-aware TLS/STARTTLS inspection for mail and other non-HTTPS services;
- boot/recovery workflows;
- certificate source -> destination -> actually-served fingerprint verification and rollout consistency.

## Later — Redacted Evidence Bundle

Only after redaction rules and threat-model work are mature enough for safe sharing/export.

## Demand-gated

Only pursue if real use proves the need:

- Authentication for direct non-loopback Web UI exposure.
- Reverse-proxy awareness beyond evidence needed for the approved roadmap.
- `.deb` packaging and signed artifacts.
- Privilege separation if future collectors genuinely justify it.
- SQLite if JSON/JSONL becomes a demonstrated limitation.
- Broader baseline-vs-incident comparison if real use proves necessary.
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
