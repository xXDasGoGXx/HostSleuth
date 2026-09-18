# M16 — STARTTLS / Mail Service Story Closeout

Status: source implementation complete and merged on 2026-09-17. Stable/public/live/recovery remain aligned on v0.8.0; publication and rollout are separate gates.

## Product question

> The SMTP/IMAP port is open; did STARTTLS actually negotiate correctly?

M16 delivers a bounded read-only pre-authentication STARTTLS story for one explicit host:port and one explicit mail protocol.

Supported protocols:

- SMTP STARTTLS;
- IMAP STARTTLS;
- POP3 STLS.

POP3 remained inside scope because it uses the same bounded capability -> upgrade -> TLS model without requiring credentials, mailbox access, or message retrieval.

## Standards model

The protocol flow is based on the STARTTLS/STLS upgrade semantics defined by:

- RFC 3207 for SMTP STARTTLS;
- RFC 9051 / RFC 2595 compatibility semantics for IMAP STARTTLS;
- RFC 2595 for POP3 STLS.

Capabilities observed before TLS are treated only as pre-upgrade evidence. HostSleuth does not present them as post-TLS or authenticated truth.

## Deterministic evidence stages

M16 reports the same ordered story across the supported protocols:

1. TCP connection;
2. service greeting;
3. STARTTLS/STLS capability advertisement;
4. protocol upgrade response;
5. TLS negotiation;
6. certificate validity;
7. certificate hostname;
8. certificate trust.

The first proven failed stage is returned as first_problem.

A proven failure outranks unresolved evidence for the overall result, while the first failed stage remains the primary diagnostic location.

## SMTP flow

HostSleuth:

1. reads a bounded 220 greeting;
2. sends the fixed command EHLO hostsleuth.invalid;
3. reads a bounded 250 capability response;
4. requires advertised STARTTLS;
5. sends fixed STARTTLS;
6. requires a 220 upgrade response;
7. performs the TLS handshake on the same connection;
8. inspects TLS/certificate metadata;
9. closes the connection.

The EHLO identity is deliberately fixed to the reserved invalid domain so the local hostname is not disclosed.

## IMAP flow

HostSleuth:

1. reads a bounded unauthenticated * OK greeting;
2. refuses to continue from PREAUTH state;
3. sends fixed a001 CAPABILITY;
4. requires a tagged OK response with STARTTLS advertised;
5. sends fixed a002 STARTTLS;
6. requires a tagged OK response;
7. performs the TLS handshake on the same connection;
8. inspects TLS/certificate metadata;
9. closes the connection.

No LOGIN, AUTHENTICATE, SELECT, EXAMINE, FETCH, SEARCH, or mailbox command is sent.

## POP3 flow

HostSleuth:

1. reads a bounded +OK greeting;
2. sends fixed CAPA;
3. reads a bounded multiline capability response;
4. requires advertised STLS;
5. sends fixed STLS;
6. requires a +OK response;
7. performs the TLS handshake on the same connection;
8. inspects TLS/certificate metadata;
9. closes the connection.

No USER, PASS, STAT, LIST, RETR, TOP, or message command is sent.

## TLS / certificate evidence

M16 reuses the existing HostSleuth TLS evidence implementation on the already-upgraded connection rather than maintaining a second certificate parser.

Returned evidence includes:

- negotiated TLS version;
- cipher suite;
- served certificate subject and issuer;
- serial;
- SANs;
- validity window and days remaining;
- SHA-256 fingerprint;
- chain subjects;
- requested-hostname match;
- trust verification against the HostSleuth host trust store.

The existing direct TLS path was refactored only enough to share the same post-handshake interpretation with STARTTLS.

## Protocol bounds and hardening

Protocol reads are deliberately bounded:

- 4 KiB maximum line length;
- 64 maximum response/capability lines;
- fixed connection and I/O deadlines;
- bounded returned evidence.

The parser fails closed on:

