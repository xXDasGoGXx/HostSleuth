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

Immediately after that cleanup, executable source on `main` matched the validated M2 code at `3352a7e8407eae855f4a88550cfaaf867f86ddf1`; only current documentation/history differed. The M3 experiment commits and branches remained available as historical work rather than active product state.

## M3.1 — Product Web UI

PR #11 replaced the original inline dashboard with a self-contained modern Web UI while deliberately leaving diagnostic semantics unchanged.

Added:

- responsive Overview / Diagnose / Changes / Host views;
- human-readable diagnosis presentation backed by the existing deterministic checks and evidence;
- recent-change timeline;
- compact host, interface, and filesystem views;
- loading, empty, and error states;
- mobile-friendly layout;
- embedded HTML/CSS/vanilla-JavaScript assets with no frontend framework or external runtime dependency.

The existing `/api/snapshot`, `/api/events`, and `/api/diagnose` endpoints remained the source of truth.

PR #11 passed formatting, `go vet`, tests, and build before merge.

Merged at:

`bed52bba90370eac53ae29137bc36bfbcf62d216`

## M3.2 — Usability

PR #13 polished the product without adding another diagnostic subsystem.

Added or improved:

- simpler user-facing wording;
- clear first-run/baseline behavior;
- friendly diagnosis titles while retaining exact deterministic conclusions and evidence;
- version plus VCS build revision visibility when available;
- newest-first event display and correct newest-five Overview behavior;
- recommended native/systemd install path at the top of the README;
- safe SSH-tunnel guidance for remote access to the loopback-only Web UI;
- JavaScript syntax validation in CI.

PR #13 passed formatting, `go vet`, tests, JavaScript syntax, and build before merge.

Merged at:

`d250019902d676f67d44b74cc122db3f40ad7e67`

## M3.3 — Supported Docker deployment

PR #14 added one Docker deployment path without turning HostSleuth into a multi-mode configuration product.

Added:

- multi-stage `Dockerfile`;
- one recommended `compose.yaml`;
- persistent `hostsleuth-data` volume;
- Linux host network/PID/UTS namespace sharing for truthful host network/listener/hostname evidence;
- narrow read-only host OS-release mount;
- Docker socket access for container inventory;
- all Linux capabilities dropped;
- `no-new-privileges` and read-only container root filesystem;
- explicit Docker deployment mode in snapshots;
- deliberate omission of host filesystem/systemd inventory when Docker isolation prevents truthful collection;
- systemd diagnostic evidence reported as `unknown` in Docker mode rather than a false clean result;
- Web UI and CLI labels for reduced Docker visibility;
- focused Docker-mode tests;
- CI validation for Compose plus Linux amd64 and arm64 container image builds.

Security documentation explicitly records that Docker socket access remains a powerful host capability even when its bind path is read-only.

Final PR #14 validation passed:

- Go formatting;
- `go vet`;
- Go tests;
- Web UI JavaScript syntax;
- native Go build;
- Compose configuration;
- Linux arm64 image build;
- Linux amd64 image build.

No image was published and the known-good native deployment was not replaced.

Merged at:

`ccaad3ef1df16e65887d1f19da44a118a6d0bb2a`

M3 Product Experience is complete. Future capability work returns to the collective product-decision list rather than continuing automatically.

## Historical source references

- First published alpha: `v0.1.0-alpha.1`
- systemd state-directory fix: `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`
- Docker semantic-event fix: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`
- M2 route evidence: `82ffe4cd81e92b8176a8f09f5e3dc2e857057475`
- M2 firewall evidence: `2fe64640093b258b3c52b148fc2e32507eeae584`
- M2 systemd/journal evidence: `37c97227a829fc341f943d0b4731242ee2a39650`
- M2 Docker correlation: `bc939ae0c6339b355bd33916629a93f4a07803c9`
- M2 evidence policy: `3352a7e8407eae855f4a88550cfaaf867f86ddf1`
- M3.1 Web UI: `bed52bba90370eac53ae29137bc36bfbcf62d216`
- M3.2 usability: `d250019902d676f67d44b74cc122db3f40ad7e67`
- M3.3 Docker deployment: `ccaad3ef1df16e65887d1f19da44a118a6d0bb2a`

Future milestone history belongs here rather than in `CURRENT-HANDOFF.md` or `TO-DO.md`.
