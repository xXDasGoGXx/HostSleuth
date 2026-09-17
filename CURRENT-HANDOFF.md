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

## Live OMV / disaster-recovery alignment — COMPLETE

The existing Arcane-managed production deployment was redeployed through the authenticated Arcane UI to:

`mjmalleo/hostsleuth:0.4.0`

No raw Docker/sudo bypass was used.

Post-redeploy acceptance on 2026-09-17 confirmed:

1. `/api/about` reports `v0.4.0`;
2. `/api/snapshot` reports schema 4 and Docker mode;
3. the live `hostsleuth` container reports image `mjmalleo/hostsleuth:0.4.0`;
4. `192.168.2.181:8787` remains reachable and healthy;
5. persistent state/events survived the redeploy, with `/api/events` still returning pre-redeploy history;
6. Optional Safe Actions remain `enabled=false` and `available=false` in the default Docker deployment;
7. the Action Web/API remains loopback-only and returns HTTP 403 from LAN access.

The disaster-recovery repository `xXDasGoGXx/OMV-Docker-Rebuild` was updated to the same pinned image. Recovery PR #4 was updated with the live acceptance evidence and merged successfully on 2026-09-17.

Recovery merge commit:

`ad2bd53ca3a469272eba6343c03936b7c1a04bc0`

Production and disaster recovery are therefore aligned on `mjmalleo/hostsleuth:0.4.0`.

## Next product milestone boundary

The **Redacted Evidence Bundle** is NOT STARTED.

Do not start it merely as release/deployment cleanup. It requires explicit owner direction. Redaction/threat-model rules must precede export implementation, and the bundle must never silently include credentials, tokens, private keys, configuration contents, or other secrets.

Do not expand Optional Safe Actions with additional action families merely because the framework exists.

## Resume order

Do not restart a full audit on every continuation. Reuse this handoff unless a consequential write depends on something that may have changed.

Current stable/public/live/recovery state is all `v0.4.0`.

If asked to continue product work, the next planned milestone is the **Redacted Evidence Bundle**, beginning with redaction/threat-model design only after explicit owner direction. Do not begin export implementation before those rules are accepted.
