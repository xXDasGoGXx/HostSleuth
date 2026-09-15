# HostSleuth — Current Handoff

Last updated: 2026-09-15

## Current objective

Build a small, local-first Linux troubleshooting tool that answers two questions better than a conventional monitoring dashboard:

1. **What changed on this Linux host?**
2. **Why can I not reach this host/service/port?**

The product targets Debian 12/13 and Ubuntu 24.04+ and remains read-only: HostSleuth observes, records, compares, and explains. It does not repair or mutate the host.

## Product direction

HostSleuth is intentionally not a full metrics/observability platform. Its differentiator is a lightweight Linux "flight recorder" plus deterministic, evidence-backed diagnostics. Core operation must not require a cloud account, AI model, external database, or API key.

Architectural rules:

- Go 1.24+.
- One Linux binary with built-in web UI.
- Default bind `127.0.0.1:8787`; remote/LAN exposure must be explicit.
- Current persistence is local JSON snapshot + append-only JSONL events with atomic snapshot writes.
- Snapshot/event schema version is `1`.
- Deterministic diagnosis first; any future AI layer may explain evidence but must not be required.
- V0.1 remains read-only with no automatic remediation.
- Target hosts do not need Go for release-binary installation.
- HostSleuth remains independent from HomeCommander. HomeCommander is only the shared owner-approved administrative transport for privileged installation/lifecycle operations.

## Milestones

### M0 — Repository foundation

Status: **complete**

### M1 — Single-host deployable MVP

Status: **COMPLETE — implementation, Debian runtime validation, managed install/update UAT, and reboot persistence all passed**

M1 provides:

- versioned snapshot/event schema;
- host/OS/kernel/CPU/memory/uptime inventory;
- filesystem/capacity inventory;
- interfaces/addresses/routes;
- listening sockets;
- systemd service inventory;
- Docker inventory when available/readable;
- atomic snapshots and append-only change history;
- semantic service/container/listener diffing;
- deterministic `diagnose host:port`;
- CLI: `snapshot`, `diagnose`, `events`, `serve`, `version`;
- built-in loopback-only dashboard/API;
- CI for format/vet/tests/build;
- systemd packaging and install/uninstall scripts;
- Linux amd64/arm64 release automation.

## M1 validation summary

### Published alpha

First published release remains **`v0.1.0-alpha.1`**.

Published amd64 SHA-256:

`8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`

It was validated on Debian 13 without system Go installed.

Important: no new public release was created during the final M1 fixes. Public release promotion remains owner-controlled and requires explicit approval. The current live validated HostSleuth is newer than `v0.1.0-alpha.1`.

### Systemd state-directory fix

The original installer created `/var/lib/hostsleuth` separately, which did not fit HomeCommander exact-artifact managed deployment. HostSleuth PR #1 added:

- `StateDirectory=hostsleuth`
- `StateDirectoryMode=0700`

PR #1 CI run `34987407246` passed and merged to `main` at:

`20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`

Systemd now safely creates `/var/lib/hostsleuth` as `root:root 0700` without HostSleuth or HomeCommander gaining generic privileged directory-write capability.

## HomeCommander Managed Administrative Deployment UAT — COMPLETE

HostSleuth was the first real consumer of HomeCommander's reviewed exact-hash Managed Administrative Deployment mechanism on live host `openmediavault`.

Initial approved artifacts:

