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

## Current state

Stable/public is now `v0.6.0` / `mjmalleo/hostsleuth:0.6.0`, and Docker `latest` resolves to the same verified v0.6.0 OCI index.

Live/recovery remain aligned on `mjmalleo/hostsleuth:0.5.0` until the separate v0.6.0 rollout passes.

## Active next step — production / recovery v0.6.0 alignment

- [x] Open M12 PR only after the branch remains locally green.
- [x] Require full GitHub CI on the exact PR head before merge.
- [x] Merge M12 PR #43 to `main` at `1c193473d9220c34ec2820526df76076bfb41ce9`.
- [x] Publish stable `v0.6.0` from exact source `fcb08be51ae3da8cd20dc3929cf9d736b15f170c`.
- [x] Independently verify v0.6.0 native assets/checksums, M12 contract CLI, and the multi-platform Docker image.
- [ ] Stage recovery image pin at 0.6.0 without merging ahead of production.
- [ ] Redeploy live Arcane production to `mjmalleo/hostsleuth:0.6.0`.
- [ ] Verify live v0.6.0, retained state/events, Docker action boundary, and M12 Expected Endpoint Contract behavior.
- [ ] Merge recovery alignment only after production acceptance.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, or automatic sharing without a separate explicit design decision.

Do not add more action families merely because M11 created the framework. A future action must independently justify its privilege cost and preserve explicit schema, allowlist, preview, confirmation, audit, and postcondition verification.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
