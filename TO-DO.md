# HostSleuth — TO-DO

## Completed milestone — M1: Single-host deployable MVP

- [x] Define versioned snapshot and event schemas (`schema_version: 1`).
- [x] Collect host identity, OS, kernel, uptime, CPU count, memory, filesystems/capacity, interfaces, addresses, routes, listeners, systemd service state, and Docker state when available/readable.
- [x] Persist snapshots atomically under a configurable state directory; system service target is `/var/lib/hostsleuth`.
- [x] Diff snapshots into concise human-readable service/container/listener change events.
- [x] Implement deterministic `diagnose host:port` checks with evidence.
- [x] Add local HTTP API and embedded web UI.
- [x] Add CLI subcommands: `serve`, `snapshot`, `diagnose`, `events`, `version`.
- [x] Add tests for diffing, diagnostics, schema, and storage behavior.
- [x] Add GitHub Actions CI definition (`gofmt`, `go vet`, `go test`, build).
- [x] Add systemd unit and release/source install + uninstall scripts.
- [x] Exercise the core runtime on a real Debian 13 host.
- [x] Add release automation for static Linux `amd64`/`arm64` binaries with checksums and version stamping.
- [x] Publish and validate `v0.1.0-alpha.1` on Debian without system Go installed.
- [x] Merge the systemd-owned state-directory fix.
  - PR #1 CI `34987407246` passed.
  - merged at `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`.
  - unit uses `StateDirectory=hostsleuth` and `StateDirectoryMode=0700`.
- [x] Use HostSleuth as the first real HomeCommander Managed Administrative Deployment UAT consumer.
  - exact-hash first install passed;
  - enable/start/status/restart passed;
  - stop/disable and re-enable/re-start passed;
  - exact uninstall/reinstall passed;
  - root-only state directory persisted;
  - audit trail captured privileged actions.
- [x] Validate Docker inventory under the final system-service privilege model.
  - 21 containers observed.
- [x] Validate installed CLI diagnosis.
  - `diagnose 127.0.0.1:22` returned reachable/high confidence.
- [x] Validate real reboot persistence on `openmediavault`.
  - HomeCommander recovered;
  - HostSleuth hashes remained valid;
  - service returned loaded, enabled, active/running;
  - dashboard HTTP 200;
  - Docker inventory remained available;
  - state directory remained `root:root 0700`.
- [x] Fix Docker uptime-only event churn found by reboot/Docker UAT.
  - pre-fix 99/100 recent events were container events caused primarily by raw `Up N minutes` progression;
  - PR #2 normalizes semantic container state only for diffing;
  - raw status remains in snapshot/UI;
  - real health/exit/restart/paused/dead/removal transitions remain meaningful;
  - CI `34991719610` passed;
  - merged at `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`.
- [x] Build and stage the merged event fix as the first real managed-update candidate.
  - version `0.1.0-dev+d702804`;
  - candidate SHA `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1`;
  - staged at `/srv/homecommander-deployments/hostsleuth/hostsleuth.candidate`;
  - unchanged unit SHA `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`.
- [x] Owner/root re-approved the new exact candidate hash with the unchanged unit/default actions.
- [x] Exercise a real HomeCommander managed update.
  - managed install replaced the old recorded executable only after prior-state verification;
  - managed restart changed PID `1360 -> 31172`;
  - final `deployment_status` shows candidate executable hash and unchanged unit hash both matching approval;
  - service remains enabled, active, running;
  - installed version `0.1.0-dev+d702804`.
- [x] Validate the Docker event-noise fix under the real root system-service model.
  - dashboard HTTP 200;
  - Docker inventory still 21 containers;
  - first fresh 60-second interval: zero new container events; only expected listener reappearance after restart;
  - second fresh 60-second interval: zero new container events; two real systemd service-state changes were still recorded;
  - historical pre-fix events were not erased.
- [x] Close M1 and the first real HomeCommander Managed Administrative Deployment UAT.

## Completed milestone — M2: Deeper deterministic diagnosis

Start with evidence that helps explain why a service/port is unreachable without turning HostSleuth into an automatic repair tool.

