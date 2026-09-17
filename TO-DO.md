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
- [x] Publish stable v0.3.0.
- [x] M7 — read-only Service Story.
- [x] M8 — read-only Incident Lens.
- [x] M9 — read-only HostSleuth Workbench.
- [x] M10 — read-only Reboot Story.
- [x] M11 — Optional Safe Actions with one fixed native `service.restart` action.
- [x] M11 explicit enablement + per-service allowlist + exact preview/confirmation + audit-before-execute + postcondition verification.
- [x] M11 loopback-only Action Web/API; Docker mode remains action-unavailable and no writable Docker socket was added.
- [x] M11 disposable real-systemd restart acceptance in CI.
- [x] Publish stable `v0.4.0` from exact source `6566c505b32cc47d96384152a988736173d3f7cd`.
- [x] v0.4.0 linux/amd64 and linux/arm64 binaries plus `SHA256SUMS` published.
- [x] v0.4.0 multi-platform Docker image published as `mjmalleo/hostsleuth:0.4.0` and `latest`.
- [x] Independently verify the published amd64 checksum/version and default-disabled Safe Actions.
- [x] Independently verify Docker `0.4.0` and `latest` share OCI digest `sha256:03b5824fddc50a707e5486033afed3f01d0be76e9adef64292d7a72743578bf0` with amd64 + arm64 manifests.
- [x] Stage `OMV-Docker-Rebuild` PR #4 for recovery image `mjmalleo/hostsleuth:0.4.0` without merging it ahead of production.

M10 implementation/acceptance: `docs/history/M10-REBOOT-STORY.md`.

M11 implementation/security/acceptance: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

v0.4.0 publication record: `docs/history/V0.4.0-PUBLICATION.md`.

## Active operational step — production/recovery alignment

Public stable is now `v0.4.0`, but the live Arcane/Docker deployment is still verified on `mjmalleo/hostsleuth:0.3.0`.

Recovery PR #4 stages `0.4.0` and remains intentionally open/unmerged until production is actually upgraded and verified.

- [ ] Redeploy the existing Arcane-managed HostSleuth project to `mjmalleo/hostsleuth:0.4.0` using an authenticated Arcane session/API.
- [ ] Verify `/api/about` = `v0.4.0`.
- [ ] Verify `/api/snapshot` = schema 4 / Docker mode and reports image `mjmalleo/hostsleuth:0.4.0`.
- [ ] Verify `192.168.2.181:8787` remains reachable and retained state/events remain present.
- [ ] Verify Actions remain unavailable in Docker mode/default deployment.
- [ ] Update recovery PR #4 documentation to mark live/recovery aligned and merge PR #4.
- [ ] Update `CURRENT-HANDOFF.md` with the final aligned live/recovery state.

Do not bypass Arcane authentication, HomeCommander Docker/sudo safeguards, or the uninstalled native `hostsleuth` deployment record to complete these items.

## Later — Redacted Evidence Bundle — NOT STARTED

Only begin after explicit owner direction. Redaction/threat-model rules come before export implementation. The bundle must never silently include credentials, tokens, private keys, configuration contents, or other secrets.

## Guardrails

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve explicit schema, allowlist, preview, confirmation, audit, and postcondition verification.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
