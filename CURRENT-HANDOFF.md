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
- M8 — Incident Lens;
- M9 — HostSleuth Workbench.

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

M7, M8, and M9 are newer source capabilities and are **not** claimed to be present in the already-published v0.3.0 artifacts.

## M7 — Service Story

M7 answers why a selected native systemd service is failed/inactive or why an expected endpoint disappeared without turning HostSleuth into a service manager.

Delivered evidence includes bounded systemd runtime/result and sanitized current-boot journal evidence, cgroup/main-PID to listener ownership, deterministic positive-evidence-only port collision, container-port context, existing Diagnose/TLS reuse, and bounded nearby retained changes. Hidden listener ownership remains `unknown` rather than becoming a false collision.

Full implementation and acceptance detail: `docs/history/M7-SERVICE-STORY.md`.

## M8 — Incident Lens

M8 answers **"what changed around the time this broke?"** with a bounded historical evidence window rather than a monitoring database.

It can anchor from an exact time, a retained event, or a completed diagnosis; uses a +/- 15 minute retained-event window; preserves all existing event categories; keeps any current endpoint Diagnose/TLS probe explicitly separate from the historical window; and labels temporal proximity as context rather than proof of causation.

Full implementation and acceptance detail: `docs/history/M8-INCIDENT-LENS.md`.

## M9 — HostSleuth Workbench

M9 answers several small but common troubleshooting questions without exposing a browser shell or generic utility launcher.

Delivered read-only workflows:

- selected-file SHA-256 and SHA-512 plus size, mode/permissions, mtime, UID/GID, and owner/group when resolvable;
- optional expected SHA-256/SHA-512 verification;
- exact file-to-file comparison by SHA-256 fingerprint;
- system-resolver DNS evidence for A/AAAA, distinct CNAME, MX, NS, TXT, and PTR;
- HTTP/HTTPS HEAD status plus a bounded redirect chain and selected response metadata;
- public PEM certificate metadata/fingerprint inspection;
- exact public certificate file fingerprint vs direct-TLS served-certificate comparison;
- CLI `hostsleuth workbench ...`, JSON APIs, and a dedicated Workbench Web UI tab.

M9 is intentionally broader than certificates. Certificate identity is one Workbench workflow alongside file integrity, DNS, and HTTP troubleshooting.

### M9 security boundary

Workbench adds capabilities that should not be exposed to an unauthenticated remote browser merely because the general HostSleuth UI was intentionally bound to a LAN address.

Therefore:

- all `/api/workbench/*` operations accept loopback clients only;
- the local UI and recommended SSH-tunnel workflow continue to work because requests arrive via loopback;
- the CLI remains available locally;
- no arbitrary command or hidden shell hook exists;
- no file editor or file-content response exists;
- HTTP inspection accepts no custom headers, cookies, request credentials, or request body;
- certificate inspection refuses a private-key PEM block encountered before a public certificate;
- file hashing reads only a user-selected readable regular file and returns metadata/fingerprints, not file contents.

### M9 validation

Focused tests cover file hashing/checksum verification, file comparison, deterministic bounded DNS output, HEAD/redirect behavior, public-certificate/private-key boundaries, exact local-file-vs-served certificate matching, and loopback-only Web/API access.

Full branch validation includes:

- `gofmt`;
- `go test ./...`;
- `go vet ./...`;
- native build;
- syntax validation of base, Service Story, Incident Lens, and Workbench JavaScript;
- syntax validation of the exact concatenated JavaScript served to consumers.

Isolated real-OMV acceptance under `/tmp/hostsleuth-m9-accept` confirmed:

- selected-file SHA-256 verification against the host's independent `sha256sum` result;
- SHA-512 output;
- exact file-to-file fingerprint match;
- real DNS A/AAAA evidence;
- real HTTPS HEAD/status evidence;
- public certificate inspection;
- exact local-public-cert vs isolated served-certificate fingerprint match using a temporary self-signed TLS listener;
- isolated Workbench API/UI smoke on loopback.

Validation found two implementation/test issues before closeout: a host umask made one test's assumed permission mode incorrect, and the first API wrapper used Go multiple-return values incorrectly. Both were fixed before acceptance. CI was also strengthened so future UI fragments cannot bypass syntax validation merely because they are concatenated at runtime.

Full detail: `docs/history/M9-WORKBENCH.md`.

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

The owner explicitly approved continuing in this order without deviation. Consumer/product research may refine later work, but it does not silently reorder or expand the active milestone.

### 1. M6 — Certificate Story / TLS Detective — COMPLETE

Read-only TLS/certificate troubleshooting is implemented and accepted.

### 2. Publish stable v0.3.0 — COMPLETE

Published and verified from exact accepted source commit `6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`. Live OMV was intentionally not migrated.

### 3. M7 — Service Story — COMPLETE

Read-only systemd/runtime/journal/listener/endpoint/change correlation is implemented and accepted.

### 4. M8 — Incident Lens — COMPLETE

Bounded, non-causal incident-window correlation is implemented and accepted.

### 5. M9 — HostSleuth Workbench — COMPLETE

Bounded file-integrity, DNS, HTTP, and public-certificate troubleshooting tools are implemented and accepted. Workbench Web/API operations are loopback-only; no browser shell or generic utility launcher was added.

### 6. M10 — Reboot Story — NEXT / ACTIVE

Explain boot/shutdown evidence and what failed to come back after a reboot without inventing reboot cause. Reuse the event/Incident Lens model, keep evidence bounded, and remain read-only.

### 7. M11 — Optional Safe Actions

This is the first planned milestone that may cross HostSleuth's read-only boundary and therefore requires an explicit design/security review before implementation. Candidate actions must be narrow, disabled by default, previewed, confirmed, audited, and postcondition-verified. No arbitrary command execution.

Certificate lifecycle is only one candidate family; consumer research should compare multiple real workflows before any safe-action set is approved.

### 8. Later — Redacted Evidence Bundle

Only after redaction rules and threat-model work are mature enough.

## Consumer/product research boundary

Current research is intentionally separate from implementation scope. The detailed research artifact is:

`docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`

The desired product direction is broader than certificates: find recurring troubleshooting workflows where users currently assemble several commands, admin screens, logs, and websites, then use HostSleuth's deterministic correlation/verification model when it can genuinely make the workflow better.

Promising areas include:

- expected-endpoint contracts;
- DNS resolver/delegation/split-view discrepancies;
- HTTP redirect/reverse-proxy/upstream mismatches;
- file permissions/ownership/deployment-path problems;
- STARTTLS-aware mail/service inspection;
- certificate source/destination/served verification and rollout consistency;
- later tightly bounded safe-action recipes with observable postcondition verification.

Research must not turn HostSleuth into an uptime dashboard, generic certificate/ACME manager, generic control panel, web shell, or multi-host orchestration system.

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
8. `docs/history/M9-WORKBENCH.md`
9. `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`
10. `docs/design/HOST-STORY-UI.md`
11. `.github/workflows/ci.yml`
12. `.github/workflows/release.yml`
13. `Dockerfile`
14. `compose.yaml`
