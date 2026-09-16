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
- [x] M6 — bounded read-only Certificate Story / TLS Detective.
- [x] M6 TLS evidence: handshake, subject/SANs/issuer/serial/validity/lifetime/SHA-256 fingerprint, hostname, trust/chain.
- [x] M6 local listener/container correlation plus native read-only Certbot lineage/renewal/timer/service evidence.
- [x] M6 local-vs-served certificate comparison with conservative stale-served detection.
- [x] M6 focused tests, normal CI, and isolated real OMV native acceptance.
- [x] Publish stable `v0.3.0` from accepted main `6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`.
- [x] v0.3.0 release workflow completed successfully with native amd64/arm64 binaries and `SHA256SUMS`.
- [x] Publish `mjmalleo/hostsleuth:0.3.0` and move `latest`.
- [x] Verify anonymous Docker registry access; `0.3.0` and `latest` resolve to the same linux/amd64 + linux/arm64 OCI index.
- [x] Verify downloaded release checksums and bounded real-consumer amd64 `version` execution.

Stable `v0.3.0` is now the current published release and contains M5 + M6. The live OMV/Arcane deployment and `OMV-Docker-Rebuild` remain intentionally pinned to known-good `mjmalleo/hostsleuth:0.1.0`; publication did not migrate production.

## Next — 3. M7 Service Story

M7 is the next approved milestone, but it has **not** started. Begin only when the owner explicitly tells HostSleuth work to continue into M7.

Correlate systemd/journal, process/listener ownership, port collisions, containers, nearby package/config changes, TLS, and listener history to answer why a service will not start or an endpoint disappeared. Evidence story only; no service controls.

## Then — exact approved order

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
