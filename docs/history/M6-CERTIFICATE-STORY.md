# M6 — Certificate Story / TLS Detective

Completed: 2026-09-16

## Goal

Make TLS/certificate failures understandable inside the existing evidence-first `host:port` Diagnose workflow without turning HostSleuth into a certificate-management product.

## Accepted implementation

M6 remains read-only and adds:

- bounded TLS connection/handshake probing;
- negotiated TLS protocol and cipher suite;
- served certificate subject, SANs, issuer, serial, valid-from, valid-until, remaining lifetime, and SHA-256 fingerprint;
- explicit hostname verification;
- bounded trust verification against the local host trust store and bounded served-chain subjects;
- positive local listener/process and Docker publication correlation for locally resolved reachable targets;
- native Certbot executable detection;
- up to 32 bounded renewal lineage configurations;
- selected non-secret renewal metadata plus certificate/fullchain paths;
- bounded `certbot.timer` and `certbot.service` state through `systemctl show`;
- local Certbot certificate metadata and fingerprint comparison with the served certificate.

HostSleuth separates successful TCP transport from TLS correctness. A reachable endpoint can therefore be reported as transport-reachable while still explaining a TLS handshake, validity, hostname, trust, or local-vs-served certificate problem.

## Conservative stale-served-certificate rule

A fingerprint difference alone is not proof that a service is stale.

HostSleuth only concludes that the certificate on disk is newer/different than the one being served when:

1. exactly one readable Certbot lineage certificate matches the requested hostname;
2. its SHA-256 fingerprint differs from the served certificate; and
3. its validity evidence is deterministically newer than the served certificate.

Ambiguous or merely different certificates remain `unknown` comparison evidence rather than an invented causal conclusion.

## Security boundary

M6 does not add:

- certificate renewal;
- Certbot write actions;
- service reload/restart;
- certificate installation;
- ACME account administration;
- private-key reads;
- arbitrary command execution;
- a generic certificate-management dashboard.

Renewal configuration reads are bounded and only selected non-secret fields are exposed. Certificate PEM reads parse public certificate blocks only.

## Validation

Focused tests cover:

- TLS handshake success with hostname verification separated from trust failure;
- certificate metadata and SHA-256 fingerprint generation;
- deterministic stale-served detection using a newer matching local Certbot certificate;
- conservative handling when certificates differ without proving staleness;
- diagnosis behavior for hostname mismatch;
- preservation of prior successful-TCP evidence when a fresh TLS diagnostic connection is unavailable.

PR #29 initially passed normal GitHub CI, but isolated native OMV acceptance exposed an environment-dependent pre-existing reachability-policy test: the test mocked the first TCP connection to `127.0.0.1:443` but did not mock the new TLS probe, so a real local listener on OMV could leak into the unit test. The test was corrected to stub the TLS probe explicitly, keeping the production TLS behavior intact and making the policy test host-independent.

After that fix:

- native `go test ./...` passed on the real OMV Debian host;
- the branch built successfully there using an isolated Go 1.24.0 toolchain under `/tmp`;
- `hostsleuth diagnose github.com:443` returned `target is reachable` / high confidence;
- TLS 1.3 / `TLS_AES_128_GCM_SHA256` was negotiated;
- hostname and trust checks passed;
- the served certificate returned subject, issuer, SANs, serial, validity/lifetime, and a 64-character SHA-256 fingerprint;
- the live HostSleuth Arcane deployment was not stopped, restarted, reconfigured, or upgraded.

## Deployment/publication boundary

M6 acceptance does not publish a release and does not migrate production.

Stable remains `v0.2.0`. The live OMV and recovery definition remain pinned to `mjmalleo/hostsleuth:0.1.0`.

The next approved step is the explicit v0.3.0 publication boundary: re-check exact accepted `main` and the release workflow, present the release state to the owner, and stop for explicit approval before creating `release/v0.3.0`, tags, release assets, or Docker images.
