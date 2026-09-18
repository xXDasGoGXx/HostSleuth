# HostSleuth — Ordered Product Roadmap

HostSleuth stays a small, local-first Linux troubleshooting tool with two jobs:

1. remember meaningful host changes;
2. explain why a host/service/port is or is not reachable using deterministic evidence.

This roadmap is intentionally ordered. Complete one bounded milestone at a time. Do not skip ahead, broaden a milestone into a generic administration platform, or add unrelated monitoring features just because they are technically possible.

Consumer/product research may identify useful later workflows, but research does not silently reorder this roadmap or expand the active milestone.

## 1. M6 — Certificate Story / TLS Detective — COMPLETE

Delivered bounded read-only TLS/certificate diagnosis including handshake, served certificate metadata/fingerprint, hostname/trust evidence, local listener/container correlation, native Certbot lineage/renewal evidence, and conservative local-vs-served certificate comparison.

No renewal, reload, certificate installation, ACME account management, private-key handling, or arbitrary command execution was added.

## 2. Publish stable v0.3.0 — COMPLETE

Stable `v0.3.0` was published from accepted source commit:

`6e6b45ca5e4a4c54897ad69a3b20a377e68fccb1`

Publication verification included native linux/amd64 and linux/arm64 binaries, `SHA256SUMS`, public `mjmalleo/hostsleuth:0.3.0` plus `latest`, anonymous multi-platform registry verification, and bounded real-consumer execution.

## 3. M7 — Service Story — COMPLETE

Delivered bounded read-only systemd runtime/journal, process/listener ownership, expected-port collision, container-port, endpoint Diagnose/TLS, and retained-change correlation.

## 4. M8 — Incident Lens — COMPLETE

Delivered a bounded read-only +/- 15 minute incident window anchored from an exact time, retained event, or completed diagnosis. Current endpoint evidence is explicitly separated from historical context, and temporal proximity is never presented as proof of causation.

## 5. M9 — HostSleuth Workbench — COMPLETE

Delivered bounded read-only file identity/checksum comparison, DNS inspection, HTTP/redirect inspection, public certificate inspection/comparison, CLI/JSON API, and a dedicated Workbench UI.

Sensitive Workbench Web/API operations are loopback-only and expose no arbitrary commands, file contents, custom HTTP credentials/headers, or private-key viewing.

## 6. M10 — Reboot Story — COMPLETE

Delivered deterministic reboot detection from kernel boot IDs, exact boot-start evidence, bounded previous/current boot journal evidence, direct-evidence-only shutdown classification, current failed-service evidence, retained post-boot recovery correlation, non-causal nearby change context, CLI/API/UI integration, and schema-upgrade protection against false reboot events.

M10 remains read-only and does not infer reboot cause from temporal proximity.

Full closeout: `docs/history/M10-REBOOT-STORY.md`.

## 7. M11 — Optional Safe Actions — COMPLETE

Delivered one fixed native action only:

`service.restart`

Security/product boundaries:

- actions disabled by default;
- explicit `--enable-actions` opt-in;
- explicit per-service `--allow-restart-service UNIT` allowlist;
- conservative `.service` unit validation;
- no arbitrary command, argv, script, shell field, or generic systemd controller;
- trusted absolute `/usr/bin/systemctl` or `/bin/systemctl` for action evidence and execution;
- deterministic preview and exact confirmation;
- durable audit before execution;
- bounded timeout/output and before/after evidence;
- success only after observed `ActiveState=active`;
- loopback-only Action Web/API operations;
- Docker mode action-unavailable and no writable Docker socket;
- no automatic remediation.

Permanent CI includes a disposable real-systemd acceptance on an ephemeral GitHub runner.

Full closeout: `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md`.

## 8. Publish stable v0.4.0 — COMPLETE

Stable `v0.4.0` was published from exact accepted source commit:

`6566c505b32cc47d96384152a988736173d3f7cd`

Publication verification included GitHub linux/amd64 and linux/arm64 binaries, `SHA256SUMS`, independent amd64 execution, and public multi-platform Docker verification.

