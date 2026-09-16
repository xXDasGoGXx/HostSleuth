# HostSleuth — Current Handoff

Last updated: 2026-09-16

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform or browser-based server administration suite.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Always re-check current `main` before starting code, a release, or a deployment rather than relying on a self-referential SHA in this file.

Completed milestones:

- M0 — repository foundation;
- M1 — deployable single-host MVP;
- M2 — deeper deterministic diagnosis;
- M3 — Product Experience;
- M3.4 — Public Container Distribution;
- M4 — package-change timeline;
- M5 — configuration fingerprinting;
- M6 — Certificate Story / TLS Detective;
- stable v0.3.0 publication;
- M7 — Service Story;
- M8 — Incident Lens.

Published stable release:

`v0.3.0`

Release source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Published Docker tags:

- `mjmalleo/hostsleuth:0.3.0`
- `mjmalleo/hostsleuth:latest`

Both resolve anonymously to the same multi-platform OCI index:

`sha256:127b388fbf794841b22d06b281fe89dc1500188fdb4215eec023b392aa98c05d`

Platforms:

- `linux/amd64`
- `linux/arm64`

M7 and M8 are newer source capabilities and are **not** claimed to be present in the already-published v0.3.0 artifacts.

## M7 — Service Story

M7 answers why a selected native systemd service is failed/inactive or why an expected endpoint disappeared without turning HostSleuth into a service manager.

Delivered evidence includes bounded systemd runtime/result and sanitized current-boot journal evidence, cgroup/main-PID to listener ownership, deterministic positive-evidence-only port collision, container-port context, existing Diagnose/TLS reuse, and bounded nearby retained changes. Hidden listener ownership remains `unknown` rather than becoming a false collision.

M7 remains read-only. Full implementation and acceptance detail: `docs/history/M7-SERVICE-STORY.md`.

## M8 — Incident Lens

M8 answers **"what changed around the time this broke?"** with a bounded historical evidence window rather than a monitoring database.

Delivered behavior:

- explicit incident anchor time;
- +/- 15 minute retained-event window;
- all existing event categories preserved when present, including package, configuration, service, container, listener, and system events;
- deterministic chronological ordering and a 100-event cap;
- optional current endpoint Diagnose/TLS re-probe for a supplied `host:port`;
- current endpoint evidence stamped separately so it is never presented as reconstructed historical state;
- explicit language that temporal proximity does not prove causation;
- CLI `hostsleuth incident --at <RFC3339> [--target host:port]`;
- API `/api/incident-lens`;
- Incident Lens embedded in the existing Changes view with **Inspect window** actions on retained events.

M8 adds no causal inference, alerting, time-series storage, historical packet/TLS reconstruction, service control, remediation, or reboot-cause analysis.

### M8 validation

Focused tests cover window boundaries, deterministic equal-time ordering, bounded results, the non-causal wording boundary, and separation of current endpoint evidence from historical event context.

Isolated acceptance on the actual OMV Debian host under `/tmp/hostsleuth-m8-accept` confirmed:

- `go test ./...` passes;
- native build passes;
- relevant JavaScript syntax passes;
- an isolated retained listener-change event is returned by the 15-minute incident window;
- optional `127.0.0.1:22` current endpoint evidence reports reachable separately from the historical anchor;
- an absent endpoint probe emits no endpoint capture timestamp;
- isolated API/UI smoke serves the Incident Lens and returns the bounded window.

The first API smoke exposed a zero-time JSON timestamp when no endpoint was supplied. M8 changed that field to an optional timestamp and added regression coverage before closeout.

Full detail: `docs/history/M8-INCIDENT-LENS.md`.

## Live OMV / recovery boundary

The known-good live HostSleuth remains managed through Arcane on OMV host `192.168.2.181` using:

`mjmalleo/hostsleuth:0.1.0`

Deployment state:

- UI/API: `http://192.168.2.181:8787`
- persistent state: `/srv/docker/volumes/hostsleuth/data`
- Compose path: `/srv/docker/volumes/compose/hostsleuth/compose.yaml`
- trusted-LAN bind: `192.168.2.181:8787`

`xXDasGoGXx/OMV-Docker-Rebuild` intentionally remains pinned to `mjmalleo/hostsleuth:0.1.0` so recovery matches the actual live deployment.

Do not change the live deployment or recovery definition merely to chase release numbers.

## Ordered roadmap — owner approved

The owner explicitly approved continuing in this order without deviation. Future consumer/product research may refine a later milestone, but it does not silently reorder or expand the active milestone.

### 1. M6 — Certificate Story / TLS Detective — COMPLETE

Read-only TLS/certificate troubleshooting is implemented and accepted.

### 2. Publish stable v0.3.0 — COMPLETE

Published and verified from exact accepted source commit `6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`. Live OMV was intentionally not migrated.

### 3. M7 — Service Story — COMPLETE

Read-only systemd/runtime/journal/listener/endpoint/change correlation is implemented and accepted.

### 4. M8 — Incident Lens — COMPLETE

Bounded, non-causal incident-window correlation is implemented and accepted.

### 5. M9 — HostSleuth Workbench — NEXT

Keep M9 deliberately small and read-only. Candidate tools are SHA-256/SHA-512, expected-checksum verification, file-to-file fingerprint comparison, path metadata/hash, deterministic DNS inspection, HTTP HEAD/redirect inspection, PEM inspection, and local-file-vs-served certificate comparison.

Consumer research may refine M9 toward the strongest real troubleshooting workflows, including possible read-only STARTTLS-aware certificate inspection, but do not turn M9 into a miscellaneous tools page, generic file manager, web shell, or admin panel.

### 6. M10 — Reboot Story

Explain boot/shutdown evidence and what failed to come back without inventing reboot cause.

### 7. M11 — Optional Safe Actions

This is the first planned milestone that may cross HostSleuth's read-only boundary and therefore still requires explicit design/security review before implementation. Candidate actions must be narrow, disabled by default, previewed, confirmed, audited, and postcondition-verified. No arbitrary command execution.

### 8. Later — Redacted Evidence Bundle

Only after redaction rules and threat-model work are mature enough.

## Consumer/product research boundary

Current research is intentionally separate from implementation scope. The detailed research artifact is:

`docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`

The strongest differentiated direction currently identified is **correlation plus postcondition verification**, not another UI wrapper around commands users already have.

Examples under later consideration:

- certificate source -> destination -> actually-served fingerprint story;
- STARTTLS-aware certificate inspection for mail protocols;
- expected-endpoint contracts;
- certificate rollout consistency across several local consumers/endpoints;
- bounded certificate deployment recipes with preview, audit, known-service reload, and endpoint re-probe.

Research must not turn HostSleuth into another uptime dashboard, generic ACME manager, generic control panel, web shell, or multi-host orchestration system. Any write action remains behind M11 or another explicit security/design gate.

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

## Repository reading order

When resuming, read:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`
4. `docs/ROADMAP.md`
5. `docs/history/DEVELOPMENT-HISTORY.md`
6. `docs/history/M7-SERVICE-STORY.md`
7. `docs/history/M8-INCIDENT-LENS.md`
8. `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`
9. `docs/design/HOST-STORY-UI.md`
10. `.github/workflows/ci.yml`
11. `.github/workflows/release.yml`
12. `Dockerfile`
13. `compose.yaml`
