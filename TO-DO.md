# HostSleuth — TO-DO

This file stays intentionally short. Completed milestone history belongs in `docs/history/DEVELOPMENT-HISTORY.md`.

## Completed

- [x] M0 — repository foundation.
- [x] M1 — deployable single-host MVP.
- [x] M2 — deeper deterministic diagnosis.
- [x] M3.1 — modern responsive Web UI.
- [x] M3.2 — usability and installation clarity.
- [x] M3.3 — supported Docker Compose deployment with amd64/arm64 CI validation.
- [x] Collective product review after M3.
- [x] Real-host native M3 acceptance on Debian 13 using a side-by-side build with isolated temporary state; the known-good installed service was not replaced.

## Current — finish M3 acceptance polish

Do these before adding another diagnostic subsystem:

- [x] Validate the merged M3 Web UI/API/diagnosis on a real supported Linux deployment.
- [ ] Smoke-test the supported Docker Compose path on a real Linux Docker host without replacing the known-good native deployment. Current approved management policy blocks Docker commands; do not bypass that control.
- [ ] Commit one real, sanitized Web UI screenshot to the README. A real Diagnose-view capture from the accepted Debian 13 build has been produced and checked for public-safe content; repository embedding remains pending.
- [x] Fix only real native-acceptance problems found by those checks. No native M3 product blocker was found.

## Next capability — M4: package-change timeline

Do not start M4 until the remaining M3 acceptance items above are deliberately resolved or explicitly accepted as external/policy-limited.

Keep M4 deliberately narrow. Its only job is to make **“what changed?”** more useful.

- [ ] Read package install/update/remove history from supported local package-manager logs.
- [ ] Start with a bounded Debian/Ubuntu `apt`/`dpkg` implementation; unsupported systems should remain truthful and quiet rather than requiring configuration.
- [ ] Add package changes to the existing event timeline instead of creating a separate package UI.
- [ ] Keep collection read-only and local-first.
- [ ] Do not add package management, update actions, alerts, repositories, or package settings.
- [ ] Add focused tests and real-host validation before merge.

## Belongs in HostSleuth, but only one at a time after M4

These fit one of HostSleuth's two core jobs, but none becomes active automatically:

- Configuration fingerprinting that records change evidence without storing configuration secrets by default.
- Bounded TLS/certificate diagnosis for common host:port failures: handshake, hostname, expiry, and trust evidence.
- General redaction rules and threat-model documentation before exporting or sharing richer diagnostic data.

## Demand-gated — keep only if real users prove the need

- Authentication for direct non-loopback Web UI exposure. Loopback + SSH tunneling remains the simple default.
- Reverse-proxy awareness for common proxies. Do not build a proxy-management layer.
- `.deb` packaging and signed artifacts if the existing installer is a real adoption problem.
- Privilege separation if future collectors actually require enough elevated access to justify the complexity.
- SQLite only if JSON/JSONL becomes a demonstrated operational limitation.
- Exposure-map UI only if listener visibility is repeatedly hard to understand from the existing views.
- Baseline-vs-incident comparison only if the event timeline proves insufficient in real incidents.
- Export/import support bundle only after redaction rules are strong enough to make sharing safe.

## Not planned unless HostSleuth changes direction

These would add substantial architecture, UI, or safety complexity beyond the current small single-host product:

- Multi-host controller/agent architecture.
- General dependency-graph platform.
- Pluggable diagnosis-rule ecosystem.
- AI explanation layer.
- Automatic or guided repair/remediation workflow.

They can be reconsidered later, but they are not standing tasks.

## Rule for new ideas

An idea goes here first. It becomes active only when it clearly improves a common HostSleuth workflow without making installation, operation, or the UI meaningfully harder.
