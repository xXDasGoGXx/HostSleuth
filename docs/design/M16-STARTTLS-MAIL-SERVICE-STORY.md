# M16 — STARTTLS / Mail Service Story

Status: implementation in progress.

## Product question

> The SMTP/IMAP port is open; did STARTTLS actually negotiate correctly?

M16 adds one bounded read-only mail-protocol story for an explicit host:port and explicit protocol.

Supported protocols:

- SMTP STARTTLS;
- IMAP STARTTLS;
- POP3 STLS.

POP3 is included because it fits the same bounded pre-authentication capability -> upgrade -> TLS model without adding mailbox or credential behavior.

## Protocol standards

The implementation follows the bounded upgrade sequences defined by:

- SMTP STARTTLS: RFC 3207;
- IMAP STARTTLS: RFC 9051 / RFC 2595 compatibility model;
- POP3 STLS: RFC 2595.

HostSleuth treats capabilities observed before TLS as pre-upgrade evidence only. It does not reuse them as authenticated or post-TLS truth.

## Inputs

CLI/API accept:

- one explicit protocol: smtp, imap, or pop3;
- one explicit host:port target.

No autodiscovery, MX lookup, account identity, username, password, OAuth token, mailbox, or message selector is accepted.

## Bounded protocol sequence

### SMTP

1. open TCP;
2. read bounded 220 greeting;
3. send fixed EHLO hostsleuth.invalid;
4. read bounded 250 capability response;
5. require advertised STARTTLS;
6. send fixed STARTTLS;
7. require 220 upgrade response;
8. perform TLS handshake on the same connection;
9. inspect certificate/TLS metadata;
10. close.

### IMAP

1. open TCP;
2. read bounded unauthenticated * OK greeting;
3. send fixed a001 CAPABILITY;
4. read bounded tagged response;
5. require advertised STARTTLS;
6. send fixed a002 STARTTLS;
7. require tagged OK;
8. perform TLS handshake on the same connection;
9. inspect certificate/TLS metadata;
10. close.

A PREAUTH greeting is not probed further because STARTTLS is a pre-authentication command.

### POP3

1. open TCP;
2. read bounded +OK greeting;
3. send fixed CAPA;
4. read bounded capability list;
5. require advertised STLS;
6. send fixed STLS;
7. require +OK;
8. perform TLS handshake on the same connection;
9. inspect certificate/TLS metadata;
10. close.

## Evidence stages

The deterministic stage order is:

1. TCP;
2. service greeting;
3. STARTTLS/STLS capability;
4. protocol upgrade response;
5. TLS handshake;
6. certificate validity;
7. certificate hostname;
8. certificate trust.

The first proven failed stage is first_problem.

A later proven failure outranks an earlier unknown when computing overall status, while first_problem remains the first failed protocol stage in order.

## TLS evidence

M16 reuses the existing HostSleuth TLS/certificate evidence model on the already-upgraded connection:

- negotiated TLS version;
- cipher suite;
- served leaf certificate metadata;
- chain subjects;
- SHA-256 fingerprint;
- validity window / days remaining;
- requested-hostname match;
- trust verification against the HostSleuth host trust store.

The existing direct-TLS probe was refactored so both implicit TLS and STARTTLS use the same certificate interpretation rather than maintaining two implementations.

## Read bounds

Protocol input is bounded:

- 4 KiB maximum protocol line;
- 64 maximum response/capability lines per bounded response;
- fixed I/O deadlines;
- bounded returned evidence strings.

Oversized or unterminated protocol responses fail closed.

## Security / privacy boundary

M16 sends only the fixed protocol commands listed above.

It does not send or accept:

- usernames;
- passwords;
- OAuth tokens;
- AUTH / AUTHENTICATE;
- LOGIN;
- USER / PASS;
- MAIL FROM / RCPT TO / DATA;
- mailbox SELECT / EXAMINE / FETCH / SEARCH;
- POP3 STAT / LIST / RETR / TOP;
- message bodies;
- attachments;
- arbitrary protocol commands;
- arbitrary client identity strings.

The SMTP EHLO identity is the fixed reserved name hostsleuth.invalid; the local hostname is not disclosed.

M16 does not submit mail, read mail, authenticate, alter server state, or remediate configuration.

## Interfaces

CLI:

    hostsleuth starttls --protocol smtp mail.example.com:25

API:

    GET /api/starttls-story?protocol=smtp&target=mail.example.com:25

Admin Console view:

- protocol selector;
- explicit host:port input;
- visual protocol -> TLS stage flow;
- greeting/capability/upgrade evidence;
- TLS/certificate identity/trust/expiry evidence;
- visible no-credentials/no-message-content boundary.

## Validation status

Passed locally:

- focused SMTP/IMAP/POP3 fixtures with real TLS upgrade;
- missing capability fixture;
- rejected-upgrade fixture;
- oversized-line fixture;
- first-failure precedence;
- Go 1.24.13 full test/race/vet/build;
- all Web JavaScript syntax;
- git diff check;
- disposable SMTP CLI/API/UI acceptance using a local self-signed localhost certificate.

The disposable fixture produced a complete eight-stage story: TCP, greeting, capability, upgrade, and TLS all passed; certificate validity and hostname passed; trust failed exactly as expected for the self-signed certificate.

Still required before source closeout:

- exact-head GitHub CI matrix;
- focused PR merge;
- post-merge closeout/history alignment.

Publication, production rollout, and recovery alignment remain separate owner-controlled gates.
