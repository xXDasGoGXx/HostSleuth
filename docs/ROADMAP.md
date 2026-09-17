# HostSleuth — Ordered Product Roadmap

HostSleuth stays a small, local-first Linux troubleshooting tool with two jobs:

1. remember meaningful host changes;
2. explain why a host/service/port is or is not reachable using deterministic evidence.

This roadmap is intentionally ordered. Complete one bounded milestone at a time. Do not skip ahead, broaden a milestone into a generic administration platform, or add unrelated monitoring features just because they are technically possible.

Consumer/product research may identify useful later workflows, but research does not silently reorder this roadmap or expand the active milestone.

## 1. M6 — Certificate Story / TLS Detective — COMPLETE

Delivered bounded read-only TLS/certificate diagnosis including handshake, served certificate metadata/fingerprint, hostname/trust evidence, local listener/container correlation, native Certbot lineage/renewal evidence, and conservative local-vs-served certificate comparison.

No renewal, reload, certificate installation, ACME account management, private-key handling, or arbitrary command execution was added.

## 2. Publish stable v0.3.0 — COMPLETE

Stable `v0.3.0` was published from accepted source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Publication verification included native linux/amd64 and linux/arm64 binaries, `SHA256SUMS`, public `mjmalleo/hostsleuth:0.3.0` plus `latest`, anonymous multi-platform registry verification, and bounded real-consumer execution.

## 3. M7 — Service Story — COMPLETE

Delivered bounded read-only systemd runtime/journal, process/listener ownership, expected-port collision, container-port, endpoint Diagnose/TLS, and retained-change correlation.

## 4. M8 — Incident Lens — COMPLETE

Delivered a bounded read-only +/- 15 minute incident window anchored from an exact time, retained event, or completed diagnosis. Current endpoint evidence is explicitly separated from historical context, and temporal proximity is never presented as proof of causation.

## 5. M9 — HostSleuth Workbench — COMPLETE

Delivered bounded read-only file identity/checksum comparison, DNS inspection, HTTP/redirect inspection, public certificate inspection/comparison, CLI/JSON API, and a dedicated Workbench UI.

Sensitive Workbench Web/API operations are loopback-only and expose no arbitrary commands, file contents, custom HTTP credentials/headers, or private-key viewing.

## 6. M10 — Reboot Story — COMPLETE

Delivered deterministic reboot detection from kernel boot IDs, exact boot-start evidence, bounded previous/current boot journal evidence, direct-evidence-only shutdown classification, current failed-service evidence, retained post-boot recovery correlation, non-causal nearby change context, CLI/API/UI integration, and schema-upgrade protection against false reboot events.

M10 remains read-only and does not infer reboot cause from temporal proximity.

Full closeout: `docs/history/M10-REBOOT-STORY.md`.

## 7. M11 — Optional Safe Actions — COMPLETE

Delivered one fixed native action only:

`service.restart`

Security/product boundaries:

- actions disabled by default;
- explicit `--enable-actions` opt-in;
- explicit per-service `--allow-restart-service UNIT` allowlist;
- conservative `.service` unit validation;
- no arbitrary command, argv, script, shell field, or generic systemd controller;
- trusted absolute `/usr/bin/systemctl` or `/bin/systemctl` for action evidence and execution;
- deterministic preview and exact confirmation;
- durable audit before execution;
- bounded timeout/output and before/after evidence;
- success only after observed `ActiveState=active`;
- loopback-only Action Web/API operations;
- Docker mode action-unavailable and no writable Docker socket;
- no automatic remediation.

Permanent CI includes a disposable real-systemd acceptance on an ephemeral GitHub runner.

Full closeout: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

## 8. Publish stable v0.4.0 — COMPLETE

Stable `v0.4.0` was published from exact accepted source commit:

`6566c505b32cc47d96384152a988736173d3f7cd`

Publication verification included GitHub linux/amd64 and linux/arm64 binaries, `SHA256SUMS`, independent amd64 execution, and public multi-platform Docker verification.

Full record: `docs/history/V0.4.0-PUBLICATION.md`.

## 9. Production / recovery v0.4.0 alignment — COMPLETE

The live Arcane-managed OMV deployment and `xXDasGoGXx/OMV-Docker-Rebuild` recovery definition were aligned on:

`mjmalleo/hostsleuth:0.4.0`

Recovery alignment merged at:

`ad2bd53ca3a469272eba6343c03936b7c1a04bc0`

Post-redeploy acceptance confirmed version/schema/mode, retained state/events, LAN health, and the Docker-mode Safe Actions boundary.

## 10. Redacted Evidence Bundle — COMPLETE

Owner accepted the threat model/redaction contract in `docs/design/REDACTED-EVIDENCE-BUNDLE.md`. PR #40 merged at:

`f1d756fa42baf71d9b762127b8d5d6ef6b77b796`

Delivered:

- typed current-Snapshot, recent-Event, and Safe-Action-audit export paths only;
- `hostsleuth evidence preview` with no archive write;
- explicit `hostsleuth evidence export` to one local ZIP;
- deterministic bundle-local aliases for sensitive host/infrastructure identifiers;
- credential/private-key/domain/IP/path/service/container/network/opaque-ID/fingerprint redaction;
- preservation of useful ports, prefix lengths, loopback/unspecified semantics, package/version data, statuses, UTC ordering, URL shape, and IPv6-CIDR shape;
- manifest, redacted payloads, SHA-256 checksums, owner-only output permissions, overwrite refusal, cleanup on failure, and bounded payload sizes;
- explicit exclusion of raw journal text, raw action command output, arbitrary file/config contents, action confirmation material, arbitrary Workbench inputs, cloud upload, and automatic sharing.

