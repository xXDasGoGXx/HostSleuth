# HostSleuth — TO-DO

This file is intentionally short. Completed milestone history belongs in `docs/history/DEVELOPMENT-HISTORY.md`.

## Completed

- [x] M0 — repository foundation.
- [x] M1 — deployable single-host MVP.
- [x] M2 — deeper deterministic diagnosis.
- [x] Real Debian 13 managed deployment and reboot validation.
- [x] Docker uptime-noise fix.
- [x] Route, nftables, systemd/journal candidate, Docker bind/network, and evidence-precedence diagnosis work.

## Active — product hardening through real use

Do not add a major subsystem until a real troubleshooting case shows why it is needed.

- [ ] Use the current HostSleuth build during real incidents and normal homelab troubleshooting.
- [ ] Capture concrete cases where the dashboard or diagnosis output is confusing, incomplete, or wrong.
- [ ] Improve wording, evidence presentation, and navigation based on those real cases.
- [ ] Add concise screenshots/examples to the README once the current UI/output is stable enough to represent the product.
- [ ] Decide separately whether the current validated build is ready for the next public alpha. Publication requires explicit owner approval.

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
