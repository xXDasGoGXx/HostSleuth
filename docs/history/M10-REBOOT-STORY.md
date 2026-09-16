# M10 — Reboot Story

Date: 2026-09-16

Status: COMPLETE in source after implementation, CI, and isolated real-host acceptance. This milestone does not authorize a new public release or a live OMV deployment change.

## Goal

Answer:

**What happened around this reboot, and what failed to come back afterward?**

M10 keeps reboot detection separate from reboot explanation. It remains read-only, deterministic, bounded, single-host, evidence-first, and non-causal unless direct evidence supports a specific classification.

## Delivered behavior

- Snapshot schema version 4 records Linux kernel boot ID when available.
- Exact boot start is derived from `/proc/stat` `btime` when available.
- A retained reboot event is emitted only when a previously known boot ID changes to another known boot ID.
- An older snapshot without boot ID is treated as a schema baseline; it does not create a false reboot event.
- Human-readable uptime is never used to infer a reboot.
- `hostsleuth reboot` provides the read-only CLI story.
- `/api/reboot-story` exposes the same story as JSON.
- The Web UI includes a Reboot tab wired into the served runtime assets.
- Previous-boot and current-boot journal evidence are bounded to 16 KiB and 48 lines per lookup.
- Journal command absence, unavailable history, command errors, and permission-only diagnostics remain explicit `unknown` evidence.
- Previous-boot shutdown classification stays unknown unless direct bounded evidence supports a distinction.
- Direct kernel-panic evidence can support an abnormal-termination classification.
- Direct systemd shutdown-target evidence can support an orderly shutdown/reboot classification.
- Current failed systemd services are surfaced in native mode.
- Retained post-boot warning events for services, listeners, and containers are correlated with the current snapshot.
- A recovery issue is shown only when retained post-boot evidence identifies the problem and the current snapshot still agrees that the service/listener/container is missing or unhealthy.
- A service/listener/container problem that later recovered is not presented as still broken.
- Nearby package, kernel/system, and configuration changes are shown as bounded context only.
- Temporal proximity is explicitly not presented as proof of reboot cause.
- No arbitrary endpoint/certificate relationship is invented when the retained evidence cannot connect it defensibly.

## Cause boundary

Reboot detection and reboot explanation are different concepts.

A changed known boot ID is deterministic evidence that a new kernel boot occurred between snapshots. Package, configuration, service, listener, container, and journal evidence around that time may explain what was observed before or after boot, but timing alone does not prove why the reboot occurred.

The Reboot Story therefore carries an explicit cause assessment stating that no reboot cause is claimed from the available evidence unless direct evidence genuinely supports a narrower statement.

## Journal permission boundary

Real OMV acceptance exposed an important environment-specific behavior: the unprivileged HomeCommander account can receive the text:

`No journal files were opened due to insufficient permissions.`

from `journalctl` without a command failure status.

M10 treats that diagnostic as unavailable journal access and reports `unknown`; it does not mislabel the diagnostic itself as usable journal history. A focused regression test covers the exact message observed on the real host for both previous-boot and current-boot evidence.

No privilege expansion was added to make journal history appear more complete.

## Validation

Closeout validation covers:

- focused Reboot Story tests;
- schema-upgrade regression proving no false reboot when an older snapshot lacks boot ID;
- direct-evidence shutdown classification tests;
- recovered-service/listener handling;
- service/listener/container recovery issue correlation;
- exact OMV journal permission-diagnostic regression;
- full `go test ./...`;
- `go vet ./...`;
- `gofmt` cleanliness;
- syntax checks for every JavaScript fragment;
- syntax validation of the exact concatenated served JavaScript;
- native Go build;
- Compose validation;
- Docker runtime smoke including the Reboot Story API/UI asset wiring;
- linux/amd64 image build;
- linux/arm64 image build;
- isolated real-OMV snapshot validation with schema 4, live boot ID, exact boot start, systemd services, and listeners;
- direct previous/current boot journal probes on the real OMV account confirming the permission boundary above.

Real-host acceptance used temporary state/build paths only. M10 work did not change the production HostSleuth deployment, production state, live Compose definition, Docker image selection, or recovery repository.

## Publication and live-deployment boundary

Stable public release remains `v0.3.0` from source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Published Docker tags remain:

- `mjmalleo/hostsleuth:0.3.0`
- `mjmalleo/hostsleuth:latest`

M7, M8, M9, and M10 are newer source work and are not claimed to be included in v0.3.0.

A final read-only closeout check of the live endpoint at `192.168.2.181:8787` reported application version `v0.3.0`. The unprivileged acceptance account could not read the live Compose file, so M10 does not claim an independently verified live image tag. The separate `xXDasGoGXx/OMV-Docker-Rebuild` repository still documents its disaster-recovery HostSleuth definition as pinned to `mjmalleo/hostsleuth:0.1.0`.

That live/recovery divergence was only recorded, not reconciled. M10 did not deploy, repin, publish, or otherwise alter either system.

## Security/product boundary

M10 added no reboot, shutdown, restart, reload, service-control, package, firewall, file-write, arbitrary-command, or remediation action.

M11 — Optional Safe Actions — remains not started and requires an explicit security/design review before HostSleuth crosses the read-only boundary.
