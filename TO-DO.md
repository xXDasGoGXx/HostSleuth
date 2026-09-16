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

## M3 — Product Experience

The sequence stays fixed: M3.1 Web UI -> M3.2 Usability -> M3.3 Docker release. Finish the active section before starting the next one.

### M3.2 — Usability — active

- [ ] Review and simplify user-facing wording.
- [ ] Make first-run state obvious and non-confusing.
- [ ] Make version/build information easy to find.
- [ ] Add concise screenshots/examples to the README once the UI is visually accepted.
- [ ] Simplify installation documentation around the recommended deployment paths.

### M3.3 — Docker release — after M3.2

- [ ] Create an official Dockerfile.
- [ ] Provide one recommended `compose.yaml` rather than many deployment variants.
- [ ] Persist HostSleuth state cleanly.
- [ ] Support amd64 and arm64 builds.
- [ ] Preserve host-aware collection rather than accidentally diagnosing only the HostSleuth container.
- [ ] Use the minimum host access necessary; do not default to unrestricted privileged mode.
- [ ] Keep the same HostSleuth UI and product behavior where technically possible.
- [ ] Add CI image-build validation.
- [ ] Do not publish a container image until explicitly approved.

## Collective review — deferred product decisions

These are not rejected ideas and they are not automatically future features. Keep them visible and review them together after the current M3 work, or earlier only when a real blocker proves one is required.

For each item, ask: **Does this solve a common HostSleuth user problem without making installation, operation, or the UI meaningfully harder?** If not, leave it out.

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

Add new ideas to the collective review list without interrupting the active milestone. An idea does not become active merely because it sounds useful.
