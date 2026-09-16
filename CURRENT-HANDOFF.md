# HostSleuth — Current Handoff

Last updated: 2026-09-15

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

The product definition and normal user-facing overview live in `README.md`.

## Current milestone state

- **M0 — Repository foundation:** complete.
- **M1 — Single-host deployable MVP:** complete.
- **M2 — Deeper deterministic diagnosis:** complete.
- **M3 — Product Experience:** active.
  - **M3.1 Web UI:** complete and merged in PR #11.
  - **M3.2 Usability:** complete and merged in PR #13 (`d250019902d676f67d44b74cc122db3f40ad7e67`).
  - **M3.3 Docker release:** active on `m3/docker-release`.

The exact checklist lives in `TO-DO.md` and is updated as work progresses.

## Product-direction rule

HostSleuth should remain easy to deploy, easy to understand, and intentionally small. New capability is not automatically good capability.

Ideas such as authentication, proxy/TLS awareness, SQLite, AI explanation, multi-host support, remediation, and other larger additions remain visible in the **collective review** section of `TO-DO.md`. They are deferred decisions, not permanent bans and not promised features.

Promote one only when it solves a common real user problem without making installation, operation, or the UI meaningfully harder.

## Native deployment

Native Linux remains the recommended HostSleuth deployment because it provides the fullest visibility into systemd and host filesystems while keeping the Web UI loopback-only by default.

The M2 diagnostic core and native systemd deployment have been validated on Debian 13. M3.1/M3.2 changed presentation and usability without changing deterministic diagnosis semantics.

## M3.3 Docker design

The Docker deployment is deliberately one supported Compose shape rather than a matrix of modes and switches.

Implemented design:

- multi-stage Linux image build;
- one `compose.yaml`;
- persistent `hostsleuth-data` volume;
- host network/PID/UTS namespaces so network/listener/hostname evidence describes the host rather than an isolated container;
- host `/etc/os-release` mounted read-only for truthful OS identity;
- Docker socket mounted for container inventory;
- all Linux capabilities dropped;
- `no-new-privileges` enabled;
- read-only container root filesystem;
- no unrestricted `privileged: true`;
- same embedded HostSleuth Web UI;
- CI definitions for Linux amd64 and arm64 image builds.

Docker isolation prevents safe, reliable access to some native evidence. Docker-mode snapshots therefore mark the deployment mode explicitly and intentionally omit host filesystem and systemd inventory. Diagnosis reports systemd evidence as unavailable instead of manufacturing a clean result. Firewall evidence may also remain unavailable without elevated network-administration privileges.

This reduced visibility is intentional. Do not add broad host-root mounts, systemd control sockets, `CAP_NET_ADMIN`, or privileged mode merely to make Docker look identical to native HostSleuth.

## Docker socket security boundary

The Compose deployment mounts `/var/run/docker.sock` for Docker inventory. This is a powerful host capability even when the socket path is mounted read-only; read-only bind mounting does not make Docker API access inherently read-only.

HostSleuth uses Docker only for read-only inventory commands, but a compromised process with socket access could potentially control the Docker daemon. This risk is documented in both `README.md` and `SECURITY.md` rather than hidden behind deployment convenience.

## What works now

### Change recorder

Native HostSleuth captures host/OS/kernel state, filesystems, interfaces/routes, listeners, systemd services, and Docker inventory. Docker mode retains truthful host identity/network/listener/Docker evidence while clearly marking native-only evidence unavailable.

### Deterministic diagnosis

`diagnose host:port` continues to use the existing deterministic precedence rules. Optional evidence that is unavailable in Docker mode remains `unknown`; it is not converted into a false failure or false pass.

### Web UI and usability

The Web UI provides Overview / Diagnose / Changes / Host views, readable diagnosis evidence, first-run guidance, version/build information, newest-first changes, and explicit Docker reduced-visibility labels.

## Active branch and scope

Active branch: `m3/docker-release`

Current scope is **M3.3 only**:

1. validate the Dockerfile and Compose definition;
2. validate amd64 and arm64 image builds;
3. validate normal Go/Web CI with Docker-mode tests;
4. fix only Docker-release blockers discovered by that validation;
5. merge after all checks are green.

Do not publish a container image. Do not deploy this branch over the known-good native service. Both actions remain owner-controlled.

## Branch hygiene

`main` remains the authoritative stable development state.

- `m3/docker-release` is the only active feature branch for the current task.
- `m3/usability` is historical after merged PR #13.
- `m3/product-experience` is historical after merged PR #11.
- Old M1/M2 feature branches are historical leftovers after merged work.
- `m3/npm-proxy-awareness` and `m3/tls-diagnostics` contain historical experimental work and are not current product state.

## Next task

**Open M3.3 for review and let CI validate native tests plus both Docker image architectures.**

If validation is green, merge M3.3. After the M3 sequence, review deferred product decisions collectively before starting another capability milestone.

## Repository source of truth

Repository: `xXDasGoGXx/HostSleuth`

Read in this order when resuming:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`

Use `docs/history/DEVELOPMENT-HISTORY.md` only when historical implementation/validation details are needed.

Public release promotion remains owner-controlled and must not happen without explicit approval.
