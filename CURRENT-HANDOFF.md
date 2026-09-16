# HostSleuth — Current Handoff

Last updated: 2026-09-16

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, read-only, local-first, single-host first, and deliberately small. It is not a generic monitoring platform.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Current `main`:

`39cb2847ea66cd04ddb9cb0f7a3111f5e659456b`

Completed milestones:

- M0 — repository foundation;
- M1 — deployable single-host MVP;
- M2 — deeper deterministic diagnosis;
- M3 — Product Experience;
- M3.4 — Public Container Distribution;
- M4 — package-change timeline;
- M5 — configuration fingerprinting.

Published stable release remains:

`v0.2.0`

Published Docker tags:

- `mjmalleo/hostsleuth:0.2.0`
- `mjmalleo/hostsleuth:latest`

Both public Docker tags resolved anonymously to OCI index digest:

`sha256:6892362d3f5fc6d30ae6bde7235ee976f169f1be8d42d181c55f53cf98600c91`

Platforms:

- `linux/amd64`
- `linux/arm64`

M5 is merged source capability and is **not** claimed to be present in the already-published v0.2.0 artifacts. A newer release remains a separate owner-controlled decision.

## M5 — configuration fingerprinting

M5 extends **what changed?** without storing configuration contents.

Delivered behavior:

- native Linux only;
- explicit default paths:
  - `/etc/hosts`
  - `/etc/fstab`
  - `/etc/ssh/sshd_config`
  - `/etc/docker/daemon.json`
  - `/etc/nftables.conf`
- snapshot stores only canonical path, state, SHA-256 fingerprint, and size for readable regular files;
- states are `present`, `missing`, or `unreadable`;
- file contents are not stored in snapshots or events;
- configuration appeared/disappeared/content-changed records reuse the existing Changes timeline as `category=configuration`;
- unreadable transitions remain quiet rather than becoming false changes;
- snapshot schema 3 establishes configuration-fingerprint awareness;
- schema 2 -> 3 baselines existing configuration fingerprints so the upgrade does not create a fake backlog;
- package-history events continue correctly across the schema bump;
- Docker mounts were not widened.

Explicitly not added:

- recursive `/etc` scanning;
- configuration editor;
- content diff viewer;
- secret storage;
- remediation;
- watcher daemon;
- alerts;
- settings framework.

## M5 validation

PR #26: `M5: add configuration fingerprint change evidence`.

Accepted code head before merge:

`6d79db5c682e633c4709893f017d35ff55368226`

CI run `35129730468` / #133 passed:

- formatting;
- `go vet`;
- focused configuration fingerprint, schema-baseline, and package-regression tests;
- Web JavaScript syntax;
- native build;
- Compose validation;
- linux/amd64 image build;
- linux/arm64 image build;
- Docker runtime smoke, API/snapshot, Web UI, self-diagnosis, and teardown.

Real OMV Debian acceptance used a temporary branch build and isolated state. It confirmed schema 3 collection of exactly five explicit candidates. `/etc/hosts`, `/etc/fstab`, `/etc/ssh/sshd_config`, and `/etc/nftables.conf` produced 64-character SHA-256 fingerprints plus size metadata; `/etc/docker/daemon.json` was truthfully reported `unreadable` without privilege escalation. The live HostSleuth deployment was not replaced.

HomeCommander's protected-path policy deliberately prevented synthetic writes using paths resembling `/etc/...`; those schema-baseline and change-generation cases are covered by the focused CI tests rather than weakening host safeguards.

## Existing live OMV deployment

The known-good live HostSleuth remains managed through Arcane on OMV host `192.168.2.181` using pinned image:

`mjmalleo/hostsleuth:0.1.0`

Deployment-specific state:

- UI/API: `http://192.168.2.181:8787`;
- persistent state: `/srv/docker/volumes/hostsleuth/data`;
- Compose path: `/srv/docker/volumes/compose/hostsleuth/compose.yaml`;
- trusted-LAN binding: `192.168.2.181:8787`.

`xXDasGoGXx/OMV-Docker-Rebuild` also remains pinned to v0.1.0 so recovery matches the actual live deployment.

Do not upgrade the live OMV deployment merely to chase the release number. M4/M5 evidence remains native-only under the current least-privilege Docker boundary.

## Accepted UI boundary

The Host Story UI from PR #21 remains accepted. Package and configuration events use the existing generic Changes timeline/category summary. M5 required no UI redesign.

Do not reopen broad UI exploration unless real use exposes a concrete defect.

## What is next

M5 is closed after PR #26 merge. **Do not automatically activate another capability.** Choose one deliberately when work resumes.

Current candidates remain in `TO-DO.md`:

- bounded TLS/certificate diagnosis;
- general redaction/threat-model work before richer exports.

Do not create a new GitHub release/tag or Docker tag/image without explicit owner approval.

## Repository reading order

When resuming, read:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`
4. `docs/history/DEVELOPMENT-HISTORY.md`
5. `docs/design/HOST-STORY-UI.md`
6. `.github/workflows/ci.yml`
7. `.github/workflows/release.yml`
8. `Dockerfile`
9. `compose.yaml`
