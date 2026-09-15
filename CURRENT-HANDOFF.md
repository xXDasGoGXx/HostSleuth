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

## M2 progress — route, firewall, systemd, Docker, and evidence policy merged

PR #3 adds bounded route-path evidence after DNS resolution:

- successful `ip route get <resolved-ip>` is `route: pass` with compact evidence;
- kernel no-route/unreachable is `route: fail`;
- command/netlink restrictions are `route: unknown`, not false failures;
- route failure can raise a remote TCP-failure conclusion to high confidence.

Validation: PR #3 CI `34996674672` passed and merged at `82ffe4cd81e92b8176a8f09f5e3dc2e857057475`. Debian Normal-mode validation reproduced the real netlink restriction while reachable TCP/22 still returned reachable/high confidence.

PR #4 adds bounded read-only nftables evidence after TCP failure:

- runs `nft -nn list ruleset` only when TCP fails;
- two-second timeout, 64 KiB captured-output cap;
- finds `nft` through PATH or standard `/usr/sbin`/`/sbin` locations;
- base-policy/direct-port matches are candidate evidence, not proven verdicts;
- command/netlink/permission failure degrades to `firewall: unknown`;
- successful TCP skips firewall collection.

Validation: PR #4 CI `34997617382` passed and merged at `2fe64640093b258b3c52b148fc2e32507eeae584`. Debian Normal-mode validation reproduced the nft netlink restriction while preserving the correct no-listener/high-confidence conclusion.

PR #5 adds bounded systemd/journal failure evidence only for local TCP failures with no listener:

- failed candidates come from the existing HostSleuth snapshot;
- maximum three failed units are examined;
- each current-boot journal read is capped at six lines, 8 KiB, and two seconds;
- common password/secret/token/API-key/authorization/cookie assignments and bearer tokens are redacted before evidence is returned;
- no failed units produces `systemd-failures: pass` and does not invoke journalctl;
- failed-unit evidence is explicitly candidate evidence, not proof that a unit owns the port;
- systemd/journal evidence is skipped when a listener already exists.

Validation: PR #5 CI `34998204162` passed and merged at `37c97227a829fc341f943d0b4731242ee2a39650`. A Debian Normal-mode snapshot collected 219 services with zero failed units; closed TCP/65534 added a passing systemd-failure check while preserving the correct high-confidence no-listener conclusion. Normal-mode journal reads are permission-restricted and this is handled as unavailable evidence.

PR #6 adds Docker port/bind/network correlation and fixes listener matching to respect the requested local address:

- parses wildcard, exact-address, IPv4/IPv6, and internal-only Docker port strings from the existing snapshot;
- distinguishes a host-port publish on the requested address, the same host port bound only to another address, internal-only container exposure, and no matching Docker exposure;
- retains Docker network names as an optional backward-compatible `networks` snapshot field and includes them in matching evidence when available;
- local listener matching is now TCP-specific, target-address-aware, and family-aware instead of matching only the port number;
- diagnosis remains read-only and does not invoke Docker during a diagnosis.

Validation: PR #6 CI `34999319318` passed and merged at `bc939ae0c6339b355bd33916629a93f4a07803c9`. Using the live root-collected snapshot without changing containers, `127.0.0.1:8789` correctly identified Chaptarr as published only on `192.168.2.181:8789`, `127.0.0.1:8192` identified FlareSolverr `8192/tcp` as internal-only, and `192.168.2.181:8789` remained reachable/high confidence.

PR #7 defines the deterministic evidence precedence and confidence policy:

1. successful TCP is definitive and stops secondary evidence collection;
2. for local failures, target-aware listener and Docker bind evidence outrank firewall/systemd candidates;
3. for remote failures, confirmed kernel no-route evidence outranks firewall inspection;
4. unavailable optional evidence remains neutral and cannot weaken a stronger conclusion;
5. contradictory snapshot-vs-current evidence lowers confidence instead of pretending certainty.

Scenario-level regressions now lock behavior for reachable targets, listener-plus-firewall candidates, failed-systemd candidates, Docker bind mismatch, contradictory Docker publication, remote no-route, and unavailable optional evidence.

