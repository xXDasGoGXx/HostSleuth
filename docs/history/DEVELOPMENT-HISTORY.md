# HostSleuth — Development History

This file preserves milestone history and validation details that do not belong in the active handoff or backlog.

## M0 — Repository foundation

Completed:

- Go module and CLI skeleton.
- Snapshot, event, diagnosis, storage, and diff structures.
- Embedded dashboard/API.
- CI for formatting, vet, tests, and build.
- systemd packaging and install/uninstall scripts.
- Linux amd64/arm64 release automation.

Initial formatting failure was corrected with `gofmt`; the quality gate remained in place.

## First public alpha

Published release: `v0.1.0-alpha.1`.

Validated on Debian 13 without system Go installed.

Published amd64 SHA-256:

`8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`

## M1 — Single-host deployable MVP

M1 established the core product:

- host/OS/kernel/CPU/memory/uptime inventory;
- filesystem/capacity inventory;
- interfaces, routes, and listeners;
- systemd service inventory;
- Docker inventory when readable;
- local JSON snapshot and append-only JSONL events;
- semantic service/container/listener change events;
- deterministic `diagnose host:port`;
- local dashboard/API;
- release/source installation and systemd service.

### Systemd state-directory fix

PR #1 added systemd-owned state-directory creation:

- `StateDirectory=hostsleuth`
- `StateDirectoryMode=0700`

Merged at:

`20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`

### Managed-deployment UAT

HostSleuth became the first real consumer of an owner-approved managed administrative deployment mechanism used in the development environment.

Validated on a Debian 13 host:

- exact-hash install;
- enable/start/status/restart;
- stop/disable and re-enable/re-start;
- uninstall/reinstall;
- preservation of `/var/lib/hostsleuth`;
- exact installed hash verification;
- reboot persistence;
- root service Docker inventory.

That deployment mechanism is external infrastructure only. It is not part of HostSleuth's product architecture.

### Docker event-noise fix

Real reboot testing exposed uptime-only Docker churn (`Up N minutes`). PR #2 changed diff semantics so raw status remains visible while uptime progression is ignored for events.

Merged at:

`d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`

The updated live build produced zero Docker uptime-noise events across fresh capture intervals while still recording real systemd changes.

M1 was then considered complete.

## M2 — Deeper deterministic diagnosis

M2 expanded diagnosis while preserving the read-only product boundary.

### Route-path evidence — PR #3

Added bounded `ip route get` evidence.

- real no-route/unreachable becomes a route failure;
- command/netlink restrictions remain `unknown`;
- missing route evidence cannot create a false failure.

Merged at:

`82ffe4cd81e92b8176a8f09f5e3dc2e857057475`

### nftables evidence — PR #4

Added bounded read-only nftables collection after TCP failure.

- two-second timeout;
- 64 KiB output cap;
- candidate evidence only, not a claimed causal verdict;
- permission/netlink restrictions degrade to `unknown`.

Merged at:

`2fe64640093b258b3c52b148fc2e32507eeae584`

### systemd/journal evidence — PR #5

Added bounded failed-unit candidate evidence for local failures with no listener.

- maximum three failed units;
- bounded current-boot journal excerpts;
- basic credential/token redaction;
- no failed units means no journal command is run.

Merged at:

`37c97227a829fc341f943d0b4731242ee2a39650`

### Docker port/bind/network correlation — PR #6

Added:

- wildcard and exact-address Docker publication parsing;
- internal-only container-port detection;
- Docker network names in snapshot evidence;
- target-aware TCP listener matching.

This fixed a correctness bug where a listener bound to one local address could previously be treated as evidence for a different local address simply because the port number matched.

Merged at:

`bc939ae0c6339b355bd33916629a93f4a07803c9`

### Evidence precedence and confidence — PR #7

Defined the deterministic rule order:

1. successful TCP is definitive;
2. local listener/Docker bind evidence outranks firewall/systemd candidates;
3. confirmed remote no-route outranks firewall inspection;
4. unavailable optional evidence stays neutral;
5. contradictory snapshot/current evidence lowers confidence.

Merged at:

`3352a7e8407eae855f4a88550cfaaf867f86ddf1`

### M2 live deployment validation

The owner-approved M2 build was installed and restarted successfully on Debian 13.

Real root-context validation confirmed:

- route lookup succeeds as root;
- nftables evidence is readable;
- Docker network names are captured;
- reachable local services produce high-confidence reachable results;
- closed local ports produce high-confidence no-listener results;
- loopback diagnosis does not misattribute a listener bound only to another local address;
- container-internal ports are distinguished from host-published ports;
- correctly published host ports are reachable on the address where they are actually bound;
- dashboard returns HTTP 200;
- normal change recording continues.

The live host had zero failed systemd services at validation time, so journal excerpt collection was not artificially triggered by breaking a service. The behavior remains covered by tests.

M2 is complete.

## Post-M2 experiments and cleanup

After M2, two experimental M3 slices were briefly merged into `main`:

- PR #8 — Nginx Proxy Manager awareness (`df7e926b0ad07c0224004c644d5c4a1d667df547`);
- PR #9 — TLS diagnostics (`5a549d977e2cc853338b728c1814732467c828d1`).

Those experiments were never deployed as the validated live HostSleuth build. During the subsequent project-clarity cleanup, they were deliberately removed from the active product source rather than allowed to blur the validated M2 boundary.

Cleanup commit:

`8dda0368e2f1e8df46325713fe1dac68df0fc9f1`

After that cleanup, executable source on `main` matches the validated M2 code at `3352a7e8407eae855f4a88550cfaaf867f86ddf1`; only current documentation/history differs. The M3 commits and branches remain available as historical/experimental work and are not active product state.

## Historical source references

- First published alpha: `v0.1.0-alpha.1`
- systemd state-directory fix: `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`
- Docker semantic-event fix: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`
- M2 route evidence: `82ffe4cd81e92b8176a8f09f5e3dc2e857057475`
- M2 firewall evidence: `2fe64640093b258b3c52b148fc2e32507eeae584`
- M2 systemd/journal evidence: `37c97227a829fc341f943d0b4731242ee2a39650`
- M2 Docker correlation: `bc939ae0c6339b355bd33916629a93f4a07803c9`
- M2 evidence policy: `3352a7e8407eae855f4a88550cfaaf867f86ddf1`

Future milestone history belongs here rather than in `CURRENT-HANDOFF.md` or `TO-DO.md`.
