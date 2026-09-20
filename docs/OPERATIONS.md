# HostSleuth Operations and Recovery Map

This document explains how HostSleuth is built, published, presented, deployed, and recovered.

Its purpose is not to duplicate implementation detail. Its purpose is to answer the future question:

> "How did I set this up, what depends on what, and how do I rebuild it safely?"

No secret values belong in this document or anywhere else in the repository.

## Source-of-truth hierarchy

Use these files in this order:

1. `CURRENT-HANDOFF.md` — exact current project/release/deployment state.
2. `ops/project.yaml` — machine-readable project/integration inventory.
3. `docs/OPERATIONS.md` — human operational and recovery map.
4. `README.md` — product and user-facing usage.
5. `SECURITY.md` — security boundaries.
6. `docs/history/` — historical implementation, release, and acceptance records.

Historical files are records of what was true at that time and should not be rewritten simply because the current release changes.

## System map

```text
                         GitHub repository
                      xXDasGoGXx/HostSleuth
                               |
            +------------------+------------------+
            |                  |                  |
            v                  v                  v
           CI             Stable release     Docker Hub metadata
     .github/workflows/    release/vX.Y.Z     docker/DOCKERHUB.md
          ci.yml                 |                  |
                                 v                  v
                         GitHub Release       mjmalleo/hostsleuth
                         native binaries       Overview/description
                                 |
                                 v
                          Docker build/push
                                 |
                                 v
                        mjmalleo/hostsleuth
                       X.Y.Z and latest tags
                                 |
               +-----------------+-----------------+
               |                                   |
               v                                   v
         Live deployment                     Recovery definition
         Arcane / Docker                  xXDasGoGXx/OMV-Docker-Rebuild
```

## GitHub workflows

### CI — `.github/workflows/ci.yml`

Purpose:
- format/vet/test/build regression coverage;
- JavaScript syntax checks;
- native Safe Action acceptance;
- native installer smoke;
- Docker build and runtime checks;
- platform and UI readiness checks.

Trigger:
- repository CI rules as defined in the workflow.

Credentials:
- no manually stored application credential should be required for ordinary CI.

Failure meaning:
- do not merge/release until the failure is understood.

### Release — `.github/workflows/release.yml`

Purpose:
- create native amd64/arm64 binaries;
- generate `SHA256SUMS`;
- publish the GitHub release/tag;
- build and push stable multi-platform Docker images.

Trigger:
- push to a branch named `release/vX.Y.Z`.

Stable behavior:
- a normal semantic version such as `v1.0.0` publishes:
  - GitHub release `v1.0.0`;
  - `mjmalleo/hostsleuth:1.0.0`;
  - `mjmalleo/hostsleuth:latest`.

Important:
- `latest` tracks the latest **stable release**, not the newest commit on `main`.

Credentials:
- GitHub Actions secret: `DOCKERHUB_TOKEN`.
- Docker Hub username is `mjmalleo`.
- the secret value is never stored in Git.

### Optional Docker Hub metadata sync — `.github/workflows/dockerhub-description.yml`

Purpose:
- keep the preferred Docker Hub short description and Overview text under version control;
- provide an optional future provider-side sync path if the owner wants it.

Source files:
- `docker/DOCKERHUB-SHORT.txt`;
- `docker/DOCKERHUB.md`.

Trigger:
- manual `workflow_dispatch` only.

Current state:
- **not required for normal operation or image publication**;
- the public Docker Hub short description was independently read during the 2026-09-19 reconciliation and the Git-tracked short-description source was aligned to that exact public value;
- automatic metadata publishing is intentionally not commissioned.

Credential when deliberately enabled later:
- GitHub Actions secret: `DOCKERHUB_METADATA_TOKEN`;
- keep it separate from `DOCKERHUB_TOKEN`, which remains dedicated to image publication;
- use a dedicated PAT with only the permissions required by the pinned metadata action.

If this optional workflow is manually enabled later and fails:
1. inspect the workflow log;
2. confirm the dedicated metadata credential exists;
3. confirm the Docker Hub repository is still `mjmalleo/hostsleuth`;
4. confirm metadata files satisfy provider size limits;
5. rerun only after the cause is understood.

## Credentials registry

Secret values are intentionally excluded.

| Credential | Stored at | Used by | Purpose |
| --- | --- | --- | --- |
| `DOCKERHUB_TOKEN` | GitHub Actions secret for HostSleuth | release workflow | publish stable Docker images |
| `DOCKERHUB_METADATA_TOKEN` | GitHub Actions secret only if optional metadata automation is enabled later | manual Docker Hub metadata workflow | optional future provider-side metadata sync |
| GitHub workflow token | supplied automatically by GitHub Actions | release workflow | create GitHub release/tag and attach artifacts |

