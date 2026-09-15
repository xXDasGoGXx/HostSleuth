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
- [ ] Exercise the privileged systemd installation/enable path on an approved host with administrative execution available; verify reboot persistence and uninstall behavior.
- [ ] Validate Docker inventory in the final service privilege model.

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
