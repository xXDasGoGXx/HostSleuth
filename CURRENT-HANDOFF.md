# HostSleuth — Current Handoff

Last updated: 2026-09-15

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
- HostSleuth remains independent from HomeCommander. HomeCommander is only the shared owner-approved administrative transport for privileged installation/lifecycle operations.

## Milestones

### M0 — Repository foundation

Status: **complete**

- Public repository created: `xXDasGoGXx/HostSleuth`.
- `CURRENT-HANDOFF.md` and `TO-DO.md` established as continuity/source-of-truth documents.
- Go module initialized.
- MIT `LICENSE` and `SECURITY.md` added.
- README documents scope, safety posture, local build, CLI, and deployment path.

### M1 — Single-host deployable MVP

Status: **release candidate validated; managed administrative deployment UAT staged; owner root approval is the current boundary**

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

## Managed Administrative Deployment UAT — 2026-09-15

HomeCommander is live on `openmediavault` with the reviewed exact-hash Managed Administrative Deployment capability. HostSleuth is the first real UAT consumer.

Verified before staging:

- HomeCommander gateway healthy on `openmediavault`;
- Normal-mode read/write policy unchanged;
- `/srv/homecommander-deployments` initially empty;
- no `hostsleuth` managed-deployment approval exists yet;
- `/usr/local/bin/hostsleuth` is not managed/installed through this path yet;
- `/var/lib/hostsleuth` does not yet exist.

Staged UAT artifacts:

- `/srv/homecommander-deployments/hostsleuth/hostsleuth`
  - source: published `v0.1.0-alpha.1` Linux amd64 release;
  - SHA-256: `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`;
  - checksum matches the published release handoff exactly.
- `/srv/homecommander-deployments/hostsleuth/hostsleuth.service`
  - SHA-256: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`.

A first-install integration issue was caught before root approval: the original `scripts/install.sh` created `/var/lib/hostsleuth`, but managed deployment intentionally promotes only the exact approved binary and systemd unit. The unit has therefore been adjusted to use:

- `StateDirectory=hostsleuth`
- `StateDirectoryMode=0700`

This lets systemd create and manage `/var/lib/hostsleuth` without broadening HomeCommander privileged capabilities. The change is isolated in branch `uat/systemd-state-directory`, commit `bb7552a43130072fb47cef30c8eed2a6e8ecb013`, PR #1. CI run `34987407246` was started for the PR.

`systemd-analyze verify` accepts the staged unit; its only warning is the expected absence of `/usr/local/bin/hostsleuth` before managed installation.

Current owner boundary: the staged artifacts must be approved out-of-band by root using `homecommander-approve-deployment`. HomeCommander must not self-approve them.

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

- The previous HomeCommander limitation around first-time privileged installation is now solved by the shared Managed Administrative Deployment mechanism. The current remaining boundary is the deliberate owner/root approval of the exact staged HostSleuth artifacts.
- HomeCommander blocks direct Docker commands and the prior test process was unprivileged, so the earlier observed `containers=0` does **not** prove that the host has no containers. Docker collection still needs validation once HostSleuth is running under the final system service privilege model.

## Important current limitations

- No authentication yet; keep the dashboard loopback-only.
- State storage is JSON/JSONL, not SQLite yet.
- Collectors intentionally degrade if optional tools (`ip`, `ss`, `systemctl`, `docker`) are absent or inaccessible.
- Diagnosis does not yet correlate nftables/UFW, reverse proxies, Docker namespaces/ports, systemd journal failures, TLS, or package/config changes.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`
First published release: `v0.1.0-alpha.1`
Active UAT branch: `uat/systemd-state-directory`
Active UAT PR: #1

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Next actions

1. Let PR #1 CI complete and merge the systemd-owned state-directory fix only if validation is green.
2. Owner/root approves the exact staged `hostsleuth` deployment; no generic sudo/root bypass.
3. Use HomeCommander `deployment_action`/`deployment_status` for install, enable/start, restart, stop/disable, uninstall/reinstall and managed update behavior as appropriate.
4. Verify service runtime, `/var/lib/hostsleuth` creation/mode, dashboard/API, reboot persistence, and Docker inventory under the actual service privilege model.
5. Close M1 only after privileged installation/reboot/uninstall/reinstall UAT is complete.
6. Then begin the next diagnosis milestone with route/firewall evidence and bounded systemd/journal failure evidence.
