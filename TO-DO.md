# HostSleuth — TO-DO

## Active milestone — M1: Single-host deployable MVP

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
- [x] Exercise the core runtime on a real Debian 13 host: snapshots, events, diagnostics, web dashboard/API, filesystem inventory, native build, and cross-builds.
- [x] Add release automation for static Linux `amd64`/`arm64` binaries with checksums, tag creation, and release version stamping.
- [x] Verify the hardened M1 code is green in GitHub Actions: run `34936129425` passed format, vet, tests, and build.
- [x] Publish `v0.1.0-alpha.1` and validate its published amd64 artifact on Debian without a system Go installation, including checksum, version, snapshot, diagnosis, and `/releases/latest/download/...` URL.
- [x] Merge the systemd-owned state-directory fix.
  - PR #1 CI run `34987407246` passed.
  - merged to `main` at `20c1793af4076f3e7fa8ea9d5cc23268fb55c6e9`.
  - unit uses `StateDirectory=hostsleuth` and `StateDirectoryMode=0700`.
- [x] Stage HostSleuth as the first HomeCommander Managed Administrative Deployment UAT target on `openmediavault`.
  - original staged release binary: `/srv/homecommander-deployments/hostsleuth/hostsleuth`
  - original binary SHA-256: `8fe9e0caf991e4a3413a98b7b1ca75063cfd748a86bcb2f6fb953edba9008c90`
  - staged unit: `/srv/homecommander-deployments/hostsleuth/hostsleuth.service`
  - unit SHA-256: `416e374c1289ca6ef020b9baa056c2213ea24c350a56c26eb511ebcbcc72ee47`
- [x] Owner/root approved the exact initial HostSleuth artifacts through `homecommander-approve-deployment`.
  - approval remains exact-hash/fixed-destination and does not add a generic root shell, unrestricted sudo, arbitrary privileged writes, or HostSleuth-specific HomeCommander privilege code.
- [x] Exercise privileged install and lifecycle through HomeCommander Managed Administrative Deployment.
  - exact-hash install passed;
  - enable/start passed;
  - restart passed and returned to active/running;
  - stop/disable passed;
  - enable/start again passed;
  - exact uninstall passed;
  - reinstall passed;
  - final state restored to active/running and enabled.
- [x] Verify systemd-owned state directory creation.
  - `/var/lib/hostsleuth` is `root:root 0700`.
  - state directory remains after managed uninstall as intended.
- [x] Validate Docker inventory in the final system-service privilege model.
  - service snapshot observed 21 Docker containers.
  - dashboard returned HTTP 200.
- [x] Validate installed CLI diagnosis after managed restart.
  - `diagnose 127.0.0.1:22` returned reachable/high confidence.
- [x] Add read-only `status` to the root-owned HomeCommander approval and verify `deployment_status`.
  - approval contains install/uninstall/enable/disable/start/stop/status/restart.
  - executable and unit both matched their approved SHA-256 exactly.
  - `recordedInstalled=true`.
  - service loaded, enabled, active, and running.
- [x] Perform one owner-authorized reboot of `openmediavault` and verify HostSleuth reboot persistence.
  - real reboot confirmed by ~635-second uptime at first post-reboot check;
  - HomeCommander recovered;
  - approved installed hashes still matched;
  - HostSleuth returned loaded + enabled + active/running with PID 1360;
  - dashboard returned HTTP 200;
  - `/var/lib/hostsleuth` remained `root:root 0700`;
  - Docker inventory remained available with 21 containers;
  - SSH diagnosis remained reachable/high confidence.
- [x] Fix Docker uptime-only flight-recorder event churn discovered during post-reboot validation.
  - 99 of the latest 100 API events were container-change records caused primarily by `Up N minutes -> Up N+1 minutes` raw Docker status text.
  - PR #2 normalizes semantic state only for diffing while keeping full raw status in snapshots/UI.
  - uptime-only and restart-timer-only changes are ignored; health/exit/restart/paused/dead/removal transitions remain meaningful.
  - regression tests cover uptime churn, restart timer churn, health transitions, and exits.
  - CI run `34991719610` passed format, vet, tests, and build.
  - merged to `main` at `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`.
