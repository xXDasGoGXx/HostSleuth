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
- [x] Align live Arcane/Docker production and `OMV-Docker-Rebuild` recovery on `mjmalleo/hostsleuth:0.4.0`.
- [x] Redacted Evidence Bundle threat model/redaction contract accepted.
- [x] Redacted Evidence Bundle bounded typed exporter implemented and adversarially tested.
- [x] Redacted Evidence Bundle PR #40 merged to `main` at `f1d756fa42baf71d9b762127b8d5d6ef6b77b796`.
- [x] Publish stable `v0.5.0` from exact source `04a53f8f0f3f48f7118a9ee9a688820cc000a340`.
- [x] v0.5.0 linux/amd64 and linux/arm64 binaries plus `SHA256SUMS` published.
- [x] v0.5.0 multi-platform Docker image published as `mjmalleo/hostsleuth:0.5.0` and `latest`.
- [x] Independently verify the published amd64 checksum/version and Evidence Bundle CLI.
- [x] Independently verify Docker `0.5.0` and `latest` share OCI index `sha256:a17325980d5e9ec9760f9003aa8a9490962bb06ffbbd5a393a0ad31a218e480d` with amd64 + arm64 manifests.

M10 implementation/acceptance: `docs/history/M10-REBOOT-STORY.md`.

M11 implementation/security/acceptance: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

v0.4.0 publication record: `docs/history/V0.4.0-PUBLICATION.md`.

Redacted Evidence Bundle design: `docs/design/REDACTED-EVIDENCE-BUNDLE.md`.

Redacted Evidence Bundle closeout: `docs/history/REDACTED-EVIDENCE-BUNDLE.md`.

v0.5.0 publication record: `docs/history/V0.5.0-PUBLICATION.md`.

## Current release/live/recovery boundary

Public stable is now:

`v0.5.0` / `mjmalleo/hostsleuth:0.5.0`

Docker `latest` also resolves to the verified v0.5.0 OCI index.

The live Arcane deployment and `OMV-Docker-Rebuild` disaster-recovery definition remain on:

`mjmalleo/hostsleuth:0.4.0`

## Next — production / recovery v0.5.0 alignment

- [ ] Stage the recovery image-pin bump from `0.4.0` to `0.5.0` without merging it ahead of production.
- [ ] Redeploy the existing Arcane-managed HostSleuth project to `mjmalleo/hostsleuth:0.5.0` through the supported authenticated path.
- [ ] Verify `/api/about`, `/api/snapshot`, LAN reachability, retained state/events, and Docker-mode Safe Actions boundary after redeploy.
- [ ] Exercise the published v0.5.0 Evidence Bundle CLI against retained HostSleuth state using a safe local output path and inspect the resulting preview/archive behavior.
- [ ] Update and merge the recovery pin only after live v0.5.0 acceptance passes.
- [ ] Record final live/recovery alignment in `CURRENT-HANDOFF.md`.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, or automatic sharing without a separate explicit design decision.

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve explicit schema, allowlist, preview, confirmation, audit, and postcondition verification.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
