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

Goal: answer "what changed around the time this broke?" without turning nearby timestamps into an invented cause.

Delivered bounded read-only correlation:

- explicit incident anchor time;
- fixed initial +/- 15 minute retained-event window;
- all existing event categories preserved inside the window, including package, configuration, service, container, listener, and system events when present;
- deterministic chronological ordering and a 100-event cap;
- optional current endpoint Diagnose/TLS evidence for a supplied `host:port`;
- current endpoint capture timestamp kept explicitly separate from the historical incident anchor;
- Changes-integrated Web UI with **Inspect window** actions on retained timeline entries;
- CLI `incident --at <RFC3339> [--target host:port]` and API `/api/incident-lens`;
- explicit language that temporal proximity is context, not proof of causation.

M8 reuses the existing JSONL event model and adds no time-series database, alerting layer, historical network reconstruction, service control, or remediation. Boot/reboot-cause analysis remains reserved for M10.

Real-host acceptance on isolated OMV state found and fixed one truthfulness defect before closeout: an absent endpoint probe initially serialized a zero capture timestamp. The timestamp is now optional and omitted when no endpoint was checked.

## 5. M9 — HostSleuth Workbench — ACTIVE

Goal: provide a small set of practical troubleshooting tools that normally force an administrator into several shell commands or websites.

Candidates for the first bounded Workbench:

- SHA-256 / SHA-512 calculation for a selected local file without storing file contents;
- expected-checksum verification;
- compare two files by fingerprint;
- inspect path owner/group/permissions/mtime/size/hash;
- DNS lookup for common records with deterministic raw evidence retained;
- HTTP HEAD / redirect-chain inspection;
- inspect a PEM certificate;
- compare a certificate file fingerprint with the certificate a remote/local service is actually presenting.

Consumer research can refine which of these solve the strongest real workflows. In particular, read-only protocol-aware certificate inspection such as STARTTLS may be considered when M9 is designed, but only if it fits the Workbench's bounded troubleshooting purpose.

M9 must not become a miscellaneous-tools junk drawer. Every tool must answer one concrete troubleshooting question. Do not add a web shell, arbitrary command box, file editor, generic system-control panel, or write action.

Validation requirements remain focused tests, normal CI, and bounded real-host acceptance.

## 6. M10 — Reboot Story

Goal: answer "why did this host reboot, and what failed to come back?"

Bounded evidence can include:

- current boot time and previous boot/shutdown evidence where available;
- orderly vs abnormal shutdown indicators only when deterministically supported;
- package/kernel/configuration changes near the reboot;
- services that are failed after boot;
- listeners that existed before but did not return;
- container state changes;
- certificate/listener/service context relevant to lost endpoints.

Do not claim a reboot cause without direct evidence.

## 7. M11 — Optional Safe Actions

Goal: carefully test whether HostSleuth can offer a very small number of surgical administrative actions without becoming Webmin, Cockpit, or a browser shell.

This milestone changes HostSleuth's current read-only boundary and therefore requires an explicit design/security review before implementation.

The first candidate action family remains certificate lifecycle because it is narrow and auditable. Consumer research may refine this beyond a simple renewal button toward a safer deployment workflow, for example:

- renewal dry-run or explicit renewal through a known certificate client;
- an explicit, predefined certificate deployment recipe rather than an arbitrary shell hook;
- preview of source/destination fingerprints and intended file/service effects;
- optional bounded reload of a known associated service only when ownership is clear;
- post-action re-probe proving which certificate the endpoint actually serves;
- audit evidence for the requested and observed operation.

Any handling of private-key material, destination writes, ownership/mode changes, ACME account credentials, or rollback semantics requires specific threat-model/design work before implementation.

Any action framework must require explicit enablement, show the exact operation before execution, require confirmation, create an audit event, expose no arbitrary command field, and default to disabled. Generic service/package/firewall administration remains out of scope unless separately justified later.

## 8. Later — Redacted Evidence Bundle

Goal: make HostSleuth evidence safely shareable after redaction rules and a threat model are mature enough.

A bundle may include selected Host Story, diagnosis, event, route/listener/service/container, package, configuration-fingerprint, certificate, Service Story, and Incident Lens evidence. It must apply documented redaction rules before export and must never silently include configuration contents, credentials, tokens, private keys, or other secrets.

Do not build export/import before the redaction/threat-model work is strong enough to support it.

## Research backlog — not an implementation milestone

Ongoing market/user-workflow research should compare HostSleuth against what people currently assemble from monitoring products, admin consoles, ACME clients, TLS scanners, scripts, and single-purpose websites.

The detailed current research artifact is `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`.

Promising differentiated questions include:

- can HostSleuth explain and verify the entire path from renewed certificate on disk to destination copy to the certificate actually being served?
- should later TLS inspection understand STARTTLS protocols such as SMTP rather than assuming immediate TLS?
- can a bounded endpoint contract tie a service, port/bind, protocol/TLS expectation, certificate fingerprint, and recent host changes together?
- can HostSleuth detect inconsistent certificate rollout across several local consumers/endpoints without becoming a multi-host controller?
- can optional safe actions use explicit schemas/recipes plus postcondition verification instead of generic scripts or command fields?

Research findings must be deliberately assigned to M9, M11, or a later approved milestone before implementation.

## Guardrails that remain in force

Do not drift into:

- multi-host controller/agent architecture;
- generic network-device/SNMP monitoring;
- time-series graphing/monitoring platform behavior;
- arbitrary web terminal or command execution;
- generic package/firewall/configuration administration;
- AI-generated causal claims;
- automatic remediation;
- broad privilege expansion simply to make a feature easier.

HostSleuth should feel powerful because it connects deterministic evidence into answers people actually need, not because it exposes every system control in a browser.
