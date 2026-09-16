# HostSleuth — Ordered Product Roadmap

HostSleuth stays a small, local-first Linux troubleshooting tool with two jobs:

1. remember meaningful host changes;
2. explain why a host/service/port is or is not reachable using deterministic evidence.

This roadmap is intentionally ordered. Complete one bounded milestone at a time. Do not skip ahead, broaden a milestone into a generic administration platform, or add unrelated monitoring features just because they are technically possible.

Consumer/product research may identify useful later workflows, but research does not silently reorder this roadmap or expand the active milestone.

## 1. M6 — Certificate Story / TLS Detective — COMPLETE

Delivered bounded read-only TLS/certificate diagnosis including handshake, served certificate metadata/fingerprint, hostname/trust evidence, local listener/container correlation, native Certbot lineage/renewal evidence, and conservative local-vs-served certificate comparison.

No renewal, reload, certificate installation, ACME account management, private-key handling, or arbitrary command execution was added.

## 2. Publish stable v0.3.0 — COMPLETE

Stable `v0.3.0` was published from accepted source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Publication verification included native linux/amd64 and linux/arm64 binaries, `SHA256SUMS`, public `mjmalleo/hostsleuth:0.3.0` plus `latest`, anonymous multi-platform registry verification, and bounded real-consumer execution. Live OMV was intentionally not migrated.

## 3. M7 — Service Story — COMPLETE

Delivered bounded read-only systemd runtime/journal, process/listener ownership, expected-port collision, container-port, endpoint Diagnose/TLS, and retained-change correlation. M7 added no service controls. Real-host acceptance found and fixed a false collision inference when listener PID metadata was hidden.

## 4. M8 — Incident Lens — COMPLETE

Delivered a bounded read-only +/- 15 minute incident window anchored from an exact time, retained event, or completed diagnosis. Existing event categories are reused with deterministic ordering, current endpoint Diagnose/TLS evidence is explicitly separated from historical context, and temporal proximity is never presented as proof of causation.

M8 added no time-series database, alerting layer, historical network reconstruction, service control, remediation, or reboot-cause analysis.

## 5. M9 — HostSleuth Workbench — COMPLETE

Goal: reduce common troubleshooting workflows that normally force an administrator across several shell commands or one-off websites, without becoming a miscellaneous utility collection or browser shell.

Delivered bounded read-only tools:

- selected-file SHA-256 and SHA-512 calculation without returning or storing file contents;
- expected SHA-256/SHA-512 verification;
- file-to-file comparison by SHA-256 fingerprint;
- file path, size, mode/permissions, mtime, UID/GID, owner/group, and fingerprints;
- common DNS evidence through the host system resolver: A, AAAA, CNAME where distinct, MX, NS, TXT, and PTR for IP input;
- direct HTTP/HTTPS HEAD inspection with a bounded redirect chain and selected response metadata;
- public PEM certificate metadata/fingerprint inspection;
- exact public certificate file fingerprint vs direct-TLS served certificate comparison;
- CLI, JSON API, and a dedicated Workbench Web UI tab.

Security/product boundaries:

- no arbitrary command field or hidden shell hook;
- no file editor or file-content display;
- no custom HTTP headers, cookies, credentials, or request body;
- certificate inspection refuses a private-key PEM block encountered before a public certificate;
- Workbench Web/API operations are loopback-only because file hashing plus server-side DNS/HTTP probing would be inappropriate on an unauthenticated LAN-visible endpoint; local CLI and the recommended SSH-tunnel workflow remain available;
- Docker Workbench file inspection only sees files actually readable inside the supported container/mounts and does not manufacture broader host-filesystem visibility.

Validation includes focused tests, full test/vet/build/format checks, syntax validation of every UI fragment and the exact concatenated served JavaScript, Docker smoke coverage, and isolated real-OMV CLI/API/UI acceptance. The real-host validation also caught an environment-specific umask assumption in a test; the test was corrected to verify actual observed permissions instead of assuming a host umask.

M9 intentionally spans file integrity, DNS, HTTP, and certificate identity. Certificate tooling is one Workbench workflow, not the product's sole enhancement direction.

## 6. M10 — Reboot Story — COMPLETE

Goal: answer **"What happened around this reboot, and what failed to come back afterward?"** using bounded deterministic evidence without inventing a reboot cause.

Delivered:

