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
- [x] M8 — read-only Incident Lens.
- [x] M9 — read-only HostSleuth Workbench.
- [x] M10 — read-only Reboot Story.
- [x] Post-M10 live deployment audit verified `mjmalleo/hostsleuth:0.3.0` in production.
- [x] With explicit owner approval, `OMV-Docker-Rebuild` recovery was repinned to `mjmalleo/hostsleuth:0.3.0` without redeploying production.
- [x] M11 — Optional Safe Actions with one fixed native `service.restart` action.
- [x] M11 actions disabled by default and gated by explicit `--enable-actions` plus per-service allowlisting.
- [x] M11 trusted absolute systemctl path, exact preview/confirmation, durable audit-before-execute, timeout, bounded before/after evidence, and `ActiveState=active` postcondition.
- [x] M11 loopback-only Action Web/API surface; Docker mode remains action-unavailable and gains no writable Docker socket.
- [x] M11 CLI, JSON API, Web UI, security/regression tests, amd64/arm64 container builds, Docker smoke, and disposable real-systemd restart acceptance.

M10 implementation and acceptance record: `docs/history/M10-REBOOT-STORY.md`.

M11 implementation, security model, and acceptance record: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

Stable `v0.3.0` remains the current published release. M7, M8, M9, M10, and M11 are newer source capabilities and are not claimed to be included in v0.3.0.

Production and disaster recovery remain aligned at the HostSleuth image-tag level on `mjmalleo/hostsleuth:0.3.0`. M11 did not restart, redeploy, publish, or enable actions on the production Docker deployment.

## Next decision — NOT STARTED

M11 source completion does not automatically authorize a release or production upgrade. Decide separately whether the accepted post-v0.3.0 source should become a new public release and whether production/recovery should later move to it.

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve the explicit schema, allowlist, preview, confirmation, audit, and postcondition model.

## Later — Redacted Evidence Bundle — NOT STARTED

Only begin after explicit owner direction. The bundle must have a mature redaction/threat model before export and must never silently include credentials, tokens, private keys, configuration contents, or other secrets.

## Consumer/product research — separate backlog, not active scope

Promising themes remain:

- endpoint-path and expected-state contracts;
- DNS resolver/delegation/split-view evidence;
- HTTP/reverse-proxy/upstream mismatch evidence;
- port/listener/bind ownership;
- file permissions/ownership/deployment-path problems;
- container disappearance/dependency evidence;
- protocol-aware TLS/STARTTLS inspection for mail and other non-HTTPS services;
- boot/recovery workflows;
- certificate source -> destination -> actually-served fingerprint verification and rollout consistency.

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
