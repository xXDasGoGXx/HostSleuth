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

Status: **managed installation UAT passed; only real reboot-persistence validation remains before M1 closure**

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

### Packaging fix discovered before install

The original `scripts/install.sh` created `/var/lib/hostsleuth`, while managed deployment intentionally promotes only the exact approved binary and systemd unit. HostSleuth PR #1 therefore added:

- `StateDirectory=hostsleuth`
- `StateDirectoryMode=0700`

This keeps HostSleuth self-contained and lets systemd create `/var/lib/hostsleuth` without broadening HomeCommander privilege.

Validation/promotion:

- PR #1: `Make systemd create HostSleuth state directory`;
- CI run `34987407246`: success;
- merged to `main` at `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`.

### Approved UAT artifacts

Staged and owner-approved artifacts:

- `/srv/homecommander-deployments/hostsleuth/hostsleuth`
  - published `v0.1.0-alpha.1` Linux amd64 release;
  - SHA-256: `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`.
- `/srv/homecommander-deployments/hostsleuth/hostsleuth.service`
  - SHA-256: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`.

The root-owned approval pins the exact artifacts, destinations, `hostsleuth.service`, and approved lifecycle actions. HostSleuth gained no sudo/root mechanism of its own.

The initial staging directory/file modes produced by generic HomeCommander Normal-mode file creation were too restrictive for the intentionally capability-stripped broker (`0700` deployment directory and `0600` unit). UAT corrected only the unprivileged staging permissions to `0750` for the deployment directory and `0640` for the unit; hashes were unchanged. This is recorded as a HomeCommander shared-infrastructure follow-up, not HostSleuth privilege code.

### Managed deployment results

Confirmed through HomeCommander `deployment_action`:

- exact-hash `install` succeeded;
- binary installed at `/usr/local/bin/hostsleuth`, mode `0755`, expected SHA-256;
- unit installed at `/etc/systemd/system/hostsleuth.service`, mode `0644`, expected SHA-256;
- `enable` succeeded;
- `start` succeeded;
- service became active/running and enabled;
- managed `restart` succeeded and produced a new main PID while remaining active/running;
- `stop` succeeded;
- `disable` succeeded;
- re-`enable` and re-`start` succeeded;
- exact managed `uninstall` removed only `/usr/local/bin/hostsleuth` and `/etc/systemd/system/hostsleuth.service`;
- `/var/lib/hostsleuth` remained present after uninstall as intended;
- exact managed reinstall succeeded;
- HostSleuth was re-enabled and re-started successfully.

Final runtime state after reinstall:

- service: active/running;
- unit: enabled;
- dashboard `http://127.0.0.1:8787/`: HTTP 200;
- binary version: `v0.1.0-alpha.1`;
- snapshot `schema_version`: 1;
- Docker containers observed under the actual root system-service privilege model: **21**;
- services observed: 219;
- listeners observed: 328;
- `diagnose 127.0.0.1:22`: reachable / high confidence.

Systemd-created state directory:

- `/var/lib/hostsleuth`;
- owner/group: `root:root`;
- mode: `0700`.

This closes the previous Docker-inventory uncertainty: Docker collection works under the current system-service privilege model.

### Remaining UAT boundary

The current owner approval was created without the optional read-only `status` action, so HomeCommander `deployment_status` correctly refuses that one operation. Lifecycle/install/uninstall operations are approved and have already passed. Re-approving the same hashes with `status` included will close that shared-control-plane check.

The only HostSleuth M1 runtime item still requiring real-world proof is **reboot persistence**. HomeCommander intentionally does not expose arbitrary reboot/root execution; the owner must perform a normal authorized reboot, after which HomeCommander can verify HostSleuth returns active/running/enabled and the dashboard/API remain healthy.

## Real Debian 13 validation

Test host: `openmediavault`, Debian GNU/Linux 13 (trixie), x86_64.

A temporary Go 1.24.13 toolchain under `/tmp` was used for source-level validation only; Go was **not** installed persistently/system-wide on the host.

Confirmed working:

- `go vet ./...`;
- `go test ./...`;
- native Linux build;
- static cross-builds for Linux `amd64` and `arm64`;
- release version stamping;
- snapshot collection;
- filesystem inventory;
- systemd service inventory;
- listener inventory;
- deterministic reachable/unreachable diagnosis;
- dashboard and `/api/snapshot`, `/api/events`, `/api/diagnose` endpoints;
- periodic flight-recorder listener change events;
- schema persistence and empty event-list behavior (`[]`, not JSON `null`);
- install/source-install/uninstall shell syntax;
- systemd unit syntax;
- real managed root-level install/lifecycle/uninstall/reinstall through HomeCommander.

## GitHub CI and release validation

- Initial CI failed only `gofmt`; corrected in commit `e8185ae4e36ba0b9d0ec621369524762c54057d5`.
- Hardened M1 code commit: `69fe6c0add132e28915672f161062f60c0765635`.
- GitHub Actions CI run `34936129425`: success.
- Release automation commit: `a97932caaccf9c0e8c6d095592dc4cb049648664`.
- Release branch: `release/v0.1.0-alpha.1`.
- Release workflow run `34936275773`: success.
- Published release: `v0.1.0-alpha.1`.
- Published assets: `hostsleuth-linux-amd64`, `hostsleuth-linux-arm64`, `SHA256SUMS`.
- Published amd64 SHA-256: `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`.
- Published amd64 artifact was validated on Debian without system Go installed.
- `/releases/latest/download/hostsleuth-linux-amd64` returned the same correct version/hash.

## Important current limitations

- No authentication yet; keep the dashboard loopback-only.
- State storage is JSON/JSONL, not SQLite yet.
- Collectors intentionally degrade if optional tools (`ip`, `ss`, `systemctl`, `docker`) are absent or inaccessible.
- Diagnosis does not yet correlate nftables/UFW, reverse proxies, Docker namespaces/ports, systemd journal failures, TLS, or package/config changes.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`
First published release: `v0.1.0-alpha.1`
State-directory fix merged at: `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Exact next actions

1. Re-approve the same HostSleuth hashes with the read-only `status` action included so `deployment_status` can be validated.
2. Perform one owner-authorized reboot of `openmediavault`.
3. After reboot, verify through HomeCommander that HostSleuth is loaded, enabled, active/running; dashboard/API are healthy; Docker inventory still works; and state persists.
4. Close M1.
5. Begin the next diagnosis milestone with route/firewall evidence and bounded systemd/journal failure evidence.
