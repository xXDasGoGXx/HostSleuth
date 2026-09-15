# HostSleuth

Local-first Linux change recorder and deterministic root-cause diagnostics.

> **Status:** early development. The current milestone is a read-only single-host MVP for Debian/Ubuntu.

HostSleuth is designed around two questions:

1. **What changed on this Linux host?**
2. **Why can I not reach this host/service/port?**

It is intentionally not a full monitoring platform. The goal is a small, inspectable Linux "flight recorder" that preserves meaningful host state, turns changes into concise events, and performs evidence-backed diagnostics without requiring a cloud account, AI model, or external database.

## Current capabilities

- Collects host identity, OS/kernel, memory, interfaces, routes, listeners, systemd services, and Docker containers when available.
- Stores the latest snapshot locally using an atomic write.
- Appends detected service/container/listener changes to a JSONL event timeline.
- Runs deterministic `host:port` diagnosis with DNS, TCP, and local-listener evidence.
- Serves a small built-in dashboard and JSON API.
- Defaults to `127.0.0.1:8787` so host data is not exposed to the LAN automatically.

## Build

Requirements: Linux and Go 1.24+.

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
go test ./...
go build -o hostsleuth ./cmd/hostsleuth
```

## Try it without installing

Capture a snapshot:

```bash
./hostsleuth snapshot
```

Diagnose a target:

```bash
./hostsleuth diagnose example.com:443
```

Show recent change events:

```bash
./hostsleuth events
```

Start the local dashboard:

```bash
./hostsleuth serve
```

Then open `http://127.0.0.1:8787` on the same machine.

## Source install with systemd

```bash
sudo ./scripts/install.sh
```

This builds `/usr/local/bin/hostsleuth`, creates `/var/lib/hostsleuth`, installs the systemd unit, and starts HostSleuth on loopback port 8787.

## Safety posture

The current MVP is **read-only**. HostSleuth does not repair, restart, reconfigure, or mutate monitored services. Remote/LAN dashboard exposure is not a supported default until authentication is implemented.

## Project continuity

- `CURRENT-HANDOFF.md` — exact current project state and next actions.
- `TO-DO.md` — active milestone and future backlog.

Those files are maintained as part of the project source of truth.

## License

A project license will be selected before the first public release. Until then, source is visible for development and evaluation but no additional reuse rights are granted beyond applicable law.
