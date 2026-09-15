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

## M3 — Product Experience

This sequence is intentionally fixed. Finish each section before starting the next one. Do not pull unrelated backlog work into M3.

### M3.1 — Web UI — active

- [ ] Modern responsive interface.
- [ ] Human-readable diagnosis results instead of raw JSON as the normal user experience.
- [ ] Clean recent-change timeline.
- [ ] Clear host overview.
- [ ] Good loading, empty, and error states.
- [ ] Mobile-friendly layout.
- [ ] Keep the UI self-contained in the Go binary with no frontend framework or external runtime dependency.

### M3.2 — Usability — next

- [ ] Review and simplify user-facing wording.
- [ ] Make first-run state obvious and non-confusing.
- [ ] Make version/build information easy to find.
- [ ] Add concise screenshots/examples to the README once the M3.1 UI is stable.
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

## Backlog — only pull forward when justified

- [ ] Authentication before normal non-loopback dashboard exposure.
- [ ] Reverse-proxy awareness for Nginx, Caddy, Traefik, and Nginx Proxy Manager.
- [ ] TLS/certificate diagnostics.
- [ ] Configuration fingerprinting without storing secrets by default.
- [ ] Package-change timeline from dpkg/apt logs.
- [ ] SQLite storage if JSON/JSONL becomes a real operational limitation.
- [ ] `.deb` package and signed release artifacts.
- [ ] Privilege separation for collectors that require elevated reads.
- [ ] General redaction rules and threat-model documentation.

## Later / research

- [ ] Exposure map for Internet/LAN/VPN-visible listeners.
- [ ] Baseline-vs-incident comparison UI.
- [ ] Dependency graph across DNS, network, proxy, app, and storage layers.
- [ ] Multi-host architecture that preserves local-first operation.
- [ ] Export/import support bundle.
- [ ] Pluggable diagnosis rules.
- [ ] Optional local AI explanation layer that cannot alter deterministic evidence.
- [ ] Any repair workflow only as explicit preview -> approval -> apply -> verify -> rollback; never as implicit automation.

## Rule for new ideas

Add ideas here without interrupting the active task. A backlog item does not become active merely because it sounds useful.
