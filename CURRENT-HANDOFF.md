# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Stable/public state — v0.5.0

Stable public release:

`v0.5.0`

Exact published source:

`04a53f8f0f3f48f7118a9ee9a688820cc000a340`

GitHub release workflow run:

`35281793336` — success

Published Docker tags:

- `mjmalleo/hostsleuth:0.5.0`
- `mjmalleo/hostsleuth:latest`

Both resolve to OCI index:

`sha256:a17325980d5e9ec9760f9003aa8a9490962bb06ffbbd5a393a0ad31a218e480d`

Platform manifests:

- linux/amd64 — `sha256:ed8fb39f4aa656e95a5cd5cfecb454c5861a77d610bbc8c3b1d27b61292dc8a7`
- linux/arm64 — `sha256:8ec337d3f0a42576e86fe08cdb69f3725414f7e1255c4b7f759841a522943d72`

GitHub release assets:

- `hostsleuth-linux-amd64` — `sha256:2626df7ae8d7ad7926b68c2be1b22beb6b489c3128e4cc15bf0e5bcde9f1960d`
- `hostsleuth-linux-arm64` — `sha256:3088738b5aad3dba76def866ec4f5aeac3eef7754a8a6107bfc175841cb1dc77`
- `SHA256SUMS` — `sha256:e758930e892e5f6d64952b11cbb9664b37ad05569aa8398d9c0799b9b8f8653e`

Independent OMV verification downloaded the public amd64 binary, validated it against `SHA256SUMS`, confirmed `v0.5.0 (04a53f8f0f3f)`, confirmed the Evidence Bundle CLI is present, and anonymously verified the Docker index/manifests above.

Full publication record: `docs/history/V0.5.0-PUBLICATION.md`.

## Live / disaster-recovery state — still v0.4.0

The existing Arcane-managed production deployment and `xXDasGoGXx/OMV-Docker-Rebuild` disaster-recovery definition remain pinned to:

`mjmalleo/hostsleuth:0.4.0`

The prior v0.4.0 recovery alignment merge is:

`ad2bd53ca3a469272eba6343c03936b7c1a04bc0`

Do not treat publication of v0.5.0 as proof that production has been upgraded. Production/recovery rollout is the active next step.

## Redacted Evidence Bundle — COMPLETE, MERGED, AND PUBLISHED

Owner accepted the threat-model/redaction contract in:

`docs/design/REDACTED-EVIDENCE-BUNDLE.md`

PR #40 merged on 2026-09-17 at:

`f1d756fa42baf71d9b762127b8d5d6ef6b77b796`

Closeout record:

`docs/history/REDACTED-EVIDENCE-BUNDLE.md`

The bounded first version is now part of stable v0.5.0.

CLI:

```text
hostsleuth evidence preview
hostsleuth evidence export [--output PATH]
```

The first exporter includes only the current Snapshot, at most 200 recent Events, and at most 100 Safe Action audit records when present. It uses deterministic bundle-local pseudonymization and credential/private-key scrubbing, emits manifest/checksum evidence, writes owner-only local archives, and excludes raw journal text, raw action command output, arbitrary file/configuration contents, action confirmation material, cloud upload, and automatic sharing.

Redaction lowers disclosure risk but cannot guarantee anonymity. Bundles must still be reviewed before sharing.

## Active next step — production / recovery v0.5.0 alignment

Resume in this order:

1. Stage the `OMV-Docker-Rebuild` HostSleuth image pin from `0.4.0` to `0.5.0` in a PR, but do not merge it ahead of production.
2. Redeploy the existing Arcane-managed HostSleuth project to `mjmalleo/hostsleuth:0.5.0` through the supported authenticated Arcane path.
3. Verify live `/api/about`, `/api/snapshot`, LAN reachability, retained events/state, and Docker-mode Safe Actions boundary.
4. Exercise the published v0.5.0 Evidence Bundle CLI against retained HostSleuth state with a safe local output path and inspect the preview/archive behavior.
5. Only after live acceptance succeeds, merge the recovery pin and record final alignment.

Do not restart a broad audit on continuation; use this handoff.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
