# HostSleuth

**A small, local-first Linux flight recorder for two questions: what changed, and why can I not reach this host/service/port?**

HostSleuth is intentionally not a full monitoring platform. It records meaningful Linux state changes and performs deterministic, evidence-backed `host:port` diagnosis without requiring a cloud account, AI model, external database, or API key.

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

HostSleuth can combine:

- DNS resolution;
- kernel route evidence;
- TCP connectivity;
- target-aware local TCP listener evidence;
- Docker port publication, bind address, and network context;
- bounded nftables candidate evidence;
- bounded failed-systemd candidate evidence.

The diagnosis engine has explicit evidence precedence: successful TCP is definitive, stronger local bind/listener evidence outranks weaker candidates, and unavailable optional evidence stays `unknown` rather than becoming a false failure.

## Product boundaries

The current product is:

- local-first;
- single-host first;
- read-only;
- deterministic before explanatory;
- loopback-only by default for the web UI.

HostSleuth does **not** automatically restart services, modify firewall rules, repair containers, or reconfigure the host.

## Try it from source

Requirements: Linux and Go 1.24+.

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
go test ./...
go build -o hostsleuth ./cmd/hostsleuth
```

Capture a snapshot:

```bash
./hostsleuth snapshot
```

Diagnose a target:

```bash
./hostsleuth diagnose example.com:443
```

Show recent events:

```bash
./hostsleuth events
```

Start the local dashboard:

```bash
./hostsleuth serve
```

Then open `http://127.0.0.1:8787` on the same machine.

## Install a release with systemd

Release installation does **not** require Go on the target host. Run the installer from a HostSleuth repository checkout so it can install the included systemd unit:

```bash
sudo ./scripts/install.sh
```

To install a specific release:

```bash
sudo HOSTSLEUTH_VERSION=v0.1.0-alpha.1 ./scripts/install.sh
```

Developers who intentionally want to build on the target host can use:

```bash
sudo ./scripts/install-source.sh
```

Uninstall the binary/service while preserving recorded state:

```bash
sudo ./scripts/uninstall.sh
```

## Current stage

M0 repository foundation, M1 deployable single-host MVP, and M2 deeper deterministic diagnosis are complete and have been validated on Debian 13.

The next phase is deliberately boring: **use the current build on real troubleshooting cases and improve the product where actual use exposes confusion or missing evidence.** New large subsystems are not the default next step.

## Security and privacy

HostSleuth can collect hostnames, IP addresses, mount paths, service names, listener addresses, and container metadata. Treat snapshots, event logs, and diagnostic output as potentially sensitive. See [`SECURITY.md`](SECURITY.md) for the current security posture and vulnerability-reporting guidance.

## Project files

- `README.md` — product definition, usage, and current scope.
- `CURRENT-HANDOFF.md` — concise current project state and next task.
- `TO-DO.md` — short active backlog.
- `docs/history/DEVELOPMENT-HISTORY.md` — milestone and validation history.

## License

HostSleuth is released under the MIT License. See `LICENSE`.
