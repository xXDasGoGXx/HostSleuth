# HostSleuth — TO-DO

This file is intentionally short. Completed milestone history belongs in `docs/history/DEVELOPMENT-HISTORY.md`.

## Completed

- [x] M0 — repository foundation.
- [x] M1 — deployable single-host MVP.
- [x] M2 — deeper deterministic diagnosis.
- [x] Real Debian 13 managed deployment and reboot validation.
- [x] Docker uptime-noise fix.
- [x] Route, nftables, systemd/journal candidate, Docker bind/network, and evidence-precedence diagnosis work.
- [x] Public-repository cleanup and documentation sanitization.
- [x] M3.1 — modern responsive Web UI with readable diagnosis, changes, and host views.
- [x] M3.2 — usability, first-run clarity, build/version visibility, and simplified installation guidance.
- [x] M3.3 — supported Docker Compose deployment with explicit reduced visibility and amd64/arm64 CI validation.

## Current — collective product review

M3 is complete. Do not start another capability milestone automatically.

Review the deferred decisions below as a group and promote only the smallest item that clearly solves a common real HostSleuth user problem without making installation, operation, or the UI meaningfully harder.

## Collective review — deferred product decisions

- [ ] Capture one real Web UI screenshot for the README after an accepted deployed build is available.
- [ ] Authentication before normal non-loopback dashboard exposure.
- [ ] Reverse-proxy awareness for Nginx, Caddy, Traefik, and Nginx Proxy Manager.
- [ ] TLS/certificate diagnostics.
- [ ] Configuration fingerprinting without storing secrets by default.
- [ ] Package-change timeline from dpkg/apt logs.
- [ ] SQLite storage only if JSON/JSONL becomes a demonstrated operational limitation.
- [ ] `.deb` package and signed release artifacts.
- [ ] Privilege separation for collectors that require elevated reads.
- [ ] General redaction rules and threat-model documentation.
- [ ] Exposure map for Internet/LAN/VPN-visible listeners.
- [ ] Baseline-vs-incident comparison UI.
- [ ] Dependency graph across DNS, network, proxy, app, and storage layers.
- [ ] Multi-host architecture that preserves local-first operation.
- [ ] Export/import support bundle.
- [ ] Pluggable diagnosis rules.
- [ ] Optional local AI explanation layer that cannot alter deterministic evidence.
- [ ] Any repair workflow only as explicit preview -> approval -> apply -> verify -> rollback; never as implicit automation.

## Rule for new ideas

Add new ideas here without interrupting the current task. An idea does not become active merely because it sounds useful.
