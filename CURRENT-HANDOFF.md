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
- M7 — Service Story.

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

M7 is newer source capability and is **not** claimed to be present in the already-published v0.3.0 artifacts.

## M7 — Service Story

M7 answers why a selected native systemd service is failed/inactive or why an expected endpoint disappeared without turning HostSleuth into a service manager.

Delivered evidence:

- bounded `systemctl show` properties for load/active/sub/unit-file/result/main-PID/cgroup/exit status;
- bounded current-boot unit journal, passed through existing secret sanitization;
- service cgroup PID membership;
- listener ownership when local permissions expose listener PIDs;
- deterministic expected-port collision only when a competing PID is positively visible;
- expected-port presence with `unknown` ownership when listener PID metadata is hidden;
- related container host-port publication context;
- reuse of the existing endpoint Diagnose engine, including TLS/certificate evidence;
- direct retained service/listener history plus package/configuration/container context within +/- 15 minutes of the newest direct event;
- CLI `hostsleuth service ...`;
- API `/api/service-story`;
- Service Story UI embedded inside the existing Diagnose view, preserving the accepted four top-level tabs.

M7 remains read-only. It adds no start/stop/restart/reload/enable/disable buttons and no arbitrary command field.

### M7 validation

Focused and full branch validation passed:

- formatting;
- `go test ./...`;
- `go vet ./...`;
- native build;
- existing Web JavaScript syntax;
- M7 JavaScript syntax;
- concatenated served JavaScript syntax.

Real-host acceptance ran on the actual OMV Debian host using an isolated `/tmp` state directory and branch binary.

The first SSH acceptance found a correctness defect: the host exposed TCP/22 but hid its listener PID from the unprivileged snapshot, and the first implementation interpreted missing ownership as a competing process. That was fixed so collision claims require positive competing PID evidence, and a regression test was added.

Corrected acceptance for `ssh.service` + `127.0.0.1:22` confirmed:

- systemd runtime `loaded / active / running`;
- cgroup/main-PID evidence available;
- journal evidence unavailable because of local permissions, truthfully `unknown`;
- TCP/22 present with hidden listener owner, truthfully `unknown` ownership;
- endpoint reachable;
- high-confidence conclusion that the service is active and expected endpoint reachable.

An isolated Web/API smoke also passed. No live HostSleuth process, Compose file, persistent production state, or recovery definition was changed.

Full implementation/acceptance detail: `docs/history/M7-SERVICE-STORY.md`.

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

### 4. M8 — Incident Lens — NEXT

Anchor a diagnosis/event/time to a bounded evidence window, initially +/- 15 minutes, and show nearby package/configuration/service/container/listener/TLS context. Temporal proximity is context, not proven causation. Reuse the event model rather than building a time-series monitoring database.

### 5. M9 — HostSleuth Workbench

Small practical troubleshooting tools only: SHA-256/SHA-512, expected checksum verification, file fingerprint comparison, path metadata/hash, DNS inspection, HTTP HEAD/redirects, PEM inspection, and local-file-vs-served-certificate comparison. No web shell or arbitrary command box.

### 6. M10 — Reboot Story

Explain boot/shutdown evidence and what failed to come back using kernel/package/config/service/listener/container context without inventing a reboot cause.

### 7. M11 — Optional Safe Actions

This is the first planned milestone that may cross HostSleuth's read-only boundary and therefore still requires explicit design/security review before implementation. Candidate actions must be narrow, disabled by default, previewed, confirmed, and audited. No arbitrary command execution.

### 8. Later — Redacted Evidence Bundle

Only after redaction rules and threat-model work are mature enough.

## Consumer/product research boundary

Current research is intentionally separate from implementation scope. It may identify differentiated later opportunities such as certificate deployment verification, STARTTLS-aware certificate inspection, or tightly bounded deployment recipes, but those ideas must be assigned to an appropriate future milestone before implementation.

Research must not turn HostSleuth into:

- another uptime/metrics dashboard;
- a generic ACME/certificate manager;
- a generic control panel;
- a web shell;
- a multi-host orchestration system.

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
7. `docs/design/HOST-STORY-UI.md`
8. `.github/workflows/ci.yml`
9. `.github/workflows/release.yml`
10. `Dockerfile`
11. `compose.yaml`
