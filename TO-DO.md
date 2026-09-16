# HostSleuth — TO-DO

This file stays intentionally short. Completed milestone history belongs in `docs/history/DEVELOPMENT-HISTORY.md`.

## Completed

- [x] M0 — repository foundation.
- [x] M1 — deployable single-host MVP.
- [x] M2 — deeper deterministic diagnosis.
- [x] M3.1 — modern responsive Web UI.
- [x] M3.2 — usability and installation clarity.
- [x] M3.3 — supported Docker Compose deployment with amd64/arm64 CI validation.
- [x] Collective product review after M3.
- [x] Real-host native M3 acceptance on Debian 13 using a side-by-side build with isolated temporary state; the known-good installed service was not replaced.
- [x] Runtime-smoke-test the supported Docker Compose deployment on an authorized ephemeral Linux Docker host: Docker mode, Web UI, self-diagnosis, and teardown all passed.
- [x] Commit one real, public-safe Web UI screenshot to the README.

M3 acceptance is complete. Do not reopen M3 unless real use exposes a concrete regression.

## Active milestone — M3.4: public container distribution

M3.4 adds no new HostSleuth functionality. It only turns the existing Docker support into a polished public distribution path.

- [x] Inspect current `main`, Dockerfile, Compose, CI, release workflow, handoff, TODO, release history, and recent development history.
- [x] Create focused branch `m3.4/public-container-distribution` from current `main`.
- [x] Keep the existing Docker security model and loopback-only default unchanged.
- [x] Make the normal Compose path consume a published image while preserving an explicit local source-build path for developers.
- [x] Add a direct `docker run` example matching the supported Compose security/runtime settings.
- [x] Confirm Docker Hub namespace/repository as `mjmalleo/hostsleuth`.
- [x] Keep Docker Hub credentials out of Git; the release workflow uses only the `DOCKERHUB_TOKEN` secret and fixed public namespace `mjmalleo`.
- [x] Integrate stable Docker publication into the existing release workflow so it does not depend on a chained `release` event created by `GITHUB_TOKEN`.
- [x] Publish Docker images only for plain stable `vX.Y.Z` release branches; prerelease versions create prerelease GitHub releases but do not publish Docker tags.
- [x] Publish stable Docker tags as `<X.Y.Z>` plus `latest` for linux/amd64 + linux/arm64.
- [x] Add OCI image metadata for source, revision, version, title, and MIT license.
- [x] Extend CI smoke testing so Compose starts from an already-built image and validates `/api/about`, `/api/snapshot`, Web UI, and diagnosis.
- [x] Run full PR CI: normal Go checks, Compose validation, amd64 image build, arm64 image build, and image-consumption smoke all passed on run 98 before the final release-pipeline integration changes.
- [ ] Confirm the final PR head remains green after the release-pipeline/documentation changes.
- [ ] Create the Docker Hub repository and GitHub Actions `DOCKERHUB_TOKEN` secret as one-time owner-controlled setup.
- [ ] Merge PR #20 after final CI is green.
- [ ] Stop at the publication boundary and obtain explicit owner approval before creating `release/v0.1.0`, the stable GitHub release/tag, or any Docker Hub image/tag.
- [ ] Test the actual published image by pulling/running `mjmalleo/hostsleuth:latest` and `mjmalleo/hostsleuth:0.1.0` as a normal user would.
- [ ] Update the separate OMV disaster-recovery repository and live OMV/Arcane deployment only after the public image is validated and the owner approves the deployment change.
- [ ] Record M3.4 completion in handoff/history after validation.

## Next application capability — M4: package-change timeline

M4 begins only after M3.4 is complete. Keep it deliberately narrow. Its only job is to make **“what changed?”** more useful.

- [ ] Read package install/update/remove history from supported local package-manager logs.
- [ ] Start with a bounded Debian/Ubuntu `apt`/`dpkg` implementation; unsupported systems should remain truthful and quiet rather than requiring configuration.
- [ ] Add package changes to the existing event timeline instead of creating a separate package UI.
- [ ] Keep collection read-only and local-first.
- [ ] Do not add package management, update actions, alerts, repositories, or package settings.
- [ ] Add focused tests and real-host validation before merge.

## Belongs in HostSleuth, but only one at a time after M4

These fit one of HostSleuth's two core jobs, but none becomes active automatically:

- Configuration fingerprinting that records change evidence without storing configuration secrets by default.
- Bounded TLS/certificate diagnosis for common host:port failures: handshake, hostname, expiry, and trust evidence.
- General redaction rules and threat-model documentation before exporting or sharing richer diagnostic data.

## Demand-gated — keep only if real users prove the need

- Authentication for direct non-loopback Web UI exposure. Loopback + SSH tunneling remains the simple default.
- Reverse-proxy awareness for common proxies. Do not build a proxy-management layer.
- `.deb` packaging and signed artifacts if the existing installer is a real adoption problem.
- Privilege separation if future collectors actually require enough elevated access to justify the complexity.
- SQLite only if JSON/JSONL becomes a demonstrated operational limitation.
- Exposure-map UI only if listener visibility is repeatedly hard to understand from the existing views.
- Baseline-vs-incident comparison only if the event timeline proves insufficient in real incidents.
- Export/import support bundle only after redaction rules are strong enough to make sharing safe.

## Not planned unless HostSleuth changes direction

These would add substantial architecture, UI, or safety complexity beyond the current small single-host product:

- Multi-host controller/agent architecture.
- General dependency-graph platform.
- Pluggable diagnosis-rule ecosystem.
- AI explanation layer.
- Automatic or guided repair/remediation workflow.

They can be reconsidered later, but they are not standing tasks.

## Rule for new ideas

An idea goes here first. It becomes active only when it clearly improves a common HostSleuth workflow without making installation, operation, or the UI meaningfully harder.
