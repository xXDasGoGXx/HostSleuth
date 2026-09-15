# HostSleuth — Current Handoff

Last updated: 2026-09-14

## Current objective

Build a small, local-first Linux troubleshooting tool that answers two questions better than a conventional monitoring dashboard:

1. **What changed on this Linux host?**
2. **Why can I not reach this host/service/port?**

The first deployable milestone targets Debian 12/13 and Ubuntu 24.04+ and stays read-only: HostSleuth observes, records, compares, and explains. It does not repair or mutate the host.

## Product direction

HostSleuth is intentionally not a full metrics/observability platform. Its differentiator is a lightweight Linux "flight recorder" plus deterministic, evidence-backed diagnostics. Core operation must not require a cloud account, AI model, external database, or API key.

## Architectural decisions

- Language: Go.
- Distribution direction: one Linux binary with built-in web UI.
- Default web bind: `127.0.0.1:8787`; remote/LAN exposure must be explicit.
- Initial persistence: local JSON snapshot plus append-only JSONL events, using atomic snapshot writes.
- SQLite remains a planned storage backend once the event/snapshot schema has stabilized.
- Diagnosis is deterministic first. AI, if ever added, may explain collected evidence but must not be required for core findings.
- V0.1 is read-only. No automatic remediation.

## Completed work

### M0 — Repository foundation

Status: **complete**

- Public repository created: `xXDasGoGXx/HostSleuth`.
- `CURRENT-HANDOFF.md` and `TO-DO.md` established as continuity/source-of-truth documents.
- Go module initialized.
- README documents scope, safety posture, local build, CLI, and deployment path.

### M1 — Single-host deployable MVP

Status: **in progress; first runnable vertical slice committed**

Implemented:

- host model and snapshot schema;
- host/OS/kernel/CPU/memory/uptime inventory;
- interface/address collection;
- route collection;
- listening socket collection via `ss` when available;
- systemd service inventory via `systemctl` when available;
- Docker container inventory when Docker is available;
- atomic local snapshot storage;
- append-only JSONL change events;
- deterministic service/container/listener snapshot diffing;
- `diagnose host:port` with target validation, DNS, TCP connection evidence, and local-listener correlation;
- CLI commands: `snapshot`, `diagnose`, `events`, `serve`, `version`;
- built-in local dashboard and JSON endpoints;
- default loopback-only web bind on `127.0.0.1:8787`;
- initial unit tests;
- GitHub Actions workflow for formatting, vet, tests, and build;
- systemd unit and source install script.

## Validation state

A GitHub Actions CI run is currently queued for the latest commit. **M1 must not be marked complete until CI has passed and deployment is exercised on a real Debian/Ubuntu host.**

## Important limitations in current code

- No authentication yet; keep the dashboard loopback-only.
- Installer currently builds from source and therefore requires Go on the target host.
- State storage is JSON/JSONL, not SQLite yet.
- Collectors intentionally degrade if optional tools (`ip`, `ss`, `systemctl`, `docker`) are absent.
- Diagnosis does not yet correlate nftables/UFW, reverse proxies, Docker namespaces, systemd journal failures, TLS, or package/config changes.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Next actions

1. Inspect the first CI result and fix any formatting/build/test failures.
2. Once CI is green, deploy to a Debian/Ubuntu test host and verify `snapshot`, `events`, `diagnose`, systemd service, and dashboard behavior.
3. Harden the installer so normal deployment does not require a Go toolchain (release binary / `.deb`).
4. Improve diagnosis with routes, firewall evidence, failed systemd state/journal evidence, and Docker port correlation.
5. Add configuration/package change tracking only after the baseline deployment is stable.
