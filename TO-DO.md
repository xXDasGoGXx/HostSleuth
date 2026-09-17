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

M10 implementation/acceptance: `docs/history/M10-REBOOT-STORY.md`.

M11 implementation/security/acceptance: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

v0.4.0 publication record: `docs/history/V0.4.0-PUBLICATION.md`.

## Redacted Evidence Bundle — implementation/acceptance complete on PR #40

Branch: `redacted-evidence-bundle`

Design contract: `docs/design/REDACTED-EVIDENCE-BUNDLE.md`

- [x] Owner accepted the redaction/threat-model contract.
- [x] Implement typed Snapshot/Event/ActionAudit redaction; no generic serialize-everything path.
- [x] Implement deterministic bundle-local pseudonymization and secret/private-key scrubbing.
- [x] Implement `hostsleuth evidence preview` with no archive write.
- [x] Implement bounded local ZIP export with manifest, checksums, owner-only permissions, overwrite refusal, and cleanup on failure.
- [x] Enforce 200-event / 100-action-audit / 1 MiB JSON / 2 MiB total payload limits.
- [x] Add adversarial tests for credentials, tokens, cookies, Authorization/Bearer, URL userinfo/query strings, full/truncated private keys, domains, IPv4/IPv6, paths, services, containers, networks, opaque IDs, and fingerprints.
- [x] Preserve useful URL and IPv6-CIDR structure after redaction.
- [x] Verify the redaction policy is independent of deployment mode.
- [x] Keep raw journal text, raw action command output, arbitrary file/config contents, action confirmation material, cloud upload, and automatic sharing out of scope.
- [x] Keep the first implementation CLI-only; no Web/API surface is required for this bounded version.
- [x] Pass full Go formatting/vet/tests/native build plus GitHub test, amd64/arm64 image build, Docker smoke, and native-actions smoke.
- [x] Pass disposable end-to-end preview/export acceptance with clean planted-secret leak scan, verified bundle checksums, and final archive mode `0600`.
- [x] Document the security boundary and closeout before merge.

Closeout: `docs/history/REDACTED-EVIDENCE-BUNDLE.md`.

## Current stable/live/recovery boundary

Stable/public/live/recovery remain on `v0.4.0` / `mjmalleo/hostsleuth:0.4.0`. The Redacted Evidence Bundle is source work on PR #40 only until a separate release/publication decision is made.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, or automatic sharing without a separate explicit design decision.

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve explicit schema, allowlist, preview, confirmation, audit, and postcondition verification.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
