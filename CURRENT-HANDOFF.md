# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Stable/public/live/recovery state

Stable public release:

`v0.4.0`

Published/live/recovery image:

`mjmalleo/hostsleuth:0.4.0`

Stable v0.4.0 source:

`6566c505b32cc47d96384152a988736173d3f7cd`

The live Arcane-managed OMV deployment and `xXDasGoGXx/OMV-Docker-Rebuild` disaster-recovery definition are aligned on 0.4.0. Recovery alignment was merged in OMV-Docker-Rebuild PR #4 at:

`ad2bd53ca3a469272eba6343c03936b7c1a04bc0`

Live acceptance previously confirmed schema 4 / Docker mode, retained state/events, and Optional Safe Actions disabled/unavailable in the default Docker deployment. Do not disturb this production baseline as part of source development.

## Completed source milestones through v0.4.0

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

v0.4.0 publication record: `docs/history/V0.4.0-PUBLICATION.md`.

## Redacted Evidence Bundle — COMPLETE IN DEVELOPMENT SOURCE

Owner accepted the threat-model/redaction contract in:

`docs/design/REDACTED-EVIDENCE-BUNDLE.md`

Development branch:

`redacted-evidence-bundle`

PR:

`#40 — Redacted Evidence Bundle: threat model and bounded exporter`

Closeout record:

`docs/history/REDACTED-EVIDENCE-BUNDLE.md`

### Delivered first version

CLI:

```text
hostsleuth evidence preview
hostsleuth evidence export [--output PATH]
```

The first exporter is deliberately typed and bounded. It includes only:

- current Snapshot;
- at most 200 recent Events;
- at most 100 Safe Action audit records when present.

It does not provide arbitrary file inclusion or a generic serialize-everything path.

`preview` builds the same redacted payload in memory and writes no archive. `export` creates one local ZIP, refuses an existing destination, writes through an owner-only temporary file, and leaves the final archive `0600` where supported.

The bundle contains a manifest, summary, redacted JSON payloads, and SHA-256 checksums. Individual JSON payloads are capped at 1 MiB and the total uncompressed bundle payload at 2 MiB.

### Redaction/security boundary

The accepted `redacted-v1` policy pseudonymizes or removes sensitive host/infrastructure identifiers and credentials while retaining useful diagnostic relationships inside the bundle.

Covered evidence includes host/domain identifiers, non-loopback IPv4/IPv6, MACs, arbitrary paths, service/container/network/interface identities, image repository identity, e-mail addresses, opaque IDs, configuration fingerprints, URL credentials/query strings, credential-like key/value forms, and full or truncated private-key material.

Useful semantics such as ports, prefix lengths, loopback/unspecified addresses, package/version data, state/status values, UTC ordering, redacted URL shape, and IPv6-CIDR shape remain available where safe.

The first version explicitly excludes raw journal text, raw Safe Action command output, ActionPreview command/confirmation material, arbitrary file/configuration contents, arbitrary Workbench file/URL/DNS inputs, cloud upload, and automatic sharing.

Redaction lowers disclosure risk but cannot guarantee anonymity. The bundle must still be reviewed before sharing.

### Validation

An isolated OMV `/tmp` clone used a temporary Go 1.24.13 toolchain. Nothing was installed system-wide and production HostSleuth was not modified.

Local validation passed:

- gofmt cleanliness;
- `go vet ./...`;
- `go test ./...`;
- native build;
- `git diff --check`.

GitHub CI on the completed implementation passed:

- test/format/vet/native build;
- linux/amd64 image build;
- linux/arm64 image build;
- supported Docker smoke;
- native-actions smoke.

Disposable end-to-end CLI acceptance used synthetic state only and confirmed:

- preview created no ZIP;
- export created the expected bounded bundle;
- final archive mode was `0600`;
- bundle checksums verified;
- planted private hostname/domain/IPv4/IPv6/interface/service/container/network/registry/path/password/Bearer-token/URL-credential/query/private-key values did not survive the leak scan;
- redacted URL and IPv6-CIDR structure remained readable.

Adversarial testing caught and fixed a real first-draft Bearer-token scrubber bug plus later domain/IPv6/truncated-private-key and evidence-structure edge cases. Regression tests cover them.

## Current release boundary

The Redacted Evidence Bundle is complete in source, but **stable v0.4.0 does not contain it**.

Do not automatically:

- publish a new release;
- move `latest`;
- redeploy the live Arcane project;
- change the recovery image pin.

Release/publication and production/recovery deployment require a separate explicit decision after PR #40 is merged.

## Resume order

Do not restart a full audit on every continuation.

1. If PR #40 is still open, verify its latest head CI is green, move it out of draft, and merge it.
2. After merge, record the merge SHA in this handoff/TO-DO if needed.
3. Keep stable/public/live/recovery on v0.4.0 until a separate release decision is made.
4. Do not invent or automatically start a new feature milestone; the research backlog remains design input only.

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.
