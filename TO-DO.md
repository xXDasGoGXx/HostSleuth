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
- [x] Align live Arcane/Docker production and `OMV-Docker-Rebuild` recovery on `mjmalleo/hostsleuth:0.4.0`.
- [x] Redacted Evidence Bundle threat model/redaction contract accepted.
- [x] Redacted Evidence Bundle bounded typed exporter implemented and adversarially tested.
- [x] Redacted Evidence Bundle PR #40 merged to `main` at `f1d756fa42baf71d9b762127b8d5d6ef6b77b796`.

M10 implementation/acceptance: `docs/history/M10-REBOOT-STORY.md`.

M11 implementation/security/acceptance: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

v0.4.0 publication record: `docs/history/V0.4.0-PUBLICATION.md`.

Redacted Evidence Bundle design: `docs/design/REDACTED-EVIDENCE-BUNDLE.md`.

Redacted Evidence Bundle closeout: `docs/history/REDACTED-EVIDENCE-BUNDLE.md`.

## Redacted Evidence Bundle — merged source state

The merged first version provides:

- `hostsleuth evidence preview` with no archive write;
- `hostsleuth evidence export [--output PATH]` for a bounded local ZIP;
- typed Snapshot/Event/ActionAudit redaction only;
- deterministic bundle-local pseudonymization and credential/private-key scrubbing;
- manifest, redaction counts, file SHA-256 values, `checksums.txt`, and final archive SHA-256;
- 200-event / 100-action-audit / 1 MiB JSON / 2 MiB total payload limits;
- owner-only archive permissions, overwrite refusal, and temporary cleanup on failure;
- explicit exclusion of raw journal text, raw action command output, arbitrary file/config contents, action confirmation material, cloud upload, and automatic sharing;
- full local and GitHub CI validation plus disposable end-to-end leak-scan acceptance.

## Current release boundary

Stable/public/live/recovery remain on:

`v0.4.0` / `mjmalleo/hostsleuth:0.4.0`

The Redacted Evidence Bundle is merged to source `main` but is **not in stable v0.4.0**.

A new release/publication decision is the next possible step only if explicitly chosen. Production/recovery rollout must remain a separate later step after release verification.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, or automatic sharing without a separate explicit design decision.

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve explicit schema, allowlist, preview, confirmation, audit, and postcondition verification.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