- [x] Add route-path evidence to diagnosis.
  - PR #3 adds bounded `ip route get` evidence after DNS resolution;
  - kernel-reported unreachable routes are surfaced as failures;
  - command/netlink restrictions are `unknown`, not false route failures;
  - PR #3 CI run `34996674672` passed format, vet, tests, and build;
  - merged at `82ffe4cd81e92b8176a8f09f5e3dc2e857057475`;
  - Normal-mode Debian validation confirmed the real netlink-restriction case still yields reachable/high confidence when TCP succeeds.
- [x] Add firewall evidence with safe bounded reads of the host's active firewall state.
  - PR #4 adds read-only `nft -nn list ruleset` evidence only after TCP failure;
  - lookup is capped at two seconds and 64 KiB of captured output;
  - PATH plus standard `/usr/sbin` and `/sbin` locations are supported;
  - base-policy/direct TCP-port matches are labeled candidate evidence, not a proven verdict;
  - unavailable netlink/permission access degrades to `unknown` without changing the stronger diagnosis conclusion;
  - PR #4 CI run `34997617382` passed format, vet, tests, and build;
  - merged at `2fe64640093b258b3c52b148fc2e32507eeae584`;
  - Debian Normal-mode validation reproduced the real nft netlink restriction and preserved the correct no-listener/high-confidence conclusion.
- [x] Add bounded systemd/journal failure evidence for relevant units.
  - PR #5 runs only for a local TCP failure with no listener;
  - selects failed units from the existing snapshot, capped at three candidates;
  - current-boot journal evidence is capped at six lines, 8 KiB, and two seconds per candidate;
  - common credential/token patterns are redacted before journal text becomes diagnosis evidence;
  - no failed units means `systemd-failures: pass` and no journal command is run;
  - failed-unit evidence is explicitly candidate evidence, not proof of port ownership;
  - PR #5 CI run `34998204162` passed format, vet, tests, and build;
  - merged at `37c97227a829fc341f943d0b4731242ee2a39650`;
  - Debian Normal-mode validation collected 219 services with zero failed units and preserved the correct high-confidence no-listener result; direct journal access was restricted as expected and is handled as unavailable evidence.
- [x] Add Docker port/bind/network correlation.
  - PR #6 parses Docker wildcard, exact-address, and internal-only TCP port mappings from the existing snapshot;
  - Docker network names are retained as an optional backward-compatible snapshot field and surfaced in matching evidence;
  - local listener matching is now TCP-specific and target-address-aware, fixing false positives where the same port was bound on a different host address;
  - PR #6 CI run `34999319318` passed format, vet, tests, and build;
  - merged at `bc939ae0c6339b355bd33916629a93f4a07803c9`;
  - live root-collected snapshot validation correctly distinguished Chaptarr bound only to `192.168.2.181:8789`, FlareSolverr `8192/tcp` as internal-only, and the reachable Chaptarr host address.
- [x] Define deterministic evidence ordering and confidence behavior when evidence is unavailable.
  - PR #7 makes successful TCP definitive; local listener/Docker evidence outranks firewall/systemd candidates; confirmed remote no-route outranks firewall; unavailable optional evidence stays neutral;
  - conflicting snapshot evidence lowers confidence instead of pretending certainty;
  - PR #7 CI run `34999939216` passed on exact head `8e2b675c1a3cf7084e19e6b966d34ba9327f5a2f`;
  - merged at `3352a7e8407eae855f4a88550cfaaf867f86ddf1`.
- [x] Add regression tests for reachable, locally blocked, service-failed, and container-port mismatch cases.
  - scenario-level tests lock check ordering and confidence for reachable, local-listener/firewall candidate, failed-systemd candidate, Docker bind mismatch, contradictory Docker publication, confirmed remote no-route, and unavailable optional evidence.
