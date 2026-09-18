# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Stable / public / live / recovery state — v0.6.0

Stable release, live Arcane production, and disaster recovery are now aligned on:

`v0.6.0` / `mjmalleo/hostsleuth:0.6.0`

Exact published source:

`fcb08be51ae3da8cd20dc3929cf9d736b15f170c`

Release workflow:

`35284864117` — success

Verified public OCI index:

`sha256:6b9f90209f477ba8213d9c4bf7db7996d6c52d2caae2cfce405822df8bc8ef2a`

Live acceptance confirmed v0.6.0, schema 4 / Docker mode, retained pre-upgrade history, the M12 Expectations UI/API, a passing live Expected Endpoint Contract, and the LAN HTTP 403 boundary for Safe Actions.

`OMV-Docker-Rebuild` recovery alignment PR #6 merged at:

`f78abbe5742c315d73cf85709ebfeba620311429`

Full publication record: `docs/history/V0.6.0-PUBLICATION.md`.

Full production/recovery record: `docs/history/V0.6.0-PRODUCTION-ALIGNMENT.md`.

## Redacted Evidence Bundle — complete, published, and live

The accepted threat model/redaction contract is in:

`docs/design/REDACTED-EVIDENCE-BUNDLE.md`

Implementation closeout:

`docs/history/REDACTED-EVIDENCE-BUNDLE.md`

CLI:

```text
hostsleuth evidence preview
hostsleuth evidence export [--output PATH]
```

The first version is deliberately bounded to current Snapshot, recent Events, and Safe Action audit records. It uses deterministic bundle-local pseudonymization, credential/private-key scrubbing, manifest/checksum evidence, owner-only local archive permissions, and explicit exclusions for raw journal text, raw command output, arbitrary file/config contents, action confirmation material, cloud upload, and automatic sharing.

Redaction lowers disclosure risk but cannot guarantee anonymity. Bundles must still be reviewed before sharing.

## M12 — Expected Endpoint Contracts

The owner selected Expected Endpoint Contracts as the next bounded milestone. PR #43 passed full CI and merged to `main` at `1c193473d9220c34ec2820526df76076bfb41ce9`.

Delivered:

- typed read-only contract with target, optional exact DNS set, TLS mode, optional systemd service, and optional container;
- TCP reachability always required;
- ordered Expected-vs-Observed checklist;
- proven failure takes precedence over unresolved evidence for `first_mismatch`;
- full existing Diagnosis included for deeper route/listener/container/TLS/firewall evidence;
- `hostsleuth contract` CLI;
- `GET /api/contract`;
- Web UI **Expectations** view with handoff to full Diagnose;
- no polling, alerts, remediation, persistent contract database, or new privilege.

Local Go 1.24.13 format/vet/full-test/build and disposable HTTP/UI acceptance passed.

Design: `docs/design/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

Closeout: `docs/history/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

## Approved forward direction

The owner explicitly approved all of the following as future HostSleuth milestones, plus continuous UI/UX polish:

1. M13 — DNS Detective / split-view resolver comparison.
2. M14 — Reverse Proxy / Upstream Story.
3. M15 — Deployment / Permissions Story.
4. M16 — STARTTLS / Mail Service Story.
5. M17 — Certificate Rollout Verification.
6. M18 — exactly one additional Safe Action after a fresh privilege/threat-model gate.

The canonical ordered plan is now in `docs/ROADMAP.md`.

## Active milestone — M13 DNS Detective + Admin Console v1

Branch:

`m13-dns-detective-admin-console`

Implemented on the branch:

- system resolver plus up to four explicit custom resolver IP comparisons;
- IP-only custom resolver validation, DNS port 53 only;
- normalized A/AAAA/CNAME or PTR evidence;
- resolver latency/error evidence;
- address-scope classification;
- bounded runtime `/etc/resolv.conf` context;
- deterministic `agree`, `diverge`, `partial`, and `single` outcomes;
- bounded split-view hint when private/local and global resolver views differ;
- `hostsleuth dns` CLI;
- `GET /api/dns-detective`;
- dedicated DNS Detective Web UI;
- Admin Console v1 desktop navigation rail;
- global Quick Target bar with Diagnose / Expectations / DNS routing;
- browser-local recent targets and density preference;
- `/` keyboard shortcut to focus the target bar;
- responsive navigation fallback;
- CI cancellation for obsolete PR runs plus M13 JS/Docker smoke coverage.

Local validation passed with Go 1.24.13: gofmt, vet, full tests, native build, and JavaScript syntax.

Disposable HTTP acceptance passed. Real resolver comparison also demonstrated resolver-specific behavior: system DNS and `192.168.3.5` agreed for the test name while `192.168.2.5` timed out; HostSleuth reported partial resolver evidence rather than collapsing that into a generic DNS error.

Design: `docs/design/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md`.

## Active next step

Perform the final M13 branch audit, write the closeout record, then open the PR. Require full CI on the exact head before merge. Stable/public/live/recovery remain on v0.6.0 until a later independently verified v0.7.0 publication and rollout.

Do not restart a broad audit on continuation; use this handoff.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
