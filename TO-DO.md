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
- [x] Publish stable v0.4.0 and align production/recovery on 0.4.0.
- [x] Redacted Evidence Bundle threat model/redaction contract accepted.
- [x] Redacted Evidence Bundle bounded typed exporter implemented and adversarially tested.
- [x] Redacted Evidence Bundle PR #40 merged to `main` at `f1d756fa42baf71d9b762127b8d5d6ef6b77b796`.
- [x] Publish stable `v0.5.0` from exact source `04a53f8f0f3f48f7118a9ee9a688820cc000a340`.
- [x] v0.5.0 native amd64/arm64 assets and `SHA256SUMS` published and independently verified.
- [x] v0.5.0 multi-platform Docker image published as `mjmalleo/hostsleuth:0.5.0` and `latest` and independently verified.
- [x] Stage `OMV-Docker-Rebuild` recovery image pin at `0.5.0` without merging ahead of production.
- [x] Redeploy live Arcane-managed HostSleuth to `mjmalleo/hostsleuth:0.5.0`.
- [x] Verify live `/api/about` = v0.5.0, schema 4 / Docker mode, retained pre-upgrade events, and LAN reachability.
- [x] Verify Docker Safe Actions remain `enabled=false` / `available=false` and Action Web/API remains loopback-only.
- [x] Exercise v0.5.0 Evidence Bundle preview/export against a disposable copy of live API evidence; mode `0600` and embedded checksums verified.
- [x] Merge `OMV-Docker-Rebuild` recovery alignment PR #5 at `9387d85acef8d19d913458cf45d48a765b6b8299`.
- [x] Shut down the temporary Arcane tunnel and remove disposable acceptance files.
- [x] M12 — Expected Endpoint Contracts core evaluator with exact DNS, TCP, TLS, optional service, and optional container expectations.
- [x] M12 CLI/API/Web UI integration with ordered Expected-vs-Observed checks and first-mismatch reporting.
- [x] M12 local Go 1.24.13 format/vet/test/build plus disposable HTTP/UI acceptance.
- [x] M15 — Deployment / Permissions Story source implementation, native/HTTP acceptance, full CI, and PR #55 merge.
- [x] M16 — STARTTLS / Mail Service Story source implementation, protocol/HTTP acceptance, full CI, and PR #57 merge.
- [x] M17 — Certificate Rollout Verification source implementation, real-TLS/native acceptance, full CI, and PR #59 merge.
- [x] M18 — Safe Actions II with owner-approved fixed native `service.reload`, real native reload acceptance, full CI, and PR #62 merge.
- [x] Publish stable `v0.9.0` from exact source `b518ed901e2d3f4e95a9bb74ade37d7b3a156540`.
- [x] Independently verify v0.9.0 native amd64/arm64 assets, SHA256SUMS, Docker `0.9.0` + `latest`, and multi-arch manifests.

## Records

