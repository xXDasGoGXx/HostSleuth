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
- Initial distribution: one static Linux binary with embedded web UI.
- Default web bind: `127.0.0.1:8787`; remote/LAN exposure must be explicit.
- Initial persistence: local JSON snapshots plus append-only JSONL events, using atomic writes. This keeps V0.1 dependency-free and makes captures inspectable by humans.
- SQLite remains a planned storage backend once the event/snapshot schema has stabilized.
- Diagnosis is deterministic first. AI, if ever added, may explain collected evidence but must not be required for core findings.
- V0.1 is read-only. No automatic remediation.

## Milestones

### M0 — Repository foundation

Status: **in progress**

Deliverables: project docs, security posture, handoff, to-do list, Go module skeleton.

### M1 — Single-host deployable MVP

Status: **planned next**

Required capabilities:

- host inventory;
- periodic state snapshots;
- meaningful change events;
- deterministic `host:port` diagnosis;
- embedded local web UI;
- systemd packaging/install path;
- CI that builds and tests on Linux.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`

This file and `TO-DO.md` must be kept synchronized with meaningful project progress.

## Next actions

1. Complete M0 documentation and project skeleton.
2. Implement core model, collectors, state store, differ, and diagnosis engine.
3. Add embedded web UI and CLI/server orchestration.
4. Add unit tests and GitHub Actions CI.
5. Add source/release deployment scripts and systemd unit.
6. Verify CI is green before declaring M1 complete.