Full record: `docs/history/V0.4.0-PUBLICATION.md`.

## 9. Production / recovery v0.4.0 alignment — COMPLETE

The live Arcane-managed OMV deployment and `xXDasGoGXx/OMV-Docker-Rebuild` recovery definition were aligned on:

`mjmalleo/hostsleuth:0.4.0`

Recovery alignment merged at:

`ad2bd53ca3a469272eba6343c03936b7c1a04bc0`

Post-redeploy acceptance confirmed version/schema/mode, retained state/events, LAN health, and the Docker-mode Safe Actions boundary.

## 10. Redacted Evidence Bundle — COMPLETE

Owner accepted the threat model/redaction contract in `docs/design/REDACTED-EVIDENCE-BUNDLE.md`. PR #40 merged at:

`f1d756fa42baf71d9b762127b8d5d6ef6b77b796`

Delivered:

- typed current-Snapshot, recent-Event, and Safe-Action-audit export paths only;
- `hostsleuth evidence preview` with no archive write;
- explicit `hostsleuth evidence export` to one local ZIP;
- deterministic bundle-local aliases for sensitive host/infrastructure identifiers;
- credential/private-key/domain/IP/path/service/container/network/opaque-ID/fingerprint redaction;
- preservation of useful ports, prefix lengths, loopback/unspecified semantics, package/version data, statuses, UTC ordering, URL shape, and IPv6-CIDR shape;
- manifest, redacted payloads, SHA-256 checksums, owner-only output permissions, overwrite refusal, cleanup on failure, and bounded payload sizes;
- explicit exclusion of raw journal text, raw action command output, arbitrary file/config contents, action confirmation material, arbitrary Workbench inputs, cloud upload, and automatic sharing.

Acceptance included adversarial leak tests, full local formatting/vet/test/build validation, GitHub test plus amd64/arm64 image builds, Docker smoke, native-actions smoke, and disposable end-to-end CLI preview/export verification.

Full closeout: `docs/history/REDACTED-EVIDENCE-BUNDLE.md`.

## 11. Publish stable v0.5.0 — COMPLETE

Stable `v0.5.0` was published from exact source:

`04a53f8f0f3f48f7118a9ee9a688820cc000a340`

Release workflow run `35281793336` completed successfully.

Independent verification confirmed:

- published amd64 checksum and execution;
- version `v0.5.0 (04a53f8f0f3f)`;
- Evidence Bundle CLI present;
- public `mjmalleo/hostsleuth:0.5.0` and `latest` share OCI index `sha256:a17325980d5e9ec9760f9003aa8a9490962bb06ffbbd5a393a0ad31a218e480d`;
- linux/amd64 and linux/arm64 manifests are present.

Full record: `docs/history/V0.5.0-PUBLICATION.md`.

## 12. Production / recovery v0.5.0 alignment — COMPLETE

The live Arcane-managed OMV deployment and `xXDasGoGXx/OMV-Docker-Rebuild` recovery definition are aligned on:

`mjmalleo/hostsleuth:0.5.0`

Live acceptance confirmed:

- `/api/about` = `v0.5.0`;
- `/api/snapshot` = schema 4 / Docker mode;
- retained pre-upgrade events survived;
- Optional Safe Actions remain disabled/unavailable;
- Action Web/API remains loopback-only with HTTP 403 from LAN;
- a disposable v0.5.0 Evidence Bundle preview/export against a copy of live API evidence succeeded with expected bounded files, mode `0600`, and verified checksums.

Recovery alignment PR #5 merged at:

`9387d85acef8d19d913458cf45d48a765b6b8299`

Full record: `docs/history/V0.5.0-PRODUCTION-ALIGNMENT.md`.

## 13. M12 — Expected Endpoint Contracts — COMPLETE

The owner selected Expected Endpoint Contracts as the next bounded milestone.

M12 adds one on-demand read-only contract that can state:

- required `host:port` target;
- optional exact DNS address set;
- TLS expectation: ignore, present, verified, or forbidden;
- optional systemd service expected active;
- optional container expected running.