- M10: `docs/history/M10-REBOOT-STORY.md`
- M11: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`
- v0.4.0 publication: `docs/history/V0.4.0-PUBLICATION.md`
- Redacted Evidence Bundle design: `docs/design/REDACTED-EVIDENCE-BUNDLE.md`
- Redacted Evidence Bundle closeout: `docs/history/REDACTED-EVIDENCE-BUNDLE.md`
- v0.5.0 publication: `docs/history/V0.5.0-PUBLICATION.md`
- v0.5.0 production/recovery alignment: `docs/history/V0.5.0-PRODUCTION-ALIGNMENT.md`
- M12 design: `docs/design/M12-EXPECTED-ENDPOINT-CONTRACTS.md`
- M12 closeout: `docs/history/M12-EXPECTED-ENDPOINT-CONTRACTS.md`
- v0.6.0 publication: `docs/history/V0.6.0-PUBLICATION.md`
- v0.6.0 production/recovery alignment: `docs/history/V0.6.0-PRODUCTION-ALIGNMENT.md`
- M13 design: `docs/design/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md`
- M13 closeout: `docs/history/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md`
- v0.7.0 publication: `docs/history/V0.7.0-PUBLICATION.md`
- v0.7.0 production/recovery alignment: `docs/history/V0.7.0-PRODUCTION-ALIGNMENT.md`
- M14 design: `docs/design/M14-REVERSE-PROXY-UPSTREAM-STORY.md`
- M14 closeout: `docs/history/M14-REVERSE-PROXY-UPSTREAM-STORY.md`
- M15 design: `docs/design/M15-DEPLOYMENT-PERMISSIONS-STORY.md`
- M15 closeout: `docs/history/M15-DEPLOYMENT-PERMISSIONS-STORY.md`
- M16 design: `docs/design/M16-STARTTLS-MAIL-SERVICE-STORY.md`
- M16 closeout: `docs/history/M16-STARTTLS-MAIL-SERVICE-STORY.md`
- M17 design: `docs/design/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`
- M17 closeout: `docs/history/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`
- M18 security review: `docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md`
- M18 closeout: `docs/history/M18-SAFE-ACTIONS-II.md`
- v0.8.0 publication: `docs/history/V0.8.0-PUBLICATION.md`
- v0.8.0 production/recovery alignment: `docs/history/V0.8.0-PRODUCTION-ALIGNMENT.md`
- v0.9.0 publication: `docs/history/V0.9.0-PUBLICATION.md`
- v0.9.0 production/recovery alignment: `docs/history/V0.9.0-PRODUCTION-ALIGNMENT.md`
- v1.0 readiness contract: `docs/design/V1.0-READINESS.md`
- v1.0 release/rollback checklist: `docs/design/V1.0-RELEASE-CHECKLIST.md`
- v1.0 readiness closeout: `docs/history/V1.0-READINESS.md`

## Current state

Stable/public/live/recovery are aligned on `v0.9.0` / `mjmalleo/hostsleuth:0.9.0`, published from exact source `b518ed901e2d3f4e95a9bb74ade37d7b3a156540`. Docker `0.9.0` and `latest` resolve to verified OCI index `sha256:c99f417419b864756d246232602a8f0fb31067fc11324430a59baec65d62dc5b`.

## v0.9.0 production/recovery alignment

- [x] Publish and independently verify v0.9.0.
- [x] Redeploy the Arcane-managed HostSleuth project from `mjmalleo/hostsleuth:0.8.0` to `mjmalleo/hostsleuth:0.9.0`.
- [x] Verify live `/api/about` = v0.9.0, schema 4 / Docker mode, running image 0.9.0, and all 100 retained events.
- [x] Verify M15 Permissions is live and native service identity remains unavailable in Docker mode.
- [x] Verify M16 with a real SMTP STARTTLS upgrade through TLS/certificate/trust.
- [x] Verify M17 with a real healthy MATCH certificate-rollout comparison.
- [x] Verify Docker Safe Actions remain disabled/unavailable and LAN Action API remains HTTP 403.
- [x] Verify M18 `service.reload` is exposed only as a fixed disabled/unavailable capability in Docker mode.
- [x] Merge `OMV-Docker-Rebuild` PR #10 only after live acceptance, at `813b7a8b76bfd987e97c534c76faf879c510bdec`.
- [x] Record final v0.9.0 production/recovery alignment.

The owner-approved roadmap through M18 is now source-complete, published, live, and recovery-aligned. No additional feature milestone is automatically approved.

## Current next step — v1.0.0 publication decision after readiness closeout

Feature-free v1.0 readiness scope:

- [x] Define bounded v1.0 readiness contract and explicit non-goals.
- [x] Add legacy persisted-state durability regression across current snapshot/event/action-audit writes.
- [x] Add dynamic tab/tabpanel semantics and keyboard navigation across the Admin Console.
- [x] Add visible focus, reduced motion, live-region behavior, and served-asset budgets.
- [x] Pass disposable headless-Chrome deep-link/accessibility and narrow-viewport no-overflow acceptance.
- [x] Pass full GitHub CI on PR #66 exact head `299fe89f81a03eec044340dc85769339c6d62ce3`.
- [x] Merge first readiness slice at `771b923c06d38ac528804468effbe56ffd4c8f78`.
- [x] Add native release checksum verification before installation.
- [x] Add restrictive browser security headers.
- [x] Add disposable real native-install acceptance using published v0.9.0.
- [x] Add Docker-smoke regression checks for browser security headers.
- [x] Pass full six-job GitHub CI on PR #67 exact head `04809c69d398a82c07a97dd1031827854b3591f9`.
- [x] Merge second readiness slice at `ec51b21aca8c4a9de50d2530d2200ba576b45989`.
- [x] Pass real disposable forward-upgrade state acceptance: published v0.9.0 -> current candidate.
- [x] Pass real disposable rollback state acceptance: current candidate -> published v0.9.0.
- [x] Review the v1 publication / production / recovery / rollback procedure against the proven v0.9.0 process.
- [x] Write `docs/design/V1.0-RELEASE-CHECKLIST.md`.
- [x] Write `docs/history/V1.0-READINESS.md`.
- [x] Pass full six-job GitHub CI on readiness closeout PR #68 exact head `2853f3a369a95817524935a93db6904d2ad3d31c`.
- [x] Merge readiness closeout PR #68 at `f93330755e229609f86add144a20d3ce2234cd04`.
- [ ] Owner explicitly approves v1.0.0 publication.
- [ ] Only after approval, publish/verify v1.0.0, roll out production, and align recovery using the accepted checklist.

Stable/public/live/recovery remain v0.9.0 until the separate v1.0.0 publication/rollout sequence is approved and completed.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, or automatic sharing without a separate explicit design decision.

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve explicit schema, allowlist, preview, confirmation, audit, and postcondition verification.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