Validation: PR #7 CI `34999939216` passed on exact PR head `8e2b675c1a3cf7084e19e6b966d34ba9327f5a2f`; PR #7 merged at `3352a7e8407eae855f4a88550cfaaf867f86ddf1`. Main CI run `35000071547` also passed on that merged commit.

### M2 managed-deployment candidate — staged, not yet approved

The exact merged M2 source was rebuilt and revalidated on Debian 13:

- source: `3352a7e8407eae855f4a88550cfaaf867f86ddf1`;
- version: `0.1.0-dev+3352a7e`;
- staged path: `/srv/homecommander-deployments/hostsleuth/hostsleuth.m2-candidate`;
- candidate SHA-256: `5d595d9db476f9cc030d052df53f6139057fdc6f74f6e440e9f833e5cee201aa`;
- staged mode: `0755`;
- unchanged unit SHA-256: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`.

The existing approved M1 candidate was deliberately **not overwritten**:

- `/srv/homecommander-deployments/hostsleuth/hostsleuth.candidate`;
- SHA-256 `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1`.

The staged M2 binary itself was replayed against the live root-collected snapshot and reproduced the expected Chaptarr bind mismatch, FlareSolverr internal-only exposure, and reachable SSH conclusions. No firewall, service, or container state was mutated for these tests.

The live managed system service has **not** been updated to M2 yet. It remains on the validated M1 build `0.1.0-dev+d702804`. Do not call `deployment_action install` until owner/root changes the managed deployment approval to the new M2 candidate path/hash.

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
- Diagnosis now correlates route, nftables candidates, local listeners, Docker port/bind/network evidence, and bounded failed-systemd candidates, but does not yet understand reverse proxies, TLS, or package/config changes.
- Firewall and failed-unit correlations are deliberately conservative candidate evidence rather than causal verdicts.
- Full M2 root-context validation is still pending because the live managed service remains on M1 until the new exact candidate hash is owner-approved.
- Journal evidence has a narrow first-pass credential redactor; general redaction rules remain a security backlog item.
- Historical pre-fix container uptime-noise events remain in the append-only history and will age out of the API window naturally; they were not deleted merely to make validation look cleaner.

## Source of truth

Repository: `xXDasGoGXx/HostSleuth`
Default branch: `main`
First published release: `v0.1.0-alpha.1`
Systemd state-directory fix: `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`
Docker semantic-event fix: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`
M2 route-path evidence: `82ffe4cd81e92b8176a8f09f5e3dc2e857057475`
M2 firewall evidence: `2fe64640093b258b3c52b148fc2e32507eeae584`
M2 systemd/journal evidence: `37c97227a829fc341f943d0b4731242ee2a39650`
M2 Docker port/bind/network correlation: `bc939ae0c6339b355bd33916629a93f4a07803c9`
M2 evidence precedence/confidence policy: `3352a7e8407eae855f4a88550cfaaf867f86ddf1`

Keep this file and `TO-DO.md` synchronized with meaningful progress.

## Next sequence

M1 is closed and the planned M2 diagnosis code is merged, CI-green, and staged as an exact candidate. Continue in this order:

1. owner/root re-approves managed deployment `hostsleuth` to use `/srv/homecommander-deployments/hostsleuth/hostsleuth.m2-candidate` with SHA-256 `5d595d9db476f9cc030d052df53f6139057fdc6f74f6e440e9f833e5cee201aa`; keep the unchanged unit approval SHA `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`;
2. re-check `deployment_status` and only then use `deployment_action install` followed by `restart`;
3. validate route, nftables, journal, Docker network-name collection, target-aware listener behavior, dashboard/API, state persistence, and change recording under the real root system-service context without manufacturing destructive failures;
4. if validation passes, mark M2 complete and decide separately whether to prepare a new public alpha;
5. after M2, continue with reverse-proxy/TLS awareness and the planned SQLite migration as appropriate.

Publishing a new public HostSleuth release remains a separate owner-controlled action and must not occur without explicit approval.
