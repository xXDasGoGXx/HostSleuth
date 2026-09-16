# HostSleuth — Current Handoff

Last updated: 2026-09-16

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

HostSleuth is not a generic monitoring platform. The product stays evidence-first, read-only, local-first, and deliberately small.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Current accepted `main`:

`3aa3cfbbb557d81d3b57dfb07ca43f2cdd2634a8`

This includes:

- completed M0, M1, M2, and M3;
- M3.4 public-container-distribution foundation from PR #20;
- the accepted Host Story evidence-first Web UI from PR #21.

No stable `v0.1.0` GitHub release/tag or Docker Hub image has been published yet.

## M3.4 — Public Container Distribution

M3.4 adds no new HostSleuth backend capability. Its purpose is to make the existing supported Docker deployment easy for normal users to pull, run, version, and update.

The accepted implementation now on `main` includes:

- `compose.yaml` using a published-image default rather than a local build;
- confirmed public image name `mjmalleo/hostsleuth`;
- `HOSTSLEUTH_IMAGE` override for pinned stable tags and local developer builds;
- documented Compose pull/update workflow;
- documented direct `docker run` workflow;
- preserved local source-build workflow;
- existing Docker security model unchanged: host network/PID/UTS, read-only root, all capabilities dropped, `no-new-privileges`, narrow host mounts, persistent state, read-only Docker socket;
- public Web UI default remains `127.0.0.1:8787`;
- optional LAN binding remains explicit and should use one specific trusted host address rather than `0.0.0.0`;
- stable Docker publication integrated directly into `.github/workflows/release.yml`;
- plain stable `vX.Y.Z` releases publish Docker tags `<X.Y.Z>` and `latest`;
- prerelease versions may create GitHub prereleases but do not publish Docker tags;
- Docker publication targets `linux/amd64` and `linux/arm64`;
- Docker Hub authentication uses GitHub Actions secret `DOCKERHUB_TOKEN` only;
- OCI image metadata is attached for source, revision, version, title, and MIT license.

A separate `release: published` Docker workflow was intentionally removed because the GitHub Release is created by the release workflow using `GITHUB_TOKEN`; GitHub suppresses most follow-on workflow events created by that token. Keeping native release and Docker publication in one workflow avoids a fragile chained release path.

## Accepted Host Story UI

PR #21 replaced the clean-but-sparse M3 presentation with one bounded evidence-first UI pass. This is the v0.1.0 UI direction and is now frozen unless a concrete defect appears.

The accepted UI uses only evidence HostSleuth already collected:

- Overview tells a plain-language Host Story instead of leading with generic counters;
- Reachability Surface shows listeners by bind scope: loopback, wildcard, or specific-address;
- listener rows can pre-fill the existing deterministic Diagnose workflow;
- listening sockets are explicitly treated as bind evidence, not proof of remote reachability;
- Diagnose shows an ordered evidence path while retaining the exact raw checks/evidence;
- recent host changes beside diagnosis are clearly labeled as context, never claimed causality;
- Host exposes routes, listeners, Docker image/status/ports/networks, interfaces, filesystems, and host facts already present in the snapshot;
- Changes includes a compact category summary while leaving the existing timeline authoritative.

No new collectors, APIs, storage, agents, SNMP, resource-graph dashboarding, alerts, remediation, AI, or multi-host architecture were added.

See `docs/design/HOST-STORY-UI.md` for the accepted design boundary.

## Validation completed

M3.4 distribution PR #20 final CI run 105 passed before merge:

- formatting;
- `go vet`;
- Go tests;
- Web JavaScript syntax;
- native build;
- Compose validation;
- linux/amd64 image build;
- linux/arm64 image build;
- image-consumption Docker smoke;
- `/api/about`;
- `/api/snapshot`;
- Web UI;
- self-diagnosis;
- clean teardown.

Host Story UI PR #21 final CI run 110 also passed after the evidence-accuracy wording fix, including the Docker runtime smoke and both image architectures.

## Version

Recommended first stable release: **v0.1.0**.

This continues the already-established `v0.1.0-alpha.1` line after completed M1/M2/M3 rather than implying a new minor feature generation with `v0.2.0`.

## Docker Hub target

Confirmed public image:

`mjmalleo/hostsleuth`

Expected first stable tags:

- `mjmalleo/hostsleuth:0.1.0`
- `mjmalleo/hostsleuth:latest`

## One-time owner-controlled setup still required

Before first publication:

1. Create the public Docker Hub repository `hostsleuth` under namespace `mjmalleo`.
2. Generate a Docker Hub access token with only the permissions needed to push the repository.
3. Add that token to GitHub repository Actions secrets as `DOCKERHUB_TOKEN`.

Never paste the token into chat or commit it to Git.

## Publication boundary

Do not perform these actions until the owner explicitly approves publication at that point:

- create `release/v0.1.0`;
- create the stable GitHub release/tag;
- push Docker Hub image/tag content;
- change the known-good live OMV deployment.

Creating the Docker Hub repository and GitHub secret is setup; creating `release/v0.1.0` is the publication trigger.

## Exact remaining endgame

1. Finish this release-prep documentation sync and merge it to `main`.
2. Complete the one-time Docker Hub repository/token/GitHub secret setup.
3. Re-check the exact final `main` SHA and release workflow.
4. Present that exact release state and obtain owner approval to publish `v0.1.0`.
5. Create `release/v0.1.0` from the exact accepted `main` commit.
6. Let the single release workflow test, build native amd64/arm64 binaries/checksums, create the GitHub release/tag, and publish Docker linux/amd64 + linux/arm64 images.
7. Verify the Docker Hub manifest and both `0.1.0` / `latest` tags.
8. Pull and run the public image exactly as a normal user would; re-check `/api/about`, `/api/snapshot`, Web UI, and diagnosis.
9. Only after public-image validation, update `xXDasGoGXx/OMV-Docker-Rebuild` and, with owner approval, convert the live OMV/Arcane HostSleuth deployment from GitHub source-build to image-pull use.
10. Record M3.4 completion in handoff/history.
11. Begin M4 package-change timeline. Do not reopen UI exploration unless real use exposes a concrete problem.

## Known OMV deployment context

Current known-good HostSleuth deployment is on OMV host `192.168.2.181`, managed through Arcane, with trusted-LAN UI binding `192.168.2.181:8787`, persistent state under `/srv/docker/volumes/hostsleuth/data`, and Compose under `/srv/docker/volumes/compose/hostsleuth/compose.yaml`.

Do not change that live deployment before the real public image is validated.

## Next application capability — M4

M4 remains the package-change timeline:

- bounded Debian/Ubuntu `apt` / `dpkg` history first;
- install/update/remove changes added to the existing event timeline;
- read-only and local-first;
- unsupported platforms remain truthful and quiet;
- no package management, update actions, alerting, repository management, or settings expansion.

## Repository reading order

When resuming, read:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`
4. `docs/design/HOST-STORY-UI.md`
5. `.github/workflows/ci.yml`
6. `.github/workflows/release.yml`
7. `Dockerfile`
8. `compose.yaml`

Use `docs/history/DEVELOPMENT-HISTORY.md` for completed historical detail.
