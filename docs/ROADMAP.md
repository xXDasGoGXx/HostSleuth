# HostSleuth — Ordered Product Roadmap

HostSleuth stays a small, local-first Linux troubleshooting tool with two jobs:

1. remember meaningful host changes;
2. explain why a host/service/port is or is not reachable using deterministic evidence.

This roadmap is intentionally ordered. Complete one bounded milestone at a time. Do not skip ahead, broaden a milestone into a generic administration platform, or add unrelated monitoring features just because they are technically possible.

## 1. M6 — Certificate Story / TLS Detective

Goal: make TLS and certificate failures understandable from the same evidence-first HostSleuth workflow.

Bounded scope:

- diagnose TLS for a `host:port` target;
- show certificate subject, SANs, issuer, serial, validity window, remaining lifetime, and SHA-256 fingerprint;
- report hostname match, handshake result, and bounded trust/chain evidence;
- correlate local listener/process/container evidence for the TLS endpoint when available;
- on native Linux, detect Certbot when present and show bounded read-only certificate/renewal/timer evidence;
- compare a local certificate fingerprint with the certificate actually being served when both are available;
- surface stale-served-certificate situations such as "certificate on disk renewed, service still presenting the previous certificate" only when evidence supports that conclusion;
- integrate into existing Diagnose/Host Story patterns rather than creating a generic certificate-management dashboard.

M6 remains read-only. No renewal button, service reload, certificate installation, ACME account management, or arbitrary command execution.

## 2. Publish stable v0.3.0

After M6 acceptance, publish one stable release containing M5 configuration fingerprinting plus M6 TLS/certificate diagnosis.

Use the existing controlled release path:

- exact accepted `main` SHA;
- owner approval at the publication boundary;
- native linux/amd64 and linux/arm64 binaries plus `SHA256SUMS`;
- public `mjmalleo/hostsleuth:0.3.0` plus `latest`;
- anonymous registry verification;
- one bounded consumer acceptance check;
- no live OMV deployment change merely to chase a version number.

## 3. M7 — Service Story

Goal: answer "why will this service not start / why did this endpoint disappear?" without becoming a service manager.

Correlate existing evidence around a selected service/listener/container:

- systemd state and bounded journal evidence;
- process/listener ownership where available;
- port collision evidence;
- related container publication/network evidence;
- nearby package changes;
- nearby configuration-fingerprint changes;
- TLS evidence when the endpoint is TLS-capable;
- listener appearance/disappearance.

Present an evidence story, not an invented causal verdict. No restart/stop/start controls in this milestone.

## 4. M8 — Incident Lens

Goal: answer "what changed around the time this broke?"

Allow a diagnosis/event/time to anchor a bounded evidence window, initially around +/- 15 minutes, showing temporally nearby:

- package events;
- configuration events;
- service/container changes;
- listener changes;
- certificate/TLS changes where evidence exists;
- reboot/boot context if already available.

Label temporal proximity as context, not proof of causation. Reuse the existing event model rather than building a time-series monitoring database.

## 5. M9 — HostSleuth Workbench

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

Every tool must answer one concrete troubleshooting question. Do not add a web shell, arbitrary command box, file editor, or generic system-control panel.

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

The first candidate action is Certbot because it is narrow and auditable:

- Certbot renewal dry-run;
- explicit certificate renewal;
- optional bounded reload of the associated known service after successful renewal, only when ownership is clear.

Any action framework must require explicit enablement, show the exact operation before execution, require confirmation, create an audit event, expose no arbitrary command field, and default to disabled. Generic service/package/firewall administration remains out of scope unless separately justified later.

## 8. Later — Redacted Evidence Bundle

Goal: make HostSleuth evidence safely shareable after redaction rules and a threat model are mature enough.

A bundle may include selected Host Story, diagnosis, event, route/listener/service/container, package, configuration-fingerprint, and certificate evidence. It must apply documented redaction rules before export and must never silently include configuration contents, credentials, tokens, private keys, or other secrets.

Do not build export/import before the redaction/threat-model work is strong enough to support it.

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
