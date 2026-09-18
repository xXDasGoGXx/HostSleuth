# M17 — Certificate Rollout Verification

Status: source implementation complete and merged. See `docs/history/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md` for acceptance details.

## Product question

> I renewed/replaced the certificate; which endpoint is still serving the old one?

M17 adds one bounded read-only certificate comparison story for multiple explicit direct-TLS endpoints.

## Expected certificate source

Exactly one expected source is accepted:

- a SHA-256 certificate fingerprint; or
- one explicit reference direct-TLS endpoint in host:port form.

Arbitrary certificate-file input is intentionally not exposed.

The owner-approved roadmap allowed a fingerprint/file/reference source, but the project also has an explicit no-private-key-read boundary. HostSleuth cannot prove that an arbitrary user-supplied file is a public certificate before opening it, so a generic certificate-file path would create a private-key-read risk. Fingerprint and reference-endpoint sources satisfy the rollout-verification goal without weakening that boundary.

## Endpoint scope

The comparison accepts:

- at least one endpoint;
- at most 16 endpoints;
- explicit direct-TLS host:port targets only;
- numeric ports only;
- no duplicate targets.

M17 does not autodiscover load balancers, DNS records, reverse proxies, mail endpoints, container ports, or certificate files.

STARTTLS endpoints remain under the separate M16 mail protocol story and are not implicitly folded into M17.

## Expected fingerprint normalization

Fingerprint input:

- is SHA-256 only;
- must represent exactly 32 bytes / 64 hexadecimal characters;
- may contain colons or whitespace;
- is normalized to uppercase hexadecimal before comparison.

No SHA-1 or weaker certificate fingerprint mode is added.

## Reference endpoint behavior

When a reference endpoint is used, HostSleuth:

1. opens one direct TLS connection to the explicit reference host:port;
2. uses the reference host for SNI/hostname evaluation;
3. captures the served leaf-certificate SHA-256 fingerprint and bounded TLS/certificate metadata;
4. uses only that fingerprint as the expected rollout identity.

If the reference endpoint cannot complete TLS or does not provide a usable leaf certificate, comparison stops. Target endpoints are not probed without a usable expected fingerprint.

Reference endpoint hostname/trust evidence is shown for context but does not silently change which certificate fingerprint is treated as expected.

## Endpoint comparison

Each explicit endpoint is probed independently.

Returned evidence includes:

- endpoint target;
- TLS handshake status/error;
- TLS version and cipher;
- served leaf-certificate subject;
- issuer;
- serial;
- SANs;
- validity window;
- days remaining;
- SHA-256 fingerprint;
- hostname verification;
- trust verification;
- rollout match state.

Rollout match state is one of:

- match — served SHA-256 fingerprint equals expected;
- mismatch — served fingerprint is known and differs;
- unknown — no usable served fingerprint could be established.

Certificate health is reported separately from rollout match.

This avoids conflating:

> The new certificate reached this endpoint.

with:

> The certificate currently served by this endpoint is valid, matches its hostname, and is trusted.

## Overall result

Result precedence is deterministic:

1. any fingerprint mismatch -> fail;
2. otherwise any unknown fingerprint -> unknown;
3. otherwise any certificate validity/hostname/trust failure -> fail;
4. otherwise any incomplete certificate-health evidence -> unknown;
5. otherwise -> pass.

The first endpoint responsible for the overall result is returned as first_problem.

## Probe bounds

- maximum 16 explicit endpoints;
- maximum 4 concurrent endpoint probes;
- existing bounded direct-TLS probe timeout per connection;
- results preserve user-supplied endpoint order;
- no retries;
- no recursive discovery;
- no mutation.

## Interfaces

CLI:

    hostsleuth cert-rollout       --fingerprint SHA256       --endpoint edge-a.example.com:443       --endpoint edge-b.example.com:443

or:

    hostsleuth cert-rollout       --reference reference.example.com:443       --endpoint edge-a.example.com:443       --endpoint edge-b.example.com:443

HTTP API:

    GET /api/certificate-rollout

Query parameters:

- fingerprint=... or reference=... exactly once as the expected source;
- endpoint=host:port repeated up to 16 times.

Admin Console:

- dedicated Cert Rollout view;
- expected source selector;
- fingerprint/reference value;
- one endpoint per line;
- rollout summary counts;
- expected-certificate card;
- endpoint matrix separating MATCH/MISMATCH/UNKNOWN from health PASS/FAIL/UNKNOWN;
- subject, expiry, hostname, trust, and fingerprint columns;
- visible safety boundary.

## Security / privacy boundary

M17 does not:

- read arbitrary certificate files;
- read private keys;
- read ACME account material;
- discover certificate paths;
- renew certificates;
- install/replace certificates;
- edit TLS configuration;
- reload/restart services;
- change reverse proxies;
- alter DNS;
- modify containers;
- add a Safe Action;
- perform automatic remediation.

All endpoint activity is read-only direct TLS negotiation against explicit targets.

## Validation status

Source acceptance passed:

- fingerprint normalization and exact-one-source tests;
- endpoint count/deduplication/host:port validation;
- reference-source establishment and failure behavior;
- deterministic match/mismatch/unknown comparison;
- separation of fingerprint match from certificate-health failure;
- CLI/API invalid-input boundary;
- real loopback TLS fixture with two different certificates;
- Admin Console JavaScript syntax and served-bundle coverage;
- full Go test/vet/build and core race test;
- Docker UI/API smoke;
- native exact-head validation on the OMV host;
- disposable fingerprint-source and reference-source CLI/API/UI acceptance.

PR #59 final exact head:

`32f0ad2dc3211e5af38a643a5d2f38081d440645`

CI run:

`35308369829` — all jobs successful.

PR #59 squash-merged to `main` at:

`f12bbd8aacb9689bb31d5b9b6535479882cb6ce6`

Publication, production rollout, recovery alignment, certificate changes, and live service changes remain separate gates.
