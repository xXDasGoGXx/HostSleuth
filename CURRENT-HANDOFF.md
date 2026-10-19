# HostSleuth — Current Handoff

Last updated: 2026-10-19 (v1.5.0)

## Stable / public / live / recovery state — v1.5.0

Stable/public/live/recovery are aligned on `v1.5.0` / `mjmalleo/hostsleuth:1.5.0`, published from exact source `7869173340f86720a712370e537a8e2c82000032` (PR #97) by release workflow `39295943111`.

v1.5.0 (record: `docs/history/V1.5.0-WHY-UNHEALTHY.md`): unhealthy containers are explained. The field-limited inspect reads each container's health-check command and Docker's newest check result (masked for secrets); fixed rules say why in Replay, Container Story and the API (`why`).

Verified live 2026-10-19: assets match `SHA256SUMS`; amd64 reports `v1.5.0 (7869173340f8)`; OMV compose sync pulled `mjmalleo/hostsleuth:1.5.0` and came up healthy (backup `compose-backups/hostsleuth/20261019T174230Z.yaml`); live compose byte-identical to GitHub; `/api/about` = `v1.5.0` / `7869173340f8`; health check and newest result recorded for all 8 health-checked containers (25 total), no secrets in them; LAN Action API 403. Nothing was unhealthy at deployment, so no live explanation was shown; the backup containers' exact checks are covered by the tests. Recovery: `OMV-Docker-Rebuild` PR #208 merged at `d73f1d47db1a9c7469dfcb9085a8146e5df24a13`; The Bridge synced to it.

## Previous release — v1.4.0

Stable/public/live/recovery are aligned on `v1.4.0` / `mjmalleo/hostsleuth:1.4.0`, published from exact source `b8c3218b3502e568b2b3ab354808c30441a69d52` (PR #95) by release workflow `38933415430`.

v1.4.0 (record: `docs/history/V1.4.0-REPLAY.md`): Replay is the home screen — containers drawn as a blueprint by compose project with bind-class connectors, a 7-day tape that rewinds the drawing, and one plain sentence per moment from fixed rules. New `GET /api/replay`, the `com.docker.compose.project` label in the field-limited inspect, self-hosted B612 fonts.

Verified live 2026-10-16: assets match `SHA256SUMS`; amd64 reports `v1.4.0 (b8c3218b3502)`; OMV compose sync pulled `mjmalleo/hostsleuth:1.4.0` and came up healthy (backup `compose-backups/hostsleuth/20261016T164025Z.yaml`); live compose byte-identical to GitHub; `/api/about` = `v1.4.0` / `b8c3218b3502`; `/api/replay` on the real host: 25 containers in 9 compose stacks, 29 events, no LAN addresses; fonts served as `font/woff`; LAN Action API 403. Replay narrated the real week including the reboot onto 7.0.14-24-pve, the qbittorrent-vpn autoheal restart, and backblaze-snapshotter recovering "after 18 hours" once the owner applied the Arcane classification fix. Recovery: `OMV-Docker-Rebuild` PR #142 merged at `2a06ca4ee92d4b2d7f9c0722945f784490dacef9`; The Bridge synced to it.

## Previous release — v1.3.0

Stable/public/live/recovery are aligned on `v1.3.0` / `mjmalleo/hostsleuth:1.3.0`, published from exact source `c9dc7bc420c6a712608f43f717c1ce435b10178c` (PR #93) by release workflow `38849643515`.

v1.3.0 (record: `docs/history/V1.3.0-CONTAINERS-SHARE-README.md`):

- container image updates, recreations, and restarts with Docker's own reason (crash and exit code, out-of-memory kill, SIGTERM/SIGKILL), from a field-limited `docker inspect` and die/oom `docker events`; never reads env, cmd, mounts or host config;
- Containers view, `/api/containers`, `/api/container-story`, `hostsleuth containers [NAME]`;
- Share redacted on every result view (`POST /api/share`, Evidence Bundle redaction, stores nothing);
- README rewritten as a short front page with screenshots from a sanitized demo dataset; reference moved to `docs/INSTALL.md`, `FEATURES.md`, `CLI.md`, `PROJECT.md`;
- CI runs a real crash-looping container in Docker smoke (it caught a `docker inspect` template bug before release).

Verified live 2026-10-15/16: assets match `SHA256SUMS`; amd64 reports `v1.3.0 (c9dc7bc420c6)`; OMV pulled `mjmalleo/hostsleuth:1.3.0` through the compose sync (healthy; live compose byte-identical to GitHub); `/api/about` = `v1.3.0` / `c9dc7bc420c6`; inspect evidence on 25 of 25 containers; no env/cmd/mount/config keys in the snapshot; Containers lists backblaze-snapshotter (unhealthy) first, then backblaze-uploader and qbittorrent-vpn (unstable); `/api/share` redacts host name, LAN address and container names; LAN Action API 403. First real events overnight: `container restarted: chaptarr (stopped cleanly, exit code 0)` and `container updated: searxng: searxng/searxng:latest, version 2026.10.15-b00212a83 → 2026.10.16-ff8f5ddcc`. Recovery: `OMV-Docker-Rebuild` PR #140 merged at `34610c7d51b7aad69305943872e0bb70ff1a31dd`; The Bridge synced to it.

Known limitations seen live:

- some images inherit `org.opencontainers.image.version` from their base image (chaptarr shows "24.04", the Ubuntu base), so a version label is a hint, not proof;
- `docker/DOCKERHUB.md` is updated, but the public Docker Hub page only changes when the optional manual metadata sync is run with a dedicated `DOCKERHUB_METADATA_TOKEN`.

## Previous state — v1.2.0

Stable/public/live/recovery are aligned on `v1.2.0` / `mjmalleo/hostsleuth:1.2.0`, published from exact source `91d31eafc9957a0a655a229200c4bfb7199d6e4a` (PRs #90 and #91) by release workflow `38844421236`.

v1.2.0 (record: `docs/history/V1.2.0-CONSOLE-AND-BLIPS.md`):

- a reboot's listener wave is one summary line; listeners that do not return keep their warning;
- after a recorded reboot, listeners on dynamically assigned ports that are replaced on the same protocol/address/port range are one line, so Reboot Story no longer reports NFS/RPC helpers as "not recovered";
- unhealthy/restarting/stopped containers lead Overview; every view linkable; Changes filters and day headings; Host search, grouping by reach, folded container networks/routes; favicon.

PR #90 was merged by the owner before review; it was reviewed afterwards and before release (see the history record).

Verified live 2026-10-15: assets match `SHA256SUMS`; amd64 reports `v1.2.0 (91d31eafc995)`; OMV pulled `mjmalleo/hostsleuth:1.2.0` through the HomeCommander-Claude compose sync (healthy); `/api/about` = `v1.2.0` / `91d31eafc995`; Reboot Story recovery issues 8 → 0 (10 dynamic-port moves paired); the Changes window now reaches back to 2026-10-09 (was about one day) and shows the 2026-10-14 reboot and kernel `7.0.14-23-pve -> 7.0.14-24-pve`; LAN Action API 403; security headers present. Recovery: `OMV-Docker-Rebuild` PR #139 merged at `50cc84a00b4c70443d47de049c5ccd52983bec2e`.

Open observations from live (not implemented):

- a local process rotates loopback TCP listeners on random high ports (about 25 appeared/disappeared lines in six days); pairing loopback dynamic-port replacements outside reboots is the next candidate;
- `192.168.2.181:8080` and `:8118` went away and came back six times on 2026-10-13/14, consistent with a container restarting; this is real evidence, not noise.

## Previous state — v1.1.0

Stable/public/live/recovery are aligned on `v1.1.0` / `mjmalleo/hostsleuth:1.1.0`, published from exact source `e2eecd68aaa236b730b43268696eb836106afd0e` (PR #88) by release workflow `38840544147`.

v1.1.0 makes the change timeline meaningful and restores build traceability:

- link-local/multicast sockets, per-interface discovery binds and ephemeral UDP client sockets are kept as Snapshot evidence but no longer recorded as changes; churn recorded by older versions is skipped on read; raw view via `events --all` / `/api/events?all=1`;
- `events.jsonl` compacts past 4 MiB to the newest 5,000 meaningful events;
- `/api/about` and `version` report the build revision in container builds (Docker smoke asserts it);
- HTTP server header/read/idle timeouts.

No schema, record format, API shape, Evidence Bundle, Safe Action or privilege change. Rollback to 1.0.1 reads the same files.

Verified 2026-10-15: native amd64/arm64 assets match `SHA256SUMS`; amd64 reports `v1.1.0 (e2eecd68aaa2)`; OMV pulled `mjmalleo/hostsleuth:1.1.0` from Docker Hub through the HomeCommander-Claude compose sync (healthy); live `/api/about` = `v1.1.0` / `e2eecd68aaa2`; schema 4 / Docker mode; LAN Action API 403; security headers present. Live timeline before: 94 of 100 events were churn. After: zero churn; container-health history visible again. Recovery: `OMV-Docker-Rebuild` PR #138 merged at `18f7f5e615fb1f6d3e6c5d4327eb878f88cc74f4`.

Open observation at the time (resolved in v1.2.0): after a reboot, real service listeners returning still filled most of the 100-event window.

Record: `docs/history/V1.1.0-EVENT-TIMELINE-QUALITY.md`.

## Previous state — v1.0.1

Stable/public/live/recovery are aligned on:

`v1.0.1` / `mjmalleo/hostsleuth:1.0.1`

Exact published source:

`9d95392882814e1c9f17643fe2664e6faaf15481`

Release workflow:

`37104898586` — success

Published Docker tags:

- `mjmalleo/hostsleuth:1.0.1`
- `mjmalleo/hostsleuth:latest`

Both resolve to verified OCI index:

`sha256:ae747aac332d953eed7826f27ee67548d8295da2638cd553461804c4ebbefe36`

Independent public verification passed for the exact GitHub tag, native amd64/arm64 binaries + SHA256SUMS, published binary version, and Docker Hub amd64/arm64 manifests.

v1.0.1 packages the owner-approved feature-free Admin Console polish from PRs #77–#84. It changes presentation and workflow ergonomics only; no collector, API contract, persisted schema/state, Evidence Bundle format, Safe Action family, privilege, monitoring/polling architecture, or remediation behavior changed.

Live Arcane acceptance confirmed:

- `/api/about` = `v1.0.1`;
- snapshot schema 4 / Docker mode;
- running container image `mjmalleo/hostsleuth:1.0.1`;
- all 100 retained events, including pre-upgrade history;
- host networking, host PID/UTS, read-only rootfs, `cap_drop: ALL`, and `no-new-privileges` preserved;
- v1 security headers preserved;
- v1.0.1 UI controls are present in the served Admin Console bundle;
- live Diagnose to `192.168.2.181:8787` = reachable / high confidence;
- real SMTP STARTTLS passed TCP, greeting, capability, upgrade, TLS, certificate validity, and trust; hostname failed only because the acceptance target intentionally used the numeric IP;
- LAN Action API = HTTP 403;
- loopback `service.restart` and `service.reload` capabilities both remain `enabled=false` / `available=false`.

`OMV-Docker-Rebuild` recovery PR #109 merged at:

`b2a88aeeec72342fc79fd29b574c541523b3b859`

Recovery `main` now pins `mjmalleo/hostsleuth:1.0.1`, with its current setup and operating snapshot recording the same live acceptance.

Publication record: `docs/history/V1.0.1-PUBLICATION.md`.

Production/recovery record: `docs/history/V1.0.1-PRODUCTION-ALIGNMENT.md`.

The v1.0.1 publication/production/recovery sequence is complete. No additional feature milestone or Safe Action family is automatically approved.

## Docker Hub storefront metadata

Docker Hub storefront copy is version-controlled separately from image publication:

- short description: `docker/DOCKERHUB-SHORT.txt`;
- Overview source: `docker/DOCKERHUB.md`;
- optional manual sync workflow: `.github/workflows/dockerhub-description.yml`.

During the 2026-09-19 reconciliation, Docker Hub's public short description was read anonymously and the Git-tracked short-description source was aligned to that exact public value.

Automatic metadata synchronization is **not an active commissioning requirement**. The workflow is manual-only and retained as optional future automation. If the owner later wants automatic provider-side metadata publishing, commission a dedicated `DOCKERHUB_METADATA_TOKEN` at that time. `DOCKERHUB_TOKEN` remains dedicated to image publication and must not be broadened or reused.

HostSleuth v1.0.1 source/release/live/recovery alignment is complete independently of that optional future automation.


## v1.0 Readiness / Hardening — COMPLETE / PUBLISHED IN v1.0.0

The owner-approved feature-free readiness cycle is complete and is now published in stable v1.0.0.

The separately gated v1.0.0 live acceptance and recovery-alignment steps are complete; production and recovery are aligned on v1.0.0.

Readiness slice 1:

- legacy persisted-state compatibility regression;
- dynamic tab/tabpanel accessibility semantics;
- keyboard navigation, visible focus, reduced-motion, and live-region handling;
- explicit HTML/CSS/JavaScript asset budgets;
- headless-Chrome deep-link and narrow-viewport acceptance.

PR #66 exact head:

`299fe89f81a03eec044340dc85769339c6d62ce3`

CI run:

`35314364419` — success

Merge:

`771b923c06d38ac528804468effbe56ffd4c8f78`

Readiness slice 2:

- native installer verifies the selected release binary against `SHA256SUMS` before installation;
- disposable CI performed a real native install of published v0.9.0 and cleanup;
- Web/API responses now carry restrictive CSP, Permissions-Policy, no-referrer, nosniff, and anti-framing headers;
- Docker smoke permanently verifies the browser security-header contract.

PR #67 exact head:

`04809c69d398a82c07a97dd1031827854b3591f9`

CI run:

`35314858546` — success

Merge:

`ec51b21aca8c4a9de50d2530d2200ba576b45989`

Final disposable upgrade/rollback acceptance also passed:

`published v0.9.0 -> current candidate -> published v0.9.0`

The same disposable state retained schema-4 snapshot identity, a synthetic retained Event, and a synthetic Safe Action audit in both directions.

Readiness contract:

`docs/design/V1.0-READINESS.md`

Release/rollback checklist:

`docs/design/V1.0-RELEASE-CHECKLIST.md`

Closeout:

`docs/history/V1.0-READINESS.md`

No feature, Safe Action family, privilege, monitoring architecture, remediation behavior, production deployment, or recovery definition was changed by readiness work.

Readiness closeout PR #68 passed the complete six-job CI matrix on exact head:

`2853f3a369a95817524935a93db6904d2ad3d31c`

CI run:

`35315452571` — success

PR #68 squash-merged to `main` at:

`f93330755e229609f86add144a20d3ce2234cd04`

The owner approved publication. v1.0.0 is now published and independently verified from exact accepted source `27971e67ad7ea875d83f5925a83c1991b0eb2b0f` under release workflow `35358798799`.

Live v1.0.0 acceptance passed and recovery PR #11 merged at `44d6a30e580e2b333841f14eade2e10cca4b5ba5`. Stable/public/live/recovery are aligned on v1.0.0.

## Product identity

HostSleuth is a small, local-first Linux troubleshooting tool with two jobs:

1. **Remember meaningful host changes.**
2. **Explain why a host/service/port is or is not reachable using deterministic evidence.**

Keep it evidence-first, local-first, single-host first, and deliberately small. It is not a generic monitoring platform, browser shell, or automatic-remediation engine.

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

M15 is published, live, and recovery-aligned in v0.9.0.

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

M16 is published, live, and recovery-aligned in v0.9.0.

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

M17 is published, live, and recovery-aligned in v0.9.0.

## M18 — Safe Actions II — SOURCE COMPLETE

The owner explicitly approved exactly `service.reload` under:

`docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md`

PR #62 passed the complete GitHub CI matrix on exact head:

`c46dd72ad806c688970244836ce26e3ca01eec88`

CI run:

`35309972305` — success

PR #62 then squash-merged to `main` at:

`cc9c9727fe19786eba03d0b7ef51c7fae7ac8ab1`

Delivered:

- fixed action ID `service.reload`;
- dedicated `--allow-reload-service UNIT` allowlist independent of restart allowlisting;
- existing `--enable-actions` global opt-in;
- native mode only; Docker exposes neither native systemd action;
- trusted absolute systemctl lookup and fixed reload argv only;
- loaded + active + `CanReload=yes` fail-closed preconditions;
- exact `RELOAD UNIT.service` confirmation;
- durable pre-execution audit;
- bounded serialized execution/output;
- no reload-or-restart helper and no restart fallback;
- postcondition requiring command success plus `ActiveState=active`;
- loopback-only Action Web/API boundary;
- server-provided restart/reload capability selection in the Admin Console.

Local Go 1.24.13 full tests, core race tests, vet, native build, Action JavaScript syntax, and diff checks passed.

The permanent native CI fixture performed a real reload on a disposable reload-capable systemd unit. Restart-only allowlisting could not authorize reload; the exact reload preview/confirmation passed; MainPID stayed unchanged across reload; the unit remained active; reload audit records were verified; and the disposable unit was removed afterward.

Design/security review: `docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md`.

Closeout: `docs/history/M18-SAFE-ACTIONS-II.md`.

M18 is published, live, and recovery-aligned in v0.9.0.

## Approved roadmap status

The owner-approved source roadmap through M18 is complete.

No additional feature milestone is currently approved. Do not invent or begin another feature automatically.

A future publication/production rollout, a new feature roadmap, or additional Safe Action family requires a separate owner decision.

## Guardrails

Do not add a raw/unredacted export mode, raw journal export, arbitrary file inclusion, cloud upload, automatic sharing, or additional Safe Action families without a separate explicit design decision.

Do not drift into a generic server-control panel, arbitrary command execution, automatic remediation, multi-host controller architecture, or unrelated monitoring work.
