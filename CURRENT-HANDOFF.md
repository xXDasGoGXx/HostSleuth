# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Current authoritative source state

Repository: `xXDasGoGXx/HostSleuth`

Completed source milestones:

- M0 — repository foundation;
- M1 — deployable single-host MVP;
- M2 — deeper deterministic diagnosis;
- M3 — Product Experience;
- M3.4 — Public Container Distribution;
- M4 — package-change timeline;
- M5 — configuration fingerprinting;
- M6 — Certificate Story / TLS Detective;
- M7 — Service Story;
- M8 — Incident Lens;
- M9 — HostSleuth Workbench;
- M10 — Reboot Story;
- M11 — Optional Safe Actions.

M10 closeout: `docs/history/M10-REBOOT-STORY.md`.

M11 closeout: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

M11 merged in PR #38 at:

`6566c505b32cc47d96384152a988736173d3f7cd`

Post-merge main CI run #216 completed successfully.

## M11 — Optional Safe Actions

Read-only behavior remains the default. The only state-changing action is:

`service.restart`

Security boundary:

- actions require explicit `--enable-actions` opt-in;
- each restart target must be explicitly allowlisted with `--allow-restart-service UNIT`;
- only conservative `.service` unit names are accepted;
- no arbitrary command, argv, script, or shell field exists;
- no generic service manager exists;
- action evidence and execution use trusted absolute `/usr/bin/systemctl` or `/bin/systemctl`;
- preview exposes exact target/effect/argv/confirmation;
- execution requires the exact preview confirmation;
- durable audit is required before execution;
- time and output are bounded;
- success requires observed post-action `ActiveState=active`;
- Action Web/API operations are loopback-only;
- state-changing Web requests are JSON-only and require `X-HostSleuth-Action: confirm`;
- Docker mode reports native systemd restart unavailable;
- no writable Docker socket, automatic remediation, package/firewall/file administration, or generic server control was added.

The permanent CI `native-actions-smoke` uses a disposable systemd service on an ephemeral GitHub runner and proves wrong-confirmation denial, real allowlisted restart, postcondition verification, audit sequence, and cleanup without touching OMV production.

## Stable public release — v0.4.0

Stable `v0.4.0` was published on 2026-09-17 from exact accepted source:

`6566c505b32cc47d96384152a988736173d3f7cd`

Release workflow run #5 (`35193948370`) completed successfully.

GitHub release assets:

- `hostsleuth-linux-amd64` — `sha256:0ea9c20ade7a96fe208aef4b29416f6ebc831dc971e87d2f6ff07d8f93b03233`
- `hostsleuth-linux-arm64` — `sha256:017144dfd5cddf5fb5a351318079d094687650d1ba6bc2759a9eceeae1c82d83`
- `SHA256SUMS` — `sha256:b7c5dc0ab4c62287e28e0c5ce051d598a3bf202f6c062e605c72c382eeb01f58`

Published Docker tags:

- `mjmalleo/hostsleuth:0.4.0`
- `mjmalleo/hostsleuth:latest`

Both resolve to multi-platform OCI index:

`sha256:03b5824fddc50a707e5486033afed3f01d0be76e9adef64292d7a72743578bf0`

Platforms:

- linux/amd64 — `sha256:f001ab1537184b4841688b5838dff5c7fca95ce7c2dfcf6eb578ec1a441b7299`
- linux/arm64 — `sha256:4e70aa4a4439801d5e14d89d1193ab87befdc96a56fad107a725fee412dd0320`

Independent consumer verification on OMV downloaded the published amd64 binary, validated it against `SHA256SUMS`, confirmed `v0.4.0 (6566c505b32c)`, and confirmed Safe Actions are disabled/unavailable by default.

Full publication record: `docs/history/V0.4.0-PUBLICATION.md`.

## Live OMV / disaster-recovery boundary

Publication did **not** redeploy the live container.

Current live production remains independently verified as:

`mjmalleo/hostsleuth:0.3.0`

Live `/api/about` reports `v0.3.0`, and `/api/snapshot` reports schema 3 / Docker mode with container `hostsleuth` on image `mjmalleo/hostsleuth:0.3.0`.

The current disaster-recovery `main` branch also remains pinned to `0.3.0`.

With explicit owner approval, recovery PR #4 now stages the future recovery pin:

`mjmalleo/hostsleuth:0.4.0`

PR #4 is intentionally **not merged yet**. Recovery must not silently move ahead of production.

### Remaining approved operational step

Redeploy the existing Arcane-managed `hostsleuth` Compose project to `mjmalleo/hostsleuth:0.4.0`, preserving its current bind/state/security settings. Then verify:

1. `/api/about` reports `v0.4.0`;
2. `/api/snapshot` reports schema 4 and Docker mode;
3. the `hostsleuth` container reports image `mjmalleo/hostsleuth:0.4.0`;
4. `192.168.2.181:8787` remains reachable;
5. state/events remain present;
6. Actions remain unavailable in Docker mode/default deployment.

Arcane itself is healthy at port 3552, but its project API requires authentication. This session has no authorized Arcane credential/connector. HomeCommander correctly blocks raw Docker/sudo access, and its `hostsleuth` managed-deployment record is an old uninstalled native-systemd path, not the live Arcane/Docker project. Do not bypass those controls.

After the authenticated Arcane redeploy is verified, update recovery PR #4 documentation to say production/recovery are aligned, merge it, and update this handoff.

## Next product milestone boundary

The **Redacted Evidence Bundle** is NOT STARTED.

Do not start it as part of release/deployment cleanup. Redaction/threat-model rules must precede export implementation. Do not expand Optional Safe Actions with additional action families merely because the framework exists.

## Resume order

Do not restart a full audit on every continuation. Reuse this handoff unless a consequential write depends on something that may have changed.

For the current continuation, the next check should be only the live HostSleuth image/version. If it is still `0.3.0`, the remaining task is the authenticated Arcane redeploy described above. If it is already `0.4.0`, verify the six acceptance points and finish recovery PR #4.
