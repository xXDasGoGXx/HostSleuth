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
- **M3 — Product Experience:** complete.
  - **M3.1 Web UI:** merged in PR #11.
  - **M3.2 Usability:** merged in PR #13 (`d250019902d676f67d44b74cc122db3f40ad7e67`).
  - **M3.3 Docker release packaging:** merged in PR #14 (`ccaad3ef1df16e65887d1f19da44a118a6d0bb2a`).

The active decision list lives in `TO-DO.md`.

## Product-direction rule

HostSleuth should remain easy to deploy, easy to understand, and intentionally small. New capability is not automatically good capability.

Authentication, proxy/TLS awareness, SQLite, AI explanation, multi-host support, remediation, and other larger ideas remain visible in the **collective review** section of `TO-DO.md`. They are deferred decisions, not permanent bans and not promised features.

Promote one only when it solves a common real user problem without making installation, operation, or the UI meaningfully harder.

## Supported deployment paths

### Native Linux — recommended

Native Linux provides the fullest HostSleuth visibility into systemd, host filesystems, network/listener state, and Docker inventory while keeping the Web UI loopback-only by default.

The diagnostic core and native systemd deployment have been validated on Debian 13. M3.1/M3.2 changed presentation and usability without changing deterministic diagnosis semantics.

### Docker Compose — convenient, reduced visibility

The repository now contains one supported `Dockerfile` and one recommended `compose.yaml`.

The Docker deployment:

- persists state in the `hostsleuth-data` volume;
- uses host network/PID/UTS namespaces for truthful host network/listener/hostname evidence;
- reads host `/etc/os-release` through a narrow read-only mount;
- uses the Docker socket for container inventory;
- drops all Linux capabilities;
- enables `no-new-privileges`;
- uses a read-only container root filesystem;
- does not use unrestricted `privileged: true`.

Docker-mode snapshots identify the deployment mode explicitly. Host filesystem and systemd inventory are intentionally unavailable rather than replaced with misleading container-local data. Systemd diagnostic evidence remains `unknown` in Docker mode. Firewall evidence may also be unavailable without elevated network-administration privileges.

PR #14 validated:

- normal Go formatting, vet, tests, and build;
- embedded Web UI JavaScript syntax;
- Compose configuration;
- Linux amd64 Docker image build;
- Linux arm64 Docker image build.

No container image was published and the known-good native service was not replaced.

## Docker socket security boundary

The Compose deployment mounts `/var/run/docker.sock` for Docker inventory. This is a powerful host capability even when the socket path is mounted read-only; read-only bind mounting does not make Docker API access inherently read-only.

HostSleuth uses Docker only for read-only inventory commands, but a compromised process with socket access could potentially control the Docker daemon. This risk is documented in both `README.md` and `SECURITY.md` rather than hidden behind deployment convenience.

## What works now

### Change recorder

Native HostSleuth captures host/OS/kernel state, filesystems, interfaces/routes, listeners, systemd services, and Docker inventory. Docker mode retains truthful host identity/network/listener/Docker evidence while clearly marking native-only evidence unavailable.

### Deterministic diagnosis

`diagnose host:port` continues to use the existing deterministic precedence rules. Optional evidence that is unavailable stays `unknown`; it is not converted into a false failure or false pass.

### Web UI and usability

The Web UI provides Overview / Diagnose / Changes / Host views, readable diagnosis evidence, first-run guidance, version/build information, newest-first changes, and explicit Docker reduced-visibility labels.

## Current limitations

- dashboard is loopback-only and has no authentication;
- storage is JSON/JSONL;
- reverse-proxy and TLS-specific diagnosis are not implemented;
- configuration/package change tracking is not implemented;
- HostSleuth is single-host first;
- no automatic remediation;
- Docker deployment has intentionally reduced systemd/filesystem/firewall visibility compared with native deployment.

These limitations remain collective product decisions; they do not automatically become the next work.

## Branch hygiene

`main` is the authoritative development state.

- `m3/docker-release`, `m3/usability`, and `m3/product-experience` are historical after merged PRs #14, #13, and #11.
- Old M1/M2 feature branches are historical leftovers after merged work.
- `m3/npm-proxy-awareness` and `m3/tls-diagnostics` contain historical experimental work and are not current product state.

## Next task

**Collectively review the deferred product decisions before choosing another capability milestone.**

Do not start one simply because it exists in the backlog. Prefer the smallest change that improves a real, common HostSleuth workflow.

A real deployed Web UI screenshot remains useful documentation polish when an accepted deployed M3 build is available to capture.

## Repository source of truth

Repository: `xXDasGoGXx/HostSleuth`

Read in this order when resuming:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`

Use `docs/history/DEVELOPMENT-HISTORY.md` only when historical implementation/validation details are needed.

Public release promotion remains owner-controlled and must not happen without explicit approval.
