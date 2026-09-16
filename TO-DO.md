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
- [x] M3.4 — public container distribution.
- [x] Publish stable `v0.1.0` native Linux amd64/arm64 release assets and checksums.
- [x] Publish public multi-platform `mjmalleo/hostsleuth:0.1.0` and `mjmalleo/hostsleuth:latest` images.
- [x] Verify anonymous public Docker retrieval and amd64/arm64 manifests.
- [x] Consumer-smoke the real published image on an ephemeral runner: API, snapshot, Web UI, diagnosis, teardown.
- [x] Update `xXDasGoGXx/OMV-Docker-Rebuild` to pin `mjmalleo/hostsleuth:0.1.0`.
- [x] Redeploy the live OMV/Arcane HostSleuth project from the pinned public image.
- [x] Verify live OMV `/api/about`, `/api/snapshot`, Web UI, and reachable/high-confidence diagnosis after redeploy.
- [x] Accept one bounded Host Story evidence-first UI pass for v0.1.0.

The v0.1.0 UI and M3/M3.4 release path are closed unless real use exposes a concrete defect.

## Active milestone — M4: package-change timeline

M4 has one job: make **“what changed?”** more useful with package install/update/remove history.

- [ ] Inspect the existing event model, collectors, persistence, tests, and current Debian/Ubuntu package logs before coding.
- [ ] Read bounded package install/update/remove history from supported local `apt` / `dpkg` logs.
- [ ] Normalize package changes into the existing event timeline instead of creating a separate package dashboard.
- [ ] Preserve read-only, local-first behavior.
- [ ] Keep unsupported platforms truthful and quiet rather than requiring configuration.
- [ ] Avoid package-management actions, update buttons, repository management, alerts, and new settings.
- [ ] Add focused parser/event tests.
- [ ] Run normal CI.
- [ ] Perform one real-host M4 acceptance pass, then merge and close the milestone if no concrete defect appears.

## After M4 — one capability at a time

These fit HostSleuth's two core jobs, but none becomes active automatically:

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