- executable: published `v0.1.0-alpha.1`
- executable SHA: `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`
- unit SHA: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`
- executable destination: `/usr/local/bin/hostsleuth`
- unit destination: `/etc/systemd/system/hostsleuth.service`

Validated through HomeCommander:

- exact-hash first install;
- enable/start;
- status;
- restart;
- stop/disable;
- re-enable/re-start;
- exact uninstall;
- exact reinstall;
- preservation of `/var/lib/hostsleuth` across uninstall;
- exact-hash verification of installed files;
- audit trail of privileged actions.

HostSleuth itself gained no sudo/root shell or HomeCommander-specific privileged code.

### Real reboot persistence — PASSED

The owner performed a real reboot of `openmediavault`.

Post-reboot validation showed:

- HomeCommander recovered and remained callable;
- HostSleuth remained recorded as installed;
- approved executable/unit hashes still matched;
- `hostsleuth.service` returned loaded, enabled, active/running;
- post-reboot PID was `1360` at first check;
- dashboard HTTP 200;
- `/var/lib/hostsleuth` remained `root:root 0700`;
- Docker inventory remained available with 21 containers;
- `diagnose 127.0.0.1:22` remained reachable/high confidence.

Normal-mode HomeCommander intentionally cannot enumerate the root-only state directory. State/runtime persistence was verified through HostSleuth's API instead of weakening permissions.

## Docker event-noise correctness fix — PASSED

The reboot/Docker UAT exposed a real HostSleuth bug: raw Docker `Status` includes elapsed uptime, so verbatim diffing produced changes such as `Up 5 minutes -> Up 6 minutes`.

Observed before the fix:

- API window contained 100 events;
- 99/100 were container events;
- most were uptime-only churn that would bury meaningful incident history.

PR #2 changed only event comparison semantics:

- full raw Docker status remains in snapshots/UI;
- diffing uses normalized semantic state;
- uptime-only and restart-timer-only changes are ignored;
- meaningful health, exit, restart, paused, dead, appearance/removal transitions remain detectable;
- warning severity remains for meaningful unhealthy/failure states;
- no schema change;
- no privilege change.

PR #2 CI run `34991719610` passed format, vet, tests, and build. PR #2 merged at:

`d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`

Regression coverage includes uptime churn, restart-timer churn, health transitions, and exits.

## First real managed update — PASSED

The merged PR #2 source was built unprivileged with a checksum-verified temporary Go 1.24.13 toolchain. Go was not installed system-wide.

Managed-update candidate:

- source commit: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`
- version: `0.1.0-dev+d702804`
- SHA-256: `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1`
- approved source: `/srv/homecommander-deployments/hostsleuth/hostsleuth.candidate`
- unit source: `/srv/homecommander-deployments/hostsleuth/hostsleuth.service`
- unchanged unit SHA: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`

The owner approved the new exact candidate hash. HomeCommander then performed a real managed update:

- pre-update `deployment_status` correctly showed the old installed binary did not match the newly approved candidate hash;
- managed `install` verified the previously recorded installation before replacing it;
- installed executable became SHA `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1`;
- unchanged unit remained SHA `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`;
- managed restart changed PID from `1360` to `31172` and remained active/running;
- final `deployment_status` reports `recordedInstalled=true`, both files present, and both exact hashes matching approval;
- unit remains enabled;
- service remains active/running;
- installed version reports `0.1.0-dev+d702804`.

Final runtime validation after update:

- dashboard HTTP 200;
- schema version 1;
- Docker inventory 21 containers;
- service inventory 219;
- listener inventory 326.

### Live event-noise proof

Historical pre-fix churn was deliberately not erased.

Fresh events were compared only after the updated service restart baseline.

Capture interval 1:

- new events: 1;
- new container events: **0**;
- only event: expected HostSleuth listener `127.0.0.1:8787` reappeared after restart.

Capture interval 2:

- new events: 2;
- new container events: **0**;
- both were real systemd service state changes (`local-working-file.service` and `playwright-mcpo.service`).

This proves the fix suppresses Docker uptime churn while HostSleuth continues recording real changes.

## HomeCommander shared-infrastructure finding

First-time UAT exposed a narrow HomeCommander staging ergonomics issue:

- generic Normal-mode directory creation produced deployment directory mode `0700`;
- generic Normal-mode file creation produced the staged unit mode `0600`;
- the capability-stripped root broker relies on the `homecommander` group for staging traversal/read access;
- first install therefore failed closed before promotion;
- changing only unprivileged staging permissions to directory `0750` and unit `0640` made the exact same approved bytes broker-readable.

This belongs in HomeCommander shared infrastructure, not in HostSleuth. Do not solve it with broader broker capabilities, generic sudo/root shell, or weak ownership.

## M2 progress — route-path evidence merged

PR #3 begins M2 by adding bounded route-path evidence to deterministic diagnosis.

Merged behavior:

- after DNS resolution, HostSleuth runs a bounded read-only `ip route get <resolved-ip>` lookup;
- a successful kernel lookup is recorded as `route: pass` with compact route evidence;
- kernel-reported unreachable/no-route results become `route: fail`;
- command absence, sandbox restrictions, or unavailable netlink access become `route: unknown` rather than a false routing failure;
- a confirmed route failure plus remote TCP failure raises the routing conclusion to high confidence;
- existing reachable and local-listener conclusions remain authoritative when those stronger signals are present.

Validation:

- PR #3 CI run `34996674672` passed format, vet, tests, and build;
- PR #3 merged to `main` at `82ffe4cd81e92b8176a8f09f5e3dc2e857057475`;
- Debian Normal-mode validation reproduced a real HomeCommander netlink restriction (`Cannot open netlink socket: Address family not supported by protocol`);
- HostSleuth correctly reported `route: unknown` while TCP/22 still produced `target is reachable` with high confidence.

The live managed system service has **not** been updated to this M2 source yet. It remains on the validated M1 build `0.1.0-dev+d702804`. Batch the next meaningful M2 diagnostics before requesting another exact-hash managed deployment approval unless live root-context validation becomes necessary sooner.

## Current live state

Host: `openmediavault`

Current installed HostSleuth:

- version: `0.1.0-dev+d702804`
- executable SHA: `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1`
- unit SHA: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`
- service: loaded, enabled, active/running
- current validated PID: `31172`
- dashboard: HTTP 200 on `127.0.0.1:8787`
- state: `/var/lib/hostsleuth`, `root:root 0700`
- Docker inventory: 21 containers

The root-owned HomeCommander approval currently points to staged source `hostsleuth.candidate`. Keep that staged candidate available while this approval is current so approved reinstall/update behavior remains reproducible.

## Important limitations

- No authentication yet; keep dashboard loopback-only.
- Storage is JSON/JSONL, not SQLite yet.
- Collectors degrade if optional tools are unavailable/inaccessible.
- Diagnosis does not yet correlate firewall rules, reverse proxies, Docker port/network paths, bounded journal failure evidence, TLS, or package/config changes.
- Historical pre-fix container uptime-noise events remain in the append-only history and will age out of the API window naturally; they were not deleted merely to make validation look cleaner.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`
First published release: `v0.1.0-alpha.1`
Systemd state-directory fix: `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`
Docker semantic-event fix: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`
M2 route-path evidence: `82ffe4cd81e92b8176a8f09f5e3dc2e857057475`

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Next sequence

M1 is closed and M2 route-path evidence is merged. Continue deterministic diagnosis in this order:

1. firewall evidence with safe bounded reads of active policy;
2. bounded systemd/journal failure evidence;
3. Docker port/bind/network correlation;
4. define final evidence ordering/confidence behavior and regression scenarios;
5. then reverse-proxy/TLS awareness and the planned SQLite migration as appropriate.

Publishing a new public HostSleuth release containing the final M1 fixes is a separate owner-controlled action and must not occur without explicit approval.
