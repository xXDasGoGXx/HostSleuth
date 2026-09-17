# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Stable/public state — v0.6.0

Stable public release:

`v0.6.0`

Exact published source:

`fcb08be51ae3da8cd20dc3929cf9d736b15f170c`

Release workflow:

`35284864117` — success

Published Docker tags:

- `mjmalleo/hostsleuth:0.6.0`
- `mjmalleo/hostsleuth:latest`

Both resolve to verified OCI index:

`sha256:6b9f90209f477ba8213d9c4bf7db7996d6c52d2caae2cfce405822df8bc8ef2a`

Independent verification passed for native checksums/execution, the M12 `contract` command, and linux/amd64 + linux/arm64 Docker manifests.

Full publication record: `docs/history/V0.6.0-PUBLICATION.md`.

## Live / disaster-recovery state — still v0.5.0

The existing Arcane-managed production deployment and `xXDasGoGXx/OMV-Docker-Rebuild` disaster-recovery definition remain pinned to:

`mjmalleo/hostsleuth:0.5.0`

Publication of v0.6.0 is not proof of production rollout. The v0.6.0 live/recovery alignment is now the active next step.

## Live v0.5.0 acceptance

Post-redeploy acceptance confirmed:

- `/api/about` reports `v0.5.0`;
- `/api/snapshot` reports schema 4 and Docker mode;
- retained event history survived the redeploy, including events captured hours before the upgrade;
- Optional Safe Actions remain `enabled=false` and `available=false` in Docker/default mode;
- the Action Web/API remains loopback-only and returns HTTP 403 from LAN access;
- the service remains reachable at `192.168.2.181:8787`.

The Redacted Evidence Bundle was exercised with the published v0.5.0 amd64 binary against a disposable copy of the live API snapshot/events. Preview reported 100 events and no action audits. Export created only the expected bounded files, wrote the archive with mode `0600`, and all embedded SHA-256 checksums verified. No privileged Docker or live state-directory access was bypassed.

The temporary Arcane Cloudflare tunnel was shut down after acceptance and the disposable local acceptance directory was removed.

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

## Active next step — production / recovery v0.6.0 alignment

Stage the `OMV-Docker-Rebuild` image pin at 0.6.0 without merging ahead of production. Then redeploy the existing authenticated Arcane-managed HostSleuth project to `mjmalleo/hostsleuth:0.6.0`, verify version/schema/mode, retained state/events, Docker Safe Actions boundaries, and an Expected Endpoint Contract in the live UI/API. Merge recovery alignment only after live acceptance passes.

Do not restart a broad audit on continuation; use this handoff.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
