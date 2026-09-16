# HostSleuth — Current Handoff

Last updated: 2026-09-15

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
- **Post-M3 collective product review:** complete on `planning/product-review`; pending documentation merge.

## Product-direction rule

HostSleuth stays easy to deploy, easy to understand, and intentionally small. New capability is not automatically good capability.

The post-M3 review split deferred ideas into three groups in `TO-DO.md`:

1. capabilities that still fit HostSleuth's two core jobs;
2. ideas that remain demand-gated until real users prove the need;
3. architecture-heavy ideas that are not planned unless the product direction changes.

This replaces the old undifferentiated backlog. Do not turn every interesting idea into a standing task.

## Supported deployment paths

### Native Linux — recommended

Native Linux provides the fullest visibility into systemd, host filesystems, network/listener state, and Docker inventory while keeping the Web UI loopback-only by default.

### Docker Compose — convenient, reduced visibility

The repository includes one supported `Dockerfile` and one recommended `compose.yaml`. The Docker deployment uses host network/PID/UTS namespaces, a persistent `hostsleuth-data` volume, a narrow read-only host OS-release mount, and Docker socket access for container inventory. It drops Linux capabilities, enables `no-new-privileges`, uses a read-only root filesystem, and does not use unrestricted `privileged: true`.

Docker-mode snapshots explicitly identify reduced visibility. Host filesystem and systemd inventory are intentionally unavailable rather than replaced with misleading container-local data. Firewall evidence may also be unavailable without elevated network-administration privileges.

The Docker socket remains a powerful host capability even when bind-mounted read-only; this risk is documented in `README.md` and `SECURITY.md`.

## What works now

### Change recorder

Native HostSleuth captures host/OS/kernel state, filesystems, interfaces/routes, listeners, systemd services, and Docker inventory. Docker mode retains truthful host identity/network/listener/Docker evidence while clearly marking native-only evidence unavailable.

### Deterministic diagnosis

`diagnose host:port` uses deterministic precedence across DNS, route, TCP, local listeners, Docker publication/bind/network evidence, bounded nftables evidence, and bounded failed-systemd evidence where available. Unavailable evidence remains `unknown` rather than becoming a false pass/fail.

### Web UI

The Web UI provides Overview / Diagnose / Changes / Host views, readable diagnosis evidence, first-run guidance, version/build information, newest-first changes, and explicit Docker reduced-visibility labels.

## Current task

**M3 acceptance polish before another capability milestone.**

1. Validate the merged M3 Web UI on a real supported Linux deployment.
2. Smoke-test the supported Docker Compose path on a real Linux Docker host without replacing the known-good native deployment.
3. Capture one real README Web UI screenshot from an accepted deployed build.
4. Fix only real acceptance problems found by those checks.

Do not start M4 until this acceptance pass is complete.

## Next capability after acceptance

**M4 — package-change timeline.**

M4 is intentionally limited to read-only package install/update/remove history, starting with bounded Debian/Ubuntu `apt`/`dpkg` logs and feeding the existing event timeline. It must not become package management, alerts, repositories, update automation, or a new settings surface.

After M4, the only currently product-aligned candidates are configuration fingerprinting, bounded TLS/certificate diagnosis, and stronger redaction/threat-model work. Even those are one-at-a-time decisions, not promises.

## Not current product direction

Multi-host controller/agent architecture, a general dependency graph, plugin ecosystem, AI explanation layer, and repair/remediation are not standing tasks. Reconsider them only if HostSleuth's product direction materially changes.

## Branch hygiene

`main` is the authoritative stable development state.

- `planning/product-review` contains only the post-M3 roadmap/handoff classification and should be merged before acceptance work begins.
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