TCP reachability is always required.

HostSleuth reuses the existing Diagnose engine, renders ordered Expected-vs-Observed checks, and reports the first proven failure. Missing Docker/systemd evidence remains `unknown`; an earlier unknown does not mask a later proven failure.

Delivered interfaces:

- `hostsleuth contract` CLI;
- `GET /api/contract`;
- Web UI Expectations view with handoff to full Diagnose.

Local Go 1.24.13 format/vet/test/build and disposable HTTP/UI acceptance passed before opening the PR. PR #43 then passed the full GitHub CI matrix and merged at `1c193473d9220c34ec2820526df76076bfb41ce9`.

M12 remains read-only and adds no scheduler, alerting, persistent contract database, arbitrary file reads, new privilege, or remediation path.

Design: `docs/design/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

Closeout: `docs/history/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

## 14. Publish stable v0.6.0 — COMPLETE

Stable `v0.6.0` was published from exact accepted source:

`fcb08be51ae3da8cd20dc3929cf9d736b15f170c`

Release workflow run `35284864117` completed successfully. Independent verification confirmed native amd64/arm64 checksums, amd64 execution/version, the M12 `contract` command, and public `0.6.0` plus `latest` on OCI index `sha256:6b9f90209f477ba8213d9c4bf7db7996d6c52d2caae2cfce405822df8bc8ef2a` with linux/amd64 and linux/arm64 manifests.

Full record: `docs/history/V0.6.0-PUBLICATION.md`.

## 15. Production / recovery v0.6.0 alignment — COMPLETE

The live Arcane-managed OMV deployment and `xXDasGoGXx/OMV-Docker-Rebuild` disaster-recovery definition are aligned on:

`mjmalleo/hostsleuth:0.6.0`

Live acceptance confirmed:

- `/api/about` = `v0.6.0`;
- `/api/snapshot` = schema 4 / Docker mode;
- 100 retained events survived, including pre-upgrade history;
- M12 Expectations UI/API is present;
- a live Expected Endpoint Contract for `192.168.2.181:8787` with `tls=forbidden` and `container=hostsleuth` passed DNS, TCP, plaintext-TLS, and container expectations;
- Action Web/API remains loopback-only and returns HTTP 403 over LAN;
- one transient post-redeploy TCP/8787 listener-disappeared event was followed by the listener reappearing on the next scheduled snapshot with live HostSleuth process ownership, matching the host socket.

Recovery alignment PR #6 merged at:

`f78abbe5742c315d73cf85709ebfeba620311429`

Full record: `docs/history/V0.6.0-PRODUCTION-ALIGNMENT.md`.

## 16. Admin-grade roadmap — OWNER APPROVED

The owner explicitly approved the following direction after v0.6.0:

- DNS Detective / split-view resolver comparison;
- Reverse Proxy / Upstream Story;
- Deployment / Permissions Story;
- STARTTLS / Mail Service Story;
- Certificate Rollout Verification;
- one additional narrowly scoped Safe Action after a fresh threat/privilege review;
- continuous UI/UX and engineering polish so HostSleuth feels like a purpose-built admin troubleshooting console.

The product standard is: **collapse the multi-tool troubleshooting sequence an experienced admin normally reconstructs by hand into one deterministic evidence story.**

UI/UX is now a cross-cutting implementation track rather than end-of-project polish.

Detailed forward design: `docs/design/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md`.

## 17. M13 — DNS Detective + Admin Console v1 — COMPLETE

Goal:

> Which DNS view is this host actually seeing, which resolver disagrees, and is the difference consistent with split-view DNS?

Bounded implementation:

- always inspect the system resolver;
- optionally compare up to four explicitly supplied resolver IPs;
- custom resolver Web/API inputs are IP-only and DNS port 53 only;
- normalized A/AAAA plus canonical CNAME evidence for names;
- PTR evidence for IP inputs;
- per-resolver duration and bounded error evidence;
- address-scope classification: loopback/private/link-local/global/other;
- bounded runtime `/etc/resolv.conf` context;
- deterministic `agree`, `diverge`, `partial`, or `single` result;
- private/local vs global disagreement may be described as **consistent with** split-view DNS, never asserted as configuration fact without direct evidence;
- CLI `hostsleuth dns`;
- `GET /api/dns-detective`;
- dedicated DNS Detective Web UI.

