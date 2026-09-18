# M14 — Reverse Proxy / Upstream Story + Admin Console v2 Closeout

Status: source implementation complete and locally accepted on branch `m14-reverse-proxy-upstream-story`. Publication and production rollout remain separate gates.

## Delivered — Reverse Proxy / Upstream Story

M14 adds one bounded read-only request-path story built from an explicit public URL and explicit expected upstream URL.

The story evaluates:

1. public DNS;
2. public route evidence;
3. public TCP;
4. public TLS for HTTPS;
5. public HTTP;
6. local listener / Docker publication context when available;
7. upstream DNS;
8. upstream route evidence;
9. upstream TCP;
10. upstream TLS for HTTPS;
11. upstream HTTP behavior.

When the public hostname differs from the upstream hostname, HostSleuth also probes the same explicit upstream address using the public HTTP Host and TLS SNI. This comparison can expose virtual-host/SNI-sensitive behavior without reading proxy configuration or accepting arbitrary custom request headers.

## Deterministic status model

Stages use:

- `pass`;
- `warn`;
- `fail`;
- `unknown`;
- `info`.

Overall outcome gives precedence to a proven failure, then warning, then unresolved evidence. A later proven failure is not hidden by an earlier unknown.

HTTP metadata classification:

- 1xx/2xx/3xx = pass;
- 4xx = warn because the request path answered but application/policy rejected it;
- 405 explicitly explains the HEAD-only diagnostic boundary;
- 5xx = fail.

## HTTP safety boundary

M14 issues HEAD only.

It sends:

- no body;
- no credentials;
- no cookies;
- no Authorization header;
- no environment proxy;
- no arbitrary custom headers.

It returns only bounded response metadata:

- status;
- Location;
- Server;
- Content-Type;
- duration.

Response bodies are never returned.

Public same-host redirects may be followed up to a fixed limit. Redirects to another hostname are recorded but not followed. Upstream redirects are recorded but never followed.

URL query strings are omitted from returned display/evidence URLs and represented as redacted. Location query strings are likewise redacted.

## Host / SNI comparison

For differing public/upstream hostnames, M14 records two upstream variants:

- native upstream Host/SNI;
- public Host/SNI against the same upstream dial target.

The UI shows these side by side.

If behavior changes between variants, HostSleuth reports a warning that Host/SNI handling may matter. It does not claim how nginx, Nginx Proxy Manager, Caddy, Traefik, Apache, HAProxy, or another proxy is configured.

## Interfaces

CLI:

```text
hostsleuth proxy --upstream http://192.0.2.40:8080/ https://photos.example.com/
```

HTTP API:

```text
GET /api/proxy-story?public=...&upstream=...
```

Web UI:

- dedicated **Proxy Path** view;
- visual request-path stages;
- first-problem highlight;
- expandable evidence;
- native vs public Host/SNI comparison;
- bounded evidence-summary copy;
- Diagnose/DNS handoff controls.

## Admin Console v2

M14 extends the console with:

- a Proxy quick-action in Quick Target;
- responsive visual path cards rather than a flat evidence form;
- front/proxy/upstream grouping;
- browser clipboard API with legacy LAN-HTTP fallback for evidence copy;
- stable `#proxy` view deep-link.

## Security / product boundary

M14 remains read-only and adds no privilege.

It does not:

- parse arbitrary proxy config;
- discover upstreams automatically;
- accept arbitrary headers;
- use body-bearing HTTP methods;
- return response bodies;
- perform authentication;
- poll endpoints;
- alert;
- reload/restart a proxy;
- remediate automatically;
- become a proxy-management UI.

A deliberately LAN-bound HostSleuth UI must still be treated as trusted-network access. Loopback remains the safer default.

## Engineering acceptance

Local acceptance used Go 1.24.13 on an isolated OMV `/tmp` clone.

Passed:

- gofmt;
- git diff --check;
- go vet ./...;
- go test ./...;
- go test -race ./internal/core;
- native build;
- syntax check for every Web JavaScript asset;
- combined served-JavaScript syntax validation.

Unit fixtures cover:

- URL safety validation;
- first-problem precedence;
- same-host redirect following;
- cross-host redirect stop;
- public Host forwarding;
- Host-sensitive upstream responses;
- HTTP 2xx/3xx/4xx/5xx classification;
- IPv6 Host normalization;
- query-string redaction;
- explicit 405/HEAD behavior.

Disposable server acceptance on `127.0.0.1:18791` confirmed:

- `/api/proxy-story` is served;
- a self-contained public→upstream path returned overall `pass`;
- CLI and API returned the same overall result;
- the public and upstream targets were preserved correctly;
- native and public-Host upstream variants both returned HTTP 200 in the neutral fixture;
- credential-bearing public URLs returned HTTP 400;
- served JavaScript contains Proxy Path, Quick Target Proxy routing, API path, and evidence-copy controls;
- served CSS contains the visual flow and Host/SNI comparison layout;
- headless-browser DOM execution produced the Proxy navigation/view, Quick Target Proxy action, path container, and Host/SNI comparison section;
- a 1440×1000 headless screenshot was generated for local visual sanity checking and not committed.

Permanent CI was extended to:

- syntax-check Proxy Path and Admin Console v2 assets;
- include them in combined served JavaScript validation;
- assert Proxy Path and Quick Target markers in Docker smoke;
- run a self-contained Docker-mode proxy-story API acceptance.

## Release boundary

Stable/public/live/recovery remain aligned on v0.7.0 until:

1. the exact M14 PR head passes full GitHub CI;
2. M14 merges to `main`;
3. v0.8.0 is published and independently verified;
4. live/recovery rollout passes separately.
