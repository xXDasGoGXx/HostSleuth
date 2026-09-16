# HostSleuth — Current Handoff

Last updated: 2026-09-16

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

The product definition and user-facing overview live in `README.md`. The active roadmap lives in `TO-DO.md`.

## Milestone state

- **M0 — Repository foundation:** complete.
- **M1 — Single-host deployable MVP:** complete.
- **M2 — Deeper deterministic diagnosis:** complete.
- **M3 — Product Experience:** complete.
  - M3.1 Web UI — PR #11.
  - M3.2 Usability — PR #13.
  - M3.3 Docker release packaging — PR #14.
- **Post-M3 collective product review:** complete and merged in PR #16.
- **M3 acceptance:** complete.
  - native side-by-side acceptance passed on a real Debian 13 host without replacing the known-good installed service;
  - supported Docker Compose runtime smoke passed on an authorized ephemeral Linux Docker host;
  - a real public-safe Diagnose-view screenshot is committed in `docs/images/hostsleuth-diagnose.png` and displayed in the README.

## Product-direction rule

HostSleuth stays easy to deploy, easy to understand, and intentionally small. New capability is not automatically good capability.

The post-M3 review split deferred ideas into three groups in `TO-DO.md`:

1. capabilities that still fit HostSleuth's two core jobs;
2. ideas that remain demand-gated until real users prove the need;
3. architecture-heavy ideas that are not planned unless the product direction changes.

Do not turn every interesting idea into a standing task.

## Supported deployment paths

### Native Linux — recommended

Native Linux provides the fullest visibility into systemd, host filesystems, network/listener state, and Docker inventory while keeping the Web UI loopback-only by default.

### Docker Compose — convenient, reduced visibility

The repository includes one supported `Dockerfile` and one recommended `compose.yaml`. The Docker deployment uses host network/PID/UTS namespaces, a persistent `hostsleuth-data` volume, a narrow read-only host OS-release mount, and Docker socket access for container inventory. It drops Linux capabilities, enables `no-new-privileges`, uses a read-only root filesystem, and does not use unrestricted `privileged: true`.

Docker-mode snapshots explicitly identify reduced visibility. Host filesystem and systemd inventory are intentionally unavailable rather than replaced with misleading container-local data. Firewall evidence may also be unavailable without elevated network-administration privileges.

The Docker socket remains a powerful host capability even when bind-mounted read-only; this risk is documented in `README.md` and `SECURITY.md`.

The public HostSleuth Compose remains portable. Deployment-specific OMV/Arcane layout belongs in the separate `xXDasGoGXx/OMV-Docker-Rebuild` source of truth rather than being baked into this repository.

## M3 acceptance result

### Native

The merged M3 build was validated side-by-side on a real Debian 13 host without stopping, replacing, or modifying the existing known-good HostSleuth service.

Acceptance confirmed:

- current source tests and build succeeded on the real host;
- the temporary build used an alternate loopback port and separate temporary state;
- `/api/about`, `/api/snapshot`, and the modern Overview / Diagnose / Changes / Host interface worked;
- diagnosis of the temporary Web UI endpoint returned reachable / high-confidence with DNS, route, and TCP checks passing;
- no native M3 product blocker was found.

The temporary acceptance process was unprivileged, so Docker inventory was unavailable in that temporary snapshot. That was expected for the side-by-side method and does not describe the normal root-managed native installation.

### Docker Compose

The supported `compose.yaml` was then exercised on an authorized ephemeral Linux Docker host in CI rather than bypassing the managed OMV host's Docker-command policy.

The runtime smoke test confirmed:

- the actual supported Compose deployment builds and starts;
- the HostSleuth API becomes ready;
- `/api/snapshot` identifies Docker mode truthfully;
- the Web UI responds;
- HostSleuth diagnoses its own running endpoint as reachable with high confidence;
- the Compose stack and test volume are torn down afterward.

The existing Linux amd64 and arm64 image-build checks also remain green.

## What works now

### Change recorder

Native HostSleuth captures host/OS/kernel state, filesystems, interfaces/routes, listeners, systemd services, and Docker inventory. Docker mode retains truthful host identity/network/listener/Docker evidence while clearly marking native-only evidence unavailable.

### Deterministic diagnosis

`diagnose host:port` uses deterministic precedence across DNS, route, TCP, local listeners, Docker publication/bind/network evidence, bounded nftables evidence, and bounded failed-systemd evidence where available. Unavailable evidence remains `unknown` rather than becoming a false pass/fail.

### Web UI

The Web UI provides Overview / Diagnose / Changes / Host views, readable diagnosis evidence, first-run guidance, version/build information, newest-first changes, and explicit Docker reduced-visibility labels. The interface has been exercised on a real supported Linux host and through the supported Docker Compose deployment.

## Current task

**M3 acceptance is complete. The next capability is M4 — package-change timeline.**

Do not broaden M4. Its only job is to strengthen **“what changed?”** by adding bounded, read-only package install/update/remove history to the existing event timeline.

Start with Debian/Ubuntu `apt`/`dpkg` logs. Unsupported systems should remain truthful and quiet. Do not add package management, update actions, alerts, repositories, or a new package settings surface.

M4 must include focused tests and real-host validation before merge.

## After M4

The only currently product-aligned candidates are configuration fingerprinting, bounded TLS/certificate diagnosis, and stronger redaction/threat-model work. Even those are one-at-a-time decisions, not promises.

## Not current product direction

Multi-host controller/agent architecture, a general dependency graph, plugin ecosystem, AI explanation layer, and repair/remediation are not standing tasks. Reconsider them only if HostSleuth's product direction materially changes.

## Branch hygiene

`main` is the authoritative stable development state.

- `acceptance/m3-finalize` contains only the final M3 acceptance/runtime-smoke/documentation closeout and becomes historical after PR #18 merges.
- `acceptance/m3-real-host` is historical native-acceptance work.
- `planning/product-review` is historical after merged PR #16.
- `m3/docker-release`, `m3/usability`, and `m3/product-experience` are historical after merged PRs.
- `m3/npm-proxy-awareness` and `m3/tls-diagnostics` remain historical experiments, not current product state.

## Repository source of truth

Repository: `xXDasGoGXx/HostSleuth`

Read in this order when resuming:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`

Use `docs/history/DEVELOPMENT-HISTORY.md` only for historical implementation/validation details.

Public release promotion, container publication, and changes to a known-good live deployment remain owner-controlled and require explicit approval.
