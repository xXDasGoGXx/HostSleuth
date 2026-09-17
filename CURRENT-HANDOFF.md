# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Stable / public / live / recovery state — v0.5.0

Stable public release:

`v0.5.0`

Exact published source:

`04a53f8f0f3f48f7118a9ee9a688820cc000a340`

GitHub release workflow:

`35281793336` — success

Published Docker tags:

- `mjmalleo/hostsleuth:0.5.0`
- `mjmalleo/hostsleuth:latest`

Both resolve to OCI index:

`sha256:a17325980d5e9ec9760f9003aa8a9490962bb06ffbbd5a393a0ad31a218e480d`

The live Arcane-managed OMV deployment and the `xXDasGoGXx/OMV-Docker-Rebuild` disaster-recovery definition are now both aligned on:

`mjmalleo/hostsleuth:0.5.0`

Recovery alignment PR #5 merged at:

`9387d85acef8d19d913458cf45d48a765b6b8299`

Full publication record: `docs/history/V0.5.0-PUBLICATION.md`.

Full production/recovery alignment record: `docs/history/V0.5.0-PRODUCTION-ALIGNMENT.md`.

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

## Current milestone boundary

v0.5.0 publication and production/recovery rollout are complete.

Do not automatically start a new feature milestone merely because this boundary is clean. The research backlog in `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md` remains design input only until the owner selects the next milestone.

Do not restart a broad audit on continuation; use this handoff.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