Privacy boundary:

- no hidden public-resolver lookup;
- a hostname is sent only to the system resolver and custom resolver IPs the operator explicitly supplies;
- no DNS configuration changes, zone transfers, updates, polling, alerts, or remediation.

### Admin Console v1 — ships with M13

- sticky left navigation rail on desktop;
- global Quick Target bar;
- Diagnose / Expectations / DNS routing from one input;
- recent targets stored browser-local only;
- comfortable/compact density toggle stored browser-local only;
- `/` keyboard shortcut to focus the global target bar;
- wider evidence workspace and stronger semantic hierarchy;
- responsive horizontal navigation on smaller screens;
- reduced-motion support;
- no decorative monitoring graphs.

PR #47 passed test/format/vet/build, Docker smoke, native-actions smoke, and both linux/amd64 + linux/arm64 image builds on exact head `4d5d8f857e362217028efbdedb20e020f7a85d48`, then squash-merged to `main` at `182a27384a090f5538bb6d76a0c4dd917ce63932`.

Full closeout: `docs/history/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md`.

## 18. Publish stable v0.7.0 — COMPLETE

Stable `v0.7.0` was published from exact accepted source:

`236002106afd6aa042fd131c0edc0f3455b9cfdf`

Release workflow run `35292714676` completed successfully.

Independent verification confirmed:

- native amd64/arm64 assets match `SHA256SUMS`;
- amd64 execution reports `v0.7.0 (236002106afd)`;
- public `hostsleuth dns` works and custom non-53 resolver input fails closed;
- public `0.7.0` and `latest` share OCI index `sha256:3663e8c483b67de72f3a0e26fd80e9e3686319d9b2bafe602979cd790e2ce2bb`;
- linux/amd64 and linux/arm64 manifests are present.

Full record: `docs/history/V0.7.0-PUBLICATION.md`.

## 19. Production / recovery v0.7.0 alignment — COMPLETE

The live Arcane-managed deployment and `xXDasGoGXx/OMV-Docker-Rebuild` recovery definition are aligned on:

`mjmalleo/hostsleuth:0.7.0`

Live acceptance confirmed:

- `/api/about` = `v0.7.0`;
- snapshot schema 4 / Docker mode;
- live container image = `mjmalleo/hostsleuth:0.7.0`;
- 100 retained events, including pre-upgrade history;
- Action Web/API remains HTTP 403 from LAN;
- loopback capability remains `enabled=false` / `available=false`;
- DNS Detective returned `agree` with an explicitly supplied local resolver;
- Admin Console and DNS Detective dynamic UI markers are present in live DOM.

Recovery PR #7 merged at:

`80e7c7c08b5905c7fbad158e4b11e6e4b4ced1f9`

Full record: `docs/history/V0.7.0-PRODUCTION-ALIGNMENT.md`.

## 20. M14 — Reverse Proxy / Upstream Story — COMPLETE

Goal:

> DNS and 443 are fine; where does the request path actually break?

Implemented v1:

- explicit public HTTP/HTTPS URL plus explicit expected upstream URL;
- deterministic public DNS/route/TCP/TLS/HTTP path;
- local listener/Docker publication context when available;
- deterministic upstream DNS/route/TCP/TLS/HTTP path;
- native upstream Host/SNI probe;
- public Host/SNI probe against the same explicit upstream when identities differ;
- first proven failure highlighted;
- 4xx warning / 5xx failure semantics;
- HEAD-only metadata probes with no body/credentials/cookies/arbitrary headers;
- same-host redirect following only; cross-host redirects are recorded and stopped;
- query strings redacted from returned URL/Location evidence;
- CLI/API/Web UI;
- Admin Console v2 visual path, Quick Target Proxy action, and bounded evidence-copy controls.

