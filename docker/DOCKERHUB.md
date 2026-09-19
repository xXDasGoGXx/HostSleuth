# HostSleuth

> **Know what changed. Understand why it broke.**

HostSleuth is a local-first Linux flight recorder and deterministic root-cause diagnostics tool for administrators who want evidence instead of guesswork.

It is built around two practical questions:

1. **What changed on this Linux host?**
2. **Why can I not reach this host, service, port, certificate, proxy path, or dependency?**

No cloud account. No external database. No API key. No AI model required.

## Why HostSleuth

Most Linux troubleshooting means manually correlating output from networking, systemd, Docker, TLS, DNS, package history, configuration files, and logs. HostSleuth collects bounded evidence from those areas and presents one deterministic story.

Where evidence is unavailable, HostSleuth says **unknown** rather than inventing an answer.

## Quick start

Pull the newest stable image:

```bash
docker pull mjmalleo/hostsleuth:latest
```

Or use the supported Compose definition from the source repository:

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
docker compose pull
docker compose up -d
```

HostSleuth binds its Web UI to loopback by default:

```text
http://127.0.0.1:8787
```

For a remote Linux server, use an SSH tunnel:

```bash
ssh -L 8787:127.0.0.1:8787 user@your-server
```

Then open `http://127.0.0.1:8787` on your computer.

## What it can investigate

HostSleuth can combine evidence for:

- DNS resolution and resolver disagreement;
- routes and TCP reachability;
- local listeners and port ownership;
- Docker containers, published ports, bind addresses, and networks;
- TLS handshakes, protocol/cipher, certificate identity, hostname, trust, and expiry;
- expected-vs-observed endpoint contracts;
- reverse-proxy and upstream request paths;
- SMTP, IMAP, and POP3 STARTTLS;
- deployment and filesystem-permission problems;
- systemd service state in native mode;
- recent package and selected configuration changes;
- reboot context and incident timelines;
- certificate rollout mismatches across explicit endpoints.

## Linux flight recorder

HostSleuth periodically records meaningful host state and keeps a bounded local change timeline.

Depending on deployment mode, that can include:

- network interfaces and routes;
- listening sockets;
- Docker container state and published ports;
- systemd services;
- package installation, update, and removal history;
- fingerprints of selected high-value configuration files;
- filesystem state;
- Linux boot identity;
- service, listener, container, package, and configuration changes.

That gives you an evidence trail for the question:

**“What changed before this stopped working?”**

## Deterministic by design

HostSleuth does not rely on an LLM to decide what happened.

Its checks use explicit evidence and bounded states such as `pass`, `fail`, `warning`, and `unknown`. Nearby events are context, not automatically treated as causation.

## Native Linux vs. Docker

Docker is supported and convenient, but container isolation necessarily reduces visibility into the host.

For the fullest HostSleuth experience, **native Linux installation is recommended**.

Native mode can provide evidence that the default Docker deployment intentionally cannot, including systemd/journal context, host package history, selected configuration fingerprints, filesystem inventory, native Certbot lineage information, and the narrowly allowlisted native Safe Actions.

Docker mode does not pretend inaccessible host evidence exists.

## Security posture

HostSleuth is intentionally:

- local-first;
- read-only by default;
- loopback-bound by default;
- deterministic before explanatory;
- explicit and bounded about state-changing actions.

The supported container is not privileged, drops Linux capabilities, uses `no-new-privileges`, and uses a read-only container filesystem.

The Docker socket is mounted read-only so HostSleuth can inventory containers. Docker daemon access is inherently security-sensitive even when the socket path is mounted read-only, so only run images and code you trust.

Native Safe Actions are unavailable in Docker mode.

See the complete security model:

https://github.com/xXDasGoGXx/HostSleuth/blob/main/SECURITY.md

## Supported platforms

Stable container images are published for:

- `linux/amd64`
- `linux/arm64`

## Tags

- `latest` — newest verified stable release.
- Numeric tags such as `1.0.0` — pinned stable releases for reproducible deployments.

Development commits do **not** automatically replace `latest`. The `latest` tag moves only as part of the stable release workflow.

## Source, releases, and documentation

Source code:

https://github.com/xXDasGoGXx/HostSleuth

Latest stable GitHub release:

https://github.com/xXDasGoGXx/HostSleuth/releases/latest

Operational/recovery map:

https://github.com/xXDasGoGXx/HostSleuth/blob/main/docs/OPERATIONS.md

MIT License:

https://github.com/xXDasGoGXx/HostSleuth/blob/main/LICENSE

## Philosophy

HostSleuth is not trying to become another generic monitoring platform, browser shell, or automatic-remediation engine.

It exists to remember meaningful host changes and explain reachability failures using deterministic evidence.

**What changed?**

**Why is this endpoint not working?**

Those are the product boundaries.
