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

Stable `v0.2.0` is the current published release. The live OMV/Arcane deployment remains intentionally pinned to known-good `mjmalleo/hostsleuth:0.1.0`; do not upgrade it merely to chase the release number because M4 package-history evidence remains unavailable in the default Docker deployment.

## Active milestone — M5: configuration fingerprinting

M5 has one job: record that important configuration files changed without storing their contents or secrets by default.

- [ ] Inspect existing snapshot/diff/event flow and real host configuration candidates before coding.
- [ ] Choose a deliberately small native-Linux default set of high-value configuration files.
- [ ] Record path plus deterministic fingerprint and minimal non-secret metadata only.
- [ ] Keep file contents out of snapshots/events by default.
- [ ] Emit configuration changes into the existing Changes timeline rather than creating a new dashboard.
- [ ] Baseline fingerprints on the first M5-aware snapshot so existing files do not create a false backlog.
- [ ] Keep missing/unreadable files quiet and truthful.
- [ ] Preserve read-only behavior.
- [ ] Do not recursively crawl `/etc`.
- [ ] Do not add a configuration editor, diff viewer, secret storage, remediation, watcher daemon, alerts, or settings framework.
- [ ] Add focused fingerprint/baseline/event tests.
- [ ] Run normal CI.
- [ ] Perform one real-host acceptance pass, then merge and close M5 if no concrete defect appears.

## After M5 — one capability at a time

These remain candidates only; none becomes active automatically:

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
