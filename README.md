# HostSleuth

**A small, local-first Linux flight recorder for two questions: what changed, and why can I not reach this host/service/port?**

HostSleuth records meaningful Linux state changes and performs deterministic, evidence-backed `host:port` diagnosis. It does not require a cloud account, external database, API key, or AI model.

## Quick start — recommended native install

HostSleuth currently works best as a native Linux service because it needs to observe the real host.

Requirements: Linux with systemd. Go is **not** required when installing a release.

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
sudo ./scripts/install.sh
```

Check that it is running:

```bash
systemctl status hostsleuth --no-pager
hostsleuth version
```

HostSleuth binds its Web UI to loopback by default for safety.

If you are using HostSleuth on the same machine, open:

```text
http://127.0.0.1:8787
```

If HostSleuth is running on a remote server, use an SSH tunnel from your computer:

```bash
ssh -L 8787:127.0.0.1:8787 user@your-server
```

Then open `http://127.0.0.1:8787` locally.

A new install starts by taking a baseline snapshot. An empty change timeline is normal until HostSleuth observes a meaningful change.

## What it does

### Remember meaningful changes

HostSleuth periodically captures useful local state including:

- host identity, OS, kernel, memory, and filesystems;
- interfaces and routes;
- listening sockets;
- systemd services;
- Docker containers, published ports, state, and network names when Docker is readable.

It compares snapshots and records meaningful service/container/listener changes in a local event timeline. Known noisy changes such as Docker uptime progression are suppressed.

### Diagnose `host:port`

Enter a target such as `192.168.1.20:443` or `example.com:443`. HostSleuth can combine:

- DNS resolution;
- kernel route evidence;
- TCP connectivity;
- target-aware local TCP listener evidence;
- Docker port publication, bind address, and network context;
- bounded nftables candidate evidence;
- bounded failed-systemd candidate evidence.

The Web UI presents the answer first and keeps the underlying evidence available for inspection. Successful TCP is definitive, stronger local bind/listener evidence outranks weaker candidates, and unavailable optional evidence stays `unknown` rather than becoming a false failure.

## CLI basics

Capture a snapshot:

```bash
hostsleuth snapshot
```

Diagnose a target:

```bash
hostsleuth diagnose example.com:443
```

Show recent events:

```bash
hostsleuth events
```

Show the installed version:

```bash
hostsleuth version
```

## Install a specific release

```bash
sudo HOSTSLEUTH_VERSION=v0.1.0-alpha.1 ./scripts/install.sh
```

## Build from source

Requirements: Linux and Go 1.24+.

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
go test ./...
go build -o hostsleuth ./cmd/hostsleuth
./hostsleuth serve
```

Developers who intentionally want the installer to build on the target host can use:

```bash
sudo ./scripts/install-source.sh
```

## Uninstall

Remove the binary and service while preserving recorded HostSleuth state:

```bash
sudo ./scripts/uninstall.sh
```

## Product boundaries

HostSleuth is intentionally:

- local-first;
- single-host first;
- read-only;
- deterministic before explanatory;
- loopback-only by default for the Web UI.

HostSleuth does **not** automatically restart services, modify firewall rules, repair containers, or reconfigure the host.

## Current stage

M0 repository foundation, M1 deployable single-host MVP, M2 deeper deterministic diagnosis, and M3.1 Web UI are complete.

M3.2 is focused on usability: clearer wording, first-run behavior, version visibility, and simpler installation guidance. M3.3 will address an official Docker deployment without turning HostSleuth into a complicated multi-mode product.

Larger ideas remain deferred for collective product review rather than automatically becoming features. See `TO-DO.md` for the current sequence and decision list.

## Security and privacy

HostSleuth can collect hostnames, IP addresses, mount paths, service names, listener addresses, and container metadata. Treat snapshots, event logs, and diagnostic output as potentially sensitive. See [`SECURITY.md`](SECURITY.md) for the current security posture and vulnerability-reporting guidance.

## Project files

- `README.md` — product definition, usage, and current scope.
- `CURRENT-HANDOFF.md` — current project state and next task.
- `TO-DO.md` — active milestone checklist and deferred product decisions.
- `docs/history/DEVELOPMENT-HISTORY.md` — milestone and validation history.

## License

HostSleuth is released under the MIT License. See `LICENSE`.
