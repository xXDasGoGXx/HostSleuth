# HostSleuth — Current Handoff

Last updated: 2026-09-16

## What HostSleuth is

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

HostSleuth is not a generic monitoring platform. Keep it evidence-first, read-only, local-first, and deliberately small.

## Current authoritative state

Repository: `xXDasGoGXx/HostSleuth`

Stable release source commit:

`bf52c51fdded40a73684171a7feb078557b9f0d5`

Published stable release:

`v0.1.0`

Published Docker image:

- `mjmalleo/hostsleuth:0.1.0`
- `mjmalleo/hostsleuth:latest`

M0, M1, M2, M3, and M3.4 are complete. M4 is now the active application milestone.

## M3.4 closeout — Public Container Distribution

M3.4 is complete.

Delivered:

- Compose consumes the published image by default while preserving a local developer build override;
- direct `docker run` instructions match the supported security/runtime model;
- stable release workflow publishes native Linux amd64/arm64 binaries and checksums;
- stable release workflow publishes one multi-platform Docker image for linux/amd64 and linux/arm64;
- plain stable releases publish both the numeric Docker tag and `latest`;
- prerelease branches do not move Docker `latest`;
- Docker Hub credentials remain only in GitHub Actions secret `DOCKERHUB_TOKEN`;
- Docker security model remains unchanged: host network/PID/UTS, read-only root, all capabilities dropped, `no-new-privileges`, narrow host mounts, persistent state, read-only Docker socket;
- public Web UI default remains `127.0.0.1:8787`;
- optional LAN exposure remains an explicit specific-address override.

## v0.1.0 publication and validation

Release workflow run `35124357891` completed successfully from the exact stable source commit.

GitHub release `v0.1.0` contains:

- `hostsleuth-linux-amd64`;
- `hostsleuth-linux-arm64`;
- `SHA256SUMS`.

Docker registry validation confirmed anonymous public retrieval and one OCI image index containing:

- `linux/amd64`;
- `linux/arm64`.

Both `mjmalleo/hostsleuth:0.1.0` and `mjmalleo/hostsleuth:latest` resolved to the same platform manifests at publication.

A one-off ephemeral GitHub runner then tested the public image exactly as a normal consumer would. Run `35124853545` passed:

- anonymous pull of `0.1.0` and `latest`;
- startup with the supported Docker runtime/security settings;
- `/api/about` reporting `v0.1.0`;
- `/api/snapshot` reporting Docker mode;
- Web UI load;
- reachable/high-confidence self-diagnosis;
- clean teardown.

The temporary validation workflow was removed afterward rather than becoming permanent project machinery.

## Live OMV deployment

HostSleuth remains managed through Arcane on OMV host `192.168.2.181`.

Deployment-specific state remains:

- UI/API: `http://192.168.2.181:8787`;
- persistent state: `/srv/docker/volumes/hostsleuth/data`;
- Compose path: `/srv/docker/volumes/compose/hostsleuth/compose.yaml`;
- trusted-LAN binding: `192.168.2.181:8787`.

The owner redeployed the Arcane project from the published pinned image `mjmalleo/hostsleuth:0.1.0` after public consumer validation.

Post-redeploy live verification passed:

- `/api/about` returned `v0.1.0`;
- `/api/snapshot` returned Docker mode with live host evidence;
- the Web UI loaded;
- diagnosis of `192.168.2.181:8787` returned `target is reachable` with high confidence.

At validation time the live snapshot exposed 328 listeners and 22 Docker containers. Those counts are observations from that moment, not product expectations.

The separate recovery repository `xXDasGoGXx/OMV-Docker-Rebuild` was also updated through PR #2 so its HostSleuth Compose definition pins `mjmalleo/hostsleuth:0.1.0` while preserving the existing LAN bind, state path, namespaces, mounts, and security settings.

## Accepted Host Story UI

PR #21 is the accepted v0.1.0 UI direction and is frozen unless real use exposes a concrete defect.

It makes existing evidence more useful without expanding HostSleuth into a monitoring platform:

- Overview presents a plain-language Host Story;
- Reachability Surface exposes listeners by bind scope;
- listener rows feed the deterministic Diagnose workflow;
- Diagnose shows an evidence path while retaining raw checks/evidence;
- recent host changes are context only, never claimed causality;
- Host exposes routes, listeners, Docker image/status/ports/networks, interfaces, filesystems, and host facts already collected;
- Changes includes a compact category summary while preserving the existing timeline.

Do not reopen broad UI exploration unless real use exposes a concrete problem.

## Active milestone — M4: package-change timeline

M4 has one job: make **“what changed?”** more useful with package install/update/remove history.

Initial scope:

- Debian/Ubuntu first;
- read existing local `apt` / `dpkg` logs;
- normalize package install/update/remove changes into the existing event timeline;
- keep collection read-only and local-first;
- unsupported platforms remain truthful and quiet;
- no package-management actions, update buttons, repository management, alerts, or new settings framework.

M4 should be implemented as one bounded branch/PR, validated with focused tests plus one real-host acceptance pass, then merged and closed. Do not turn it into an open-ended package subsystem.

## Repository reading order

When resuming, read:

1. `README.md`
2. `CURRENT-HANDOFF.md`
3. `TO-DO.md`
4. `docs/design/HOST-STORY-UI.md`
5. `docs/history/DEVELOPMENT-HISTORY.md`
6. `.github/workflows/ci.yml`
7. `.github/workflows/release.yml`
8. `Dockerfile`
9. `compose.yaml`
