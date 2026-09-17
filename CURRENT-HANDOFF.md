# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform or browser-based server administration suite.

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

M10 implementation/acceptance history:

`docs/history/M10-REBOOT-STORY.md`

M11 implementation/security/acceptance history:

`docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`

M11 was developed on `m11-safe-actions` in PR #38.

## M11 — Optional Safe Actions — complete in source

M11 is the first source milestone that crosses HostSleuth's read-only boundary, but read-only operation remains the default.

Only one real action is implemented:

`service.restart`

It is intentionally narrow:

- actions are disabled unless `--enable-actions` is supplied;
- a target service must also be explicitly listed through repeated `--allow-restart-service UNIT` flags;
- only conservative `.service` unit names are accepted;
- no arbitrary command, argv, script, or shell field exists;
- no generic service manager exists;
- action pre/post evidence and execution use trusted absolute `/usr/bin/systemctl` or `/bin/systemctl`, not `$PATH` resolution;
- preview exposes the exact target, effect, argv, and confirmation value;
- execution requires the exact confirmation returned by preview;
- a durable audit record is required before restart execution;
- command output and execution time are bounded;
- action execution is serialized;
- success requires post-action `ActiveState=active` evidence;
- Action Web/API endpoints are loopback-only;
- state-changing Web requests require JSON plus `X-HostSleuth-Action: confirm`;
- Docker mode reports native systemd restart unavailable;
- M11 does not add a writable Docker socket, automatic remediation, package/firewall/file administration, or generic server control.

Delivered interfaces:

```text
hostsleuth action list
hostsleuth action preview
hostsleuth action run
hostsleuth action audit
```

plus the loopback-only JSON Action API and dedicated Actions Web UI.

### M11 validation

Exact-branch validation in an isolated `/tmp` clone on OMV passed:

- `gofmt` cleanliness;
- `go vet ./...`;
- `go test ./...`;
- native build;
- JavaScript syntax;
- exact concatenated served-JavaScript syntax.

No OMV production action or deployment change was performed during that validation.

CI run #205 validated functional source at:

`25318fed095f311221b44d67a427f033b1a46993`

All jobs passed:

- test / format / vet / Go tests / JS / native build;
- `native-actions-smoke`;
- Docker runtime smoke;
- linux/amd64 image build;
- linux/arm64 image build.

The native action smoke used a disposable systemd unit on an ephemeral GitHub-hosted Ubuntu runner. It proved preview, exact-confirmation enforcement, denied action without PID change, a real allowlisted restart with PID change, `ActiveState=active` postcondition verification, expected audit records, and cleanup.

The OMV production host was not used for a real action restart, which avoids bypassing HomeCommander administrative safeguards merely to satisfy acceptance.

## Published release boundary

Stable public release remains:

`v0.3.0`

Release source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Published Docker tags remain:

- `mjmalleo/hostsleuth:0.3.0`
- `mjmalleo/hostsleuth:latest`

Published multi-platform OCI index:

`sha256:127b388fbf794841b22d06b281fe89dc1500188fdb4215eec023b392aa98c05d`

Platforms:

- `linux/amd64`
- `linux/arm64`

M7, M8, M9, M10, and M11 are newer source capabilities and are **not** claimed to be included in v0.3.0.

Do not publish a new release merely for version-number alignment. M11 completion is not release authorization.

## Live OMV / recovery boundary

The active production HostSleuth image remains independently verified as:

`mjmalleo/hostsleuth:0.3.0`

The separate `xXDasGoGXx/OMV-Docker-Rebuild` disaster-recovery source of truth was explicitly repinned in PR #3 to the same image:

`mjmalleo/hostsleuth:0.3.0`

Production and recovery therefore remain aligned at the image-tag level.

M11 did **not**:

- restart or redeploy the live HostSleuth container;
- change the production Compose definition;
- enable actions in production;
- change the recovery image tag;
- publish an image or release.

Future release, production, or recovery image changes remain explicit owner-approved actions.

## Next decision boundary

There is no automatic M12 start.

The immediate decision is whether the accepted post-v0.3.0 source should become a new public release. Release publication and any later OMV/recovery upgrade are separate explicit decisions.

The next planned product feature after that decision is the **Redacted Evidence Bundle**, but it is NOT STARTED and must not begin without explicit owner direction. Redaction/threat-model rules come before export implementation.

Do not expand Optional Safe Actions with more action families merely because the framework exists. A future action must independently justify its privilege cost and preserve the M11 explicit-schema, allowlist, preview, confirmation, audit, and postcondition model.

## Consumer/product research boundary

Research remains separate from implementation scope. The detailed artifact is:

`docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md`

Promising later areas include endpoint-path/expected-state contracts, resolver/delegation/split-view DNS discrepancies, HTTP/reverse-proxy/upstream problems, permissions/ownership/deployment-path reasoning, container disappearance/dependencies, protocol-aware STARTTLS inspection, and certificate source -> destination -> actually-served verification.

## Product guardrails

Do not drift into:

- multi-host controller/agent architecture;
- generic network-device/SNMP monitoring;
- time-series monitoring/graph platform behavior;
- arbitrary web terminal or command execution;
- generic package/firewall/configuration administration;
- AI-generated causal claims;
- automatic remediation;
- broad privilege expansion merely to make features easier.

## Resume order for future work

Do not restart a full audit on every continuation. Reuse this handoff unless a consequential write genuinely depends on something that may have changed.

Before a consequential release/deployment or new milestone branch, check the directly relevant current state, then read:

1. `CURRENT-HANDOFF.md`
2. `TO-DO.md`
3. `docs/ROADMAP.md`
4. the most recent milestone history document
5. `.github/workflows/ci.yml` when changing executable source
6. `.github/workflows/release.yml` only for release work

M11 closeout: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.
