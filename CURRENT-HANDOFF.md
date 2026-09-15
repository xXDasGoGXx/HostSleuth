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

- Language: Go 1.24+.
- Distribution direction: one Linux binary with built-in web UI.
- Default web bind: `127.0.0.1:8787`; remote/LAN exposure must be explicit.
- Initial persistence: local JSON snapshot plus append-only JSONL events, using atomic snapshot writes.
- Snapshot and event formats now carry explicit `schema_version: 1`.
- SQLite remains a planned storage backend once the event/snapshot schema has stabilized further.
- Diagnosis is deterministic first. AI, if ever added, may explain collected evidence but must not be required for core findings.
- V0.1 is read-only. No automatic remediation.
- Normal release installation must not require Go on the target host; source installation remains available for developers.

## Completed work

### M0 — Repository foundation

Status: **complete**

- Public repository created: `xXDasGoGXx/HostSleuth`.
- `CURRENT-HANDOFF.md` and `TO-DO.md` established as continuity/source-of-truth documents.
- Go module initialized.
- MIT `LICENSE` and `SECURITY.md` added.
- README documents scope, safety posture, local build, CLI, and deployment path.

### M1 — Single-host deployable MVP

Status: **in progress; core runtime validated on Debian 13, CI revalidation pending**

Implemented:

- explicit snapshot/event schema version 1;
- host/OS/kernel/CPU/memory/uptime inventory;
- filesystem/mount/capacity inventory;
- interface/address collection;
- route collection;
- listening socket collection via `ss` when available;
- systemd service inventory via `systemctl` when available;
- Docker container inventory when Docker is available and readable by the running user;
- atomic local snapshot storage;
- append-only JSONL change events;
- deterministic service/container/listener snapshot diffing;
- `diagnose host:port` with target validation, DNS, TCP connection evidence, and local-listener correlation;
- CLI commands: `snapshot`, `diagnose`, `events`, `serve`, `version`;
- built-in local dashboard and JSON endpoints;
- default loopback-only web bind on `127.0.0.1:8787`;
- unit tests including storage/schema behavior;
- GitHub Actions CI for formatting, vet, tests, and build;
- systemd unit;
- release-binary installer, source installer, and uninstall script;
- tagged-release workflow that builds static Linux `amd64` and `arm64` binaries, stamps the release version, writes SHA256 sums, and publishes a GitHub release.

## Real Debian 13 validation

Test host: `openmediavault`, Debian GNU/Linux 13 (trixie), x86_64. Testing was performed through HomeCommander in Normal mode.

Validated with a temporary Go 1.24.13 toolchain under `/tmp` only; Go was **not** installed persistently on the host.

Confirmed working:

- `go vet ./...`;
- `go test ./...`;
- native Linux build;
- static cross-builds for Linux `amd64` and `arm64`;
- release version stamping (`v0.1.0-alpha.1` test build);
- snapshot collection;
- filesystem inventory: 117 mounts were observed on the test host; `/` was identified as ext4 on `/dev/nvme0n1p1` with capacity data;
- systemd inventory: 219 services observed;
- listener inventory: 329 listeners at baseline;
- `diagnose 127.0.0.1:22` correctly returned reachable/high confidence;
- `diagnose 127.0.0.1:65530` correctly returned no local listener/high confidence;
- dashboard and `/api/snapshot`, `/api/events`, `/api/diagnose` endpoints;
- periodic flight-recorder behavior: starting the web listener produced `listener appeared: tcp 127.0.0.1:8787`; stopping it and taking the next snapshot produced `listener disappeared: tcp 127.0.0.1:8787`;
- schema persistence and empty event-list behavior (`[]`, not JSON `null`);
- install/source-install/uninstall shell syntax;
- systemd unit syntax. `systemd-analyze verify` only warned that `/usr/local/bin/hostsleuth` was not installed on the test machine, which is expected.

Important validation limitation:

- HomeCommander Normal mode does not permit privileged writes to `/usr/local/bin`, `/etc/systemd/system`, or service installation, so the actual root-level installer/systemd enable path has **not** yet been exercised through HomeCommander.
- HomeCommander also blocks direct Docker commands and the test process is unprivileged, so the observed `containers=0` does **not** prove that the host has no containers; Docker collection needs validation in a context where the HostSleuth service can read Docker state.

## CI history

- Initial CI run failed only the formatting gate (`gofmt`); vet/test/build were skipped by that run.
- Formatting was corrected and independently verified on Debian with Go 1.24.13.
- Commit `e8185ae4e36ba0b9d0ec621369524762c54057d5` preserved the gofmt correction.
- Commit `69fe6c0add132e28915672f161062f60c0765635` preserved schema/filesystem/release/deployment hardening.
- A fresh normal GitHub contents commit is being used after that milestone to trigger CI again. **Do not mark M1 complete until the new CI run is green.**

## Important limitations in current code

- No authentication yet; keep the dashboard loopback-only.
- State storage is JSON/JSONL, not SQLite yet.
- Collectors intentionally degrade if optional tools (`ip`, `ss`, `systemctl`, `docker`) are absent or inaccessible.
- Diagnosis does not yet correlate nftables/UFW, reverse proxies, Docker namespaces/ports, systemd journal failures, TLS, or package/config changes.
- Release installer depends on a published GitHub release; the first tagged release has not yet been published.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Next actions

1. Confirm the fresh GitHub Actions CI run is green; fix any remaining CI/build issue before closing M1.
2. Reconcile `TO-DO.md` after CI.
3. Publish the first alpha release so the no-Go binary installer can be exercised end-to-end.
4. Exercise the privileged systemd install path on an approved test host when administrative execution is available.
5. After the M1 baseline is locked, improve diagnosis with route evidence, firewall evidence, failed systemd/journal evidence, and Docker port correlation.