Local format/vet/full-test/race-test/build and disposable API/UI/headless-browser acceptance passed before opening a PR.

M14 remains read-only and adds no config parser, upstream discovery, authentication flow, polling, proxy reload/edit, or remediation.

Design: `docs/design/M14-REVERSE-PROXY-UPSTREAM-STORY.md`.

Closeout: `docs/history/M14-REVERSE-PROXY-UPSTREAM-STORY.md`.

PR #51 passed test/format/vet/build, Docker smoke including the proxy-story API, native-actions smoke, and both linux/amd64 + linux/arm64 image builds on exact head `e7737c3342c628d80bf17a3b60a2cc5b4ead48c0`, then squash-merged to `main` at `135a73ffb333c1e4ac5135f93db7bac3dac5cdae`.

Full closeout: `docs/history/M14-REVERSE-PROXY-UPSTREAM-STORY.md`.

## 21. Publish stable v0.8.0 — COMPLETE

Stable `v0.8.0` was published from exact accepted source:

`28b8d88ec018782353840dee8528d765c4056e89`

Release workflow run `35294912224` completed successfully.

Independent verification confirmed:

- native amd64/arm64 assets match `SHA256SUMS`;
- amd64 execution reports `v0.8.0 (28b8d88ec018)`;
- public `hostsleuth proxy` command is present and invalid public URL input fails closed;
- public `0.8.0` and `latest` share OCI index `sha256:152255f44e451995767cd40b64b99879feb0ed0ae28a4fddfb4a269a55fc11df`;
- linux/amd64 and linux/arm64 manifests are present.

Full record: `docs/history/V0.8.0-PUBLICATION.md`.

## 22. Production / recovery v0.8.0 alignment — COMPLETE

The live Arcane-managed deployment and `xXDasGoGXx/OMV-Docker-Rebuild` disaster-recovery definition are aligned on:

`mjmalleo/hostsleuth:0.8.0`

Live acceptance confirmed:

- `/api/about` = `v0.8.0`;
- snapshot schema 4 / Docker mode;
- live container image = `mjmalleo/hostsleuth:0.8.0`;
- 100 retained events, including pre-upgrade history;
- Action Web/API remains HTTP 403 from LAN;
- loopback action capability remains `enabled=false` / `available=false`;
- a live Proxy Path from Arcane's LAN endpoint to its loopback upstream passed all tested public and upstream DNS/route/TCP/HTTP stages;
- native upstream and public-Host variants both returned HTTP 200;
- Proxy Path and Admin Console v2 dynamic UI markers are present in the live DOM.

Recovery PR #8 merged at:

`290e7e6a6781ad29cbc7c8296d00877fb8489fd0`

Full record: `docs/history/V0.8.0-PRODUCTION-ALIGNMENT.md`.

## 23. M15 — Deployment / Permissions Story — COMPLETE

Goal:

> The process is running; why can it not use this path/socket/port?

Delivered bounded source implementation:

- one native systemd service plus one explicit absolute path/socket path;
- observed PID, effective UID/GID, supplementary groups, working directory, executable, and effective capability mask;
- exact ownership/mode parent-directory traversal chain;
- deterministic read/write/execute/traverse/connect decisions;
- capability-aware UID 0 DAC reasoning;
- relevant Docker bind-mount metadata when already available;
- CLI/API/dedicated Permissions Web UI;
- searchable evidence table and visual permission-chain story;
- no recursive filesystem crawl, file contents, remediation, or privilege expansion;
- Docker deployment mode fails closed for native service identity rather than crossing the container boundary.

Native pass/fail acceptance and disposable HTTP/UI acceptance passed.

PR #55 passed the full GitHub CI matrix on exact head:

`78bd31d8cdb34f9f5e592e57f14f2e8537e79b41`

and squash-merged to `main` at:

`d64bb387c8f9efcf5dcb9f814f54b519ee231205`

Design: `docs/design/M15-DEPLOYMENT-PERMISSIONS-STORY.md`.

