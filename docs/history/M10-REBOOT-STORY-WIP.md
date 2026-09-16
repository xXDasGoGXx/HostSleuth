# M10 — Reboot Story — WIP Checkpoint

Date: 2026-09-16

Status: SUPERSEDED HISTORICAL CHECKPOINT. M10 is complete in source. Use `docs/history/M10-REBOOT-STORY.md` for the final implementation and acceptance record. The remaining-work list below is preserved only as the original mid-milestone handoff and is no longer current.

## Starting point

M10 branch: `m10-reboot-story`

Base `main` at M10 start:

`5745f9f0ed675e05e219b7a1c9147d25c049bc98`

That commit is the completed M9 — HostSleuth Workbench merge.

Checkpoint source before this handoff commit:

`164aff7691c15831ed1d49ac94a0473191e5c3eb`

## Goal

Explain boot/shutdown evidence and what failed to come back after a reboot without inventing a reboot cause. Reuse HostSleuth's retained event and Incident Lens model, keep evidence bounded, and remain read-only.

## Implemented so far

- Snapshot schema advanced to version 4 for deterministic boot identity evidence.
- Host snapshot evidence now records the Linux kernel boot ID when available.
- Host snapshot evidence now records an exact boot start time derived from `/proc/stat` `btime` when available.
- Snapshot diff logic detects a reboot only when both old and new boot IDs are known and differ; it does not infer reboot from uptime text alone.
- Reboot detection emits retained system evidence that can anchor later correlation.
- Initial Reboot Story core and focused tests are committed.
- Initial Reboot Story Web UI JavaScript/CSS/assets are committed.
- No reboot, shutdown, restart, service-control, package, firewall, or arbitrary-command action has been added.

## Real-host evidence already checked

On the actual OMV Debian host, isolated validation under `/tmp` confirmed:

- `go test ./...` passes on the then-current M10 core;
- native build passes;
- current host boot ID is readable through `/proc/sys/kernel/random/boot_id`;
- exact boot start is derivable from `/proc/stat` `btime`;
- a schema-version-4 snapshot contains `boot_id` and `boot_started_at`;
- journal boot history is not readable by the unprivileged HomeCommander acceptance account.

The journal permission result is an expected evidence boundary. M10 must represent unavailable journal evidence as unknown/unavailable rather than treating it as proof of failure or inventing a reboot cause.

## Important implementation boundary

M10 must keep two concepts separate:

1. **Reboot detection** — a changed kernel boot ID is deterministic evidence that a new boot occurred between snapshots.
2. **Reboot explanation/recovery evidence** — retained service/listener/container/package/configuration events and bounded journal evidence may show what happened around that boot, but temporal proximity alone does not prove why the reboot occurred.

Do not claim reboot cause unless explicit deterministic evidence exists.

## Work still remaining before M10 can be called complete

Historical checkpoint list only; all applicable items were completed before M10 closeout. See `docs/history/M10-REBOOT-STORY.md`.

- finish wiring the Reboot Story into CLI/API/runtime assets/UI navigation;
- verify bounded journal/current-boot/previous-boot evidence behavior where permissions allow;
- verify graceful `unknown` behavior where journal access is denied or history is unavailable;
- verify recovery correlation for services/listeners/containers that fail to return after boot;
- ensure no first-schema-upgrade false reboot event is emitted when an older snapshot lacks boot ID;
- run focused Go tests plus full `go test ./...`, `go vet ./...`, formatting, JavaScript syntax, native build, Docker smoke, amd64 and arm64 image builds;
- perform one bounded isolated real-OMV acceptance without touching the live HostSleuth deployment;
- update README / CURRENT-HANDOFF / TO-DO / ROADMAP and create final M10 history only after acceptance;
- open/complete the normal M10 PR and merge only when CI and acceptance are clean.

## Publication / deployment boundary

Do **not** publish a new HostSleuth release or Docker tag from this WIP checkpoint.

Stable public release remains `v0.3.0` from source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Public Docker tags remain:

- `mjmalleo/hostsleuth:0.3.0`
- `mjmalleo/hostsleuth:latest`

Live OMV remains intentionally pinned to:

`mjmalleo/hostsleuth:0.1.0`

Do not alter the live OMV Compose deployment, persistent state, or `xXDasGoGXx/OMV-Docker-Rebuild` recovery pin merely to advance source/version numbers.

## Roadmap boundary

At this historical checkpoint M10 was active. M10 is now complete in source; M11 — Optional Safe Actions — has **not** started.

Consumer-opportunity research remains a separate input and is intentionally broader than certificate/Certbot work.

## Resume instructions

This file is no longer the resume source of truth. Use `CURRENT-HANDOFF.md`, `TO-DO.md`, `docs/ROADMAP.md`, and `docs/history/M10-REBOOT-STORY.md`.