- snapshot schema v4 with Linux kernel boot ID and exact `/proc/stat` `btime` when available;
- reboot detection only when a previously known boot ID changes to another known boot ID;
- no inference from human-readable uptime;
- schema-upgrade protection so an older snapshot with no boot ID cannot create a false reboot event;
- bounded previous-boot and current-boot journal evidence;
- explicit `unknown` behavior when journal history is unavailable or permissions prevent access;
- orderly-vs-abnormal shutdown classification only when direct bounded evidence supports it;
- current failed-service evidence;
- retained post-boot service/listener/container recovery correlation;
- a recovery issue only when retained post-boot evidence and current snapshot state agree the problem remains;
- nearby package/kernel/system/configuration changes as context, never timing-based proof of cause;
- CLI, JSON API, and a dedicated Reboot Web UI tab;
- reuse of the existing retained event/Incident Lens primitives rather than a second history store.

Real-host acceptance on OMV found that `journalctl` may return `No journal files were opened due to insufficient permissions.` as output. M10 now recognizes that diagnostic as unavailable evidence and reports `unknown`; no privilege expansion was added. Full source validation included tests, vet, formatting, JavaScript fragments plus exact served-script syntax, native build, Docker smoke, and linux/amd64 + linux/arm64 image builds.

M10 remains read-only. No reboot cause is claimed from temporal proximity. Full closeout detail is in `docs/history/M10-REBOOT-STORY.md`.

## 7. M11 — Optional Safe Actions — NOT STARTED

Goal: carefully test whether HostSleuth can offer a very small number of surgical administrative actions without becoming Webmin, Cockpit, or a browser shell.

This milestone changes HostSleuth's current read-only boundary and therefore requires an explicit design/security review **before implementation**. Do not start M11 automatically after M10; explicit owner direction is required.

Certificate lifecycle remains one candidate because it is narrow and auditable, but it is not the only possible safe-action family. Consumer research should compare multiple real troubleshooting workflows before any action set is approved.

Possible design properties include:

- explicit predefined schemas/recipes rather than arbitrary commands;
- exact preview of intended effects;
- optional bounded operation only when ownership/target is clear;
- post-action verification against an observable condition;
- audit evidence for requested and observed behavior;
- disabled-by-default action capability and explicit confirmation.

Any handling of private keys, destination writes, ownership/mode changes, credentials/tokens, package/service/firewall state, or rollback semantics requires specific threat-model/design work before implementation. Generic administration remains out of scope unless separately justified later.

## 8. Later — Redacted Evidence Bundle

Goal: make HostSleuth evidence safely shareable after redaction rules and a threat model are mature enough.

A bundle may include selected Host Story, diagnosis, event, route/listener/service/container, package, configuration-fingerprint, certificate, Service Story, Incident Lens, Workbench, and Reboot Story evidence. It must apply documented redaction rules before export and must never silently include configuration contents, credentials, tokens, private keys, or other secrets.

Do not build export/import before the redaction/threat-model work is strong enough to support it.

## Research backlog — not an implementation milestone

Ongoing market/user-workflow research should compare HostSleuth against what people currently assemble from CLI tools, monitoring products, admin consoles, log viewers, DNS/HTTP/TLS sites, package tools, scripts, and single-purpose utilities.

The detailed current research artifact is `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`.

Research is deliberately broader than certificate management. Promising differentiated questions include:

- can an expected-endpoint contract tie service ownership, listener/bind, DNS, protocol/TLS behavior, local file/certificate evidence, and retained changes into one troubleshooting story?
- can HostSleuth explain resolver/delegation or split-view DNS mismatches without becoming a DNS server manager?
- can it explain redirect, reverse-proxy, Host-header, or upstream mismatches with bounded HTTP evidence without becoming a proxy manager?
- can permissions/ownership/path evidence explain why a service cannot consume a file it is expected to use?
- should later protocol inspection understand STARTTLS services such as SMTP/IMAP rather than assuming immediate TLS?
- can certificate source -> destination -> actually-served comparison verify deployment and rollout without becoming a generic ACME manager?
- can optional safe actions use explicit schemas plus observable postcondition verification instead of generic scripts or command fields?

Research findings must be deliberately assigned to an approved milestone before implementation.

## Guardrails that remain in force

Do not drift into:

- multi-host controller/agent architecture;
- generic network-device/SNMP monitoring;
- time-series graphing/monitoring platform behavior;
- arbitrary web terminal or command execution;
- generic package/firewall/configuration administration;
- AI-generated causal claims;
- automatic remediation;
- broad privilege expansion simply to make features easier.

HostSleuth should feel powerful because it connects deterministic evidence into answers people actually need, not because it exposes every system control in a browser.
