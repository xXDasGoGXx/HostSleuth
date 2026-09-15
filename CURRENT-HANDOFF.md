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
- Distribution: one Linux binary with built-in web UI.
- Default web bind: `127.0.0.1:8787`; remote/LAN exposure must be explicit.
- Initial persistence: local JSON snapshot plus append-only JSONL events, using atomic snapshot writes.
- Snapshot and event formats carry explicit `schema_version: 1`.
- SQLite remains a planned storage backend once the event/snapshot schema has stabilized further.
- Diagnosis is deterministic first. AI, if ever added, may explain collected evidence but must not be required for core findings.
- V0.1 is read-only. No automatic remediation.
- Normal release installation must not require Go on the target host; source installation remains available for developers.
- Release automation is driven by `release/v*` branches. The workflow tests, cross-builds, creates SHA256 sums, creates the matching `v*` tag, and publishes the GitHub release.

## Milestones

### M0 — Repository foundation

Status: **complete**

- Public repository created: `xXDasGoGXx/HostSleuth`.
- `CURRENT-HANDOFF.md` and `TO-DO.md` established as continuity/source-of-truth documents.
- Go module initialized.
- MIT `LICENSE` and `SECURITY.md` added.
- README documents scope, safety posture, local build, CLI, and deployment path.

### M1 — Single-host deployable MVP

Status: **release candidate validated; one privileged installation check remains environment-blocked**

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
- release workflow for static Linux `amd64` and `arm64` binaries, version stamping, SHA256 sums, tag creation, and GitHub release publication.

## Real Debian 13 validation

Test host: `openmediavault`, Debian GNU/Linux 13 (trixie), x86_64. Testing was performed through HomeCommander in Normal mode.

A temporary Go 1.24.13 toolchain under `/tmp` was used for source-level validation only; Go was **not** installed persistently/system-wide on the host.

Confirmed working:

- `go vet ./...`;
- `go test ./...`;
- native Linux build;
- static cross-builds for Linux `amd64` and `arm64`;
- release version stamping;
- snapshot collection;
- filesystem inventory: 117 mounts observed; `/` identified as ext4 on `/dev/nvme0n1p1` with capacity data;
- systemd inventory: 219 services observed;
- listener inventory: 329 listeners at baseline;
- `diagnose 127.0.0.1:22` returned reachable/high confidence;
- `diagnose 127.0.0.1:65530` returned no local listener/high confidence;
- dashboard and `/api/snapshot`, `/api/events`, `/api/diagnose` endpoints;
- periodic flight-recorder behavior: starting the web listener emitted `listener appeared: tcp 127.0.0.1:8787`; stopping it and taking the next snapshot emitted `listener disappeared: tcp 127.0.0.1:8787`;
- schema persistence and empty event-list behavior (`[]`, not JSON `null`);
- install/source-install/uninstall shell syntax;
- systemd unit syntax. `systemd-analyze verify` only warned that `/usr/local/bin/hostsleuth` was absent on the non-installed test machine.

## GitHub CI and release validation

- Initial CI failed only `gofmt`; that was corrected and preserved in commit `e8185ae4e36ba0b9d0ec621369524762c54057d5`.
- Hardened M1 code commit: `69fe6c0add132e28915672f161062f60c0765635`.
- GitHub Actions CI run **34936129425** passed every step: format, vet, tests, and build.
- Release automation commit: `a97932caaccf9c0e8c6d095592dc4cb049648664`.
- Release branch: `release/v0.1.0-alpha.1`.
- Release workflow run **34936275773** passed tests, built both architectures, and successfully published the tag/release.
- Published release: **`v0.1.0-alpha.1`**.
- Published assets: `hostsleuth-linux-amd64`, `hostsleuth-linux-arm64`, `SHA256SUMS`.
- Published amd64 SHA-256: `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`.
- The published amd64 binary was downloaded onto the Debian 13 host using the exact release URL. `sha256sum -c` passed.
- The same host confirmed `go` is not installed system-wide, yet the downloaded binary ran successfully, reported `v0.1.0-alpha.1`, captured a real snapshot, and diagnosed SSH correctly.
- The default installer-style `/releases/latest/download/hostsleuth-linux-amd64` URL was also tested and returned the same correct version and SHA-256.

## Important validation limitation

- HomeCommander Normal mode does not permit privileged writes to `/usr/local/bin`, `/etc/systemd/system`, or administrative `systemctl enable --now`, so the final root-level copy/unit-enable portion of `scripts/install.sh` has **not** been executed through this environment.
- HomeCommander also blocks direct Docker commands and the test process is unprivileged, so the observed `containers=0` does **not** prove that the host has no containers. Docker collection still needs validation in a context where the HostSleuth process can read Docker state.

## Important current limitations

- No authentication yet; keep the dashboard loopback-only.
- State storage is JSON/JSONL, not SQLite yet.
- Collectors intentionally degrade if optional tools (`ip`, `ss`, `systemctl`, `docker`) are absent or inaccessible.
- Diagnosis does not yet correlate nftables/UFW, reverse proxies, Docker namespaces/ports, systemd journal failures, TLS, or package/config changes.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`
First published release: `v0.1.0-alpha.1`

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Next actions

1. Exercise the privileged `scripts/install.sh` systemd installation path on an approved host whenever administrative execution is available; verify restart/reboot persistence and uninstall behavior.
2. Validate Docker collection under the eventual service privilege model.
3. Once that M1 installation check is closed, begin the next diagnosis milestone with route/firewall evidence and bounded systemd/journal failure evidence.
4. Continue with Docker port/network correlation, then package/config change tracking.
