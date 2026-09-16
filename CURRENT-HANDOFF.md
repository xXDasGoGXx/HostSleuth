# HostSleuth — Current Handoff

Last updated: 2026-09-16

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform or browser-based server administration suite.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Important source anchors:

- M5 code merge: `39cb2847ea66cd04ddb9cb0f7a3111f5e659456b`
- M5 documentation closeout: `f688b3c8aedb3e4143c17b6dd5049a86f098ed20`
- ordered roadmap file: `docs/ROADMAP.md`

Always re-check current `main` before starting code or a release rather than relying on a self-referential SHA in this file.

Completed milestones:

- M0 — repository foundation;
- M1 — deployable single-host MVP;
- M2 — deeper deterministic diagnosis;
- M3 — Product Experience;
- M3.4 — Public Container Distribution;
- M4 — package-change timeline;
- M5 — configuration fingerprinting.

Published stable release:

`v0.2.0`

Published Docker tags:

- `mjmalleo/hostsleuth:0.2.0`
- `mjmalleo/hostsleuth:latest`

Both public tags were verified anonymously as a multi-platform OCI index for:

- `linux/amd64`
- `linux/arm64`

Verified v0.2.0/latest index digest:

`sha256:6892362d3f5fc6d30ae6bde7235ee976f169f1be8d42d181c55f53cf98600c91`

M5 is merged source capability and is **not** claimed to be present in the already-published v0.2.0 artifacts.

## M5 — configuration fingerprinting

M5 extends **what changed?** without storing configuration contents.

Native Linux fingerprints this deliberately small explicit set:

- `/etc/hosts`
- `/etc/fstab`
- `/etc/ssh/sshd_config`
- `/etc/docker/daemon.json`
- `/etc/nftables.conf`

Snapshots store only canonical path, state, SHA-256 fingerprint, and size for readable regular files. States are `present`, `missing`, or `unreadable`. File contents are not stored.

Configuration appeared/disappeared/content-changed records reuse the existing Changes timeline as `category=configuration`. Unreadable transitions remain quiet rather than becoming false changes.

Snapshot schema 3 baselines existing configuration fingerprints across the schema 2 -> 3 upgrade while preserving M4 package-event behavior.

PR #26 CI run `35129730468` / #133 passed formatting, vet, focused tests, JS syntax, native build, Compose validation, amd64/arm64 builds, and Docker runtime smoke.

Real OMV Debian acceptance confirmed the five explicit candidates using an isolated branch build/state. Four readable files produced 64-character SHA-256 fingerprints and size metadata; `/etc/docker/daemon.json` was truthfully reported `unreadable` without privilege escalation. The live deployment was not replaced.

## Live OMV / recovery boundary

The known-good live HostSleuth remains managed through Arcane on OMV host `192.168.2.181` using:

`mjmalleo/hostsleuth:0.1.0`

Deployment state:

- UI/API: `http://192.168.2.181:8787`
- persistent state: `/srv/docker/volumes/hostsleuth/data`
- Compose path: `/srv/docker/volumes/compose/hostsleuth/compose.yaml`
- trusted-LAN bind: `192.168.2.181:8787`

`xXDasGoGXx/OMV-Docker-Rebuild` intentionally remains pinned to `mjmalleo/hostsleuth:0.1.0` so recovery matches the actual live deployment.

Do not change the live deployment or recovery definition merely to chase release numbers. M4/M5 evidence is native-only under the current least-privilege Docker boundary.

## Ordered roadmap — owner approved

The owner explicitly approved the following order. Stay in this order unless the owner deliberately changes it. Full detail is in `docs/ROADMAP.md`.

### 1. M6 — Certificate Story / TLS Detective — ACTIVE NEXT

Read-only TLS/certificate troubleshooting integrated into HostSleuth diagnosis:

- certificate subject/SANs/issuer/serial/validity/days remaining/SHA-256 fingerprint;
- handshake, hostname match, bounded trust/chain evidence;
- local listener/process/container correlation;
- native Certbot discovery and read-only renewal/timer evidence when present;
- local certificate vs actually served certificate fingerprint comparison;
- evidence-backed detection of stale-served-certificate situations.

No renewal button or service reload in M6.

### 2. Publish stable v0.3.0

After M6 acceptance, publish M5 + M6 together using the existing controlled release path. Stop at the publication boundary for explicit owner approval.

### 3. M7 — Service Story

Answer why a service will not start or why an endpoint disappeared by correlating systemd/journal, process/listener ownership, port collisions, containers, package/config changes, TLS, and listener history. Evidence story only; no service controls.

### 4. M8 — Incident Lens

Anchor a bounded time window around a diagnosis/event/time and show temporally nearby package/config/service/container/listener/TLS evidence. Temporal proximity is context, not proven causation.

### 5. M9 — HostSleuth Workbench

Small practical troubleshooting tools only: SHA-256/SHA-512, expected checksum verification, file fingerprint comparison, path metadata/hash, DNS inspection, HTTP HEAD/redirects, PEM inspection, and local-file-vs-served-certificate comparison. No web shell or arbitrary command box.

### 6. M10 — Reboot Story

Explain boot/shutdown evidence and what failed to come back using kernel/package/config/service/listener/container context without inventing a reboot cause.

### 7. M11 — Optional Safe Actions

This is the first planned milestone that may cross HostSleuth's read-only boundary, so it requires explicit design/security review first. Initial candidate: Certbot dry-run, explicit renewal, and tightly bounded associated service reload. Actions must be disabled by default, explicitly enabled, previewed, confirmed, and audited. No arbitrary command execution.

### 8. Later — Redacted Evidence Bundle

Only after redaction rules and threat-model work are mature enough. Export selected HostSleuth evidence while excluding secrets/config contents/private keys/tokens.

## Product guardrails

Do not drift into:

- multi-host controller/agent architecture;
- generic network-device/SNMP monitoring;
- time-series monitoring/graph platform behavior;
- arbitrary web terminal or command execution;
- generic package/firewall/configuration administration;
- AI-generated causal claims;
- automatic remediation;
- broad privilege expansion merely to make features easier.

The product should feel powerful because it connects deterministic evidence into answers people actually need.

## Release/deployment boundary

Do not create a new GitHub release/tag, Docker tag/image, or change the live OMV deployment without explicit owner approval at the relevant boundary.

## Repository reading order

When resuming, read:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`
4. `docs/ROADMAP.md`
5. `docs/history/DEVELOPMENT-HISTORY.md`
6. `docs/design/HOST-STORY-UI.md`
7. `.github/workflows/ci.yml`
8. `.github/workflows/release.yml`
9. `Dockerfile`
10. `compose.yaml`
