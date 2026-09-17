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

Release workflow run #5 (`35193948370`) completed successfully.

Publication verification includes:

- GitHub linux/amd64 and linux/arm64 binaries;
- `SHA256SUMS`;
- independent amd64 checksum/version execution;
- confirmation that Safe Actions remain disabled by default;
- public `mjmalleo/hostsleuth:0.4.0` plus `latest`;
- both Docker tags resolving to OCI index `sha256:03b5824fddc50a707e5486033afed3f01d0be76e9adef64292d7a72743578bf0`;
- linux/amd64 and linux/arm64 manifests.

Full record: `docs/history/V0.4.0-PUBLICATION.md`.

## 9. Production / recovery v0.4.0 alignment — COMPLETE

The live Arcane-managed OMV deployment was redeployed through the authenticated Arcane UI to:

`mjmalleo/hostsleuth:0.4.0`

Post-redeploy acceptance confirmed:

- `/api/about` = `v0.4.0`;
- `/api/snapshot` = schema 4 / Docker mode;
- live container image = `mjmalleo/hostsleuth:0.4.0`;
- `192.168.2.181:8787` remains healthy;
- retained state/events survived the redeploy;
- Optional Safe Actions remain disabled/unavailable in Docker/default mode;
- Action Web/API remains loopback-only.

The disaster-recovery repository `xXDasGoGXx/OMV-Docker-Rebuild` was aligned to the same pinned image and PR #4 was merged at:

`ad2bd53ca3a469272eba6343c03936b7c1a04bc0`

No Arcane authentication, HomeCommander Docker/sudo safeguard, or deployment-control boundary was bypassed.

## 10. Redacted Evidence Bundle — COMPLETE IN SOURCE

Owner accepted the threat model/redaction contract in `docs/design/REDACTED-EVIDENCE-BUNDLE.md` on 2026-09-17. The bounded first implementation is complete on PR #40.

Delivered:

- typed current-Snapshot, recent-Event, and Safe-Action-audit export paths only;
- `hostsleuth evidence preview` with no archive write;
- explicit `hostsleuth evidence export` to one local ZIP;
- deterministic bundle-local aliases for sensitive host/infrastructure identifiers;
- secret, credential, URL-userinfo/query, full/truncated-private-key, domain, IPv4/IPv6, path, service/container/network, opaque-ID, and fingerprint redaction;
- preservation of useful ports, prefix lengths, loopback/unspecified semantics, package/version data, statuses, UTC ordering, URL shape, and IPv6-CIDR shape;
- `manifest.json`, summary, redacted JSON payloads, SHA-256 checksums, and final archive SHA-256;
- owner-only output permissions, overwrite refusal, temporary-file cleanup on failure, and bounded record/payload sizes;
- explicit exclusion of raw journal text, raw action command output, arbitrary file/config contents, action confirmation material, arbitrary Workbench inputs, cloud upload, and automatic sharing.

Acceptance includes adversarial leak tests, full local formatting/vet/test/build validation, GitHub test plus amd64/arm64 image builds, Docker smoke, native-actions smoke, and a disposable end-to-end CLI preview/export run whose planted-secret scan was clean and whose checksums and `0600` archive permissions verified.

Full closeout: `docs/history/REDACTED-EVIDENCE-BUNDLE.md`.

This milestone is **not in stable v0.4.0**. Release/publication and production/recovery deployment remain separate future decisions.

No raw/unredacted mode, journal inclusion, arbitrary file inclusion, broader sharing/upload mechanism, or additional Safe Action family is authorized by this milestone.

## Research backlog — not an implementation milestone

The detailed current research artifact is `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`.

Promising differentiated questions include expected-endpoint contracts, resolver/delegation/split-view DNS mismatches, reverse-proxy/upstream problems, permissions/ownership/deployment-path reasoning, STARTTLS inspection, certificate rollout verification, and only narrowly justified future actions using the M11 security model.

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

HostSleuth should feel powerful because it connects deterministic evidence into answers people actually need, not because it exposes every system control in a browser.