Closeout: `docs/history/M15-DEPLOYMENT-PERMISSIONS-STORY.md`.

Stable/public/live/recovery remain on v0.8.0; M15 is development source until a separate publication/rollout decision.

## 24. M16 — STARTTLS / Mail Service Story — COMPLETE

Goal:

> The SMTP/IMAP port is open; did STARTTLS actually negotiate correctly?

Delivered bounded source implementation:

- SMTP STARTTLS, IMAP STARTTLS, and POP3 STLS;
- explicit protocol plus explicit host:port input only;
- deterministic TCP, greeting, capability, upgrade, TLS, certificate validity, hostname, and trust stages;
- fixed pre-authentication protocol commands only;
- bounded 4 KiB lines, 64 response lines, and fixed deadlines;
- shared existing TLS/certificate interpretation;
- CLI/API/dedicated STARTTLS Admin Console view;
- no credentials, authentication, mail submission, mailbox access, message contents, arbitrary protocol commands, or remediation.

Local full test/race/vet/build and disposable SMTP CLI/API/UI acceptance passed.

PR #57 passed the complete GitHub CI matrix on exact head:

`579fff79889d5ab0c3430135d3f5ceba4dc28400`

and squash-merged to `main` at:

`e5663d5883acdd2d859ff57b39a447c0018790b3`

Design: `docs/design/M16-STARTTLS-MAIL-SERVICE-STORY.md`.

Closeout: `docs/history/M16-STARTTLS-MAIL-SERVICE-STORY.md`.

Stable/public/live/recovery remain on v0.8.0; M16 is development source until a separate publication/rollout decision.

## 25. M17 — Certificate Rollout Verification — COMPLETE

Goal:

> I renewed/replaced the certificate; which endpoint is still serving the old one?

Delivered bounded source implementation:

- exactly one expected source: SHA-256 fingerprint or explicit reference direct-TLS endpoint;
- up to 16 explicit direct-TLS host:port endpoints;
- numeric-port validation and duplicate rejection;
- deterministic MATCH / MISMATCH / UNKNOWN rollout identity;
- separate certificate validity, hostname, and trust health;
- bounded four-worker probing with user endpoint order preserved;
- CLI/API/dedicated Cert Rollout Admin Console matrix;
- no arbitrary certificate-file reads, private-key reads, renewal, install, reload, ACME management, or remediation.

Arbitrary certificate-file input was intentionally omitted because an arbitrary path cannot be proven safe before opening it; fingerprint/reference sources preserve the no-private-key-read boundary.

PR #59 passed the complete GitHub CI matrix on final exact head:

`32f0ad2dc3211e5af38a643a5d2f38081d440645`

Native two-certificate CLI/API/UI acceptance also passed.

PR #59 squash-merged to `main` at:

`f12bbd8aacb9689bb31d5b9b6535479882cb6ce6`

Design: `docs/design/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`.

Closeout: `docs/history/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`.

Stable/public/live/recovery remain on v0.8.0; M17 is development source until a separate publication/rollout decision.

## 26. M18 — Safe Actions II — COMPLETE

The owner explicitly approved exactly one additional action: `service.reload`, under the accepted M18 security review. No other candidate was approved.

Delivered bounded source implementation:

- fixed action ID `service.reload`;
- independent `--allow-reload-service UNIT` allowlist;
- existing global `--enable-actions` opt-in;
- native mode only;
- trusted absolute systemctl path and fixed reload argv only;
- loaded + active + `CanReload=yes` fail-closed precondition;
- exact preview and `RELOAD UNIT.service` confirmation;
- durable requested audit before execution;
- bounded serialized execution/output;
- no reload-or-restart helper and no restart fallback;
- success only after command success and `ActiveState=active`;
- existing loopback-only Action Web/API boundary;
- Actions UI driven only by fixed server-provided capabilities and allowlists;
- Docker mode remains unavailable for both native systemd actions.

Local full test/race/vet/build/JS/diff validation passed.

PR #62 passed the complete GitHub CI matrix on exact head:

