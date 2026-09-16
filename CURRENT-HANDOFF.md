# HostSleuth — Current Handoff

Last updated: 2026-09-16

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, read-only, local-first, single-host first, and deliberately small. It is not a generic monitoring platform.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Current `main` release source:

`365753ff513ff155d832df7119892090b903569d`

Completed milestones:

- M0 — repository foundation;
- M1 — deployable single-host MVP;
- M2 — deeper deterministic diagnosis;
- M3 — Product Experience;
- M3.4 — Public Container Distribution;
- M4 — package-change timeline.

Published stable release:

`v0.2.0`

Published Docker tags:

- `mjmalleo/hostsleuth:0.2.0`
- `mjmalleo/hostsleuth:latest`

Both public Docker tags resolved anonymously to OCI index digest:

`sha256:6892362d3f5fc6d30ae6bde7235ee976f169f1be8d42d181c55f53cf98600c91`

Platforms advertised by the public index:

- `linux/amd64`
- `linux/arm64`

## v0.2.0 publication closeout

Owner approval was obtained before publication.

Release branch:

`release/v0.2.0`

Release source commit:

`365753ff513ff155d832df7119892090b903569d`

Release workflow run:

`35128525862`

The release workflow completed successfully through:

- version validation;
- Go tests;
- native linux/amd64 and linux/arm64 release builds;
- `SHA256SUMS` generation;
- public GitHub `v0.2.0` release/tag creation;
- Docker Hub login;
- public multi-platform Docker push for `0.2.0` and `latest`.

GitHub `v0.2.0` release assets:

- `hostsleuth-linux-amd64`;
- `hostsleuth-linux-arm64`;
- `SHA256SUMS`.

Anonymous Docker Registry verification confirmed both `0.2.0` and `latest` are public, expose linux/amd64 and linux/arm64, and point to the same image index digest.

A separate runtime re-smoke of the published image was not forced after publication because the normal CI path had already passed Docker runtime smoke on the exact release source, the release workflow itself completed the final public push, and the available OMV administration path intentionally blocks raw Docker execution. Do not weaken those controls merely to repeat a test already covered by CI.

## M4 — package-change timeline

M4 makes the existing **Changes** timeline more useful without creating a package-management subsystem.

Delivered behavior:

- native Debian/Ubuntu collection reads local `dpkg` package history;
- `apt` history is a fallback when usable dpkg history is unavailable;
- current and recent rotated logs are read with bounded file sizes and bounded retained history;
- install, update, and remove records preserve package name, architecture, versions, and original package timestamp;
- package changes appear through the existing event pipeline as `category=package`;
- only newly observed package records become events;
- snapshot schema 2 establishes package-history awareness;
- upgrading from a schema-1 snapshot baselines existing package history instead of replaying historical records as new events;
- unsupported or unavailable logs remain quiet;
- collection is read-only.

### Docker boundary

The supported Docker deployment remains intentionally reduced-visibility. It does not mount host `apt` / `dpkg` logs, so host package-history evidence is unavailable there by default. M4 did not widen Docker host mounts merely to expose package logs.

Native installation remains the full-evidence path for package history.

## Existing live OMV deployment

The known-good live HostSleuth remains managed through Arcane on OMV host `192.168.2.181` using pinned image:

`mjmalleo/hostsleuth:0.1.0`

Deployment-specific state:

- UI/API: `http://192.168.2.181:8787`;
- persistent state: `/srv/docker/volumes/hostsleuth/data`;
- Compose path: `/srv/docker/volumes/compose/hostsleuth/compose.yaml`;
- trusted-LAN binding: `192.168.2.181:8787`.

`xXDasGoGXx/OMV-Docker-Rebuild` also remains pinned to the same v0.1.0 image so recovery matches the actual live deployment.

Do not upgrade the live OMV deployment merely to chase the new version number. M4's user-visible package-history capability is native-only under the current least-privilege Docker boundary.

## Accepted UI boundary

The Host Story UI from PR #21 remains accepted. Package events use the existing generic Changes timeline/category summary; M4 required no new dashboard or UI redesign.

Do not reopen broad UI exploration unless real use exposes a concrete defect.

## Active milestone — M5: configuration fingerprinting

M5 has one job: make **“what changed?”** more useful by recording that important configuration files changed, without storing their contents or secrets by default.

Initial bounded scope:

- native Linux first;
- fingerprint a small, explicit set of high-value configuration files that HostSleuth can safely read;
- record path plus non-secret metadata/fingerprint only;
- emit configuration-change events into the existing Changes timeline;
- baseline existing fingerprints on first M5-aware capture so upgrades do not create a fake backlog;
- missing/unreadable files remain truthful and quiet;
- preserve read-only behavior;
- no configuration editor, diff viewer, secret storage, remediation, watcher daemon, broad recursive `/etc` crawl, or settings framework.

M5 should be one bounded branch/PR with focused tests and one real-host acceptance pass, then merge and close if no concrete defect appears.

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
