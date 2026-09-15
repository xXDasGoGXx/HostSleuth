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
- **Next phase:** product hardening through real use. Do not start another large subsystem until real usage exposes a concrete gap.

Detailed milestone history is archived in `docs/history/DEVELOPMENT-HISTORY.md`.

## Current live deployment

Host: `openmediavault`

Installed HostSleuth:

- version: `0.1.0-dev+3352a7e`
- executable SHA-256: `5d595d9db476f9cc030d052df53f6139057fdc6f74f6e440e9f833e5cee201aa`
- systemd unit SHA-256: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`
- service: loaded, enabled, active/running
- validated PID after M2 update: `443538`
- dashboard: HTTP 200 on `127.0.0.1:8787`
- state directory: `/var/lib/hostsleuth`
- Docker inventory at validation: 21 containers, all with network names
- systemd inventory at validation: 219 services
- failed services at validation: 0

HomeCommander is only the owner-approved administrative transport used to install/restart the live service. It is not part of HostSleuth itself.

## What works now

### Change recorder

HostSleuth captures and compares:

- host/OS/kernel state;
- filesystems and capacity;
- interfaces and routes;
- listening sockets;
- systemd services;
- Docker containers, ports, state, and network names.

It records meaningful service/container/listener events while suppressing Docker uptime-only churn.

### Deterministic diagnosis

`diagnose host:port` can currently use:

- DNS resolution;
- kernel route evidence;
- TCP connectivity;
- target-aware local TCP listener evidence;
- Docker publication/bind/network evidence;
- bounded nftables candidate evidence;
- bounded failed-systemd candidate evidence.

Evidence precedence is deliberate: successful TCP is definitive; strong local bind/listener evidence outranks weaker firewall/systemd candidates; unavailable optional evidence stays neutral.

## Real M2 validation

The current installed build was validated on Debian 13 without manufacturing destructive failures.

Confirmed examples:

- `127.0.0.1:22` -> reachable / high confidence.
- closed local port -> no listener / high confidence.
- `127.0.0.1:8789` -> Chaptarr is published on `192.168.2.181:8789`, not loopback.
- `127.0.0.1:8192` -> FlareSolverr exposes `8192/tcp` internally but does not publish it on the host.
- `192.168.2.181:8789` -> reachable / high confidence.
- root-context route lookup and nftables evidence work.
- dashboard/API and change recording remain healthy.

The live snapshot contained zero failed systemd services, so journal excerpts were not forced by intentionally breaking a service. That path remains covered by tests.

## Current limitations

- dashboard is loopback-only and has no authentication;
- storage is still JSON/JSONL;
- reverse-proxy and TLS-specific diagnosis are not implemented;
- configuration/package change tracking is not implemented;
- HostSleuth is single-host first;
- no automatic remediation.

These are backlog items, not reasons to expand the product immediately.

## Branch hygiene

`main` is the only authoritative development state.

- Old M1/M2 feature branches are historical leftovers after merged work.
- `m3/npm-proxy-awareness` and `m3/tls-diagnostics` contain unmerged experimental work and are **not part of the product**.
- Do not merge, continue, or treat those M3 branches as active unless a future real-world need explicitly justifies reviving them.
- `m3/generic-nginx-awareness` contains no work ahead of `main`.

## Next task

**Use the current build as the product.**

For the next development work:

1. run HostSleuth during real troubleshooting;
2. note where the dashboard, event timeline, or diagnosis output is confusing or insufficient;
3. improve those concrete user-facing gaps first;
4. add a new collector/subsystem only when a real case proves it is needed.

Do not begin SQLite, reverse-proxy/TLS work, multi-host support, AI explanation, or repair automation just because they exist in the backlog.

## Repository source of truth

Repository: `xXDasGoGXx/HostSleuth`

Read in this order when resuming:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`

Use `docs/history/DEVELOPMENT-HISTORY.md` only when historical implementation/validation details are needed.

Public release promotion remains owner-controlled and must not happen without explicit approval.