- [x] Validate the merged M2 build under the real root system-service privilege model.
  - merged M2 source is `3352a7e8407eae855f4a88550cfaaf867f86ddf1`; main CI run `35000071547` passed format, vet, tests, and build;
  - owner/root approved `/srv/homecommander-deployments/hostsleuth/hostsleuth.m2-candidate` with SHA-256 `5d595d9db476f9cc030d052df53f6139057fdc6f74f6e440e9f833e5cee201aa`;
  - managed install and restart succeeded; PID changed `31172 -> 443538`;
  - final `deployment_status` reports executable and unchanged unit hashes both matching approval; service remains enabled, active, and running;
  - installed version is `0.1.0-dev+3352a7e`;
  - root-service `ip route get` evidence is readable and reports `route: pass` for local targets;
  - root-service nftables reads are available and produce bounded candidate evidence rather than permission/netlink errors;
  - all 21 Docker containers now carry network names in the live snapshot;
  - `127.0.0.1:8789` correctly identifies Chaptarr as published only on `192.168.2.181:8789`; `127.0.0.1:8192` correctly identifies FlareSolverr as internal-only; `192.168.2.181:8789` and `127.0.0.1:22` remain reachable/high confidence;
  - dashboard returns HTTP 200; schema remains 1; service inventory is 219 and Docker inventory is 21;
  - live snapshot contains zero failed systemd services, so the no-failure branch is validated and journal excerpt collection was intentionally not forced by breaking a service merely for testing;
  - no firewall, service, or container state was mutated to manufacture diagnosis evidence.
- [x] Close M2 after merged-code, CI, exact-hash managed deployment, and live root-context validation all passed.

## Active milestone — M3: Reverse-proxy and TLS awareness

- [ ] Add read-only reverse-proxy awareness for Nginx, Caddy, Traefik, and Nginx Proxy Manager.
- [ ] Correlate proxy listener/frontend evidence with backend host:port targets without editing proxy configuration.
- [ ] Add deterministic TLS/certificate diagnostics after proxy correlation.
- [ ] Add regression scenarios for direct backend success, proxy-to-backend mismatch, certificate expiry/not-yet-valid, hostname mismatch, and unavailable proxy/TLS evidence.
- [ ] Validate M3 on the real Debian 13 host without mutating production proxy configuration merely for tests.

## Planned after M3

- [ ] SQLite storage backend and migration path from JSON/JSONL.
- [ ] Configuration fingerprinting for selected `/etc` files without storing secrets by default.
- [ ] Package-change timeline from dpkg/apt logs.
- [ ] `.deb` package and signed release artifacts.
- [ ] Multi-host architecture research without compromising local-first operation.

## Release / distribution boundary

- [ ] When explicitly authorized by the owner, prepare the next public alpha release containing the final M1 fixes.
  - Do not publish automatically.
  - Current published release remains `v0.1.0-alpha.1` and predates the Docker semantic-event fix.
  - Current live validated build is `0.1.0-dev+3352a7e` from commit `3352a7e8407eae855f4a88550cfaaf867f86ddf1`.

## Shared-infrastructure finding recorded in HomeCommander

- Generic HomeCommander Normal-mode staging initially created the HostSleuth deployment directory as `0700` and staged unit as `0600`.
- The intentionally capability-stripped broker relies on the `homecommander` group for staging access, so first install failed closed before promotion.
- UAT corrected only staging permissions to directory `0750` and unit `0640`; artifact bytes/hashes did not change.
- This belongs in HomeCommander shared-infrastructure hardening, not HostSleuth privileged code.

## Security / safety backlog

- [ ] Authentication before supporting non-loopback web binding as a normal workflow.
- [ ] Privilege separation: run unprivileged by default and isolate collectors requiring elevated reads.
- [ ] Redaction rules for environment variables, command lines, socket endpoints, config files, and logs.
- [ ] Threat model document.
- [ ] Optional audit mode that reports what HostSleuth itself can read.

## Future ideas — do not derail current milestone

- [ ] Dependency graph: DNS -> route/firewall -> proxy -> app -> database/storage.
- [ ] Exposure map showing Internet/LAN/VPN-visible listeners.
- [ ] Baseline-vs-incident comparison UI.
- [ ] Safe repair plans: preview -> snapshot -> user approval -> apply -> verify -> rollback.
- [ ] Self-updating `ARCHITECTURE.md` / recovery documentation generated from observed state.
- [ ] Pluggable diagnosis rules.
- [ ] Optional local LLM explanation layer that cannot alter underlying deterministic evidence.
- [ ] Export/import support bundle for offline troubleshooting.
- [ ] RPM/Arch packaging after Debian/Ubuntu support is mature.
