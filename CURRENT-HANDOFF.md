# HostSleuth — Current Handoff

Last updated: 2026-09-16

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

The product definition and user-facing overview live in `README.md`. The active roadmap lives in `TO-DO.md`.

## Milestone state

- **M0 — Repository foundation:** complete.
- **M1 — Single-host deployable MVP:** complete.
- **M2 — Deeper deterministic diagnosis:** complete.
- **M3 — Product Experience:** complete and accepted.
- **M3.4 — Public Container Distribution:** active on branch `m3.4/public-container-distribution`.
- **M4 — Package-change timeline:** remains the next application-feature milestone after M3.4.

M3.4 does not add HostSleuth functionality. Its only purpose is to make the existing supported Docker deployment easy for normal users to pull, run, version, and update.

## Authoritative repository state before M3.4

M3.4 started from current `main` at:

`2ab8189b37f45712dca3832d81e96a990764be18`

That commit is the post-M3 documentation update that explains optional trusted-LAN binding while preserving loopback-only as the public default.

The only verified published GitHub release at M3.4 start was:

`v0.1.0-alpha.1`

That release predates completed M3 and must not be treated as the stable M3 container release.

## Existing container foundation

Before M3.4, HostSleuth already had:

- multi-stage Docker build;
- static CGO-disabled binary;
- TARGETOS/TARGETARCH support;
- Alpine runtime;
- Linux amd64 and arm64 CI image builds;
- actual Docker Compose runtime smoke testing;
- host network/PID/UTS namespace sharing;
- persistent HostSleuth state;
- narrow read-only `/etc/os-release` mount;
- read-only Docker socket mount;
- read-only root filesystem;
- all Linux capabilities dropped;
- `no-new-privileges`;
- truthful reduced Docker visibility.

M3.4 must not weaken those boundaries to gain additional visibility.

## M3.4 implementation currently on branch

The active branch now contains:

- `compose.yaml` changed from local `build:` to `image:` consumption;
- confirmed public image name `mjmalleo/hostsleuth`;
- `HOSTSLEUTH_IMAGE` override so explicit stable tags and local developer builds remain easy;
- README pull/Compose usage;
- README direct `docker run` usage matching the supported Compose security/runtime settings;
- explicit developer source-build instructions;
- `.github/workflows/docker-publish.yml` for release-gated Docker Hub publication;
- stable-tag policy where only plain `vX.Y.Z` GitHub releases publish images;
- Docker image tags `<X.Y.Z>` and `latest` for stable releases;
- linux/amd64 + linux/arm64 multi-platform publication;
- Docker Hub authentication via GitHub Actions `DOCKERHUB_TOKEN` secret and `DOCKERHUB_NAMESPACE` variable only;
- CI smoke changed to build a local image first, then make Compose consume that image without `--build`;
- CI checks `/api/about`, `/api/snapshot`, Web UI, and self-diagnosis from the image-consumption path.

No Docker Hub repository, Docker Hub image/tag, GitHub release, or release Git tag has been created by M3.4 work so far.

## Version recommendation

The recommended first stable version is **v0.1.0**.

Reasoning:

- the project already established the `0.1.0` line with `v0.1.0-alpha.1`;
- completed M1/M2/M3 now represent the matured form of that initial product line;
- moving directly to `v0.2.0` would imply a new minor feature generation rather than stabilization of the existing 0.1 line;
- `v0.1.0` provides the simplest SemVer transition from the existing alpha to the first stable release.

Do not create the tag/release until the owner explicitly approves publication.

## Docker Hub repository

Confirmed repository/name:

`mjmalleo/hostsleuth`

The Docker Hub personal namespace is `mjmalleo`. Keep that exact namespace in Compose, README examples, and GitHub Actions repository variable `DOCKERHUB_NAMESPACE`.

## One-time owner-controlled setup still required

Before the first public image can be published:

1. Create the public Docker Hub repository `hostsleuth` under namespace `mjmalleo`.
2. Generate a Docker Hub access token with only the permissions needed to push this repository.
3. Add that token to GitHub Actions as repository secret `DOCKERHUB_TOKEN`.
4. Add `mjmalleo` as GitHub Actions repository variable `DOCKERHUB_NAMESPACE`.

Never paste the token into chat or commit it to Git.

## Publication boundary

Do not perform any of these actions without explicit owner approval at that point:

- create the Docker Hub repository;
- create/push a public Docker image or Docker Hub tag;
- create a stable GitHub release;
- create/push the public release Git tag;
- publish any package or artifact beyond the already-public repository branch/PR work.

The release workflow already in the repository publishes native GitHub release binaries when a `release/v*` branch is pushed. Because that is a public publication action, do not create a stable release branch such as `release/v0.1.0` until approval is given.

## Validation still required

Before M3.4 can be called complete:

- PR CI must pass normal Go checks;
- Compose config must validate;
- linux/amd64 and linux/arm64 image builds must pass;
- runtime image-consumption smoke must pass;
- `/api/about`, `/api/snapshot`, Web UI, and diagnosis must pass;
- after explicit publication approval, the actual public multi-arch image must be pulled and run in the same style documented for normal users;
- only after that validation should the separate OMV disaster-recovery repository be updated to consume the published image.

## Product-direction rule

HostSleuth stays easy to deploy, easy to understand, and intentionally small. M3.4 must not introduce Watchtower behavior, automatic updating, Kubernetes, Swarm, privileged mode, broad host mounts/capabilities, authentication redesign, remediation, M4 package work, or unrelated product features.

## Supported deployment paths

### Native Linux — recommended

Native Linux provides the fullest visibility into systemd, host filesystems, network/listener state, and Docker inventory while keeping the Web UI loopback-only by default.

### Docker — convenient, reduced visibility

The Docker deployment uses host network/PID/UTS namespaces, persistent state, a narrow host OS-release mount, and Docker socket access for inventory. It intentionally reports unavailable evidence as unavailable instead of escalating privileges.

The public default remains `127.0.0.1:8787`. Optional trusted-LAN binding should use one specific host address rather than `0.0.0.0`.

## Repository source of truth

Repository: `xXDasGoGXx/HostSleuth`

Read in this order when resuming:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`
4. `.github/workflows/ci.yml`
5. `.github/workflows/docker-publish.yml`
6. `.github/workflows/release.yml`
7. `Dockerfile`
8. `compose.yaml`

Use `docs/history/DEVELOPMENT-HISTORY.md` for completed historical details.
