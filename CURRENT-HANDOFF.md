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

Status: **reboot persistence passed; one Docker event-noise correctness update is staged and awaiting owner exact-hash approval before M1 closure**

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

### Initial approved UAT artifacts

Owner-approved initial artifacts:

- `/srv/homecommander-deployments/hostsleuth/hostsleuth`
  - published `v0.1.0-alpha.1` Linux amd64 release;
  - SHA-256: `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`.
- `/srv/homecommander-deployments/hostsleuth/hostsleuth.service`
  - SHA-256: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`.

The root-owned approval pins the exact artifacts, destinations, `hostsleuth.service`, and approved lifecycle actions. HostSleuth gained no sudo/root mechanism of its own.

The owner re-approved the same initial hashes using the helper's normal default action set so read-only `status` is also approved. Current approved actions are install, uninstall, enable, disable, start, stop, status, and restart.

The initial staging directory/file modes produced by generic HomeCommander Normal-mode file creation were too restrictive for the intentionally capability-stripped broker (`0700` deployment directory and `0600` unit). UAT corrected only the unprivileged staging permissions to `0750` for the deployment directory and `0640` for the unit; hashes were unchanged. This is recorded as a HomeCommander shared-infrastructure follow-up, not HostSleuth privilege code.

### Managed deployment lifecycle results

Confirmed through HomeCommander:

- exact-hash `install` succeeded;
- binary installed at `/usr/local/bin/hostsleuth`, mode `0755`, expected SHA-256;
- unit installed at `/etc/systemd/system/hostsleuth.service`, mode `0644`, expected SHA-256;
- enable/start succeeded;
- managed restart succeeded and produced a new main PID while remaining active/running;
- stop/disable succeeded;
- re-enable/re-start succeeded;
- exact managed uninstall removed only `/usr/local/bin/hostsleuth` and `/etc/systemd/system/hostsleuth.service`;
- `/var/lib/hostsleuth` remained present after uninstall as intended;
- exact managed reinstall succeeded;
- HostSleuth was re-enabled and re-started successfully;
- corrected `deployment_status` succeeded and verified both installed artifact hashes exactly against the root-owned approval.

## Reboot persistence — PASSED

The owner performed a real reboot of `openmediavault` on 2026-09-15.

Post-reboot verification through HomeCommander:

- host uptime at first check: about 635 seconds, confirming a real reboot;
- HomeCommander recovered and remained callable;
- `deployment_status` succeeded;
- `recordedInstalled=true`;
- installed executable SHA still exactly matches the approved `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`;
- installed unit SHA still exactly matches `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`;
- `hostsleuth.service` loaded automatically;
- unit remained enabled;
- service returned active/running with post-reboot PID `1360`;
- dashboard returned HTTP 200;
- installed version remained `v0.1.0-alpha.1`;
- `/var/lib/hostsleuth` remained `root:root 0700`;
- API snapshot schema version remained 1;
- Docker inventory remained available with 21 containers;
- post-reboot snapshot also observed 216 services, 326 listeners, 100 filesystems, 46 interfaces, and 12 routes;
- `diagnose 127.0.0.1:22` remained reachable/high confidence.

Normal-mode HomeCommander cannot directly enumerate `/var/lib/hostsleuth` because the directory is intentionally root-only `0700`; persistence/runtime validation therefore uses HostSleuth's own API rather than weakening permissions.

## Docker event-noise bug found during post-reboot validation

The first Docker-enabled reboot test exposed a real M1 correctness issue: Docker's raw status string includes elapsed uptime (`Up 5 minutes`, `Up 6 minutes`, etc.), and HostSleuth was comparing that string verbatim when diffing containers.

Observed impact:

- `/api/events` returned its 100-event window;
- 99 of those 100 events were container-change events;
- the repeated changes were primarily uptime-only transitions such as `Up 5 minutes -> Up 6 minutes`;
- this would rapidly bury meaningful incident history, so M1 is not being declared closed with that behavior.

Root cause was confirmed in `internal/core/diff.go`: `containerMap` mapped container name directly to raw `c.Status`.

### PR #2 — fixed and merged

Fix strategy:

- keep the full raw Docker `Status` in snapshots/UI;
- normalize only the semantic state used for event diffing;
- ignore uptime-only running changes and restart-timer-only changes;
- preserve meaningful transitions such as healthy -> unhealthy, running -> exited, restarting, paused, dead, and removal states;
- mark unhealthy/exited/restarting/paused/dead transitions as warnings where appropriate;
- no snapshot/event schema change;
- no privilege change.

Validation:

- branch: `fix/container-event-uptime-churn`;
- PR #2: `Suppress Docker uptime-only event churn`;
- regression tests cover uptime-only churn, restart-timer churn, health changes, and exit transitions;
- GitHub Actions CI run `34991719610`: format, vet, tests, and build all passed;
- merged to `main` at `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`.

### Managed-update candidate staged

No new public release was created. The existing published release remains `v0.1.0-alpha.1`.

A temporary Go 1.24.13 toolchain was downloaded under `/tmp`, its official Linux amd64 SHA-256 was verified, and it was not installed system-wide. Exact merged source commit `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e` was cloned and locally tested successfully.

Candidate build:

- source commit: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`;
- version stamp: `0.1.0-dev+d702804`;
- SHA-256: `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1`;
- staged at `/srv/homecommander-deployments/hostsleuth/hostsleuth.candidate`;
- mode `0755`;
- original staged published release `/srv/homecommander-deployments/hostsleuth/hostsleuth` remains untouched at its original hash;
- staged unit remains unchanged at SHA-256 `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`.

The candidate is intentionally not installable through HomeCommander until the owner re-approves the new exact binary hash. This is also the first real managed **update** UAT opportunity.

## Real Debian 13 validation

Test host: `openmediavault`, Debian GNU/Linux 13 (trixie), x86_64.

A temporary Go 1.24.13 toolchain under `/tmp` has been used for source-level validation only; Go is not installed persistently/system-wide on the host.

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
- real managed root-level install/lifecycle/uninstall/reinstall through HomeCommander;
- exact managed status verification through HomeCommander;
- real reboot persistence of the managed installation.

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
- PR #2 Docker event-normalization CI run `34991719610`: success.

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
Docker event-noise fix merged at: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Exact next actions

1. Owner re-approves the staged `hostsleuth.candidate` exact hash, with the same unit and normal default action set.
2. Use HomeCommander managed `install` to exercise a real exact-hash update from `v0.1.0-alpha.1` to candidate `0.1.0-dev+d702804`, then managed restart.
3. Verify `deployment_status` reports the new approved/installed binary hash and unchanged unit hash.
4. Verify dashboard/API/Docker inventory remain healthy.
5. Observe at least two normal 60-second capture intervals and confirm Docker uptime text no longer generates container-change event churn while meaningful state changes remain detectable by regression tests.
6. Close M1 and the first real HomeCommander Managed Administrative Deployment UAT if green.
7. Begin the next diagnosis milestone with route/firewall evidence and bounded systemd/journal failure evidence.
