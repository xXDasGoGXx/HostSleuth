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
- **M3.4 — Public Container Distribution:** active on branch `m3.4/public-container-distribution` / PR #20.
- **M4 — Package-change timeline:** remains the next application-feature milestone after M3.4.

M3.4 does not add HostSleuth functionality. Its only purpose is to make the existing supported Docker deployment easy for normal users to pull, run, version, and update.

## Authoritative repository state before M3.4

M3.4 started from `main` at:

`2ab8189b37f45712dca3832d81e96a990764be18`

The only verified published GitHub release at M3.4 start was `v0.1.0-alpha.1`, which predates completed M3 and is not the stable M3 container release.

## Existing container foundation

Before M3.4, HostSleuth already had a multi-stage static Go/Alpine image, TARGETOS/TARGETARCH support, linux/amd64 and linux/arm64 CI builds, runtime Compose smoke testing, persistent state, host network/PID/UTS namespaces, narrow host mounts, a read-only root filesystem, all Linux capabilities dropped, `no-new-privileges`, and truthful reduced Docker visibility.

M3.4 must not weaken those boundaries to gain additional visibility.

## M3.4 implementation currently on branch

The active branch now contains:

- `compose.yaml` changed from local `build:` to `image:` consumption;
- confirmed public image name `mjmalleo/hostsleuth`;
- `HOSTSLEUTH_IMAGE` override so explicit stable tags and local developer builds remain easy;
- README pull/Compose usage;
- README direct `docker run` usage matching the supported Compose security/runtime settings;
- explicit developer source-build instructions;
- the existing `.github/workflows/release.yml` extended to publish the Docker image in the same release workflow;
- plain stable `vX.Y.Z` versions publish Docker tags `<X.Y.Z>` and `latest`;
- prerelease version branches may create prerelease GitHub releases but do not publish Docker tags;
- linux/amd64 + linux/arm64 multi-platform Docker publication;
- Docker Hub authentication via GitHub Actions `DOCKERHUB_TOKEN` secret only;
- fixed Docker Hub namespace `mjmalleo` in the release workflow;
- OCI metadata for title, source, revision, version, and MIT license;
- release reruns refuse to continue if an existing release tag points at a different commit;
- CI smoke builds a local image first, then makes Compose consume that image without `--build`;
- CI checks `/api/about`, `/api/snapshot`, Web UI, and self-diagnosis from the image-consumption path.

A separate `release: published` Docker workflow was deliberately removed after confirming GitHub's `GITHUB_TOKEN` recursion rule: the existing release workflow creates GitHub Releases with `github.token`, and events produced by that token do not generally trigger another workflow. Keeping Docker publication inside the same release workflow avoids adding another GitHub credential solely to chain workflows.

No Docker Hub repository, Docker Hub image/tag, stable GitHub release, or stable release Git tag has been created by M3.4 work so far.

## Validation result so far

GitHub Actions CI run 98 passed on the confirmed-namespace implementation head before the final release-pipeline integration changes:

- formatting passed;
- `go vet` passed;
- Go tests passed;
- Web UI JavaScript syntax passed;
- native Go build passed;
- Compose validation passed;
- linux/amd64 image build passed;
- linux/arm64 image build passed;
- image-consumption Docker smoke passed;
- `/api/about`, `/api/snapshot`, Web UI, and self-diagnosis passed;
- clean teardown passed.

The final PR head must remain green after the release-pipeline/documentation changes before merge.

## Version recommendation

The recommended first stable version is **v0.1.0**.

This continues the already-established `v0.1.0-alpha.1` line after completed M1/M2/M3 instead of implying a new feature generation with `v0.2.0`.

Do not create the stable release branch/tag/release until the owner explicitly approves publication.

## Docker Hub repository

Confirmed repository/name:

`mjmalleo/hostsleuth`

Keep that exact namespace in Compose, README examples, and the release workflow.

## One-time owner-controlled setup still required

Before the first public image can be published:

1. Create the public Docker Hub repository `hostsleuth` under namespace `mjmalleo`.
2. Generate a Docker Hub access token with only the permissions needed to push this repository.
3. Add that token to GitHub Actions as repository secret `DOCKERHUB_TOKEN`.

Never paste the token into chat or commit it to Git.

## Publication boundary

Do not perform any of these actions without explicit owner approval at that point:

- create the Docker Hub repository;
- create/push a public Docker image or Docker Hub tag;
- create `release/v0.1.0` or another public release branch intended to publish;
- create a stable GitHub release or release Git tag;
- publish any package or artifact beyond the already-public repository branch/PR work;
- change the known-good live OMV deployment.

## Exact remaining M3.4 sequence

1. Confirm the final PR head is green.
2. Complete the one-time Docker Hub repository/token/GitHub secret setup.
3. Merge PR #20 to `main` after final validation.
4. Re-check `main` and the release workflow after merge.
5. Stop and obtain explicit owner approval to publish `v0.1.0`.
6. After approval, create `release/v0.1.0` from the exact accepted `main` commit. The release workflow will test, build native amd64/arm64 binaries and checksums, create the GitHub release/tag, then build and push Docker linux/amd64 + linux/arm64 images as `mjmalleo/hostsleuth:0.1.0` and `mjmalleo/hostsleuth:latest`.
7. Verify the real Docker Hub tags/manifest and pull both tags as a normal user would.
8. Start the published image and re-check `/api/about`, `/api/snapshot`, Web UI, and diagnosis.
9. Only after public-image validation, update the separate OMV disaster-recovery repository and, with owner approval, convert the live OMV/Arcane deployment from GitHub source-build to image-pull use.
10. Record M3.4 completion in handoff/history and then begin M4.

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
5. `.github/workflows/release.yml`
6. `Dockerfile`
7. `compose.yaml`

Use `docs/history/DEVELOPMENT-HISTORY.md` for completed historical details.
