# HostSleuth — TO-DO

## Active milestone — M1: Single-host deployable MVP

- [ ] Define stable snapshot and event schemas.
- [ ] Collect host identity, OS, kernel, uptime, CPU count, memory, filesystems, interfaces, addresses, routes, listeners, systemd service state, and Docker state when available.
- [ ] Persist snapshots atomically under `/var/lib/hostsleuth` (configurable).
- [ ] Diff snapshots into concise human-readable change events.
- [ ] Implement deterministic `diagnose host:port` checks with evidence.
- [ ] Add local HTTP API and embedded web UI.
- [ ] Add CLI subcommands: `serve`, `snapshot`, `diagnose`, `events`, `version`.
- [ ] Add tests for parsing, diffing, diagnostics, and storage behavior.
- [ ] Add GitHub Actions CI (`go test`, `go vet`, build).
- [ ] Add systemd unit and install/uninstall scripts.
- [ ] Document first deployment on Debian/Ubuntu.
- [ ] Verify CI green and record exact validation in `CURRENT-HANDOFF.md`.

## Planned immediately after M1

- [ ] SQLite storage backend and migration path from JSON/JSONL.
- [ ] Configuration fingerprinting for selected `/etc` files without storing secrets by default.
- [ ] Package-change timeline from dpkg/apt logs.
- [ ] Improved systemd failure evidence using bounded journal excerpts.
- [ ] Docker port/bind/network correlation.
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
