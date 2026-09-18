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

## M15 — Deployment / Permissions Story — SOURCE COMPLETE

PR #55 passed the complete GitHub CI matrix on exact head:

`78bd31d8cdb34f9f5e592e57f14f2e8537e79b41`

CI run:

`35302370815` — success

PR #55 then squash-merged to `main` at:

`d64bb387c8f9efcf5dcb9f814f54b519ee231205`

Delivered:

- one native systemd service plus one explicit absolute path/socket path;
- observed PID, effective UID/GID, supplementary groups, working directory, executable, and effective capabilities;
- exact parent-directory ownership/mode traversal chain;
- deterministic read/write/execute/traverse/connect reasoning with owner/group/other class selection;
- capability-aware UID 0 handling without assuming unconditional root bypass;
- bounded relevant Docker bind-mount metadata when available;
- `hostsleuth permissions --service UNIT /absolute/path`;
- `GET /api/permissions-story`;
- dedicated Permissions Web UI with identity cards, visual chain, and searchable evidence;
- permanent Docker-mode CI coverage that requires native service identity to remain unavailable rather than crossing the container boundary.

Local/native acceptance passed without host mutation:

- `ssh.service` → `/etc/ssh`: permitted chain;
- `dbus.service` → `/root`: deterministic failure at `read:/root`;
- disposable HTTP/API/UI acceptance reproduced the failure and confirmed served Permissions assets.

M15 reads no file contents, performs no recursive crawl, and adds no remediation or privilege.

Design: `docs/design/M15-DEPLOYMENT-PERMISSIONS-STORY.md`.

Closeout: `docs/history/M15-DEPLOYMENT-PERMISSIONS-STORY.md`.

Stable/public/live/recovery remain aligned on v0.8.0. M15 is development source on `main`; it has not been published or deployed.

## M16 — STARTTLS / Mail Service Story — SOURCE COMPLETE

PR #57 passed the full GitHub CI matrix on exact head:

`579fff79889d5ab0c3430135d3f5ceba4dc28400`

CI run:

`35303552155` — success

PR #57 then squash-merged to `main` at:

`e5663d5883acdd2d859ff57b39a447c0018790b3`

Delivered:

- SMTP STARTTLS, IMAP STARTTLS, and POP3 STLS within one bounded pre-authentication model;
- deterministic TCP → greeting → capability → upgrade → TLS → certificate validity → hostname → trust stages;
- fixed protocol commands only, with no credentials, authentication, mail submission, mailbox access, or message contents;
- bounded protocol reads (4 KiB line, 64 lines, fixed deadlines);
- shared TLS/certificate interpretation with the existing TLS engine;
- `hostsleuth starttls --protocol smtp|imap|pop3 host:port`;
- `GET /api/starttls-story`;
- dedicated STARTTLS Admin Console view;
- permanent JavaScript/Docker-smoke coverage for the new interface and invalid-input boundary.

Local Go 1.24.13 full tests, core race tests, vet, native build, all Web JavaScript syntax checks, and `git diff --check` passed.

Disposable SMTP CLI/API/UI acceptance passed with a real STARTTLS upgrade: TCP, greeting, capability, upgrade, TLS, certificate validity, and hostname passed; the deliberately self-signed fixture correctly failed trust as the first problem. The disposable processes were stopped afterward.

Design: `docs/design/M16-STARTTLS-MAIL-SERVICE-STORY.md`.

Closeout: `docs/history/M16-STARTTLS-MAIL-SERVICE-STORY.md`.

Stable/public/live/recovery remain aligned on v0.8.0. M16 is development source on `main`; it has not been published or deployed.

## M17 — Certificate Rollout Verification — SOURCE COMPLETE

PR #59 passed the complete GitHub CI matrix on final exact head:

`32f0ad2dc3211e5af38a643a5d2f38081d440645`

CI run:

`35308369829` — success

PR #59 then squash-merged to `main` at:

`f12bbd8aacb9689bb31d5b9b6535479882cb6ce6`

Delivered:

- exactly one expected source: SHA-256 fingerprint or explicit reference direct-TLS endpoint;
- up to 16 explicit direct-TLS host:port endpoints with numeric-port validation and duplicate rejection;
- deterministic MATCH / MISMATCH / UNKNOWN rollout identity;
- separate certificate validity, hostname, and trust health;
- bounded four-worker probing while preserving user endpoint order;
- `hostsleuth cert-rollout` CLI;
- `GET /api/certificate-rollout`;
- dedicated Cert Rollout Admin Console matrix;
- focused decision tests and a real loopback TLS mismatch fixture;
- permanent JavaScript/bundle/Docker-smoke coverage.

Arbitrary certificate-file reads remain intentionally unexposed because HostSleuth cannot prove an arbitrary path is not a private key before opening it. This preserves the no-private-key-read boundary.

Native exact-head Go 1.24.13 validation passed: format, full tests, core race tests, vet, native build, JavaScript syntax, and diff check.

Disposable native acceptance also passed in both expected-source modes. Two loopback TLS endpoints served different certificates; fingerprint-source and reference-source comparisons both identified `localhost:19444` as the mismatched endpoint, and the served API/UI reproduced the same matrix. All temporary processes were confirmed stopped afterward.

Design: `docs/design/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`.

Closeout: `docs/history/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`.

Stable/public/live/recovery remain aligned on v0.8.0. M17 is development source on `main`; it has not been published or deployed.

## Active next step — M18 Safe Actions II — Security Review

Do not implement a second Safe Action immediately.

The required next work is:

1. compare concrete high-value action candidates;
2. choose exactly one;
3. write a fresh threat model and privilege-cost analysis;
4. define the narrow allowlist and exact command shape;
5. preserve explicit enablement, deterministic preview, exact confirmation, durable pre-execution audit, bounded execution, and postcondition verification;
6. reject any design that weakens Docker security merely to make the action convenient.

No generic command runner, generic systemd controller, generic container controller, or broad privilege mechanism.

Stable/public/live/recovery remain on v0.8.0.

Do not restart a broad audit on continuation; begin M18 with candidate comparison and security review only.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
