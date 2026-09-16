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
- **Post-M3 collective product review:** complete and merged in PR #16.
- **Native M3 acceptance:** complete on a real Debian 13 host using a side-by-side temporary build; the known-good installed service remained untouched.

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

## Native M3 acceptance result

The merged `main` revision was validated side-by-side on a real Debian 13 host without stopping, replacing, or modifying the existing known-good HostSleuth service.

Acceptance procedure and result:

- confirmed the existing managed service remained enabled/running and retained its previously approved binary/unit state;
- confirmed that installed service was still the older validated pre-M3 build, so the acceptance test did not mistake it for the new UI;
- cloned current `main` into a temporary directory;
- downloaded a temporary Go 1.24 toolchain, verified it against the publisher-provided SHA-256, and installed nothing system-wide;
- ran the full Go test suite successfully on the real host;
- built the current `main` revision successfully;
- started the M3 build on an alternate loopback port with a separate temporary state directory;
- verified `/api/about`, `/api/snapshot`, and the modern Overview / Diagnose / Changes / Host interface;
- diagnosed the temporary Web UI endpoint and received a reachable / high-confidence result with DNS, route, and TCP checks passing;
- found no native M3 product blocker.

The temporary acceptance process ran unprivileged, so Docker inventory was unavailable in that temporary snapshot. That is expected for this side-by-side acceptance method and does not describe the normal root-managed native installation.

A real Diagnose-view screenshot was captured from this accepted build at a public-safe viewport. It contains no real hostname, LAN address, mount path, or activity timeline. Committing that image to the README remains a separate repository step until the binary asset is transferred cleanly.

## Remaining acceptance limit

The real-host Docker Compose smoke test has **not** been performed through the current approved management gateway because its policy blocks Docker commands. This is an environment/policy limitation, not a HostSleuth test failure.

Do not bypass that policy. Complete the Docker smoke test only on an authorized Linux Docker environment where Docker execution is allowed, or explicitly decide that existing Compose validation plus amd64/arm64 image-build CI is sufficient for this milestone.

## What works now

### Change recorder

Native HostSleuth captures host/OS/kernel state, filesystems, interfaces/routes, listeners, systemd services, and Docker inventory. Docker mode retains truthful host identity/network/listener/Docker evidence while clearly marking native-only evidence unavailable.

### Deterministic diagnosis

`diagnose host:port` uses deterministic precedence across DNS, route, TCP, local listeners, Docker publication/bind/network evidence, bounded nftables evidence, and bounded failed-systemd evidence where available. Unavailable evidence remains `unknown` rather than becoming a false pass/fail.

### Web UI

The Web UI provides Overview / Diagnose / Changes / Host views, readable diagnosis evidence, first-run guidance, version/build information, newest-first changes, and explicit Docker reduced-visibility labels. The merged interface has now been exercised on a real supported Linux host.

## Current task

**Finish the two remaining M3 acceptance-polish items before M4:**

1. commit the sanitized real Web UI screenshot to the public README;
2. resolve the Docker smoke-test requirement without bypassing management policy.

Do not start M4 until those are completed or deliberately accepted as externally limited.

## Next capability after acceptance

**M4 — package-change timeline.**

M4 is intentionally limited to read-only package install/update/remove history, starting with bounded Debian/Ubuntu `apt`/`dpkg` logs and feeding the existing event timeline. It must not become package management, alerts, repositories, update automation, or a new settings surface.

After M4, the only currently product-aligned candidates are configuration fingerprinting, bounded TLS/certificate diagnosis, and stronger redaction/threat-model work. Even those are one-at-a-time decisions, not promises.

## Not current product direction

Multi-host controller/agent architecture, a general dependency graph, plugin ecosystem, AI explanation layer, and repair/remediation are not standing tasks. Reconsider them only if HostSleuth's product direction materially changes.

## Branch hygiene

`main` is the authoritative stable development state.

- `acceptance/m3-real-host` contains only M3 acceptance documentation/polish and must not modify or deploy the known-good service.
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