- oversized lines;
- unterminated lines;
- malformed SMTP reply codes;
- malformed SMTP multiline separators;
- missing STARTTLS/STLS capability;
- rejected upgrade commands;
- invalid protocol or host:port input.

M16 never falls through from a failed capability/upgrade stage into TLS.

## Interfaces

CLI:

    hostsleuth starttls --protocol smtp mail.example.com:25
    hostsleuth starttls --protocol imap mail.example.com:143
    hostsleuth starttls --protocol pop3 mail.example.com:110

HTTP API:

    GET /api/starttls-story?protocol=smtp&target=mail.example.com:25

Web UI:

- dedicated STARTTLS navigation/view;
- SMTP / IMAP / POP3 protocol selector;
- explicit host:port target;
- visual protocol-to-TLS stage flow;
- greeting / capability / upgrade evidence;
- TLS version/cipher/hostname/trust evidence;
- served-certificate metadata;
- visible no-credentials/no-mail safety boundary.

## Security / privacy boundary

M16 remains read-only and adds no privilege.

It does not accept or send:

- usernames;
- passwords;
- OAuth tokens;
- SMTP AUTH;
- IMAP AUTHENTICATE or LOGIN;
- POP3 USER/PASS;
- MAIL FROM / RCPT TO / DATA;
- mailbox selection or message retrieval commands;
- message bodies;
- attachments;
- arbitrary protocol commands;
- arbitrary client identity strings.

It does not:

- submit mail;
- read mail;
- authenticate;
- edit mail-server configuration;
- reload/restart mail services;
- renew/install certificates;
- access private keys;
- remediate automatically.

## Local engineering acceptance

Local validation used Go 1.24.13 on an isolated /tmp clone.

Passed:

- gofmt;
- git diff --check;
- go test ./...;
- go test -race ./internal/core;
- go vet ./...;
- native build;
- syntax check for every Web JavaScript asset.

Focused fixtures cover:

- successful SMTP STARTTLS upgrade;
- successful IMAP STARTTLS upgrade;
- successful POP3 STLS upgrade;
- missing STARTTLS capability;
- rejected STARTTLS command;
- oversized line rejection;
- unterminated line rejection;
- malformed SMTP reply separator;
- POP3-specific STLS wording;
- first-failure precedence.

The protocol fixtures assert the exact bounded pre-authentication commands expected by each protocol.

## Disposable CLI / API / UI acceptance

A disposable SMTP fixture was started on loopback with a one-day self-signed localhost certificate.

The built M16 CLI and a temporary HostSleuth server both exercised the same fixture.

Observed story:

- TCP: pass;
- SMTP greeting: pass;
- STARTTLS capability: pass;
- STARTTLS command: pass;
- TLS handshake: pass;
- certificate validity: pass;
- hostname: pass;
- trust: fail, exactly as expected for the self-signed fixture.

The returned stage order was:

    tcp
    greeting
    capability
    upgrade
    tls
    certificate
    hostname
    trust

The served JavaScript and CSS also contained the STARTTLS view and visual flow assets.

Both disposable processes were stopped after acceptance. Production HostSleuth was not touched.

## GitHub CI acceptance

PR #57 exact head:

    579fff79889d5ab0c3430135d3f5ceba4dc28400

CI run:

    35303552155

All jobs passed:

- format / vet / full tests / Web JavaScript syntax / native build;
- native Safe Actions regression smoke;
- Docker smoke, including STARTTLS UI marker and invalid-input API boundary;
- linux/amd64 image build;
- linux/arm64 image build.

## Merge acceptance

PR #57 squash-merged to main at:

    e5663d5883acdd2d859ff57b39a447c0018790b3

## Release boundary

M16 source is complete on main.

Stable/public/live/recovery remain:

    v0.8.0 / mjmalleo/hostsleuth:0.8.0

No image publication, Arcane redeploy, disaster-recovery change, mail-server change, certificate change, or Safe Action expansion was performed for M16.

The next approved source milestone is M17 — Certificate Rollout Verification.
