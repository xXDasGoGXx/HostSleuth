# HostSleuth — Current Handoff

Last updated: 2026-09-17

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

## Stable / public / live / recovery state — v0.8.0

Stable/public, the live Arcane-managed deployment, and `xXDasGoGXx/OMV-Docker-Rebuild` disaster recovery are aligned on:

`v0.8.0` / `mjmalleo/hostsleuth:0.8.0`

Exact published source:

`28b8d88ec018782353840dee8528d765c4056e89`

Release workflow:

`35294912224` — success

Published Docker tags:

- `mjmalleo/hostsleuth:0.8.0`
- `mjmalleo/hostsleuth:latest`

Both resolve to verified OCI index:

`sha256:152255f44e451995767cd40b64b99879feb0ed0ae28a4fddfb4a269a55fc11df`

Live acceptance confirmed v0.8.0, schema 4 / Docker mode, retained pre-upgrade events, the Safe Actions loopback/default-disabled boundary, a passing live M14 Proxy Path, and Admin Console v2 in the actual live DOM.

Recovery PR #8 merged at:

`290e7e6a6781ad29cbc7c8296d00877fb8489fd0`

Full publication record: `docs/history/V0.8.0-PUBLICATION.md`.

Full production/recovery record: `docs/history/V0.8.0-PRODUCTION-ALIGNMENT.md`.

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

## M13 — DNS Detective + Admin Console v1 — SOURCE COMPLETE

PR #47 passed the full GitHub CI matrix and squash-merged to `main` at:

`182a27384a090f5538bb6d76a0c4dd917ce63932`

Delivered:

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

Disposable HTTP acceptance passed. Real resolver comparison also demonstrated resolver-specific behavior: the system resolver and one explicitly supplied local resolver agreed for the test name while another local resolver timed out; HostSleuth reported partial resolver evidence rather than collapsing that into a generic DNS error.

Design: `docs/design/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md`.

## M14 — Reverse Proxy / Upstream Story + Admin Console v2 — COMPLETE AND LIVE

PR #51 passed the full GitHub CI matrix and squash-merged to `main` at:

`135a73ffb333c1e4ac5135f93db7bac3dac5cdae`.

Delivered:

- explicit public HTTP/HTTPS URL plus explicit expected upstream URL;
- deterministic public DNS/route/TCP/TLS/HTTP and upstream DNS/route/TCP/TLS/HTTP stages;
- local listener/Docker publication context when the public endpoint is proven local;
- native upstream Host/SNI probe;
- public Host/SNI probe against the same explicit upstream when identities differ;
- first proven failure / warning / unknown precedence;
- HTTP 4xx warning and 5xx failure semantics;
- HEAD-only metadata probes with no bodies, credentials, cookies, Authorization, or arbitrary headers;
- bounded same-host redirects and cross-host redirect stop;
- query-string redaction in returned URL/Location evidence;
- `hostsleuth proxy` CLI;
- `GET /api/proxy-story`;
- dedicated visual **Proxy Path** Web UI;
- Admin Console v2 Proxy Quick Target action;
- bounded copyable evidence summary with LAN-HTTP clipboard fallback;
- permanent CI coverage for Proxy Path/Admin Console v2 JavaScript and Docker-mode proxy-story smoke.

Local validation passed with Go 1.24.13: gofmt, diff-check, vet, full tests, `go test -race ./internal/core`, native build, every Web JS syntax check, and disposable CLI/API/UI/headless-browser acceptance.

Design: `docs/design/M14-REVERSE-PROXY-UPSTREAM-STORY.md`.

Closeout: `docs/history/M14-REVERSE-PROXY-UPSTREAM-STORY.md`.

## Active next step — M15 Deployment / Permissions Story

v0.8.0 publication, live acceptance, and disaster-recovery alignment are complete.

Begin M15 from the approved roadmap. The bounded product question is:

> The process is running; why can it not use this path/socket/port?

Initial M15 scope remains:

- native service identity, UID/GID, working directory, and executable identity;
- one explicit user-supplied path;
- ownership/mode plus parent-directory traversal chain;
- deterministic read/write/execute/traverse reasoning;
- relevant Docker bind-mount metadata when available;
- no recursive filesystem crawl;
- no file contents;
- searchable evidence and a visual permission-chain story in the Admin Console.

Do not restart a broad audit on continuation; use this handoff.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