Recovery rule:
- if a required active credential is lost, create/rotate it at the provider and replace the named GitHub Actions secret;
- do not create `DOCKERHUB_METADATA_TOKEN` merely for completeness; it is needed only if the owner deliberately enables the optional metadata-sync workflow.
- never put the token value in `README.md`, `CURRENT-HANDOFF.md`, `ops/project.yaml`, issues, PRs, or logs.

## Distribution endpoints

### GitHub

Repository:
- https://github.com/xXDasGoGXx/HostSleuth

Stable releases:
- https://github.com/xXDasGoGXx/HostSleuth/releases

What GitHub owns:
- source code;
- documentation;
- workflow definitions;
- native release artifacts;
- release tags;
- current/history/recovery documentation.

### Docker Hub

Repository:
- https://hub.docker.com/r/mjmalleo/hostsleuth

What Docker Hub owns:
- published container manifests/layers;
- versioned stable image tags;
- `latest` stable tag;
- public repository short description and Overview.

What it does **not** own:
- canonical storefront text. That lives in Git under `docker/`.

## Deployment modes

### Native Linux

Recommended when full host visibility matters.

Installer:
```bash
sudo ./scripts/install.sh
```

Default behavior:
- resolves the latest published GitHub release unless `HOSTSLEUTH_VERSION` is explicitly set;
- downloads the architecture-specific binary;
- verifies it against the release `SHA256SUMS`;
- installs the systemd unit;
- binds the Web UI to loopback by default.

### Docker

Recommended when containerized deployment convenience is more important than full host visibility.

Stable image:
```text
mjmalleo/hostsleuth:latest
```

Pinned production/recovery deployments should prefer a numeric stable tag.

Docker mode intentionally has reduced visibility and exposes no native systemd Safe Actions.

## Production and recovery

Current exact state belongs in `CURRENT-HANDOFF.md`.

The recovery repository is:

- `xXDasGoGXx/OMV-Docker-Rebuild`

Release ordering is deliberate:

1. source is accepted;
2. stable release is published and independently verified;
3. recovery update may be staged but remains unmerged;
4. live production is upgraded and accepted;
5. recovery definition is merged/aligned only after live acceptance.

This prevents the recovery definition from getting ahead of the last known-good production state.

## Rebuild from zero

If HostSleuth itself, Docker Hub metadata, or the live deployment had to be reconstructed:

### 1. Recover source

Clone:

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
```

Read:
- `CURRENT-HANDOFF.md`;
- `ops/project.yaml`;
- this document;
- `SECURITY.md`.

### 2. Validate source

Run:

```bash
go test ./...
go build -o hostsleuth ./cmd/hostsleuth
```

CI remains the broader acceptance gate.

### 3. Recover native distribution

The stable release workflow can reconstruct release binaries from an accepted release branch.

Do not repoint an existing immutable release to different source.

### 4. Recover Docker distribution

A stable release branch runs the release workflow and publishes both:
- the numeric stable image tag;
- `latest`.

Confirm both architectures and image labels before production rollout.

### 5. Recover Docker Hub storefront

The canonical text is already in:
- `docker/DOCKERHUB-SHORT.txt`;
- `docker/DOCKERHUB.md`.

The public storefront can be maintained manually from the Git-tracked source files. Automatic provider-side sync is optional. If the owner deliberately enables it later, create a dedicated `DOCKERHUB_METADATA_TOKEN` and manually dispatch **Sync Docker Hub metadata**.

### 6. Recover live deployment

Use the exact stable image documented in `CURRENT-HANDOFF.md` and the recovery repository.

Do not infer a production image from whatever happens to be newest on `main`.

### 7. Recover the recovery definition

Compare `OMV-Docker-Rebuild` against the accepted live image and only align it after the live deployment has passed acceptance.

## Documentation rule for future work

Any new external integration, publishing target, deployment path, privileged action, credential dependency, or recovery dependency is incomplete until all of the following are recorded:

- **what** it is;
- **why** it exists;
- **where** its configuration lives;
- **what triggers it**;
- **which named credentials it uses**;
- **what it depends on**;
- **how to validate it**;
- **how to roll it back or recreate it**;
- **what must never be stored in Git**.

For machine-readable inventory, update `ops/project.yaml`.

For operational explanation, update this file.

For exact present state, update `CURRENT-HANDOFF.md`.

For completed historical work, add or update the appropriate record under `docs/history/`.

## Disaster principle

A future administrator should be able to understand the architecture and rebuild the system using:

1. the Git repositories;
2. access to the external providers;
3. newly issued credentials with the documented secret names.

The design must not depend on remembering a chat conversation, an undocumented browser click, or a token that exists only on one machine.