`c46dd72ad806c688970244836ce26e3ca01eec88`

CI run:

`35309972305` — success

The native CI fixture performed a real reload on a disposable reload-capable systemd service, proved restart-only allowlisting could not authorize reload, verified exact reload argv/confirmation and audit records, confirmed the MainPID stayed unchanged, and confirmed the service remained active afterward.

PR #62 squash-merged to `main` at:

`cc9c9727fe19786eba03d0b7ef51c7fae7ac8ab1`

Security review:

`docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md`

Closeout:

`docs/history/M18-SAFE-ACTIONS-II.md`

M18 is published, live, and recovery-aligned in stable v0.9.0.

The owner-approved feature roadmap through M18 is complete. No additional feature milestone is approved.

## v1.0 Readiness / Hardening — COMPLETE / PUBLISHED IN v1.0.0

The owner-approved feature-free hardening cycle is complete and published in stable v1.0.0.

This is not M19 and added no troubleshooting story, Safe Action family, privilege surface, remediation feature, monitoring architecture, or cloud behavior.

Readiness contract:

`docs/design/V1.0-READINESS.md`

Accepted slice 1:

- persisted-state compatibility regression;
- Admin Console keyboard/accessibility semantics;
- visible focus and reduced motion;
- responsive browser acceptance;
- explicit served-asset budgets.

PR #66 passed full CI and merged at:

`771b923c06d38ac528804468effbe56ffd4c8f78`

Accepted slice 2:

- native release binary checksum verification against `SHA256SUMS`;
- real disposable native-install acceptance of published v0.9.0;
- restrictive browser security headers;
- permanent Docker-smoke security-header checks.

PR #67 passed the complete six-job CI matrix and merged at:

`ec51b21aca8c4a9de50d2530d2200ba576b45989`

Final disposable state acceptance passed in both directions:

`published v0.9.0 -> current candidate -> published v0.9.0`

Release/rollback procedure:

`docs/design/V1.0-RELEASE-CHECKLIST.md`

Readiness closeout:

`docs/history/V1.0-READINESS.md`

Readiness closeout PR #68 passed the complete six-job CI matrix on exact head `2853f3a369a95817524935a93db6904d2ad3d31c` and squash-merged to `main` at `f93330755e229609f86add144a20d3ce2234cd04`.

The owner approved publication. Stable/public v1.0.0 was published from exact source `27971e67ad7ea875d83f5925a83c1991b0eb2b0f`; release workflow `35358798799` succeeded and public native/Docker artifacts were independently verified.

Live Arcane production and recovery `main` remain on v0.9.0 during the production-acceptance gate. Recovery PR #11 stages only the image pin to `mjmalleo/hostsleuth:1.0.0` and remains intentionally unmerged until live acceptance passes.

Publication record: `docs/history/V1.0.0-PUBLICATION.md`.

## Continuous polish / quality track — APPROVED

Bounded non-feature improvements may proceed between milestones:

- CI concurrency/cancellation for obsolete PR runs;
- stronger DNS/proxy/TLS integration fixtures;
- keyboard/accessibility regression checks;
- responsive UI acceptance;
- public-safe screenshots matching the current console;
- copy-to-clipboard for bounded evidence blocks;
- consistent status vocabulary across all stories;
- stable deep links where useful;
- browser-local recent targets/preferences only;
- performance budget for initial UI load and large retained-event rendering.

## Guardrails that remain in force

Do not drift into:

- multi-host controller/agent architecture;
- generic network-device/SNMP monitoring;
- time-series graphing/monitoring platform behavior;
- arbitrary web terminal or command execution;
- generic package/firewall/configuration administration;
- AI-generated causal claims presented as evidence;
- automatic remediation;
- broad privilege expansion simply to make features easier.

Do not add a raw/unredacted evidence mode, raw journal export, arbitrary file inclusion, hidden cloud upload/telemetry, automatic sharing, or new Safe Action family without a separate explicit design decision.

HostSleuth should feel powerful because it connects deterministic evidence into answers people actually need, not because it exposes every system control in a browser.
