# HostSleuth — Current Handoff

Last updated: 2026-09-16

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform or browser-based server administration suite.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Completed source milestones:

- M0 — repository foundation;
- M1 — deployable single-host MVP;
- M2 — deeper deterministic diagnosis;
- M3 — Product Experience;
- M3.4 — Public Container Distribution;
- M4 — package-change timeline;
- M5 — configuration fingerprinting;
- M6 — Certificate Story / TLS Detective;
- M7 — Service Story;
- M8 — Incident Lens;
- M9 — HostSleuth Workbench;
- M10 — Reboot Story.

M10 was developed on `m10-reboot-story` in PR #34. Its final implementation and acceptance record is:

`docs/history/M10-REBOOT-STORY.md`

The earlier `docs/history/M10-REBOOT-STORY-WIP.md` is historical only.

No M11 implementation has started.

## M10 — Reboot Story — complete in source

M10 answers:

**What happened around this reboot, and what failed to come back afterward?**

Delivered behavior includes:

- snapshot schema version 4;
- Linux kernel boot ID capture when available;
- exact boot-start capture from `/proc/stat` `btime` when available;
- deterministic reboot detection only when a known boot ID changes to another known boot ID;
- no reboot inference from human-readable uptime;
- regression coverage proving an older snapshot with no boot ID does not emit a false reboot during schema upgrade;
- bounded previous-boot and current-boot journal evidence;
- explicit `unknown` behavior when journal history is unavailable or inaccessible;
- orderly-vs-abnormal shutdown classification only when direct bounded evidence supports it;
- current failed-service evidence;
- retained post-boot service/listener/container problem correlation;
- recovery issues only when retained post-boot evidence and current snapshot state agree the problem remains;
- package/kernel/system/configuration context around boot without causal overclaiming;
- CLI, JSON API, runtime Web asset, and Reboot-tab integration;
- no reboot/shutdown/restart/reload/service-control/package/firewall/file-write/arbitrary-command action.

M10 remains read-only and explicitly separates reboot detection from reboot explanation. Temporal proximity is context, not proof of cause.

### M10 validation

Validation included:

- focused M10 tests;
- full `go test ./...`;
- `go vet ./...`;
- `gofmt` cleanliness;
- individual JavaScript syntax checks;
- syntax validation of the exact concatenated served script;
- native build;
- Docker runtime smoke including Reboot Story API/UI wiring;
- linux/amd64 image build;
- linux/arm64 image build;
- isolated real-OMV snapshot and journal acceptance.

Real OMV acceptance confirmed schema 4, live boot ID, exact boot start, native service/listener evidence, and the real journal permission boundary. The unprivileged acceptance account receives `No journal files were opened due to insufficient permissions.` for previous/current boot journal reads. M10 now recognizes that diagnostic as unavailable evidence and reports `unknown`; no privilege expansion was added.

Full details: `docs/history/M10-REBOOT-STORY.md`.

## Published release boundary

Stable public release remains:

`v0.3.0`

Release source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Published Docker tags remain:

- `mjmalleo/hostsleuth:0.3.0`
- `mjmalleo/hostsleuth:latest`

Published multi-platform OCI index:

`sha256:127b388fbf794841b22d06b281fe89dc1500188fdb4215eec023b392aa98c05d`

Platforms:

- `linux/amd64`
- `linux/arm64`

M7, M8, M9, and M10 are newer source capabilities and are **not** claimed to be included in v0.3.0.

Do not publish a new release merely for version-number alignment.

## Live OMV / recovery boundary

The known-good live HostSleuth remains managed through Arcane on OMV host `192.168.2.181` using:

`mjmalleo/hostsleuth:0.1.0`

Deployment state:

- UI/API: `http://192.168.2.181:8787`
- persistent state: `/srv/docker/volumes/hostsleuth/data`
- Compose path: `/srv/docker/volumes/compose/hostsleuth/compose.yaml`
- trusted-LAN bind: `192.168.2.181:8787`

`xXDasGoGXx/OMV-Docker-Rebuild` intentionally remains pinned to `mjmalleo/hostsleuth:0.1.0` so recovery matches the actual live deployment.

Do not change the live deployment, Compose, persistent state, Docker tag, or recovery pin merely to chase source/release numbers.

## Next milestone boundary

### M11 — Optional Safe Actions — NOT STARTED

M11 is the first planned milestone that may cross HostSleuth's read-only boundary. Before implementing it, perform an explicit security/design review.

Any future action must be narrow, disabled by default, previewed, explicitly confirmed, audited, postcondition-verified, and free of arbitrary shell/command fields.

Do not start M11 automatically after M10. It requires explicit owner direction.

## Consumer/product research boundary

Research remains separate from implementation scope. The detailed artifact is:

`docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`

Certificates/Certbot are one opportunity among many, not the product direction. Promising areas include endpoint-path reasoning, expected-state contracts, DNS resolver/delegation/split-view discrepancies, HTTP/reverse-proxy/upstream problems, port/listener ownership, file permissions/ownership/deployment paths, container disappearance/dependencies, mail/STARTTLS inspection, boot/recovery workflows, and certificate delivery/rollout verification.

A future feature should generally provide correlation, verification, and boundedness without turning HostSleuth into a generic administration platform.

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

## Resume order for future work

Before consequential writes, re-check current `main`, open PRs, and the active milestone branch. Then read:

1. `CURRENT-HANDOFF.md`
2. `TO-DO.md`
3. `docs/ROADMAP.md`
4. `docs/history/DEVELOPMENT-HISTORY.md`
5. the most recent milestone history document
6. `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`
7. `.github/workflows/ci.yml`
8. `.github/workflows/release.yml`
9. `Dockerfile`
10. `compose.yaml`

For M10 implementation/acceptance history specifically, use `docs/history/M10-REBOOT-STORY.md`.
