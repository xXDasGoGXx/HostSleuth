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
  - **M3.1 Web UI:** active now.
  - **M3.2 Usability:** next, only after M3.1.
  - **M3.3 Docker release:** after M3.2.

The exact M3 checklist is locked in `TO-DO.md`. Do not broaden M3 with unrelated collectors, storage changes, proxy/TLS work, AI, multi-host work, or remediation.

Detailed milestone history is archived in `docs/history/DEVELOPMENT-HISTORY.md`.

## Current validated deployment state

The current M2 product build has been validated on Debian 13 as a managed systemd service with the local dashboard available on loopback. Validation included Docker and systemd inventory, route and nftables evidence, listener/bind correlation, and normal event recording.

Environment-specific hostnames, addresses, process IDs, inventory counts, and local deployment hashes are intentionally omitted from this public repository handoff.

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

The current product code was validated on Debian 13 without manufacturing destructive failures.

Confirmed classes of behavior include:

- reachable local services -> reachable / high confidence;
- closed local ports -> no listener / high confidence;
- listeners bound to a different local address are not treated as evidence for loopback;
- container-internal ports are distinguished from host-published ports;
- published host ports are diagnosed correctly on the address where they are actually bound;
- root-context route lookup and nftables evidence work;
- dashboard/API and change recording remain healthy.

The validation host had no failed systemd services at the time, so journal excerpts were not forced by intentionally breaking a service. That path remains covered by tests.

## Current limitations

- dashboard is loopback-only and has no authentication;
- storage is still JSON/JSONL;
- reverse-proxy and TLS-specific diagnosis are not implemented;
- configuration/package change tracking is not implemented;
- HostSleuth is single-host first;
- no automatic remediation.

These are backlog items, not reasons to expand the active M3 scope.

## Active branch and scope

Active branch: `m3/product-experience`

M3.1 changes only the user experience around capabilities HostSleuth already has:

- modern responsive web interface;
- readable diagnosis presentation;
- recent-change timeline;
- host overview;
- loading/empty/error states;
- mobile-friendly layout;
- self-contained assets embedded in the Go binary with no frontend framework.

Do not begin Docker packaging on this branch. Once M3.1 is merged and M3.2 is complete, create the Docker release branch from the then-current `main`.

## Branch hygiene

`main` remains the authoritative stable development state.

- `m3/product-experience` is the only active feature branch for the current task.
- Old M1/M2 feature branches are historical leftovers after merged work.
- `m3/npm-proxy-awareness` and `m3/tls-diagnostics` contain unmerged experimental work and are **not part of the product**.
- Do not merge, continue, or treat those experimental branches as active unless a future real-world need explicitly justifies reviving them.

## Next task

**Complete M3.1 Web UI and nothing else.**

Use the current APIs and deterministic diagnosis output. Improve presentation rather than adding diagnostic engines or backend subsystems.

## Repository source of truth

Repository: `xXDasGoGXx/HostSleuth`

Read in this order when resuming:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`

Use `docs/history/DEVELOPMENT-HISTORY.md` only when historical implementation/validation details are needed.

Public release promotion remains owner-controlled and must not happen without explicit approval.
