# HostSleuth — Current Handoff

Last updated: 2026-09-16

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, read-only, local-first, single-host first, and deliberately small. It is not a generic monitoring platform.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Completed milestones:

- M0 — repository foundation;
- M1 — deployable single-host MVP;
- M2 — deeper deterministic diagnosis;
- M3 — Product Experience;
- M3.4 — Public Container Distribution;
- M4 — package-change timeline.

Published stable release remains `v0.1.0`, sourced from commit `bf52c51fdded40a73684171a7feb078557b9f0d5`.

Published Docker tags remain:

- `mjmalleo/hostsleuth:0.1.0`
- `mjmalleo/hostsleuth:latest`

M4 is merged source capability and is **not** claimed to be present in the already-published `v0.1.0` artifacts. Creating a newer release/tag/image remains a separate owner-controlled publication decision.

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

Explicitly not added:

- package install/update/remove actions;
- update buttons;
- repository management;
- package alerts;
- a package dashboard;
- a settings framework.

### Docker boundary

The supported Docker deployment remains intentionally reduced-visibility. It does not mount host `apt` / `dpkg` logs, so host package-history evidence is unavailable there by default. M4 did not widen Docker host mounts merely to expose package logs.

Native installation remains the full-evidence path.

## M4 validation

PR #24: `M4: add package-change timeline evidence`.

Final code-head CI run `35127119772` passed:

- formatting;
- `go vet`;
- Go tests, including apt/dpkg parsing and schema-upgrade baseline regression;
- Web JavaScript syntax;
- native build;
- Compose validation;
- linux/amd64 image build;
- linux/arm64 image build;
- Docker runtime smoke, API, Web UI, self-diagnosis, and teardown.

Real-host acceptance was performed on the actual OMV Debian environment from exact code head `e12686b07f7725840c6fb58f00060b8831faf820` using a temporary user-space Go toolchain and isolated state.

Acceptance confirmed:

- real `/var/log/dpkg.log` history parsed successfully;
- retained history was bounded at 200 package records;
- a real Docker CE version update was parsed with correct old/new versions;
- schema-1 -> schema-2 first capture emitted **zero** historical package events;
- one synthetic install appended only to a copied dpkg log produced exactly one `package` event;
- the real `/var/log/dpkg.log` SHA-256 was unchanged before and after acceptance;
- no real package was installed, removed, or upgraded;
- the live HostSleuth deployment was not replaced for M4 acceptance.

## Existing v0.1.0 deployment state

The known-good live HostSleuth remains managed through Arcane on OMV host `192.168.2.181` using pinned image `mjmalleo/hostsleuth:0.1.0`.

Deployment-specific state:

- UI/API: `http://192.168.2.181:8787`;
- persistent state: `/srv/docker/volumes/hostsleuth/data`;
- Compose path: `/srv/docker/volumes/compose/hostsleuth/compose.yaml`;
- trusted-LAN binding: `192.168.2.181:8787`.

`xXDasGoGXx/OMV-Docker-Rebuild` also pins the same v0.1.0 image. M4 did not change this production deployment or recovery definition because no newer public image has been approved/published.

## Accepted UI boundary

The Host Story UI from PR #21 remains accepted. Package events use the existing generic Changes timeline/category summary; M4 required no new dashboard or UI redesign.

Do not reopen broad UI exploration unless real use exposes a concrete defect.

## What is next

M4 is closed after PR #24 merge. **Do not automatically activate another capability.** The candidate list remains in `TO-DO.md`; choose one deliberately when work resumes.

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
