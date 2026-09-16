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
  - **M3.2 Usability:** implementation complete on PR #13; pending final green CI/merge.
  - **M3.3 Docker release:** next, only after M3.2 merge.

The exact checklist lives in `TO-DO.md` and is updated as work progresses.

## Product-direction rule

HostSleuth should remain easy to deploy, easy to understand, and intentionally small. New capability is not automatically good capability.

Ideas such as authentication, proxy/TLS awareness, SQLite, AI explanation, multi-host support, remediation, and other larger additions remain visible in the **collective review** section of `TO-DO.md`. They are deferred decisions, not permanent bans and not promised features.

Promote one only when it solves a common real user problem without making installation, operation, or the UI meaningfully harder.

## Current validated deployment state

The M2 diagnostic core has been validated on Debian 13 as a managed systemd service with the local dashboard available on loopback. Validation included Docker and systemd inventory, route and nftables evidence, listener/bind correlation, and normal event recording.

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

### M3.1 Web UI

The merged UI provides responsive Overview / Diagnose / Changes / Host views, readable diagnosis presentation, a recent-change timeline, host/interface/filesystem views, and self-contained HTML/CSS/vanilla JavaScript embedded in the Go binary.

### M3.2 usability implementation

PR #13 keeps product capability unchanged while improving use and onboarding:

- simpler user-facing wording;
- clear first-run / empty-history explanation;
- friendly diagnosis titles while preserving the exact deterministic conclusion and evidence;
- running version and VCS revision visible through CLI/Web UI when available;
- newest changes shown first, including the newest five on the Overview;
- recommended native/systemd install path moved to the top of the README;
- safe SSH-tunnel instructions for remote access to the loopback-only UI;
- CI now parses the embedded JavaScript in addition to Go formatting, vet, tests, and build.

A real README UI screenshot is intentionally deferred until an accepted deployed build can be captured; no mock screenshot will be used just to satisfy documentation.

## Current limitations

- dashboard is loopback-only and has no authentication;
- storage is JSON/JSONL;
- reverse-proxy and TLS-specific diagnosis are not implemented;
- configuration/package change tracking is not implemented;
- HostSleuth is single-host first;
- no automatic remediation.

These limitations stay visible for collective product review; they do not automatically become the next work.

## Active branch and scope

Active branch: `m3/usability`

Open PR: `#13 — M3.2: simplify HostSleuth usability`

Do not start Docker packaging on this branch. After PR #13 is green and merged, create the Docker release branch from the then-current `main`.

## Branch hygiene

`main` remains the authoritative stable development state.

- `m3/usability` is the only active feature branch for the current task.
- `m3/product-experience` is historical after merged PR #11.
- Old M1/M2 feature branches are historical leftovers after merged work.
- `m3/npm-proxy-awareness` and `m3/tls-diagnostics` contain historical experimental work and are not current product state.

## Next task

**Merge M3.2 after final green CI, then begin M3.3 Docker packaging from current `main`.**

After the M3 sequence, review the deferred product decisions collectively and decide what, if anything, truly belongs in the product.

## Repository source of truth

Repository: `xXDasGoGXx/HostSleuth`

Read in this order when resuming:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`

Use `docs/history/DEVELOPMENT-HISTORY.md` only when historical implementation/validation details are needed.

Public release promotion remains owner-controlled and must not happen without explicit approval.
