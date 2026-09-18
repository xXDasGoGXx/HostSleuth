# M13 — DNS Detective + Admin Console v1 Closeout

Status: source implementation complete and merged on 2026-09-17. Publication and production rollout remain separate gates.

## Delivered — DNS Detective

M13 adds a dedicated read-only resolver comparison story.

One request can compare:

- the normal system resolver; and
- up to four explicitly supplied resolver IPs.

Custom resolver inputs are bounded to IP addresses on DNS port 53.

For hostname input, HostSleuth records:

- normalized A/AAAA answers;
- canonical CNAME evidence when different from the requested name;
- per-resolver lookup duration;
- bounded resolver errors;
- address family and scope classification.

For IP input, HostSleuth records bounded PTR evidence.

Runtime resolver context includes bounded resolver/search/options metadata visible through `/etc/resolv.conf`, with an explicit warning that Docker mode may expose container-runtime resolver configuration rather than the host-native file.

Deterministic outcomes:

- `agree`;
- `diverge`;
- `partial`;
- `single`.

A private/local-vs-global divergence may be described as **consistent with** split-view DNS or resolver-specific overrides. HostSleuth does not assert split-horizon intent without direct configuration evidence.

Interfaces:

- `hostsleuth dns [--resolver LABEL=IP] NAME_OR_IP`;
- `GET /api/dns-detective`;
- dedicated **DNS Detective** Web UI.

## Delivered — Admin Console v1

The existing feature views remain intact, but the shell now behaves more like a purpose-built troubleshooting console:

- sticky desktop navigation rail;
- wider evidence workspace;
- global Quick Target bar;
- one-input routing to Diagnose, Expectations, or DNS Detective;
- browser-local recent targets;
- browser-local Comfortable / Compact density preference;
- `/` keyboard shortcut to focus the global target bar;
- responsive horizontal navigation on smaller screens;
- stronger focus states and reduced-motion support.

No new server-side user-preference store was added.

## Privacy / security boundary

DNS Detective:

- never silently queries a hard-coded public resolver;
- sends a hostname only to the system resolver and resolver IPs explicitly supplied for that request;
- does not persist server-side resolver history;
- does not perform DNS updates, zone transfers, polling, alerting, or remediation;
- restricts custom Web/API resolver endpoints to IP addresses on port 53;
- adds no privilege.

Admin Console v1 stores recent targets and density preference only in browser local storage.

## Engineering polish

CI now uses per-PR/ref concurrency with cancellation of obsolete in-progress runs.

Permanent CI syntax/smoke coverage includes:

- Endpoint Contracts JavaScript;
- DNS Detective JavaScript;
- Admin Console JavaScript;
- combined served JavaScript;
- Docker smoke presence for Expected Endpoint Contracts, DNS Detective, and Quick Target;
- Docker-mode `/api/dns-detective?name=localhost` acceptance.

## Local validation

Using Go 1.24.13 on an isolated OMV `/tmp` clone:

- `gofmt`;
- `git diff --check`;
- `go vet ./...`;
- `go test ./...`;
- native build;
- standalone syntax checks for every Web JavaScript asset;
- combined served-JavaScript syntax check.

Disposable HTTP server acceptance confirmed:

- DNS Detective API is served;
- system resolver + one explicit reachable local resolver returned `agree` for the test name;
- another explicit local resolver timed out and produced `partial` evidence instead of a generic DNS failure;
- invalid custom resolver port returned HTTP 400;
- served JavaScript contains DNS Detective and the Admin Console Quick Target flow;
- served CSS contains the desktop console rail and DNS resolver-card layout;
- headless browser DOM execution produced the Quick Target deck, DNS navigation/view, density control, and recent-target container.

A local headless screenshot was also produced for visual sanity checking but was not committed.

## Merge acceptance

PR #47 ran on exact head:

`4d5d8f857e362217028efbdedb20e020f7a85d48`

The full CI matrix passed:

- format/vet/test/native build;
- served JavaScript syntax;
- Docker smoke including DNS Detective and Admin Console markers;
- native Safe Actions real-systemd smoke;
- linux/amd64 image build;
- linux/arm64 image build.

PR #47 then squash-merged to `main` at:

`182a27384a090f5538bb6d76a0c4dd917ce63932`

## Release boundary

Stable/public/live/recovery remain aligned on v0.6.0 until:

1. the exact M13 PR head passes full GitHub CI;
2. M13 merges to `main`;
3. v0.7.0 is published and independently verified;
4. production/recovery rollout passes separately.
