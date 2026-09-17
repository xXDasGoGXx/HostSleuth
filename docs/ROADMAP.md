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

Publication verification included native linux/amd64 and linux/arm64 binaries, `SHA256SUMS`, public `mjmalleo/hostsleuth:0.3.0` plus `latest`, anonymous multi-platform registry verification, and bounded real-consumer execution.

## 3. M7 — Service Story — COMPLETE

Delivered bounded read-only systemd runtime/journal, process/listener ownership, expected-port collision, container-port, endpoint Diagnose/TLS, and retained-change correlation. M7 added no service controls.

## 4. M8 — Incident Lens — COMPLETE

Delivered a bounded read-only +/- 15 minute incident window anchored from an exact time, retained event, or completed diagnosis. Existing event categories are reused with deterministic ordering, current endpoint evidence is explicitly separated from historical context, and temporal proximity is never presented as proof of causation.

## 5. M9 — HostSleuth Workbench — COMPLETE

Delivered bounded read-only file identity/checksum comparison, DNS inspection, HTTP/redirect inspection, public certificate inspection/comparison, CLI/JSON API, and a dedicated Workbench UI.

Workbench Web/API operations are loopback-only because selected-file hashing plus server-side DNS/HTTP probing would be inappropriate on an unauthenticated LAN-visible endpoint. It does not expose file contents, arbitrary commands, custom HTTP credentials/headers, or private-key viewing.

## 6. M10 — Reboot Story — COMPLETE

Delivered deterministic reboot detection from kernel boot IDs, exact boot-start evidence, bounded previous/current boot journal evidence, direct-evidence-only shutdown classification, current failed-service evidence, retained post-boot recovery correlation, non-causal nearby change context, CLI/API/UI integration, and schema-upgrade protection against false reboot events.

M10 remains read-only and does not infer a reboot cause from temporal proximity.

Full closeout: `docs/history/M10-REBOOT-STORY.md`.

## 7. M11 — Optional Safe Actions — COMPLETE

Goal: prove that HostSleuth can offer one surgical administrative action without becoming Webmin, Cockpit, a browser shell, or an automatic-remediation engine.

Delivered one fixed action only:

`service.restart`

Security/product boundaries:

- actions are disabled by default;
- explicit `--enable-actions` opt-in is required;
- each restart target must be explicitly allowlisted with `--allow-restart-service UNIT`;
- conservative `.service` unit validation;
- no arbitrary command, argv, script, or shell field;
- no generic systemd controller;
- trusted absolute `/usr/bin/systemctl` or `/bin/systemctl` for action evidence and execution rather than `$PATH` resolution;
- deterministic preview of target/effect/argv;
- exact confirmation value required before execution;
- durable audit write required before the restart command runs;
- bounded before/after evidence and action timeout;
- success only after observing `ActiveState=active`;
- loopback-only Action Web/API operations;
- JSON-only state-changing Web requests plus explicit `X-HostSleuth-Action: confirm` header;
- Docker mode reports native systemd restart unavailable;
- no writable Docker socket added;
- no automatic remediation.

Delivered surfaces:

- `hostsleuth action list|preview|run|audit`;
- loopback-only Action JSON API;
- dedicated Actions UI with allowlisted target selection, preview, exact confirmation, result, and audit display.

Validation includes focused security/regression tests, full format/vet/test/build checks, exact served-JavaScript syntax, Docker smoke, linux/amd64 and linux/arm64 image builds, and a disposable real-systemd acceptance on an ephemeral GitHub runner. The real acceptance proved wrong-confirmation denial without PID change, a successful allowlisted restart with PID change, `ActiveState=active` postcondition verification, the expected audit sequence, and cleanup.

No OMV production service was restarted or redeployed for M11 acceptance.

Full closeout: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

## 8. Release / deployment decision — NOT STARTED

M11 completion does not automatically authorize a new public release, Docker publication, OMV production upgrade, or disaster-recovery repin.

Stable public production/recovery remains `mjmalleo/hostsleuth:0.3.0` until an explicit owner decision changes that boundary.

If a new release is approved, validate the exact accepted source, native assets, checksums, amd64/arm64 container images, Docker smoke, upgrade notes, and then separately decide whether live OMV/recovery should move to the new tag.

## 9. Later — Redacted Evidence Bundle — NOT STARTED

Goal: make HostSleuth evidence safely shareable only after redaction rules and a threat model are mature enough.

A bundle may include selected Host Story, diagnosis, event, route/listener/service/container, package, configuration-fingerprint, certificate, Service Story, Incident Lens, Workbench, Reboot Story, and safe-action audit evidence.

Requirements before implementation:

- explicit inclusion rules;
- deterministic documented redaction rules;
- preview before export;
- no silent configuration-content export;
- no credentials, tokens, cookies, private keys, or other secrets;
- clear manifest of what was included/redacted;
- bounded output and integrity/checksum information.

Do not build export/import before the redaction/threat-model work is strong enough, and do not start this milestone without explicit owner direction.

## Research backlog — not an implementation milestone

The detailed current research artifact is `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`.

Promising differentiated questions include:

- expected-endpoint contracts tying ownership, listener/bind, DNS, protocol/TLS, local files/certificates, and retained changes together;
- resolver/delegation or split-view DNS mismatches without becoming a DNS manager;
- redirect, reverse-proxy, Host-header, or upstream mismatches without becoming a proxy manager;
- permissions/ownership/path evidence explaining why a service cannot consume an expected file;
- protocol-aware STARTTLS inspection for SMTP/IMAP and similar services;
- certificate source -> destination -> actually-served verification;
- narrowly justified future safe actions using the M11 explicit-schema/allowlist/preview/confirmation/audit/postcondition model.

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
