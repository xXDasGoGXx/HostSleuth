# HostSleuth — Admin-Grade Product Roadmap

This roadmap promotes the owner's approved direction after v0.6.0: HostSleuth should answer the troubleshooting questions experienced admins repeatedly have to reconstruct by hand, while keeping deterministic evidence, local-first operation, bounded privileges, and a UI that feels like a purpose-built troubleshooting console rather than a collection of forms.

Stable/public/live/recovery currently remain aligned on v0.6.0 until the next release gate.

## Product rule — "Finally, someone who gets it"

A feature is worth adding when it collapses a frustrating multi-tool troubleshooting sequence into one deterministic story.

HostSleuth should preferentially answer questions such as:

- Which resolver gave this answer, and who disagrees?
- DNS is right and 443 is open, so where does the reverse-proxy path actually break?
- The service is "active"; can its runtime identity actually read/write/execute what it needs?
- STARTTLS says the port is open; did negotiation, hostname, trust, and served certificate all succeed?
- I renewed the certificate; which endpoint is still serving the old one?
- If a safe corrective action exists, can HostSleuth preview exactly what will happen, require explicit confirmation, audit it, and prove the postcondition?

The tool must not become a generic monitoring suite, shell, or automatic-remediation platform.

## Cross-cutting UX track — Admin Console

UI/UX is an implementation track, not end-of-project polish.

Every milestone should improve operator flow where it naturally fits.

### Admin Console v1 — M13

- persistent desktop navigation rail;
- global target bar available from every view;
- one-click routing from a target to Diagnose, Expectations, or DNS Detective;
- recent-target memory stored only in the browser;
- comfortable/compact density toggle stored only in the browser;
- stronger semantic status hierarchy and evidence cards;
- keyboard shortcut / to focus the global target bar;
- responsive fallback to horizontal/mobile navigation;
- no graphing for graphing's sake.

### Admin Console v2+ — later milestones

- visual request-path story for reverse proxy/upstream diagnosis;
- filter/search for dense evidence;
- copyable evidence blocks and stable deep links;
- comparison matrices for rollout verification;
- protocol-step storytelling for STARTTLS;
- accessibility, keyboard flow, reduced-motion support, and performance budgets as release gates.

## M13 — DNS Detective + Admin Console v1 — ACTIVE

Goal: answer "Which DNS view is this host actually seeing, and how does it differ from the resolver I expected?"

Bounded v1:

- always inspect the system resolver;
- optionally compare up to four explicitly supplied DNS resolver IPs;
- custom Web/API resolvers are IP addresses on DNS port 53 only;
- collect normalized A/AAAA address answers plus canonical CNAME evidence;
- PTR when the input is an IP;
- record per-resolver success/error and lookup duration;
- classify returned addresses as loopback/private/link-local/global/other;
- parse bounded runtime resolver/search-domain evidence from /etc/resolv.conf;
- deterministically report agree/diverge/partial/single-resolver;
- label a private-vs-global divergence as "consistent with split-view DNS" without claiming intent;
- CLI, API, dedicated Web UI view;
- no polling, no DNS changes, no hidden public-resolver query, and no automatic hostname disclosure to third-party DNS.

Admin Console v1 ships with M13.

Release target after accepted M13 source: v0.7.0.

## M14 — Reverse Proxy / Upstream Story

Goal: answer "The proxy is reachable; where does the request path break?"

Bounded design direction:

- user supplies public-facing endpoint and expected upstream endpoint;
- deterministic path story: DNS -> route/TCP -> proxy listener -> served TLS/HTTP -> upstream DNS/route/TCP -> upstream TLS/HTTP;
- show redirect behavior and host/SNI distinctions;
- correlate local Docker publication/listener evidence where available;
- never parse arbitrary proxy configs in v1;
- no proxy reconfiguration or reload;
- visual request-path UI that highlights the first proven failure.

## M15 — Deployment / Permissions Story

Goal: answer "The process exists, so why can't it use this path/socket/port?"

Bounded design direction:

- native-mode service identity, UID/GID, working directory, selected path metadata, ownership/mode, parent-directory traversal permissions, executable identity, and relevant bind-mount metadata;
- explicit user-supplied path only; no recursive filesystem crawl;
- no file contents;
- deterministic read/write/execute/traverse reasoning from observed metadata;
- Docker/native distinctions remain explicit;
- searchable evidence table and "permission chain" UI.

## M16 — STARTTLS / Mail Service Story

Goal: answer "SMTP/IMAP is reachable; did STARTTLS actually negotiate correctly?"

Initial protocols:

- SMTP STARTTLS;
- IMAP STARTTLS;
- POP3 STARTTLS only if protocol design remains equally bounded.

Evidence:

- greeting;
- advertised STARTTLS capability;
- negotiation success/failure;
- negotiated TLS version/cipher;
- served certificate;
- hostname/trust/expiry evidence;
- protocol stage at first failure.

No authentication credentials, mail submission, mailbox access, or message contents.

## M17 — Certificate Rollout Verification

Goal: answer "I replaced the certificate; which endpoint has not picked it up?"

Bounded design direction:

- compare one expected certificate fingerprint/file or one reference endpoint against multiple explicitly supplied endpoints;
- show served fingerprint, subject/SAN, validity, hostname/trust, and match/mismatch;
- deterministic endpoint matrix;
- no renewal, installation, reload, private-key reads, or ACME management;
- exportable redacted comparison evidence only through an explicitly designed typed exporter extension.

## M18 — Safe Actions II

Goal: add one additional action only if its operational value justifies the privilege cost.

Before implementation:

1. compare concrete candidate actions;
2. select exactly one;
3. write threat model and privilege analysis;
4. preserve explicit enablement, allowlist, preview, exact confirmation, durable audit, bounded execution, and postcondition proof;
5. do not weaken the Docker security posture merely to make an action convenient.

No generic command runner or generic service/container controller.

## Later polish / quality track

Approved non-feature work can proceed between milestones when bounded:

- CI concurrency/cancellation to suppress obsolete runs;
- stronger integration fixtures for DNS/proxy/TLS edge cases;
- keyboard/accessibility regression checks;
- responsive UI acceptance;
- public-safe screenshots that reflect the current console;
- copy-to-clipboard for bounded evidence blocks;
- consistent status language across Diagnose, Expectations, DNS, Reboot, Incident, Workbench, and Safe Actions;
- performance budget for initial UI load and large retained-event rendering.

## Guardrails

Still prohibited without a separate explicit design decision:

- arbitrary shell/command execution;
- automatic remediation;
- generic server-control-panel behavior;
- multi-host controller/agent architecture;
- broad recursive config/filesystem ingestion;
- unredacted support export;
- hidden cloud upload/telemetry;
- polling/alerting/time-series monitoring as a product direction;
- broad privilege expansion for convenience.
