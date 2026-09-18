# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Stable / public / live / recovery state — v0.6.0

Stable release, live Arcane production, and disaster recovery are now aligned on:

`v0.6.0` / `mjmalleo/hostsleuth:0.6.0`

Exact published source:

`fcb08be51ae3da8cd20dc3929cf9d736b15f170c`

Release workflow:

`35284864117` — success

Verified public OCI index:

`sha256:6b9f90209f477ba8213d9c4bf7db7996d6c52d2caae2cfce405822df8bc8ef2a`

Live acceptance confirmed v0.6.0, schema 4 / Docker mode, retained pre-upgrade history, the M12 Expectations UI/API, a passing live Expected Endpoint Contract, and the LAN HTTP 403 boundary for Safe Actions.

`OMV-Docker-Rebuild` recovery alignment PR #6 merged at:

`f78abbe5742c315d73cf85709ebfeba620311429`

Full publication record: `docs/history/V0.6.0-PUBLICATION.md`.

Full production/recovery record: `docs/history/V0.6.0-PRODUCTION-ALIGNMENT.md`.

## Redacted Evidence Bundle — complete, published, and live

The accepted threat model/redaction contract is in:

`docs/design/REDACTED-EVIDENCE-BUNDLE.md`

Implementation closeout:

`docs/history/REDACTED-EVIDENCE-BUNDLE.md`

CLI:

```text
hostsleuth evidence preview
hostsleuth evidence export [--output PATH]
```

The first version is deliberately bounded to current Snapshot, recent Events, and Safe Action audit records. It uses deterministic bundle-local pseudonymization, credential/private-key scrubbing, manifest/checksum evidence, owner-only local archive permissions, and explicit exclusions for raw journal text, raw command output, arbitrary file/config contents, action confirmation material, cloud upload, and automatic sharing.

Redaction lowers disclosure risk but cannot guarantee anonymity. Bundles must still be reviewed before sharing.

## M12 — Expected Endpoint Contracts

The owner selected Expected Endpoint Contracts as the next bounded milestone. PR #43 passed full CI and merged to `main` at `1c193473d9220c34ec2820526df76076bfb41ce9`.

Delivered:

- typed read-only contract with target, optional exact DNS set, TLS mode, optional systemd service, and optional container;
- TCP reachability always required;
- ordered Expected-vs-Observed checklist;
- proven failure takes precedence over unresolved evidence for `first_mismatch`;
- full existing Diagnosis included for deeper route/listener/container/TLS/firewall evidence;
- `hostsleuth contract` CLI;
- `GET /api/contract`;
- Web UI **Expectations** view with handoff to full Diagnose;
- no polling, alerts, remediation, persistent contract database, or new privilege.

Local Go 1.24.13 format/vet/full-test/build and disposable HTTP/UI acceptance passed.

Design: `docs/design/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

Closeout: `docs/history/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

## Current milestone boundary

v0.6.0 publication and production/recovery rollout are complete.

Do not automatically start a new feature milestone merely because this boundary is clean. Use `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md` as design input only when the owner selects the next bounded milestone.

Do not restart a broad audit on continuation; use this handoff.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