Acceptance included adversarial leak tests, full local formatting/vet/test/build validation, GitHub test plus amd64/arm64 image builds, Docker smoke, native-actions smoke, and disposable end-to-end CLI preview/export verification.

Full closeout: `docs/history/REDACTED-EVIDENCE-BUNDLE.md`.

## 11. Publish stable v0.5.0 — COMPLETE

Stable `v0.5.0` was published from exact source:

`04a53f8f0f3f48f7118a9ee9a688820cc000a340`

Release workflow run `35281793336` completed successfully.

Independent verification confirmed:

- published amd64 checksum and execution;
- version `v0.5.0 (04a53f8f0f3f)`;
- Evidence Bundle CLI present;
- public `mjmalleo/hostsleuth:0.5.0` and `latest` share OCI index `sha256:a17325980d5e9ec9760f9003aa8a9490962bb06ffbbd5a393a0ad31a218e480d`;
- linux/amd64 and linux/arm64 manifests are present.

Full record: `docs/history/V0.5.0-PUBLICATION.md`.

## 12. Production / recovery v0.5.0 alignment — COMPLETE

The live Arcane-managed OMV deployment and `xXDasGoGXx/OMV-Docker-Rebuild` recovery definition are aligned on:

`mjmalleo/hostsleuth:0.5.0`

Live acceptance confirmed:

- `/api/about` = `v0.5.0`;
- `/api/snapshot` = schema 4 / Docker mode;
- retained pre-upgrade events survived;
- Optional Safe Actions remain disabled/unavailable;
- Action Web/API remains loopback-only with HTTP 403 from LAN;
- a disposable v0.5.0 Evidence Bundle preview/export against a copy of live API evidence succeeded with expected bounded files, mode `0600`, and verified checksums.

Recovery alignment PR #5 merged at:

`9387d85acef8d19d913458cf45d48a765b6b8299`

Full record: `docs/history/V0.5.0-PRODUCTION-ALIGNMENT.md`.

## 13. M12 — Expected Endpoint Contracts — COMPLETE

The owner selected Expected Endpoint Contracts as the next bounded milestone.

M12 adds one on-demand read-only contract that can state:

- required `host:port` target;
- optional exact DNS address set;
- TLS expectation: ignore, present, verified, or forbidden;
- optional systemd service expected active;
- optional container expected running.

TCP reachability is always required.

HostSleuth reuses the existing Diagnose engine, renders ordered Expected-vs-Observed checks, and reports the first proven failure. Missing Docker/systemd evidence remains `unknown`; an earlier unknown does not mask a later proven failure.

Delivered interfaces:

- `hostsleuth contract` CLI;
- `GET /api/contract`;
- Web UI Expectations view with handoff to full Diagnose.

Local Go 1.24.13 format/vet/test/build and disposable HTTP/UI acceptance passed before opening the PR. PR #43 then passed the full GitHub CI matrix and merged at `1c193473d9220c34ec2820526df76076bfb41ce9`.

M12 remains read-only and adds no scheduler, alerting, persistent contract database, arbitrary file reads, new privilege, or remediation path.

Design: `docs/design/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

Closeout: `docs/history/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

## 14. Publish stable v0.6.0 — COMPLETE

Stable `v0.6.0` was published from exact accepted source:

`fcb08be51ae3da8cd20dc3929cf9d736b15f170c`

Release workflow run `35284864117` completed successfully. Independent verification confirmed native amd64/arm64 checksums, amd64 execution/version, the M12 `contract` command, and public `0.6.0` plus `latest` on OCI index `sha256:6b9f90209f477ba8213d9c4bf7db7996d6c52d2caae2cfce405822df8bc8ef2a` with linux/amd64 and linux/arm64 manifests.

Full record: `docs/history/V0.6.0-PUBLICATION.md`.

## 15. Production / recovery v0.6.0 alignment — ACTIVE NEXT STEP

Public stable is v0.6.0 while live Arcane production and the disaster-recovery definition remain pinned to v0.5.0.

Ordered rollout:

1. stage the recovery image-pin bump to `mjmalleo/hostsleuth:0.6.0` without merging ahead of production;
2. redeploy the existing Arcane-managed project through the supported authenticated UI;
3. verify live version/schema/mode, retained state/events, LAN health, and Docker-mode Safe Actions boundary;
4. exercise M12 Expected Endpoint Contracts against the live v0.6.0 service;
5. merge the recovery pin only after live acceptance passes;
6. record final v0.6.0 production/recovery alignment.

## Research backlog — not automatically scheduled

The detailed research artifact remains:

`docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`

Remaining promising differentiated questions include resolver/delegation/split-view DNS mismatches, reverse-proxy/upstream problems, permissions/ownership/deployment-path reasoning, STARTTLS inspection, certificate rollout verification, and only narrowly justified future actions using the M11 security model.

Research remains design input only; it does not silently become implementation scope.

## Guardrails that remain in force

Do not drift into:

- multi-host controller/agent architecture;
- generic network-device/SNMP monitoring;
- time-series graphing/monitoring platform behavior;
- arbitrary web terminal or command execution;
- generic package/firewall/configuration administration;
- AI-generated causal claims;
- automatic remediation;
- broad privilege expansion simply to make features easier.

Do not add a raw/unredacted evidence mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or new Safe Action family without a separate explicit design decision.

HostSleuth should feel powerful because it connects deterministic evidence into answers people actually need, not because it exposes every system control in a browser.