- [x] Build and stage the merged Docker-event fix as a non-release managed-update candidate.
  - no public release was created;
  - source commit: `d7028044fcb0fa3396b37c621a33fd5c4c1f2c5e`;
  - version: `0.1.0-dev+d702804`;
  - candidate SHA-256: `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1`;
  - staged path: `/srv/homecommander-deployments/hostsleuth/hostsleuth.candidate`;
  - original `v0.1.0-alpha.1` staged binary remains untouched for recovery/reference;
  - systemd unit bytes/hash are unchanged.
- [ ] Owner/root re-approves the new `hostsleuth.candidate` exact hash with the unchanged unit and normal default actions.
- [ ] Exercise a real HomeCommander managed update to candidate `0.1.0-dev+d702804`.
  - managed `install` should update the recorded deployment only after verifying the current installed files still match their prior state;
  - managed restart after install;
  - verify `deployment_status` pins/observes candidate SHA `1014a482ea00812bbf0ae816e55494caa31e49d7bfde6bb867dd0cca132da4e1` and unchanged unit SHA.
- [ ] Validate the Docker event-noise fix under the real root system-service model.
  - verify dashboard/API and 21-container Docker inventory remain healthy;
  - observe at least two normal 60-second capture intervals;
  - confirm uptime progression no longer creates repeated container-change events;
  - do not erase existing history merely to make validation easier.
- [ ] Close M1 after the managed update + live event-noise validation are green.

## Shared-infrastructure finding recorded in HomeCommander

- Generic HomeCommander Normal-mode staging initially created `/srv/homecommander-deployments/hostsleuth` as `0700` and the staged unit as `0600`.
- The intentionally capability-stripped root broker relies on the `homecommander` group for staging access, so those worker-private modes caused the first install attempt to fail closed with permission denied.
- UAT corrected only staging permissions to `0750` for the deployment directory and `0640` for the unit; hashes did not change.
- This belongs in HomeCommander shared-infrastructure hardening/documentation, not as HostSleuth privileged code.

## Planned immediately after M1

- [ ] SQLite storage backend and migration path from JSON/JSONL.
- [ ] Configuration fingerprinting for selected `/etc` files without storing secrets by default.
- [ ] Package-change timeline from dpkg/apt logs.
- [ ] Improved systemd failure evidence using bounded journal excerpts.
- [ ] Docker port/bind/network correlation.
- [ ] Route and firewall evidence in diagnosis.
- [ ] Reverse-proxy awareness for Nginx, Caddy, Traefik, and Nginx Proxy Manager.
- [ ] TLS/certificate diagnostics.
- [ ] `.deb` package and signed release artifacts.
- [ ] Multi-host architecture research without compromising local-first operation.

## Security / safety backlog

- [ ] Authentication before supporting non-loopback web binding as a normal workflow.
- [ ] Privilege separation: run unprivileged by default and isolate collectors requiring elevated reads.
- [ ] Redaction rules for environment variables, command lines, socket endpoints, config files, and logs.
- [ ] Threat model document.
- [ ] Optional audit mode that reports what HostSleuth itself can read.

## Future ideas — do not derail current milestone

- [ ] Dependency graph: DNS → route/firewall → proxy → app → database/storage.
- [ ] Exposure map showing Internet/LAN/VPN-visible listeners.
- [ ] Baseline-vs-incident comparison UI.
- [ ] Safe repair plans: preview → snapshot → user approval → apply → verify → rollback.
- [ ] Self-updating `ARCHITECTURE.md` / recovery documentation generated from observed state.
- [ ] Pluggable diagnosis rules.
- [ ] Optional local LLM explanation layer that cannot alter underlying deterministic evidence.
- [ ] Export/import support bundle for offline troubleshooting.
- [ ] RPM/Arch packaging after Debian/Ubuntu support is mature.
