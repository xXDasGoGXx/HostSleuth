# M17 — Certificate Rollout Verification Closeout

Status: source implementation complete and merged on 2026-09-17. Stable/public/live/recovery remain aligned on v0.8.0; publication and rollout are separate gates.

## Product question

> I renewed/replaced the certificate; which endpoint is still serving the old one?

M17 delivers a bounded read-only certificate rollout comparison for multiple explicit direct-TLS endpoints.

## Expected source

Exactly one expected certificate source is accepted:

- a SHA-256 certificate fingerprint; or
- one explicit reference direct-TLS endpoint in host:port form.

Arbitrary certificate-file paths are intentionally not exposed. HostSleuth cannot prove an arbitrary user-supplied file is a public certificate rather than a private key before opening it. Fingerprint/reference sources satisfy rollout verification while preserving the existing no-private-key-read boundary.

## Endpoint scope

The rollout story accepts:

- at least one endpoint;
- at most 16 endpoints;
- explicit direct-TLS host:port targets;
- numeric ports only;
- no duplicate endpoints.

At most four endpoint probes run concurrently, while returned rows preserve user-supplied endpoint order.

There is no endpoint autodiscovery, DNS expansion, reverse-proxy discovery, certificate-path discovery, or STARTTLS autodetection. Mail STARTTLS remains under M16.

## Comparison model

Expected fingerprint input is SHA-256 only and normalizes optional colons/whitespace to uppercase 64-character hexadecimal.

When a reference endpoint is selected, HostSleuth first establishes a usable served leaf certificate from that explicit endpoint. If the reference certificate cannot be observed, target comparison stops rather than inventing an expected identity.

Every endpoint returns independent rollout and certificate-health evidence:

- TLS handshake status/error;
- TLS version and cipher;
- served leaf subject;
- issuer;
- serial;
- SANs;
- validity window;
- days remaining;
- SHA-256 fingerprint;
- hostname verification;
- trust verification;
- rollout match state.

Rollout match state is:

- match;
- mismatch;
- unknown.

Certificate health remains separate. A matching fingerprint therefore does not hide an expired certificate, hostname mismatch, or trust failure.

## Deterministic result precedence

1. any known fingerprint mismatch -> fail;
2. otherwise any unknown fingerprint -> unknown;
3. otherwise any validity/hostname/trust failure -> fail;
4. otherwise incomplete certificate-health evidence -> unknown;
5. otherwise -> pass.

The first endpoint responsible for the overall result is returned as first_problem.

## Interfaces

CLI:

    hostsleuth cert-rollout --fingerprint SHA256 --endpoint edge-a.example.com:443 --endpoint edge-b.example.com:443

or:

    hostsleuth cert-rollout --reference reference.example.com:443 --endpoint edge-a.example.com:443 --endpoint edge-b.example.com:443

HTTP API:

    GET /api/certificate-rollout

Web UI:

- dedicated Cert Rollout navigation/view;
- expected fingerprint/reference selector;
- explicit endpoint list;
- rollout summary;
- expected-certificate card;
- endpoint matrix separating rollout identity from certificate health;
- subject, expiry, hostname, trust, and fingerprint evidence;
- visible safety boundary.

## Security boundary

M17 does not:

- open arbitrary certificate files;
- read private keys;
- read ACME account material;
- renew or install certificates;
- edit TLS configuration;
- reload/restart services;
- change reverse proxies;
- change DNS;
- mutate containers;
- add a Safe Action;
- perform automatic remediation.

All endpoint activity is read-only direct TLS negotiation against explicit targets.

## GitHub CI acceptance

PR #59 final exact head:

    32f0ad2dc3211e5af38a643a5d2f38081d440645

CI run:

    35308369829

All jobs passed:

- format;
- vet;
- full Go tests;
- Web JavaScript syntax;
- native build;
- native Safe Actions regression smoke;
- Docker UI/API smoke;
- linux/amd64 image build;
- linux/arm64 image build.

The Go tests include a real loopback TLS integration fixture with two endpoints serving different certificates and assert the rollout mismatch deterministically.

## Native exact-head validation

The exact PR head was fetched into an isolated /tmp working copy on the OMV host.

Using Go 1.24.13, all passed:

- gofmt check;
- go test ./...;
- go test -race ./internal/core;
- go vet ./...;
- native build;
- certificate-rollout JavaScript syntax;
- git diff --check.

No live HostSleuth container or production service was changed.

## Disposable two-certificate acceptance

Two temporary loopback TLS endpoints were created:

- localhost:19443 served certificate A;
- localhost:19444 served a deliberately different certificate B.

A temporary M17 HostSleuth server listened only on 127.0.0.1:18794.

Acceptance passed in both expected-source modes.

Fingerprint source:

- expected fingerprint = certificate A;
- localhost:19443 -> match;
- localhost:19444 -> mismatch;
- first_problem = localhost:19444.

Reference endpoint source:

- reference = localhost:19443;
- localhost:19443 -> match;
- localhost:19444 -> mismatch;
- first_problem = localhost:19444.

The served HTTP API reproduced the same matrix and the served JavaScript/CSS contained the Cert Rollout UI.

All temporary TLS/HostSleuth processes were confirmed stopped afterward.

## Merge acceptance

PR #59 squash-merged to main at:

    f12bbd8aacb9689bb31d5b9b6535479882cb6ce6

## Release boundary

M17 source is complete on main.

Stable/public/live/recovery remain:

    v0.8.0 / mjmalleo/hostsleuth:0.8.0

No image publication, Arcane redeploy, disaster-recovery change, certificate change, mail-server change, proxy/DNS change, or Safe Action expansion was performed for M17.

The next approved source milestone is M18 — Safe Actions II, beginning with a fresh candidate comparison, threat model, and privilege-cost review before any implementation.
