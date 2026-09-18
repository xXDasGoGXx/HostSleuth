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
- v0.8.0 publication: `docs/history/V0.8.0-PUBLICATION.md`
- v0.8.0 production/recovery alignment: `docs/history/V0.8.0-PRODUCTION-ALIGNMENT.md`

## Current state

Stable/public/live/recovery are aligned on `v0.8.0` / `mjmalleo/hostsleuth:0.8.0`. Docker `latest` resolves to the same verified v0.8.0 OCI index.

## Active next step — M17 Certificate Rollout Verification

M16 source closeout is complete:

- [x] Bounded M16 protocol/security design.
- [x] SMTP STARTTLS, IMAP STARTTLS, and POP3 STLS deterministic evidence stages.
- [x] Shared TLS/certificate identity/trust/expiry evidence.
- [x] CLI/API/dedicated STARTTLS Admin Console view.
- [x] Focused protocol parser and upgrade fixtures.
- [x] Local Go 1.24.13 full-test/race/vet/build, all Web JS syntax, and diff validation.
- [x] Disposable SMTP CLI/API/UI acceptance with real TLS upgrade.
- [x] Full GitHub CI on exact PR #57 head `579fff79889d5ab0c3430135d3f5ceba4dc28400`.
- [x] Squash-merge PR #57 to `main` at `e5663d5883acdd2d859ff57b39a447c0018790b3`.
- [x] M16 closeout and source documentation alignment.

Begin only the approved M17 scope:

- [x] Write bounded Certificate Rollout Verification design and safety boundary.
- [x] Define exactly one expected source: SHA-256 fingerprint or explicit reference direct-TLS endpoint.
- [x] Intentionally omit arbitrary certificate-file reads to preserve the no-private-key-read boundary.
- [x] Compare up to 16 explicit endpoints with served fingerprint/identity/SAN/validity/hostname/trust evidence.
- [x] Build deterministic endpoint MATCH / MISMATCH / UNKNOWN matrix with separate health status.
- [x] Add CLI/API/dedicated Cert Rollout Admin Console view.
- [x] Add focused decision fixtures and a real loopback TLS integration fixture.
- [x] Extend permanent Web/Docker smoke coverage.
- [ ] Require full GitHub CI on the exact M17 PR head.
- [ ] Perform native/disposable OMV acceptance if the authorized execution connector returns before closeout.
- [ ] Write M17 closeout record and source documentation alignment only after CI/merge acceptance.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, or automatic sharing without a separate explicit design decision.

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve explicit schema, allowlist, preview, confirmation, audit, and postcondition verification.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
